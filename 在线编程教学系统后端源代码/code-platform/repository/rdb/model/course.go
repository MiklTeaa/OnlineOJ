package model

import (
	"context"
	"database/sql"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Course struct {
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	PicURL      string         `db:"pic_url"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	SecretKey   sql.NullString `db:"secret_key"`
	ID          uint64         `db:"id"`
	TeacherID   uint64         `db:"teacher_id"`
	NeedAudit   bool           `db:"need_audit"`
	IsClosed    bool           `db:"is_closed"`
	Language    int8           `db:"language"`
}

func QueryTotalAmountCoursesByTeacherID(ctx context.Context, rdbClient storage.RDBClient, teacherID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM course WHERE teacher_id = ?`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, teacherID); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryCoursesByTeacherID(ctx context.Context, rdbClient storage.RDBClient, teacherID uint64, offset, limit int) ([]*Course, error) {
	const sqlStr = `
SELECT course.*
FROM course INNER JOIN
(SELECT id
FROM course
WHERE teacher_id = ?
ORDER BY id DESC
LIMIT ?, ?) AS c
ON course.id = c.id
`
	courses := make([]*Course, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &courses, sqlStr, teacherID, offset, limit); err != nil {
		return nil, err
	}
	return courses, nil
}

func QueryCourses(ctx context.Context, rdbClient storage.RDBClient, offset, limit int) ([]*Course, error) {
	const sqlStr = `
SELECT *
FROM course INNER JOIN
(SELECT id
FROM course
ORDER BY id DESC
LIMIT ?, ?) AS c
ON course.id = c.id
`
	courses := make([]*Course, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &courses, sqlStr, offset, limit); err != nil {
		return nil, err
	}
	return courses[:len(courses):len(courses)], nil
}

func QueryCoursesByStudentID(ctx context.Context, rdbClient storage.RDBClient, studentID uint64, offset, limit int) ([]*Course, error) {
	const sqlStr = `
SELECT course.*
FROM course INNER JOIN
(SELECT course.id
FROM course INNER JOIN
arrange_course
ON arrange_course.user_id = ?
AND arrange_course.course_id = course.id
AND arrange_course.is_pass = TRUE
ORDER BY id DESC
LIMIT ?, ?) AS ca
ON course.id = ca.id
`
	courses := make([]*Course, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &courses, sqlStr, studentID, offset, limit); err != nil {
		return nil, err
	}
	return courses[:len(courses):len(courses)], nil
}

func QueryCoursesByName(ctx context.Context, rdbClient storage.RDBClient, keyword string, offset, limit int) ([]*Course, error) {
	const sqlStr = `
SELECT *
FROM course INNER JOIN
(SELECT id
FROM course
WHERE MATCH (name, description) AGAINST (?)
ORDER BY id DESC
LIMIT ?, ?) AS c
ON course.id = c.id
`
	courses := make([]*Course, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &courses, sqlStr, keyword, offset, limit); err != nil {
		return nil, err
	}
	return courses, nil
}

func QueryCourseMapsByIDs(ctx context.Context, rdbClient storage.RDBClient, IDs []uint64) (map[uint64]*Course, error) {
	if len(IDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("course").
		Where(squirrel.Eq{"id": IDs}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var courses []*Course
	if err := sqlx.SelectContext(ctx, rdbClient, &courses, query, args...); err != nil {
		return nil, err
	}
	m := make(map[uint64]*Course)
	for _, course := range courses {
		m[course.ID] = course
	}
	return m, nil
}

func QueryCourseByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) (*Course, error) {
	const sqlStr = `SELECT * FROM course WHERE id = ?`
	var course Course
	if err := sqlx.GetContext(ctx, rdbClient, &course, sqlStr, ID); err != nil {
		return nil, err
	}
	return &course, nil
}

func (c *Course) Update(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Update("course").
		SetMap(map[string]interface{}{
			"teacher_id":  c.TeacherID,
			"name":        c.Name,
			"description": c.Description,
			"pic_url":     c.PicURL,
			"secret_key":  c.SecretKey,
			"need_audit":  c.NeedAudit,
			"is_closed":   c.IsClosed,
			"language":    c.Language,
			"created_at":  c.CreatedAt,
			"updated_at":  c.UpdatedAt,
		}).Where(squirrel.Eq{"id": c.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, sqlStr, args...)
	return err
}

func (c *Course) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("course").
		Columns("teacher_id", "name", "description", "pic_url", "secret_key", "need_audit", "is_closed", "language", "created_at", "updated_at").
		Values(c.TeacherID, c.Name, c.Description, c.PicURL, c.SecretKey, c.NeedAudit, c.IsClosed, c.Language, c.CreatedAt, c.UpdatedAt).
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

func QueryTotalAmountOfCourseByFuzzyCourseName(ctx context.Context, rdbClient storage.RDBClient, keyword string) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM course WHERE MATCH (name, description) AGAINST (?)`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, keyword); err != nil {
		return 0, err
	}
	return total, nil
}

func DeleteCourseByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) error {
	const sqlStr = `DELETE FROM course WHERE id = ? LIMIT 1`
	_, err := rdbClient.ExecContext(ctx, sqlStr, ID)
	return err
}

func QueryAllCourses(ctx context.Context, rdbClient storage.RDBClient, offset, limit int) ([]*Course, error) {
	const sqlStr = `SELECT * FROM course LIMIT ?, ?`
	courses := make([]*Course, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &courses, sqlStr, offset, limit); err != nil {
		return nil, err
	}
	return courses, nil
}

func QueryTotalAmountCourses(ctx context.Context, rdbClient storage.RDBClient) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM course`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr); err != nil {
		return 0, err
	}
	return total, nil
}

func BatchInsertCourses(ctx context.Context, rdbClient storage.RDBClient, courses []*Course) error {
	if len(courses) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO course
(teacher_id, name, description, pic_url, secret_key, need_audit, is_closed, language, created_at, updated_at)
VALUES (:teacher_id, :name, :description, :pic_url, :secret_key, :need_audit, :is_closed, :language, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, courses)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range courses {
		courses[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func QueryCourseTeacherIDByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (uint64, error) {
	const sqlStr = `SELECT teacher_id FROM course WHERE id = ?`
	var teacherID uint64
	if err := sqlx.GetContext(ctx, rdbClient, &teacherID, sqlStr, courseID); err != nil {
		return 0, err
	}
	return teacherID, nil
}
