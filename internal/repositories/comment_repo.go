package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sirini/goapi/internal/configs"
	"github.com/sirini/goapi/pkg/models"
)

type CommentRepository interface {
	FindPostUserUidByUid(commentUid uint) (uint, uint)
	GetComments(param models.CommentListParameter) ([]models.CommentItem, error)
	GetLikedCountForLoop(stmt *sql.Stmt, commentUid uint) uint
	GetLikedCount(commentUid uint) uint
	GetPostStatus(postUid uint) models.Status
	GetPostWriterUid(postUid uint) uint
	GetMaxUid() uint
	HasReplyComment(commentUid uint) bool
	IsLikedComment(commentUid uint, userUid uint) bool
	InsertComment(param models.CommentWriteParameter) (uint, error)
	InsertLikeComment(param models.CommentLikeParameter)
	RemoveComment(commentUid uint) error
	UpdateComment(commentUid uint, content string)
	UpdateLikeComment(param models.CommentLikeParameter)
	UpdateReplyUid(commentUid uint, replyUid uint)
}

type TsboardCommentRepository struct {
	db    *sql.DB
	board BoardRepository
}

// sql.DB, board 포인터 주입받기
func NewTsboardCommentRepository(db *sql.DB, board BoardRepository) *TsboardCommentRepository {
	return &TsboardCommentRepository{db: db, board: board}
}

