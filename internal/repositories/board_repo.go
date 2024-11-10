package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type BoardRepository interface {
	CheckLikedPost(postUid uint, userUid uint) bool
	CheckLikedComment(commentUid uint, userUid uint) bool
	FindPostsByTitleContent(param *models.BoardListParameter) ([]*models.BoardListItem, error)
	FindPostsByNameCategory(param *models.BoardListParameter) ([]*models.BoardListItem, error)
	FindPostsByHashtag(param *models.BoardListParameter) ([]*models.BoardListItem, error)
	GetBoardUidById(id string) uint
	GetBoardCategories(boardUid uint) []models.Pair
	GetCategoryByUid(categoryUid uint) models.Pair
	GetCoverImage(postUid uint) string
	GetCountByTable(table models.Table, postUid uint) uint
	GetGroupAdminUid(boardUid uint) uint
	GetNoticePosts(boardUid uint, actionUserUid uint) ([]*models.BoardListItem, error)
	GetNormalPosts(param *models.BoardListParameter) ([]*models.BoardListItem, error)
	GetMaxUid() uint
	GetTagUids(names string) (string, int)
	GetTotalPostCount(boardUid uint) uint
	GetUidByTable(table models.Table, name string) uint
	GetWriterInfo(userUid uint) *models.BoardWriter
	LoadBoardConfig(boardUid uint) *models.BoardConfig
	MakeListItem(actionUserUid uint, rows *sql.Rows) ([]*models.BoardListItem, error)
}

type TsboardBoardRepository struct {
	db *sql.DB
}

// sql.DB 포인터 주입받기
func NewTsboardBoardRepository(db *sql.DB) *TsboardBoardRepository {
	return &TsboardBoardRepository{db: db}
}

