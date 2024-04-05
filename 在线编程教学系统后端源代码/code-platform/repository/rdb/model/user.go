package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type User struct {
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	Major        string    `db:"major"`
	Number       string    `db:"number"`
	Name         string    `db:"name"`
	Password     string    `db:"password"`
	Avatar       string    `db:"avatar"`
	College      string    `db:"college"`
	Organization string    `db:"organization"`
	ID           uint64    `db:"id"`
	Grade        uint16    `db:"grade"`
	Role         uint16    `db:"role"`
	Gender       int8      `db:"gender"`
}

func QueryUsersByIDs(ctx context.Context, rdbClient storage.RDBClient, IDs []uint64) ([]*User, error) {
	if len(IDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("user").
		Where(squirrel.Eq{"id": IDs}).
		ToSql()
	if err != nil {
		return nil, err
	}

	users := make([]*User, 0, len(IDs))
	if err := sqlx.SelectContext(ctx, rdbClient, &users, query, args...); err != nil {
		return nil, err
	}
	return users[:len(users):len(users)], nil
}

func QueryUserMapByIDs(ctx context.Context, rdbClient storage.RDBClient, IDs []uint64) (map[uint64]*User, error) {
	users, err := QueryUsersByIDs(ctx, rdbClient, IDs)
	if err != nil {
		return nil, err
	}
	m := make(map[uint64]*User)
	for _, user := range users {
		m[user.ID] = user
	}
	return m, nil
}

func QueryUserByID(ctx context.Context, rdbClient storage.RDBClient, ID uint64) (*User, error) {
	const sqlStr = `SELECT * FROM user WHERE id = ?`
	var user User
	if err := sqlx.GetContext(ctx, rdbClient, &user, sqlStr, ID); err != nil {
		return nil, err
	}
	return &user, nil
}

func QueryUserByNumber(ctx context.Context, rdbClient storage.RDBClient, num string) (*User, error) {
	const sqlStr = `SELECT * FROM user WHERE number = ?`
	var user User
	if err := sqlx.GetContext(ctx, rdbClient, &user, sqlStr, num); err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *User) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("user").
		Columns("role", "number", "name", "password", "avatar", "gender", "college", "grade", "major", "organization", "created_at", "updated_at").
		Values(u.Role, u.Number, u.Name, u.Password, u.Avatar, u.Gender, u.College, u.Grade, u.Major, u.Organization, u.CreatedAt, u.UpdatedAt).
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

func (u *User) Update(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Update("user").SetMap(map[string]interface{}{
		"role":         u.Role,
		"number":       u.Number,
		"name":         u.Name,
		"password":     u.Password,
		"avatar":       u.Avatar,
		"gender":       u.Gender,
		"college":      u.College,
		"grade":        u.Grade,
		"major":        u.Major,
		"organization": u.Organization,
		"created_at":   u.CreatedAt,
		"updated_at":   u.UpdatedAt,
	}).Where(squirrel.Eq{"id": u.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, sqlStr, args...)
	return err
}

func QueryUsersByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, offset, limit int) ([]*User, error) {
	const sqlStr = `
SELECT user.*
FROM user INNER JOIN
(SELECT user_id
FROM arrange_course
WHERE course_id = ?
AND is_pass = TRUE
ORDER BY user_id ASC
LIMIT ?, ?) AS a
ON user.id = a.user_id
`
	var users []*User
	if err := sqlx.SelectContext(ctx, rdbClient, &users, sqlStr, courseID, offset, limit); err != nil {
		return nil, err
	}
	return users, nil
}

func BatchInsertUsers(ctx context.Context, rdbClient storage.RDBClient, users []*User) error {
	if len(users) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO user
(role, number, name, password, avatar, gender, college, grade, major, organization, created_at, updated_at)
VALUES (:role, :number, :name, :password, :avatar, :gender, :college, :grade, :major, :organization, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, users)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range users {
		users[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func QueryUserNumberToUserIDPointerMapByNumber(ctx context.Context, rdbClient storage.RDBClient, numbers []string) (map[string]*uint64, error) {
	if len(numbers) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("number", "id").
		From("user").
		Where(squirrel.Eq{"number": numbers}).
		ToSql()
	if err != nil {
		return nil, err
	}
	type userIDAndNumber struct {
		Number string `db:"number"`
		UserID uint64 `db:"id"`
	}

	var numbersExist []userIDAndNumber
	if err := sqlx.SelectContext(ctx, rdbClient, &numbersExist, query, args...); err != nil {
		return nil, err
	}
	m := make(map[string]*uint64, len(numbers))
	for _, number := range numbers {
		m[number] = nil
	}
	for _, v := range numbersExist {
		userID := v.UserID
		m[v.Number] = &userID
	}
	return m, nil
}

func QueryUserNumberExistsMapByNumber(ctx context.Context, rdbClient storage.RDBClient, numbers []string) (map[string]struct{}, error) {
	if len(numbers) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("number").
		From("user").
		Where(squirrel.Eq{"number": numbers}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var numberExists []string
	if err := sqlx.SelectContext(ctx, rdbClient, &numberExists, query, args...); err != nil {
		return nil, err
	}
	m := make(map[string]struct{}, len(numbers))
	for _, number := range numbers {
		m[number] = struct{}{}
	}
	for _, number := range numberExists {
		m[number] = struct{}{}
	}
	return m, nil
}

func QueryTotalAmountOfUsers(ctx context.Context, rdbClient storage.RDBClient) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM user`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryAllUsers(ctx context.Context, rdbClient storage.RDBClient, offset, limit int) ([]*User, error) {
	const sqlStr = `
SELECT user.*
FROM user INNER JOIN
(SELECT id
FROM user
ORDER BY id ASC
LIMIT ?, ?) AS u
ON user.id = u.id
`
	users := make([]*User, 0, limit)
	if err := sqlx.SelectContext(ctx, rdbClient, &users, sqlStr, offset, limit); err != nil {
		return nil, err
	}
	return users, nil
}
