package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
)

type Upload struct {
	CreatedAt  time.Time `db:"created_at"`
	BucketName string    `db:"bucket_name"`
	ObjectName string    `db:"object_name"`
	ID         uint64    `db:"id"`
	UserID     uint64    `db:"user_id"`
}

func (u *Upload) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("upload").Columns("user_id", "bucket_name", "object_name", "created_at").
		Values(u.UserID, u.BucketName, u.ObjectName, u.CreatedAt).
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
	u.ID = uint64(lastID)
	return nil
}
