package model

import (
	"context"
	"time"

	"code-platform/pkg/slicex"
	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type CourseComment struct {
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	CommentText string    `db:"comment_text"`
	CourseID    uint64    `db:"course_id"`
	PID         uint64    `db:"pid"`
	ReplyUserID uint64    `db:"reply_user_id"`
	ID          uint64    `db:"id"`
	UserID      uint64    `db:"user_id"`
}

func QuerySubCourseCommentsMapByParentCommentIDs(ctx context.Context, rdbClient storage.RDBClient, parentCommentIDs []uint64) (m map[uint64][]*CourseCommentWithUserInfo, replyUserIDs []uint64, err error) {
	if len(parentCommentIDs) == 0 {
		return nil, nil, nil
	}

	const sqlStr = `
SELECT course_comment.*, user.name ,user.avatar
FROM course_comment
INNER JOIN
user
ON course_comment.pid IN (?)
AND course_comment.user_id = user.id
ORDER BY course_comment.id DESC
`
	query, args, err := sqlx.In(sqlStr, parentCommentIDs)
	if err != nil {
		return nil, nil, err
	}

	var subComments []*CourseCommentWithUserInfo
	if err := sqlx.SelectContext(ctx, rdbClient, &subComments, query, args...); err != nil {
		return nil, nil, err
	}

	replyUserIDs = make([]uint64, len(subComments))
	m = make(map[uint64][]*CourseCommentWithUserInfo)
	for index, comment := range subComments {
		m[comment.PID] = append(m[comment.PID], comment)
		replyUserIDs[index] = comment.ReplyUserID
	}
	return m, slicex.DistinctUint64Slice(replyUserIDs), nil
}

func (c *CourseComment) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("course_comment").
		Columns("course_id", "comment_text", "pid", "user_id", "reply_user_id", "created_at", "updated_at").
		Values(c.CourseID, c.CommentText, c.PID, c.UserID, c.ReplyUserID, c.CreatedAt, c.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}
	result, err := rdbClient.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = uint64(lastID)
	return nil
}

func DeleteCourseCommentByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) error {
	const sqlStr = `DELETE FROM course_comment WHERE id = ?`
	_, err := rdbClient.ExecContext(ctx, sqlStr, ID)
	return err
}

func QueryTotalAmountInCourseParentCommentByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM course_comment WHERE course_id = ? AND pid = 0`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, courseID); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryCourseParentCommentWithUserInfosByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, offset, limit int) ([]*CourseCommentWithUserInfo, error) {
	const sqlStr = `
SELECT course_comment.*, user.name, user.avatar
FROM user INNER JOIN course_comment INNER JOIN
(SELECT id, user_id
FROM course_comment
WHERE course_comment.course_id = ?
AND course_comment.pid = 0
ORDER BY course_comment.id DESC
LIMIT ?, ?) AS cc
ON user.id = cc.user_id
AND course_comment.id = cc.id
`
	comments := make([]*CourseCommentWithUserInfo, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &comments, sqlStr, courseID, offset, limit); err != nil {
		return nil, err
	}
	return comments, nil
}

func QueryCourseCommentByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) (*CourseComment, error) {
	const sqlStr = `SELECT * FROM course_comment WHERE id = ?`
	var comment CourseComment
	if err := sqlx.GetContext(ctx, rdbClient, &comment, sqlStr, ID); err != nil {
		return nil, err
	}
	return &comment, nil
}

func (c *CourseComment) IsParent() bool {
	return c.PID == 0
}

func DeleteSubCourseCommentsByParentCommentID(ctx context.Context, rdbClient storage.RDBClient, PID uint64) error {
	const sqlStr = `DELETE FROM course_comment WHERE pid = ?`
	_, err := rdbClient.ExecContext(ctx, sqlStr, PID)
	return err
}

func BatchInsertCourseComments(ctx context.Context, rdbClient storage.RDBClient, courseComments []*CourseComment) error {
	if len(courseComments) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO
course_comment
(course_id, comment_text, pid, user_id, reply_user_id, created_at, updated_at)
VALUES (:course_id, :comment_text, :pid, :user_id, reply_user_id, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, courseComments)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range courseComments {
		courseComments[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}
