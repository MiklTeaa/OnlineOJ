package model

import (
	"context"
	"database/sql"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Lab struct {
	CreatedAt     time.Time    `db:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
	DeadLine      sql.NullTime `db:"dead_line"`
	Title         string       `db:"title"`
	Content       string       `db:"content"`
	AttachMentURL string       `db:"attachment_url"`
	ID            uint64       `db:"id"`
	CourseID      uint64       `db:"course_id"`
}

func (l *Lab) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("lab").
		Columns("course_id", "title", "content", "attachment_url", "dead_line", "created_at", "updated_at").
		Values(l.CourseID, l.Title, l.Content, l.AttachMentURL, l.DeadLine, l.CreatedAt, l.UpdatedAt).
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
	l.ID = uint64(lastID)
	return nil
}

// UpdateIDToZero Only For Test
func (l *Lab) UpdateIDToZero(ctx context.Context, rdbClient storage.RDBClient) error {
	const sqlStr = `update lab SET id = 0 WHERE id = ?`
	if _, err := rdbClient.ExecContext(ctx, sqlStr, l.ID); err != nil {
		return err
	}
	l.ID = 0
	return nil
}

func (l *Lab) Update(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Update("lab").SetMap(squirrel.Eq{
		"course_id":      l.CourseID,
		"title":          l.Title,
		"content":        l.Content,
		"attachment_url": l.AttachMentURL,
		"dead_line":      l.DeadLine,
		"created_at":     l.CreatedAt,
		"updated_at":     l.UpdatedAt,
	}).Where(squirrel.Eq{"id": l.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, sqlStr, args...)
	return err
}

func DeleteLabByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) error {
	const sqlStr = `DELETE FROM lab WHERE id = ?`
	_, err := rdbClient.ExecContext(ctx, sqlStr, ID)
	return err
}

func QueryLabByID(ctx context.Context, rdbClient storage.RDBClient, labID uint64) (*Lab, error) {
	const sqlStr = `SELECT * FROM lab WHERE id = ?`
	var lab Lab
	if err := sqlx.GetContext(ctx, rdbClient, &lab, sqlStr, labID); err != nil {
		return nil, err
	}
	return &lab, nil
}

func QueryTotalAmountInLabByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM lab WHERE course_id = ?`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, courseID); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryTotalAmountInLabByCourseIDs(ctx context.Context, rdbClient storage.RDBClient, courseIDs []uint64) (int, error) {
	if len(courseIDs) == 0 {
		return 0, nil
	}
	query, args, err := squirrel.Select("COUNT(1)").
		From("lab").
		Where(squirrel.Eq{"course_id": courseIDs}).
		ToSql()
	if err != nil {
		return 0, err
	}
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryLabSubmitInfosByUserIDAndCourseID(ctx context.Context, rdbClient storage.RDBClient, userID, courseID uint64, offset, limit int) ([]*LabInfoByUserIDAndCourseID, error) {
	const sqlStr = `
SELECT lab.*, lab_submit.report_url, lab_submit.score, lab_submit.is_finish, lab_submit.comment
FROM
lab INNER JOIN lab_submit INNER JOIN
(SELECT lab_submit.id
FROM lab
INNER JOIN
lab_submit
ON lab.course_id = ?
AND lab.id = lab_submit.lab_id
AND lab_submit.user_id = ?
ORDER BY lab.id DESC
LIMIT ?, ?) AS ll
ON lab_submit.lab_id = lab.id
AND lab_submit.id = ll.id
`
	infos := make([]*LabInfoByUserIDAndCourseID, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, sqlStr, courseID, userID, offset, limit); err != nil {
		return nil, err
	}
	return infos, nil
}

func QueryLabSubmitInfosByUserIDAndCourseIDs(ctx context.Context, rdbClient storage.RDBClient, userID uint64, courseIDs []uint64, offset, limit int) ([]*LabInfoByUserIDAndCourseID, error) {
	if len(courseIDs) == 0 {
		return nil, nil
	}
	const sqlStr = `
SELECT lab.*, lab_submit.report_url, lab_submit.score, lab_submit.is_finish, lab_submit.comment
FROM lab INNER JOIN lab_submit INNER JOIN
(SELECT lab_submit.id
FROM lab
INNER JOIN
lab_submit
ON lab.course_id IN (?)
AND lab.id = lab_submit.lab_id
AND lab_submit.user_id = ?
ORDER BY lab.id DESC
LIMIT ?, ?) AS ll
ON lab_submit.lab_id = lab.id
AND lab_submit.id = ll.id
`
	query, args, err := sqlx.In(sqlStr, courseIDs, userID, offset, limit)
	if err != nil {
		return nil, err
	}

	infos := make([]*LabInfoByUserIDAndCourseID, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, query, args...); err != nil {
		return nil, err
	}
	return infos, nil
}

func BatchInsertLabs(ctx context.Context, rdbClient storage.RDBClient, labs []*Lab) error {
	if len(labs) == 0 {
		return nil
	}

	const sqlStr = `
INSERT INTO lab
(course_id, title, content, attachment_url, dead_line, created_at, updated_at)
VALUES (:course_id, :title, :content, :attachment_url, :dead_line, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, labs)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for index := range labs {
		labs[index].ID = uint64(lastID) + uint64(index)
	}

	return nil
}

func QueryLabsByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, offset, limit int) ([]*Lab, error) {

	const sqlStr = `
SELECT lab.*
FROM lab INNER JOIN
(SELECT id
FROM lab
WHERE course_id = ?
ORDER BY id DESC
LIMIT ?, ?) AS l
ON lab.id = l.id
`
	labs := make([]*Lab, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &labs, sqlStr, courseID, offset, limit); err != nil {
		return nil, err
	}

	return labs, nil
}

func QueryLabIDToCourseIDMapByLabIDs(ctx context.Context, rdbClient storage.RDBClient, labIDs []uint64) (map[uint64]uint64, []uint64, error) {
	if len(labIDs) == 0 {
		return nil, nil, nil
	}
	query, args, err := squirrel.Select("id", "course_id").
		From("lab").
		Where(squirrel.Eq{"id": labIDs}).
		ToSql()
	if err != nil {
		return nil, nil, err
	}

	type IDToCourseID struct {
		ID       uint64 `db:"id"`
		CourseID uint64 `db:"course_id"`
	}
	IDInfos := make([]*IDToCourseID, 0, len(labIDs))
	if err := sqlx.SelectContext(ctx, rdbClient, &IDInfos, query, args...); err != nil {
		return nil, nil, err
	}
	if len(IDInfos) == 0 {
		return nil, nil, nil
	}

	m := make(map[uint64]uint64, len(IDInfos))
	courseIDs := make([]uint64, len(IDInfos))
	for index, info := range IDInfos {
		m[info.ID] = info.CourseID
		courseIDs[index] = info.CourseID
	}
	return m, courseIDs, nil
}

func QueryLabMapsByIDs(ctx context.Context, rdbClient storage.RDBClient, IDs []uint64) (map[uint64]*Lab, error) {
	if len(IDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("lab").
		Where(squirrel.Eq{"id": IDs}).
		ToSql()
	if err != nil {
		return nil, err
	}

	labs := make([]*Lab, 0, len(IDs))
	if err := sqlx.SelectContext(ctx, rdbClient, &labs, query, args...); err != nil {
		return nil, err
	}

	if len(labs) == 0 {
		return nil, nil
	}

	m := make(map[uint64]*Lab, len(labs))
	for _, lab := range labs {
		m[lab.ID] = lab
	}
	return m, nil
}

func QueryLabIDsByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) ([]uint64, error) {
	const sqlStr = `SELECT id FROM lab WHERE course_id = ?`
	var ids []uint64
	if err := sqlx.SelectContext(ctx, rdbClient, &ids, sqlStr, courseID); err != nil {
		return nil, err
	}
	return ids, nil
}

func QueryCourseIDByLabID(ctx context.Context, rdbClient storage.RDBClient, labID uint64) (uint64, error) {
	const sqlStr = `SELECT course_id FROM lab WHERE id = ?`
	var courseID uint64
	if err := sqlx.GetContext(ctx, rdbClient, &courseID, sqlStr, labID); err != nil {
		return 0, err
	}
	return courseID, nil
}

func QueryLabIDToDeadlineMapAfterDeadline(ctx context.Context, rdbClient storage.RDBClient, labIDs []uint64) (map[uint64]time.Time, error) {
	if len(labIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("id", "dead_line").
		From("lab").
		Where(squirrel.Eq{"id": labIDs}).
		Where(squirrel.Lt{"dead_line": time.Now()}).
		ToSql()
	if err != nil {
		return nil, err
	}
	type labIDAndDeadline struct {
		Deadline time.Time `db:"dead_line"`
		LabID    uint64    `db:"id"`
	}
	var infos []*labIDAndDeadline
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, query, args...); err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, nil
	}
	m := make(map[uint64]time.Time)
	for _, info := range infos {
		m[info.LabID] = info.Deadline
	}
	return m, nil
}
