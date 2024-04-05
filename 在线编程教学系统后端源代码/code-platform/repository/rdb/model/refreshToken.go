package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type RefreshToken struct {
	ExpireAt  time.Time `db:"expire_at"`
	CreatedAt time.Time `db:"created_at"`
	Token     string    `db:"token"`
	ID        uint64    `db:"id"`
	UserID    uint64    `db:"user_id"`
}

func (r *RefreshToken) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("refresh_token").
		Columns("user_id", "token", "expire_at", "created_at").
		Values(r.UserID, r.Token, r.ExpireAt, r.CreatedAt).
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
	r.ID = uint64(lastID)
	return nil
}

func QueryRefreshTokenByToken(ctx context.Context, rdbClient storage.RDBClient, token string) (*RefreshToken, error) {
	const sqlStr = `SELECT * FROM refresh_token WHERE token = ?`
	var refreshToken RefreshToken
	if err := sqlx.GetContext(ctx, rdbClient, &refreshToken, sqlStr, token); err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func BatchInsertRefreshTokens(ctx context.Context, rdbClient storage.RDBClient, refreshTokens []*RefreshToken) error {
	if len(refreshTokens) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO refresh_token
(user_id, token, expire_at, created_at)
VALUES (:user_id, :token, :expire_at, :created_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, refreshTokens)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range refreshTokens {
		refreshTokens[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func (r *RefreshToken) Update(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Update("refresh_token").
		SetMap(map[string]interface{}{
			"user_id":    r.UserID,
			"token":      r.Token,
			"expire_at":  r.ExpireAt,
			"created_at": r.CreatedAt,
		}).Where(squirrel.Eq{"id": r.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, sqlStr, args...)
	return err
}
