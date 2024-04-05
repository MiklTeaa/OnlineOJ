package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type CheckInDetail struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	ID        uint64    `db:"id"`
	RecordID  uint64    `db:"record_id"`
	UserID    uint64    `db:"user_id"`
	IsCheckIn bool      `db:"is_check_in"`
}

func (c *CheckInDetail) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("check_in_detail").
		Columns("record_id", "user_id", "is_check_in", "created_at", "updated_at").
		Values(c.RecordID, c.UserID, c.IsCheckIn, c.CreatedAt, c.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}
	result, err := rdbClient.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	c.ID = uint64(lastID)
	if err != nil {
		return err
	}
	return nil
}

func QueryTotalAmountInCheckInDetailByCheckRecordID(ctx context.Context, rdbClient storage.RDBClient, recordID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM check_in_detail WHERE record_id = ?`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, recordID); err != nil {
		return 0, err
	}
	return total, nil
}

func BatchInsertCheckInDetails(ctx context.Context, rdbClient storage.RDBClient, details []*CheckInDetail) error {
	if len(details) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO
check_in_detail
(record_id, user_id, is_check_in, created_at, updated_at)
VALUES (:record_id, :user_id, :is_check_in, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, details)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range details {
		details[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func QueryCheckInRecordIDToAmountMapByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (map[uint64]int, error) {
	const sqlStr = `
SELECT check_in_record.id, COUNT(1) AS total
FROM check_in_detail
INNER JOIN
check_in_record
ON check_in_record.course_id = ?
AND check_in_record.id = check_in_detail.record_id
AND check_in_detail.is_check_in = TRUE
GROUP BY check_in_record.id
`
	type checkInAmount struct {
		ID    uint64 `db:"id"`
		Total int    `db:"total"`
	}
	var checkInAmounts []*checkInAmount
	if err := sqlx.SelectContext(ctx, rdbClient, &checkInAmounts, sqlStr, courseID); err != nil {
		return nil, err
	}
	m := make(map[uint64]int)
	for _, v := range checkInAmounts {
		m[v.ID] = v.Total
	}
	return m, nil
}

func QuerySuccessCheckInAmountMapByCourseIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, userIDs []uint64) (map[uint64]int, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	const sqlStr = `
SELECT user_id, COUNT(1) AS total
FROM check_in_record
INNER JOIN
check_in_detail
ON check_in_record.course_id = ?
AND check_in_detail.record_id = check_in_record.id 
AND check_in_detail.is_check_in = TRUE
AND check_in_detail.user_id IN (?)
GROUP BY check_in_record.course_id, check_in_detail.user_id
`
	query, args, err := sqlx.In(sqlStr, courseID, userIDs)
	if err != nil {
		return nil, err
	}

	type checkInAmount struct {
		UserID uint64 `db:"user_id"`
		Total  int    `db:"total"`
	}
	var checkInAmounts []*checkInAmount
	if err := sqlx.SelectContext(ctx, rdbClient, &checkInAmounts, query, args...); err != nil {
		return nil, err
	}

	m := make(map[uint64]int)
	for _, v := range checkInAmounts {
		m[v.UserID] = v.Total
	}
	return m, nil
}

func QueryCheckInAboutUser(ctx context.Context, rdbClient storage.RDBClient, recordID uint64, offset, limit int) ([]*CheckInDetailWithUserData, error) {
	const sqlStr = `
SELECT user.*, c.record_id, c.is_check_in
FROM user INNER JOIN
(SELECT user_id, record_id, is_check_in
FROM check_in_detail
WHERE record_id = ?
ORDER BY id DESC
LIMIT ?, ?) AS c
ON c.user_id = user.id
`
	var infos []*CheckInDetailWithUserData
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, sqlStr, recordID, offset, limit); err != nil {
		return nil, err
	}
	return infos, nil
}

func (c *CheckInDetail) Update(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Update("check_in_detail").
		SetMap(map[string]interface{}{
			"record_id":   c.RecordID,
			"user_id":     c.UserID,
			"is_check_in": c.IsCheckIn,
			"created_at":  c.CreatedAt,
			"updated_at":  c.UpdatedAt,
		}).
		Where(squirrel.Eq{"id": c.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, sqlStr, args...)
	return err
}

func QueryUserIDToCheckInDetailsMapByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (map[uint64][]*CheckInDetail, error) {
	const sqlStr = `
SELECT check_in_detail.*
FROM check_in_detail
INNER JOIN
check_in_record
ON check_in_record.course_id = ?
AND check_in_record.id = check_in_detail.record_id
GROUP BY check_in_detail.user_id, check_in_detail.record_id
`
	var infos []*CheckInDetail
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, sqlStr, courseID); err != nil {
		return nil, err
	}

	if len(infos) == 0 {
		return nil, nil
	}

	m := make(map[uint64][]*CheckInDetail)
	for i, info := range infos {
		m[info.UserID] = append(m[info.UserID], infos[i])
	}
	return m, nil
}

func QueryCheckInDetailByRecordIDAndUserID(ctx context.Context, rdbClient storage.RDBClient, recordID, userID uint64) (*CheckInDetail, error) {
	const sqlStr = `SELECT * FROM check_in_detail WHERE user_id = ? AND record_id = ?`
	var detail CheckInDetail
	if err := sqlx.GetContext(ctx, rdbClient, &detail, sqlStr, userID, recordID); err != nil {
		return nil, err
	}
	return &detail, nil
}

func QueryIsCheckInMapByUserIDAndRecordIDs(ctx context.Context, rdbClient storage.RDBClient, userID uint64, recordIDs []uint64) (map[uint64]bool, error) {
	if len(recordIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("check_in_detail").
		Where(squirrel.Eq{"user_id": userID, "record_id": recordIDs}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var checkInDetails []*CheckInDetail
	if err := sqlx.SelectContext(ctx, rdbClient, &checkInDetails, query, args...); err != nil {
		return nil, err
	}

	m := make(map[uint64]bool, len(checkInDetails))
	for _, checkInDetail := range checkInDetails {
		m[checkInDetail.RecordID] = checkInDetail.IsCheckIn
	}
	return m, nil
}