// 댓글 고유 번호로 댓글 작성자의 고유 번호 반환하기
func (r *TsboardCommentRepository) FindPostUserUidByUid(commentUid uint) (uint, uint) {
	var postUid, userUid uint
	query := fmt.Sprintf("SELECT post_uid, user_uid FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, commentUid).Scan(&postUid, &userUid)
	return postUid, userUid
}

// 댓글들 가져오기
func (r *TsboardCommentRepository) GetComments(param models.CommentListParameter) ([]models.CommentItem, error) {
	arrow, _ := param.Direction.Query()
	query := fmt.Sprintf(`SELECT uid, reply_uid, user_uid, content, submitted, modified, status 
												FROM %s%s WHERE post_uid = ? AND status != ? AND uid %s ?
												ORDER BY reply_uid ASC LIMIT ?`, configs.Env.Prefix, models.TABLE_COMMENT, arrow)

	rows, err := r.db.Query(query, param.PostUid, models.CONTENT_REMOVED, param.SinceUid, param.Bunch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 게시글 작성자 정보 가져오는 쿼리문 준비
	query = fmt.Sprintf("SELECT name, profile, signature FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_USER)
	stmtWriter, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtWriter.Close()

	// 댓글에 대한 좋아요 수 반환하는 쿼리문 준비
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE comment_uid = ? AND liked = ?",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	stmtLikedCount, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtLikedCount.Close()

	// 댓글에 좋아요를 클릭했는지 확인하는 쿼리문 준비
	query = fmt.Sprintf("SELECT liked FROM %s%s WHERE comment_uid = ? AND user_uid = ? AND liked = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)
	stmtLiked, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmtLiked.Close()

	items := make([]models.CommentItem, 0)
	for rows.Next() {
		item := models.CommentItem{}
		err = rows.Scan(&item.Uid, &item.ReplyUid, &item.Writer.UserUid, &item.Content, &item.Submitted, &item.Modified, &item.Status)
		if err != nil {
			return nil, err
		}
		item.PostUid = param.PostUid
		item.Writer = r.board.GetWriterInfoForLoop(stmtWriter, item.Writer.UserUid)
		item.Like = r.GetLikedCountForLoop(stmtLikedCount, item.Uid)
		item.Liked = r.board.CheckLikedCommentForLoop(stmtLiked, item.Uid, param.UserUid)
		items = append(items, item)
	}
	return items, nil
}

// 반복문에서 사용하는 댓글에 대한 좋아요 수 반환
func (r *TsboardCommentRepository) GetLikedCountForLoop(stmt *sql.Stmt, commentUid uint) uint {
	var count uint
	stmt.QueryRow(commentUid, 1).Scan(&count)
	return count
}

// 댓글에 대한 좋아요 수 반환
func (r *TsboardCommentRepository) GetLikedCount(commentUid uint) uint {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s%s WHERE comment_uid = ? AND liked = ?",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	var count uint
	r.db.QueryRow(query, commentUid, 1).Scan(&count)
	return count
}

// 게시글 상태 가져오기
func (r *TsboardCommentRepository) GetPostStatus(postUid uint) models.Status {
	var status int8
	query := fmt.Sprintf("SELECT status FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)

	r.db.QueryRow(query, postUid).Scan(&status)
	return models.Status(status)
}

// 게시글 작성자의 고유 번호 반환하기
func (r *TsboardCommentRepository) GetPostWriterUid(postUid uint) uint {
	var userUid uint
	query := fmt.Sprintf("SELECT user_uid FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_POST)

	r.db.QueryRow(query, postUid).Scan(&userUid)
	return userUid
}

// 가장 마지막 댓글 고유 번호 가져오기
func (r *TsboardCommentRepository) GetMaxUid() uint {
	var uid uint
	query := fmt.Sprintf("SELECT MAX(uid) FROM %s%s", configs.Env.Prefix, models.TABLE_COMMENT)
	r.db.QueryRow(query).Scan(&uid)
	return uid
}

// 이 댓글에 답글이 하나라도 있는지 확인하기
func (r *TsboardCommentRepository) HasReplyComment(commentUid uint) bool {
	var replyUid uint
	query := fmt.Sprintf("SELECT reply_uid FROM %s%s WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, commentUid).Scan(&replyUid)
	if replyUid != commentUid {
		return false
	}

	var uid uint
	query = fmt.Sprintf("SELECT uid FROM %s%s WHERE reply_uid = ? AND uid != ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.QueryRow(query, commentUid, commentUid).Scan(&uid)
	return uid > 0
}

// 이미 이 댓글에 좋아요를 클릭한 적이 있는지 확인하기
func (r *TsboardCommentRepository) IsLikedComment(commentUid uint, userUid uint) bool {
	var uid uint
	query := fmt.Sprintf("SELECT comment_uid FROM %s%s WHERE comment_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	r.db.QueryRow(query, commentUid, userUid).Scan(&uid)
	return uid > 0
}

// 새로운 댓글 작성하기
func (r *TsboardCommentRepository) InsertComment(param models.CommentWriteParameter) (uint, error) {
	query := fmt.Sprintf(`INSERT INTO %s%s 
												(reply_uid, board_uid, post_uid, user_uid, content, submitted, modified, status) 
												VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_COMMENT)

	result, err := r.db.Exec(
		query,
		0,
		param.BoardUid,
		param.PostUid,
		param.UserUid,
		param.Content,
		time.Now().UnixMilli(),
		0,
		models.CONTENT_NORMAL,
	)
	if err != nil {
		return models.FAILED, err
	}
	insertId, _ := result.LastInsertId()
	return uint(insertId), nil
}

// 이 댓글에 대한 좋아요 추가하기
func (r *TsboardCommentRepository) InsertLikeComment(param models.CommentLikeParameter) {
	query := fmt.Sprintf(`INSERT INTO %s%s (board_uid, comment_uid, user_uid, liked, timestamp) 
												VALUES (?, ?, ?, ?, ?)`, configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	r.db.Exec(query, param.BoardUid, param.CommentUid, param.UserUid, param.Liked, time.Now().UnixMilli())
}

// 댓글을 삭제 상태로 변경하기
func (r *TsboardCommentRepository) RemoveComment(commentUid uint) error {
	query := fmt.Sprintf("UPDATE %s%s SET status = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)
	_, err := r.db.Exec(query, models.CONTENT_REMOVED, commentUid)
	return err
}

// 기존 댓글 수정하기
func (r *TsboardCommentRepository) UpdateComment(commentUid uint, content string) {
	query := fmt.Sprintf("UPDATE %s%s SET content = ?, modified = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.Exec(query, content, time.Now().UnixMilli(), commentUid)
}

// 이 댓글에 대한 좋아요 변경하기
func (r *TsboardCommentRepository) UpdateLikeComment(param models.CommentLikeParameter) {
	query := fmt.Sprintf("UPDATE %s%s SET liked = ?, timestamp = ? WHERE comment_uid = ? AND user_uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT_LIKE)

	r.db.Exec(query, param.Liked, time.Now().UnixMilli(), param.CommentUid, param.UserUid)
}

// 답글 고유 번호 업데이트
func (r *TsboardCommentRepository) UpdateReplyUid(commentUid uint, replyUid uint) {
	query := fmt.Sprintf("UPDATE %s%s SET reply_uid = ? WHERE uid = ? LIMIT 1",
		configs.Env.Prefix, models.TABLE_COMMENT)

	r.db.Exec(query, replyUid, commentUid)
}
