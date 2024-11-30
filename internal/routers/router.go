package routers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/sirini/goapi/internal/handlers"
)

// 라우터들 등록하기
func RegisterRouters(api fiber.Router, h *handlers.Handler) {
	RegisterAuthRouters(api, h)
	RegisterBoardRouters(api, h)
	RegisterChatRouters(api, h)
	RegisterCommentRouters(api, h)
	RegisterHomeRouters(api, h)
	RegisterNotiRouters(api, h)
	RegisterUserRouters(api, h)
}
