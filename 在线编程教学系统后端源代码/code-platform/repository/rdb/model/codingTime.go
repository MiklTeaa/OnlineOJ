package model

import (
	"context"
	"time"

	"code-platform/pkg/timex"
	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type CodingTime struct {
	CreatedAt     time.Time `db:"created_at"`
	CreatedAtDate time.Time `db:"created_at_date"`
	ID            uint64    `db:"id"`
	LabID         uint64    `db:"lab_id"`
	UserID        uint64    `db:"user_id"`
	Duration      uint32    `db:"duration"`
}

func QueryCodingTimesByCourseIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, userIDs []uint64) ([]*CodingTimeInfoWithUserID, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	const sqlStr = `
SELECT coding_time.user_id, SUM(coding_time.duration) AS duration, coding_time.created_at_date
FROM lab INNER JOIN coding_time
ON lab.course_id = ?
AND lab.id = coding_time.lab_id
AND coding_time.user_id IN (?)
GROUP BY coding_time.user_id, coding_time.created_at_date
ORDER BY coding_time.created_at_date ASC
`
	query, args, err := sqlx.In(sqlStr, courseID, userIDs)
	if err != nil {
		return nil, err
	}

	var codingTimeInfos []*CodingTimeInfoWithUserID
	if err := sqlx.SelectContext(ctx, rdbClient, &codingTimeInfos, query, args...); err != nil {
		return nil, err
	}
	return codingTimeInfos, nil
}

func QueryCodingTimeInfosByUserIDInNaturalYear(ctx context.Context, rdbClient storage.RDBClient, userID uint64) ([]*CodingTimeInfo, error) {
	now := time.Now()
	query, args, err := squirrel.Select("SUM(duration) AS duration", "created_at_date").
		From("coding_time").
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.Expr("created_at_date BETWEEN ? AND ?", timex.StartOfYear(now), timex.EndOfYear(now))).
		GroupBy("user_id", "created_at_date").
		ToSql()
	if err != nil {
		return nil, err
	}
	var codingTimeInfos []*CodingTimeInfo
	if err := sqlx.SelectContext(ctx, rdbClient, &codingTimeInfos, query, args...); err != nil {
		return nil, err
	}
	return codingTimeInfos, nil
}

func QueryCodingTimesMapByLabIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, labID uint64, userIDs []uint64) (map[uint64]uint64, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("user_id", "SUM(duration) AS duration").
		From("coding_time").
		Where(squirrel.Eq{"lab_id": labID, "user_id": userIDs}).
		GroupBy("user_id").
		ToSql()
	if err != nil {
		return nil, err
	}

	type UserIDWithDuration struct {
		UserID   uint64 `db:"user_id"`
		Duration uint64 `db:"duration"`
	}
	var infos []*UserIDWithDuration
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, query, args...); err != nil {
		return nil, err
	}

	if len(infos) == 0 {
		return nil, nil
	}

	m := make(map[uint64]uint64)
	for _, codingTime := range infos {
		m[codingTime.UserID] = codingTime.Duration
	}
	return m, nil
}

func BatchInsertCodingTimes(ctx context.Context, rdbClient storage.RDBClient, codingTimes []*CodingTime) error {
	if len(codingTimes) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO coding_time
(lab_id, user_id, duration, created_at, created_at_date)
VALUES (:lab_id, :user_id, :duration, :created_at, :created_at_date)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, codingTimes)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range codingTimes {
		codingTimes[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}
