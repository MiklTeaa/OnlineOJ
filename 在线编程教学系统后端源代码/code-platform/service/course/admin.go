package course

import (
	"context"
	"database/sql"

	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/repository/rdb/model"
)

func coursesToCourseInfos(courses []*model.Course) []*CourseInfo {
	courseInfos := make([]*CourseInfo, len(courses))
	for i, course := range courses {
		courseInfos[i] = &CourseInfo{
			CourseID:          course.ID,
			CourseName:        course.Name,
			CourseDescription: course.Description,
			PictureURL:        course.PicURL,
			SecretKey:         course.SecretKey.String,
			IsClose:           course.IsClosed,
			Language:          course.Language,
			NeedAudit:         course.NeedAudit,
			CreatedAt:         course.CreatedAt,
			UpdatedAt:         course.UpdatedAt,
		}
	}
	return courseInfos
}

func (c *CourseService) ListAllCoursesByAdmin(ctx context.Context, offset, limit int) (*PageResponse, error) {
	var (
		total   int
		courses []*model.Course
	)
	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountCourses(ctx, c.Dao.Storage.RDB)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountCourses is canceled")
				return err
			default:
				c.Logger.Error(err, "query total amount of course failed")
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			courses, err = model.QueryAllCourses(ctx, c.Dao.Storage.RDB, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryAllCourses is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryAllCourses by offset[%d] and limit[%d] failed")
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		PageInfo: &PageInfo{Total: total},
		Records:  coursesToCourseInfos(courses),
	}, nil
}

func (c *CourseService) UpdateCourseByAdmin(ctx context.Context, id uint64, name string, description string, picURL string, secretKey sql.NullString, isClosed, needAudit bool) error {
	course, err := model.QueryCourseByID(ctx, c.Dao.Storage.RDB, id)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("course is not found by id[%d]", id)
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "Query Course By id[%d] failed", id)
		return errorx.InternalErr(err)
	}

	course.Name = name
	course.Description = description
	course.PicURL = picURL
	course.SecretKey = sql.NullString{
		Valid:  secretKey.String != "",
		String: secretKey.String,
	}
	course.IsClosed = isClosed
	course.NeedAudit = needAudit

	err = course.Update(ctx, c.Dao.Storage.RDB)
	if err != nil {
		c.Logger.Errorf(err, "update course %+v failed", course)
		return errorx.InternalErr(err)
	}

	return nil
}
