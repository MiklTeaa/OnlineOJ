package course

import (
	"context"
	"database/sql"

	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/repository/rdb/model"
)

func (c *CourseService) AuthAddComment(ctx context.Context, userID, courseID uint64) error {
	var (
		isStudentForCourse bool
		isTeacherForCourse bool
	)

	tasks := []func() error{
		func() (err error) {
			err = model.QueryArrangeCourseExistsByCourseIDAndUserID(ctx, c.Dao.Storage.RDB, courseID, userID)
			switch err {
			case nil:
				isStudentForCourse = true
			case sql.ErrNoRows:
				isStudentForCourse = false
			case context.Canceled:
				c.Logger.Debug("QueryArrangeCourseByCourseIDAndUserID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "query arrange course by courseID[%d] and userID[%d] failed", courseID, userID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			teacherID, err := model.QueryCourseTeacherIDByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
				isTeacherForCourse = teacherID == userID
			case sql.ErrNoRows:
				return errorx.ErrIsNotFound
			case context.Canceled:
				c.Logger.Debug("QueryArrangeCourseByCourseIDAndUserID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "query course by id[%d] failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}
	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return err
	}

	if !isStudentForCourse && !isTeacherForCourse {
		return errorx.ErrFailToAuth
	}
	return nil
}

func (c *CourseService) AuthCourseForTeacher(ctx context.Context, courseID, userID uint64) error {
	teacherID, err := model.QueryCourseTeacherIDByCourseID(ctx, c.Dao.Storage.RDB, courseID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("course is not found by id[%d]", courseID)
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "query course failed by id[%d]", courseID)
		return errorx.InternalErr(err)
	}
	if teacherID != userID {
		return errorx.ErrFailToAuth
	}
	return nil
}

func (c *CourseService) QueryWhetherStudentInCourse(ctx context.Context, courseID uint64, studentID uint64) error {
	err := model.QueryArrangeCourseExistsByCourseIDAndUserID(ctx, c.Dao.Storage.RDB, courseID, studentID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "query arrange course exists by courseID[%d] and studentID[%d] failed", courseID, studentID)
		return errorx.InternalErr(err)
	}
	return nil
}
