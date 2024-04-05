package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type CheckInRecord struct {
	DeadLine  time.Time `db:"dead_line"`
	CreatedAt time.Time `db:"created_at"`
	Name      string    `db:"name"`
	ID        uint64    `db:"id"`
	CourseID  uint64    `db:"course_id"`
}

func QueryTotalAmountInCheckInRecordByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM check_in_record WHERE course_id = ?`
	var total int
	err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, courseID)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func QueryCheckInRecordsByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, offset, limit int) ([]*CheckInRecord, error) {
	const sqlStr = `
SELECT check_in_record.*
FROM check_in_record
INNER JOIN
(SELECT id
FROM check_in_record
WHERE course_id = ?
ORDER BY id DESC
LIMIT ?, ?) AS c
ON check_in_record.id = c.id
`
	var records []*CheckInRecord
	if err := sqlx.SelectContext(ctx, rdbClient, &records, sqlStr, courseID, offset, limit); err != nil {
		return nil, err
	}
	return records, nil
}

func (c *CheckInRecord) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("check_in_record").
		Columns("course_id", "name", "created_at", "dead_line").
		Values(c.CourseID, c.Name, c.CreatedAt, c.DeadLine).
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

func QueryCheckInRecordWithIsCheckInStatusByCourseIDAndUserID(ctx context.Context, rdbClient storage.RDBClient, courseID, userID uint64, offset, limit int) ([]*CheckInRecordWithIsCheckInStatus, error) {
	const sqlStr = `
SELECT check_in_record.*, cc.is_check_in
FROM check_in_record INNER JOIN
(SELECT check_in_record.id ,check_in_detail.is_check_in
FROM check_in_record
INNER JOIN
check_in_detail
ON check_in_detail.user_id = ?
AND check_in_detail.record_id = check_in_record.id 
AND check_in_record.course_id = ?
ORDER BY check_in_record.id DESC
LIMIT ?, ?
) AS cc
ON check_in_record.id = cc.id
`
	infos := make([]*CheckInRecordWithIsCheckInStatus, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, sqlStr, userID, courseID, offset, limit); err != nil {
		return nil, err
	}
	return infos[:len(infos):len(infos)], nil
}

func QueryCheckInRecordByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) (*CheckInRecord, error) {
	const sqlStr = `SELECT * FROM check_in_record WHERE id = ?`
	var record CheckInRecord
	err := sqlx.GetContext(ctx, rdbClient, &record, sqlStr, ID)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func DeleteCheckInRecordByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) error {
	const sqlStr = `DELETE FROM check_in_record WHERE id = ? LIMIT 1`
	_, err := rdbClient.ExecContext(ctx, sqlStr, ID)
	return err
}

func BatchInsertCheckInRecords(ctx context.Context, rdbClient storage.RDBClient, checkInRecords []*CheckInRecord) error {
	if len(checkInRecords) == 0 {
		return nil
	}

	const sqlStr = `
INSERT INTO check_in_record
(course_id, name, created_at, dead_line)
VALUES (:course_id, :name, :created_at, :dead_line)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, checkInRecords)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range checkInRecords {
		checkInRecords[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func QueryTotalAmountInCheckInRecordByCourseIDs(ctx context.Context, rdbClient storage.RDBClient, courseIDs []uint64) (int, error) {
	if len(courseIDs) == 0 {
		return 0, nil
	}
	query, args, err := squirrel.Select("COUNT(1)").
		From("check_in_record").
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

func QueryCheckInRecordsByCourseIDs(ctx context.Context, rdbClient storage.RDBClient, courseIDs []uint64, offset, limit int) ([]*CheckInRecord, error) {
	if len(courseIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("check_in_record").
		Where(squirrel.Eq{"course_id": courseIDs}).
		OrderBy("id DESC").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var checkInRecords []*CheckInRecord
	if err := sqlx.SelectContext(ctx, rdbClient, &checkInRecords, query, args...); err != nil {
		return nil, err
	}
	return checkInRecords, nil
}

func QueryCheckInRecordsByCourseIDsWithoutTimeout(ctx context.Context, rdbClient storage.RDBClient, courseIDs []uint64) ([]*CheckInRecord, error) {
	if len(courseIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("check_in_record").
		Where(squirrel.Eq{"course_id": courseIDs}).
		Where(squirrel.Gt{"dead_line": time.Now()}).
		OrderBy("id DESC").ToSql()
	if err != nil {
		return nil, err
	}

	var checkInRecords []*CheckInRecord
	if err := sqlx.SelectContext(ctx, rdbClient, &checkInRecords, query, args...); err != nil {
		return nil, err
	}
	return checkInRecords, nil
}

func QueryCheckInRecordNamesByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) ([]string, error) {
	const sqlStr = `SELECT name FROM check_in_record WHERE course_id = ?`
	var names []string
	if err := sqlx.SelectContext(ctx, rdbClient, &names, sqlStr, courseID); err != nil {
		return nil, err
	}
	return names, nil
}

func QueryCheckInRecordIDsByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) ([]uint64, error) {
	const sqlStr = `SELECT id FROM check_in_record WHERE course_id = ?`
	var ids []uint64
	if err := sqlx.SelectContext(ctx, rdbClient, &ids, sqlStr, courseID); err != nil {
		return nil, err
	}
	return ids, nil
}

func QueryCourseIDByRecordID(ctx context.Context, rdbClient storage.RDBClient, recordID uint64) (uint64, error) {
	const sqlStr = `SELECT course_id FROM check_in_record WHERE id = ?`
	var courseID uint64
	if err := sqlx.GetContext(ctx, rdbClient, &courseID, sqlStr, recordID); err != nil {
		return 0, err
	}
	return courseID, nil
}
