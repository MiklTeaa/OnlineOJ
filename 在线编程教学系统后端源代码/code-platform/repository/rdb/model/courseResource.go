package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type CourseResource struct {
	UpdatedAt     time.Time `db:"updated_at"`
	CreatedAt     time.Time `db:"created_at"`
	Title         string    `db:"title"`
	Content       string    `db:"content"`
	AttachMentURL string    `db:"attachment_url"`
	CourseID      uint64    `db:"course_id"`
	ID            uint64    `db:"id"`
}

func (c *CourseResource) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("course_resource").
		Columns("course_id", "title", "content", "attachment_url", "created_at", "updated_at").
		Values(c.CourseID, c.Title, c.Content, c.AttachMentURL, c.CreatedAt, c.UpdatedAt).
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

func (c *CourseResource) Update(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Update("course_resource").
		SetMap(map[string]interface{}{
			"course_id":      c.CourseID,
			"title":          c.Title,
			"content":        c.Content,
			"attachment_url": c.AttachMentURL,
			"created_at":     c.CreatedAt,
			"updated_at":     c.UpdatedAt,
		}).Where(squirrel.Eq{"id": c.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, sqlStr, args...)
	return err
}

func QueryCourseResourcesByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, offset, limit int) ([]*CourseResource, error) {
	const sqlStr = `
SELECT course_resource.*
FROM course_resource INNER JOIN
(SELECT id
FROM course_resource
WHERE course_id = ?
ORDER BY created_at DESC
LIMIT ?, ?) AS c
ON course_resource.id = c.id
`
	courseResources := make([]*CourseResource, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &courseResources, sqlStr, courseID, offset, limit); err != nil {
		return nil, err
	}
	return courseResources[:len(courseResources):len(courseResources)], nil
}

func QueryCourseResourceByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) (*CourseResource, error) {
	const sqlStr = `SELECT * FROM course_resource WHERE id = ?`
	var courseResource CourseResource
	if err := sqlx.GetContext(ctx, rdbClient, &courseResource, sqlStr, ID); err != nil {
		return nil, err
	}
	return &courseResource, nil
}

func DeleteCourseResourceByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) error {
	const sqlStr = `DELETE FROM course_resource WHERE id = ? LIMIT 1`
	_, err := rdbClient.ExecContext(ctx, sqlStr, ID)
	return err
}

func QueryTotalAmountOfCourseResourceByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM course_resource WHERE course_id = ?`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, courseID); err != nil {
		return 0, err
	}
	return total, nil
}

func BatchInsertCourseResources(ctx context.Context, rdbClient storage.RDBClient, courseResources []*CourseResource) error {
	if len(courseResources) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO
course_resource
(course_id, title, content, attachment_url, created_at, updated_at)
VALUES (:course_id, :title, :content, :attachment_url, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, courseResources)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range courseResources {
		courseResources[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func QueryCourseIDByCourseResourceID(ctx context.Context, rdbClient storage.RDBClient, courseResourceID uint64) (uint64, error) {
	const sqlStr = `SELECT course_id FROM course_resource WHERE id = ?`
	var courseID uint64
	if err := sqlx.GetContext(ctx, rdbClient, &courseID, sqlStr, courseResourceID); err != nil {
		return 0, err
	}
	return courseID, nil
}
