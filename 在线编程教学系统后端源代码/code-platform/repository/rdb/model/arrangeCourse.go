package model

import (
	"context"
	"time"

	"code-platform/storage"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type ArrangeCourse struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	ID        uint64    `db:"id"`
	UserID    uint64    `db:"user_id"`
	CourseID  uint64    `db:"course_id"`
	IsPass    bool      `db:"is_pass"`
}

func QueryTotalAmountOfArrangeCourseByStudentIDWithPass(ctx context.Context, rdbClient storage.RDBClient, studentID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM arrange_course WHERE user_id = ? AND is_pass = TRUE`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, studentID); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryArrangeCourseExistsByCourseIDAndUserID(ctx context.Context, rdbClient storage.RDBClient, courseID, userID uint64) error {
	const sqlStr = `SELECT 1 FROM arrange_course WHERE course_id = ? AND user_id = ? AND is_pass = TRUE`
	var i uint8
	if err := sqlx.GetContext(ctx, rdbClient, &i, sqlStr, courseID, userID); err != nil {
		return err
	}
	return nil
}

func (a *ArrangeCourse) Insert(ctx context.Context, rdbClient storage.RDBClient) error {
	sqlStr, args, err := squirrel.Insert("arrange_course").
		Columns("user_id", "course_id", "is_pass", "created_at", "updated_at").
		Values(a.UserID, a.CourseID, a.IsPass, a.CreatedAt, a.UpdatedAt).
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
	a.ID = uint64(lastID)
	return nil
}

func QueryTotalAmountOfArrangeCourseWithoutPassByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM arrange_course WHERE course_id = ? AND is_pass = FALSE`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, courseID); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryTotalAmountOfArrangeCourseWithPassByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) (int, error) {
	const sqlStr = `SELECT COUNT(1) FROM arrange_course WHERE course_id = ? AND is_pass = TRUE`
	var total int
	if err := sqlx.GetContext(ctx, rdbClient, &total, sqlStr, courseID); err != nil {
		return 0, err
	}
	return total, nil
}

func QueryAllUsersInArrangeCourseWithPassByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) ([]*User, error) {
	const sqlStr = `
SELECT user.*
FROM arrange_course
INNER JOIN
user
ON arrange_course.course_id = ?
AND arrange_course.user_id = user.id
AND arrange_course.is_pass = TRUE
`
	var users []*User
	if err := sqlx.SelectContext(ctx, rdbClient, &users, sqlStr, courseID); err != nil {
		return nil, err
	}
	return users, nil
}

func QueryUsersInArrangeCourseWithoutPassByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, offset, limit int) ([]*User, error) {
	const sqlStr = `
SELECT *
FROM user INNER JOIN
(SELECT user.id
FROM arrange_course
INNER JOIN
user
ON arrange_course.course_id = ?
AND arrange_course.user_id = user.id
AND arrange_course.is_pass = FALSE
ORDER BY arrange_course.id DESC
LIMIT ?, ?) AS au
ON user.id = au.id
`
	var users []*User
	if err := sqlx.SelectContext(ctx, rdbClient, &users, sqlStr, courseID, offset, limit); err != nil {
		return nil, err
	}
	return users, nil
}

func QueryUsersInArrangeCourseWithPassByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, offset, limit int) ([]*User, error) {
	const sqlStr = `
SELECT user.*
FROM user
INNER JOIN
(SELECT user.id
FROM arrange_course
INNER JOIN
user
ON arrange_course.course_id = ?
AND arrange_course.user_id = user.id
AND arrange_course.is_pass = TRUE
ORDER BY arrange_course.id DESC
LIMIT ?, ?
) as au
ON user.id = au.id
`
	var users []*User
	if err := sqlx.SelectContext(ctx, rdbClient, &users, sqlStr, courseID, offset, limit); err != nil {
		return nil, err
	}
	return users, nil
}

