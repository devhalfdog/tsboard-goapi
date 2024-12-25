package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/templates"
	"github.com/sirini/goapi/pkg/utils"
)

type BlogHandler interface {
	BlogRssLoadHandler(c fiber.Ctx) error
}

type TsboardBlogHandler struct {
	service *services.Service
}

// services.Service 주입 받기
func NewTsboardBlogHandler(service *services.Service) *TsboardBlogHandler {
	return &TsboardBlogHandler{service: service}
}

// RSS 불러오기 핸들러
func (h *TsboardBlogHandler) BlogRssLoadHandler(c fiber.Ctx) error {
	c.Set("Content-Type", "application/rss+xml; charset=UTF-8")
	id := c.Params("id")
	boardUid := h.service.Board.GetBoardUid(id)
	config := h.service.Board.GetBoardConfig(boardUid)

	if config.Uid < 1 {
		return c.SendString(`<error>Invalid board id.</error>`)
	}

	posts, err := h.service.Blog.GetLatestPosts(boardUid, 50)
	if err != nil {
		return c.SendString(`<error>Unable to load the latest posts from server, please visit website instead.</error>`)
	}

	latestDate := ""
	var items []string
	for _, post := range posts {
		writer, err := h.service.User.GetUserInfo(post.UserUid)
		if err != nil {
			return c.SendString(`<error>Unable to find the information of writer.</error>`)
		}

		t := time.UnixMilli(int64(post.Submitted))
		pubDate := t.Format(time.RFC1123)
		item := fmt.Sprintf(`<item>
          <title>%s</title>
          <link>%s%s/blog/%s/%d</link>
          <description>%s</description>
          <author>%s</author>
          <pubDate>%s</pubDate>
          <guid isPermaLink="true">%s%s/blog/%s/%d</guid>
        </item>`,
			utils.Unescape(post.Title),
			configs.Env.URL, configs.Env.URLPrefix, id, post.Uid,
			utils.Unescape(post.Content),
			writer.Name,
			pubDate,
			configs.Env.URL, configs.Env.URLPrefix, id, post.Uid,
		)
		items = append(items, item)

		if len(latestDate) < 1 {
			latestDate = pubDate
		}
	}

	var rss string
	rss = strings.ReplaceAll(templates.RssBody, "#BLOG.TITLE#", utils.Unescape(config.Name))
	rss = strings.ReplaceAll(rss, "#BLOG.LINK#", fmt.Sprintf("%s%s/blog/%s", configs.Env.URL, configs.Env.URLPrefix, id))
	rss = strings.ReplaceAll(rss, "#BLOG.INFO#", utils.Unescape(config.Info))
	rss = strings.ReplaceAll(rss, "#BLOG.LANG#", "ko-kr")
	rss = strings.ReplaceAll(rss, "#BLOG.DATE#", latestDate)
	rss = strings.ReplaceAll(rss, "#BLOG.GENERATOR#", fmt.Sprintf("TSBOARD %s [tsboard.dev]", configs.Env.Version))
	rss = strings.ReplaceAll(rss, "#BLOG.ITEM#", strings.Join(items, ""))

	return c.SendString(rss)
}
