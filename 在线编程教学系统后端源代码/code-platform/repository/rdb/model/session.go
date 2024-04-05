package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type Session struct {
	CreatedAt time.Time `db:"created_at"`
	ExpireAt  time.Time `db:"expire_at"`
	Token     string    `db:"token"`
	ID        uint64    `db:"id"`
	UserID    uint64    `db:"user_id"`
}

func (s *Session) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("session").
		Columns("user_id", "token", "created_at", "expire_at").
		Values(s.UserID, s.Token, s.CreatedAt, s.ExpireAt).
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
	s.ID = uint64(lastID)
	return nil
}

func QuerySessionByToken(ctx context.Context, rdbClient storage.RDBClient, token string) (*Session, error) {
	const sqlStr = `SELECT * FROM session WHERE token = ?`
	var session Session
	if err := sqlx.GetContext(ctx, rdbClient, &session, sqlStr, token); err != nil {
		return nil, err
	}
	return &session, nil
}

func BatchInsertSessions(ctx context.Context, rdbClient storage.RDBClient, sessions []*Session) error {
	if len(sessions) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO session
(user_id, token, created_at, expire_at)
VALUES (:user_id, :token, :created_at, :expire_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, sessions)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range sessions {
		sessions[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}
