package model

import (
	"context"
	"database/sql"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type LabSubmit struct {
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
	Comment   string        `db:"comment"`
	ReportURL string        `db:"report_url"`
	ID        uint64        `db:"id"`
	LabID     uint64        `db:"lab_id"`
	UserID    uint64        `db:"user_id"`
	Score     sql.NullInt32 `db:"score"`
	IsFinish  bool          `db:"is_finish"`
}

func (l *LabSubmit) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("lab_submit").
		Columns("lab_id", "user_id", "report_url", "score", "is_finish", "comment", "created_at", "updated_at").
		Values(l.LabID, l.UserID, l.ReportURL, l.Score, l.IsFinish, l.Comment, l.CreatedAt, l.UpdatedAt).
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

func (l *LabSubmit) Update(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Update("lab_submit").
		SetMap(map[string]interface{}{
			"lab_id":     l.LabID,
			"user_id":    l.UserID,
			"report_url": l.ReportURL,
			"score":      l.Score,
			"is_finish":  l.IsFinish,
			"comment":    l.Comment,
			"created_at": l.CreatedAt,
			"updated_at": l.UpdatedAt,
		}).Where(squirrel.Eq{"id": l.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, sqlStr, args...)
	return err
}

func QueryAverageScoreMapByCourseIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, userIDs []uint64) (map[uint64]float64, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	const sqlStr = `
SELECT IFNULL(AVG(score), 0) AS score, user_id
FROM lab_submit
INNER JOIN
lab
ON lab.course_id = ?
AND lab.id = lab_submit.lab_id
AND lab_submit.user_id IN (?)
GROUP BY lab_submit.user_id
`
	query, args, err := sqlx.In(sqlStr, courseID, userIDs)
	if err != nil {
		return nil, err
	}
	type labSubmitAvgScore struct {
		UserID uint64  `db:"user_id"`
		Score  float64 `db:"score"`
	}
	var userAvgScores []*labSubmitAvgScore
	if err := sqlx.SelectContext(ctx, rdbClient, &userAvgScores, query, args...); err != nil {
		return nil, err
	}
	m := make(map[uint64]float64)
	for _, v := range userAvgScores {
		m[v.UserID] = v.Score
	}
	return m, nil
}

func DeleteLabSubmitsByLabID(ctx context.Context, rdbClient storage.RDBClient, labID uint64) error {
	const sqlStr = `DELETE FROM lab_submit WHERE lab_id = ?`
	_, err := rdbClient.ExecContext(ctx, sqlStr, labID)
	return err
}

func QueryLabSubmitsByLabIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, labID int32, userIDs []int32) ([]*LabSubmit, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("lab_submit").
		Where(squirrel.Eq{"lab_id": labID, "user_id": userIDs}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var labSubmits []*LabSubmit
	if err := sqlx.SelectContext(ctx, rdbClient, &labSubmits, query, args...); err != nil {
		return nil, err
	}
	return labSubmits, nil
}

func QueryLabSubmitsByUserIDAndLabIDs(ctx context.Context, rdbClient storage.RDBClient, userID int32, labIDs []int32) ([]*LabSubmit, error) {
	if len(labIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("lab_submit").
		Where(squirrel.Eq{"user_id": userID, "lab_id": labIDs}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var labSubmits []*LabSubmit
	if err := sqlx.SelectContext(ctx, rdbClient, &labSubmits, query, args...); err != nil {
		return nil, err
	}
	return labSubmits, nil
}

func QueryLabSubmitByLabIDAndUserID(ctx context.Context, rdbClient storage.RDBClient, labID, userID uint64) (*LabSubmit, error) {
	const sqlStr = `SELECT * FROM lab_submit WHERE lab_id = ? AND user_id = ?`
	var labSubmit LabSubmit
	if err := sqlx.GetContext(ctx, rdbClient, &labSubmit, sqlStr, labID, userID); err != nil {
		return nil, err
	}
	return &labSubmit, nil
}

func QueryLabWithUserLabDatasByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) ([]*LabWithUserData, error) {
	const sqlStr = `
SELECT lab.*, lab_submit.score, lab_submit.user_id
FROM lab
INNER JOIN
lab_submit
ON lab.course_id = ?
AND lab.id = lab_submit.lab_id
`
	var labWithUserDatas []*LabWithUserData
	if err := sqlx.SelectContext(ctx, rdbClient, &labWithUserDatas, sqlStr, courseID); err != nil {
		return nil, err
	}
	return labWithUserDatas, nil
}

func BatchInsertLabSubmits(ctx context.Context, rdbClient storage.RDBClient, labSubmits []*LabSubmit) error {
	if len(labSubmits) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO
lab_submit
(lab_id, user_id, report_url, score, is_finish, comment, created_at, updated_at)
VALUES (:lab_id, :user_id, :report_url, :score, :is_finish, :comment, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, labSubmits)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range labSubmits {
		labSubmits[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func QueryLabSubmitInfosByLabID(ctx context.Context, rdbClient storage.RDBClient, labID uint64, offset, limit int) ([]*LabSubmitInfoByLabID, error) {
	const sqlStr = `
SELECT lab_submit.*, user.name, user.number
FROM lab_submit INNER JOIN user INNER JOIN
(SELECT lab_submit.id
FROM lab_submit INNER JOIN user
ON lab_submit.lab_id = ?
AND lab_submit.user_id = user.id
ORDER BY lab_submit.id DESC
LIMIT ?, ?) AS lu
ON lab_submit.user_id = user.id
AND lab_submit.id = lu.id
`
	infos := make([]*LabSubmitInfoByLabID, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &infos, sqlStr, labID, offset, limit); err != nil {
		return nil, err
	}
	return infos[:len(infos):len(infos)], nil
}