func BatchInsertArrangeCourses(ctx context.Context, rdbClient storage.RDBClient, arrangeCourses []*ArrangeCourse) error {
	if len(arrangeCourses) == 0 {
		return nil
	}
	const sqlStr = `
INSERT INTO
arrange_course
(user_id, course_id, is_pass, created_at, updated_at)
VALUES (:user_id, :course_id, :is_pass, :created_at, :updated_at)
`
	result, err := sqlx.NamedExecContext(ctx, rdbClient, sqlStr, arrangeCourses)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for index := range arrangeCourses {
		arrangeCourses[index].ID = uint64(lastID) + uint64(index)
	}
	return nil
}

func BatchUpdateIsPassInArrangeCourseByCourseIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, userIDs []uint64) error {
	if len(userIDs) == 0 {
		return nil
	}

	query, args, err := squirrel.Update("arrange_course").
		Set("is_pass", true).
		Where(squirrel.Eq{"course_id": courseID, "user_id": userIDs}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = rdbClient.ExecContext(ctx, query, args...)
	return err
}

func BatchDeleteArrangeCourseByCourseIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, userIDs []uint64) error {
	if len(userIDs) == 0 {
		return nil
	}
	query, args, err := squirrel.Delete("arrange_course").
		Where(squirrel.Eq{"course_id": courseID, "user_id": userIDs}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = rdbClient.ExecContext(ctx, query, args...)
	return err
}

func QueryUserIDsInArrangeCourseByCourseID(ctx context.Context, rdbClient storage.RDBClient, courseID uint64) ([]uint64, error) {
	const sqlStr = `SELECT user_id FROM arrange_course WHERE course_id = ? AND is_pass = TRUE`
	var userIDs []uint64
	if err := sqlx.SelectContext(ctx, rdbClient, &userIDs, sqlStr, courseID); err != nil {
		return nil, err
	}
	return userIDs, nil
}

func QueryCourseIDsInArrangeCourseByUserID(ctx context.Context, rdbClient storage.RDBClient, userID uint64) ([]uint64, error) {
	const sqlStr = `SELECT course_id FROM arrange_course WHERE user_id = ? AND is_pass = TRUE`
	var courseIDs []uint64
	if err := sqlx.SelectContext(ctx, rdbClient, &courseIDs, sqlStr, userID); err != nil {
		return nil, err
	}
	return courseIDs, nil
}

func BatchUpdateIsPassInArrangeCourseByIDs(ctx context.Context, rdbClient storage.RDBClient, arrangeCourseIDs []uint64, isPass bool) error {
	if len(arrangeCourseIDs) == 0 {
		return nil
	}
	query, args, err := squirrel.Update("arrange_course").
		Set("is_pass", true).
		Where(squirrel.Eq{"id": arrangeCourseIDs}).
		ToSql()
	if err != nil {
		return err
	}
	if _, err := rdbClient.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	return nil
}

func QueryArrangeCourseExistMapByCourseIDAndUserIDs(ctx context.Context, rdbClient storage.RDBClient, courseID uint64, userIDs []uint64) (map[uint64]*ArrangeCourse, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	query, args, err := squirrel.Select("*").
		From("arrange_course").
		Where(squirrel.Eq{"course_id": courseID, "user_id": userIDs}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var arrangeCourses []*ArrangeCourse
	if err := sqlx.SelectContext(ctx, rdbClient, &arrangeCourses, query, args...); err != nil {
		return nil, err
	}

	m := make(map[uint64]*ArrangeCourse, len(userIDs))
	// 默认以为不存在
	for _, userID := range userIDs {
		m[userID] = nil
	}

	// 更新键值对
	for index := range arrangeCourses {
		m[arrangeCourses[index].UserID] = arrangeCourses[index]
	}
	return m, nil
}

func DeleteArrangeCourseByCourseIDAndUserID(ctx context.Context, rdbClient storage.RDBClient, courseID, userID uint64) error {
	const sqlStr = `DELETE FROM arrange_course WHERE course_id = ? AND user_id = ? LIMIT 1`
	_, err := rdbClient.ExecContext(ctx, sqlStr, courseID, userID)
	return err
}
