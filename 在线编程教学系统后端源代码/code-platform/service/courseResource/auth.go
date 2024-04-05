package courseResource

import (
	"context"
	"database/sql"

	"code-platform/pkg/errorx"
	"code-platform/repository/rdb/model"
)

func (c *CourseResourceService) GetCourseIDByCourseResourceID(ctx context.Context, courseResourceID uint64) (uint64, error) {
	courseID, err := model.QueryCourseIDByCourseResourceID(ctx, c.Dao.Storage.RDB, courseResourceID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return 0, errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "query course resource by id[%d] failed", courseResourceID)
		return 0, errorx.InternalErr(err)
	}
	return courseID, nil
}