// 게시글에 좋아요를 클릭했었는지 확인하기
func (r *TsboardBoardRepository) CheckLikedPost(postUid uint, userUid uint) bool {
	if userUid < 1 {
		return false
	}
	var liked uint8
	query := fmt.Sprintf("SELECT liked FROM %s%s WHERE post_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST_LIKE)
	r.db.QueryRow(query, postUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 댓글에 좋아요를 클릭했었는지 확인하기
func (r *TsboardBoardRepository) CheckLikedComment(commentUid uint, userUid uint) bool {
	if userUid < 1 {
		return false
	}
	var liked uint8
	query := fmt.Sprintf("SELECT liked FROM %s%s WHERE comment_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	r.db.QueryRow(query, commentUid, userUid, 1).Scan(&liked)
	return liked > 0
}

// 게시글 제목 혹은 내용으로 검색해서 가져오기
func (r *TsboardBoardRepository) FindPostsByTitleContent(param *models.BoardListParameter) ([]*models.BoardListItem, error) {
	option := param.Option.String()
	keyword := "%" + param.Keyword + "%"
	arrow, order := param.Direction.Query()
	query := fmt.Sprintf(`SELECT uid, user_uid, category_uid, title, content, submitted, modified, hit, status
												FROM %s%s WHERE board_uid = ? AND status = ? AND %s LIKE ? AND uid %s ?
												ORDER BY uid %s LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, option, arrow, order)
	rows, err := r.db.Query(query, param.BoardUid, models.POST_NORMAL, keyword, param.SinceUid, param.Bunch-param.NoticeCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시글 작성자 혹은 분류명으로 검색해서 가져오기
func (r *TsboardBoardRepository) FindPostsByNameCategory(param *models.BoardListParameter) ([]*models.BoardListItem, error) {
	option := param.Option.String()
	arrow, order := param.Direction.Query()
	table := models.TABLE_USER
	if param.Option == models.SEARCH_CATEGORY {
		table = models.TABLE_BOARD_CAT
	}
	uid := r.GetUidByTable(table, param.Keyword)
	query := fmt.Sprintf(`SELECT uid, user_uid, category_uid, title, content, submitted, modified, hit, status
												FROM %s%s WHERE board_uid = ? AND status = ? AND %s = ? AND uid %s ?
												ORDER BY uid %s LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, option, arrow, order)
	rows, err := r.db.Query(query, param.BoardUid, models.POST_NORMAL, uid, param.SinceUid, param.Bunch-param.NoticeCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시글 태그로 검색해서 가져오기
func (r *TsboardBoardRepository) FindPostsByHashtag(param *models.BoardListParameter) ([]*models.BoardListItem, error) {
	arrow, order := param.Direction.Query()
	tagUidStr, tagCount := r.GetTagUids(param.Keyword)
	query := fmt.Sprintf(`SELECT p.uid, p.user_uid, p.category_uid, p.title, p.content, 
												p.submitted, p.modified, p.hit, p.status
												FROM %s%s AS p JOIN %s%s AS ph ON p.uid = ph.post_uid
												WHERE p.board_uid = ? AND p.status = ? AND p.uid %s ? AND ph.hashtag_uid IN (%s)
												GROUP BY ph.post_uid HAVING (COUNT(ph.hashtag_uid) = ?)
												ORDER BY p.uid %s LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, configs.Env.Prefix, models.TABLE_POST_HASHTAG, arrow, tagUidStr, order)
	rows, err := r.db.Query(query, param.BoardUid, models.POST_NORMAL, param.SinceUid, tagCount, param.Bunch-param.NoticeCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시판 아이디로 게시판 고유 번호 가져오기
func (r *TsboardBoardRepository) GetBoardUidById(id string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE id = ? LIMIT 1", configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, id).Scan(&uid)
	return uid
}

// 지정된 게시판에서 사용중인 카테고리 목록들 반환
func (r *TsboardBoardRepository) GetBoardCategories(boardUid uint) []models.Pair {
	var items []models.Pair
	query := fmt.Sprintf("SELECT uid, name FROM %s%s WHERE board_uid = ?", configs.Env.Prefix, models.TABLE_BOARD_CAT)
	rows, err := r.db.Query(query, boardUid)
	if err != nil {
		return items
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Pair{}
		err := rows.Scan(&item.Uid, &item.Name)
		if err != nil {
			return items
		}
		items = append(items, item)
	}
	return items
}

// 카테고리 이름 가져오기
func (r *TsboardBoardRepository) GetCategoryByUid(categoryUid uint) models.Pair {
	cat := models.Pair{}
	query := fmt.Sprintf("SELECT uid, name FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_BOARD_CAT)
	r.db.QueryRow(query, categoryUid).Scan(&cat.Uid, &cat.Name)
	return cat
}

// 게시글 대표 커버 썸네일 이미지 가져오기
func (r *TsboardBoardRepository) GetCoverImage(postUid uint) string {
	var path string
	query := fmt.Sprintf("SELECT path FROM %s%s WHERE post_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_FILE_THUMB)
	r.db.QueryRow(query, postUid).Scan(&path)
	return path
}

// 댓글 혹은 좋아요 개수 가져오기
func (r *TsboardBoardRepository) GetCountByTable(table models.Table, postUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s WHERE post_uid = ?", configs.Env.Prefix, table)
	r.db.QueryRow(query, postUid).Scan(&count)
	return count
}

// 게시판이 속한 그룹의 관리자 고유 번호값 가져오기
func (r *TsboardBoardRepository) GetGroupAdminUid(boardUid uint) uint {
	var adminUid uint
	query := fmt.Sprintf(`SELECT g.admin_uid FROM %s%s AS g JOIN %s%s AS b 
												ON g.uid = b.group_uid WHERE b.uid = ? LIMIT 1`,
		configs.Env.Prefix, models.TABLE_GROUP, configs.Env.Prefix, models.TABLE_BOARD)
	r.db.QueryRow(query, boardUid).Scan(&adminUid)
	return adminUid
}

// 게시판 공지글만 가져오기
func (r *TsboardBoardRepository) GetNoticePosts(boardUid uint, actionUserUid uint) ([]*models.BoardListItem, error) {
	query := fmt.Sprintf(`SELECT uid, user_uid, category_uid, title, content, submitted, modified, hit, status
												FROM %s%s WHERE board_uid = ? AND status = ?`, configs.Env.Prefix, models.TABLE_POST)
	rows, err := r.db.Query(query, boardUid, models.POST_NOTICE)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(actionUserUid, rows)
}

// 비밀글을 포함한 일반 게시글들 가져오기
func (r *TsboardBoardRepository) GetNormalPosts(param *models.BoardListParameter) ([]*models.BoardListItem, error) {
	arrow, order := param.Direction.Query()
	query := fmt.Sprintf(`SELECT uid, user_uid, category_uid, title, content, submitted, modified, hit, status
												FROM %s%s WHERE board_uid = ? AND status IN (?, ?) AND uid %s ?
												ORDER BY uid %s LIMIT ?`,
		configs.Env.Prefix, models.TABLE_POST, arrow, order)
	rows, err := r.db.Query(query, param.BoardUid, models.POST_NORMAL, models.POST_SECRET, param.SinceUid, param.Bunch-param.NoticeCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.MakeListItem(param.UserUid, rows)
}

// 게시판의 현재 uid 값 반환하기
func (r *TsboardBoardRepository) GetMaxUid() uint {
	var max uint
	query := fmt.Sprintf("SELECT MAX(uid) AS last FROM %s%s", configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query).Scan(&max)
	return max
}

// 스페이스로 구분된 태그 이름들을 가져와서 태그 고유번호 문자열로 변환
func (r *TsboardBoardRepository) GetTagUids(keyword string) (string, int) {
	tags := strings.Split(keyword, " ")
	var strUids []string
	for _, tag := range tags {
		var uid uint
		query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? LIMIT 1",
			configs.Env.Prefix, models.TABLE_HASHTAG)
		if err := r.db.QueryRow(query, tag).Scan(&uid); err != nil {
			continue
		}
		strUids = append(strUids, fmt.Sprintf("'%d'", uid))
	}
	result := strings.Join(strUids, ",")
	return result, len(strUids)
}

// 게시판에 등록된 글 갯수 반환
func (r *TsboardBoardRepository) GetTotalPostCount(boardUid uint) uint {
	var count uint
	query := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s%s WHERE board_uid = ? AND status != ?",
		configs.Env.Prefix, models.TABLE_POST)
	r.db.QueryRow(query, boardUid, models.POST_REMOVED).Scan(&count)
	return count
}

// 이름으로 고유 번호 가져오기 (회원 번호 혹은 카테고리 번호 등)
func (r *TsboardBoardRepository) GetUidByTable(table models.Table, name string) uint {
	var uid uint
	query := fmt.Sprintf("SELECT uid FROM %s%s WHERE name = ? ORDER BY uid DESC LIMIT 1", configs.Env.Prefix, table)
	r.db.QueryRow(query, name).Scan(&uid)
	return uid
}

// (댓)글 작성자 기본 정보 가져오기
func (r *TsboardBoardRepository) GetWriterInfo(userUid uint) *models.BoardWriter {
	writer := &models.BoardWriter{}
	writer.UserUid = userUid
	query := fmt.Sprintf("SELECT name, profile, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	r.db.QueryRow(query, userUid).Scan(&writer.Name, &writer.Profile, &writer.Signature)
	return writer
}

// 게시판 설정값 가져오기
func (r *TsboardBoardRepository) LoadBoardConfig(boardUid uint) *models.BoardConfig {
	config := &models.BoardConfig{}
	query := fmt.Sprintf(`SELECT admin_uid, type, name, info, row_count, width, use_category,
												level_list, level_view, level_write, level_comment, level_download,
												point_view, point_write, point_comment, point_download 
												FROM %s%s WHERE uid = ? LIMIT 1`, configs.Env.Prefix, models.TABLE_BOARD)
	var useCategory uint8
	r.db.QueryRow(query, boardUid).Scan(&config.Admin.Board, &config.Type, &config.Name, &config.Info,
		&config.RowCount, &config.Width, &useCategory, &config.Level.List, &config.Level.View,
		&config.Level.Write, &config.Level.Comment, &config.Level.Download, &config.Point.View,
		&config.Point.Write, &config.Point.Comment, &config.Point.Download)
	config.UseCategory = useCategory > 0
	config.Category = r.GetBoardCategories(boardUid)
	config.Admin.Group = r.GetGroupAdminUid(boardUid)
	return config
}

// 게시글 목록 만들어서 반환
func (r *TsboardBoardRepository) MakeListItem(actionUserUid uint, rows *sql.Rows) ([]*models.BoardListItem, error) {
	var items []*models.BoardListItem
	for rows.Next() {
		item := &models.BoardListItem{}
		var writerUid uint
		err := rows.Scan(&item.Uid, &writerUid, &item.Category.Uid, &item.Title, &item.Content,
			&item.Submitted, &item.Modified, &item.Hit, &item.Status)
		if err != nil {
			return nil, err
		}
		item.Category = r.GetCategoryByUid(item.Category.Uid)
		item.Cover = r.GetCoverImage(item.Uid)
		item.Comment = r.GetCountByTable(models.TABLE_COMMENT, item.Uid)
		item.Like = r.GetCountByTable(models.TABLE_POST_LIKE, item.Uid)
		item.Liked = r.CheckLikedPost(item.Uid, actionUserUid)
		item.Writer = r.GetWriterInfo(writerUid)
		items = append(items, item)
	}
	return items, nil
}
