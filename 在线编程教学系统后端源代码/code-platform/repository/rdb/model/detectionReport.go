package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type DetectionReport struct {
	CreatedAt time.Time `db:"created_at"`
	Data      []byte    `db:"data"`
	ID        uint64    `db:"id"`
	LabID     uint64    `db:"lab_id"`
}

func (d *DetectionReport) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("detection_report").
		Columns("lab_id", "data", "created_at").
		Values(d.LabID, d.Data, d.CreatedAt).
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
	d.ID = uint64(lastID)
	return nil
}

func QueryDetectionReportsByLabID(ctx context.Context, rdbClient storage.RDBClient, labID uint64, offset, limit int) ([]*DetectionReportIDWithCreatedAt, error) {
	const sqlStr = `
SELECT detection_report.id, detection_report.created_at
FROM detection_report INNER JOIN
(SELECT id
FROM detection_report
WHERE lab_id = ?
ORDER BY id DESC
LIMIT ?, ?
) AS d
ON d.id = detection_report.id
`
	var detectionReports []*DetectionReportIDWithCreatedAt
	if err := sqlx.SelectContext(ctx, rdbClient, &detectionReports, sqlStr, labID, offset, limit); err != nil {
		return nil, err
	}
	return detectionReports, nil
}

func QueryTotalAmountDetectionReportByLabID(ctx context.Context, rdbClient storage.RDBClient, labID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM detection_report WHERE lab_id = ?`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, labID); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryDetectionReportByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) (*DetectionReport, error) {
	const sqlStr = `SELECT * FROM detection_report WHERE id = ?`
	var info DetectionReport
	if err := sqlx.GetContext(ctx, rdbClient, &info, sqlStr, ID); err != nil {
		return nil, err
	}
	return &info, nil
}

func BatchInsertDetectionReports(ctx context.Context, rdbClient storage.RDBClient, detectionReports []*DetectionReport) error {
	if len(detectionReports) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO detection_report
(lab_id, data, created_at)
VALUES (:lab_id, :data,:created_at)
`
	if _, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, detectionReports); err != nil {
		return err
	}
	return nil
}
