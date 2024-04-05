package courseResource

import (
	"context"
	"database/sql"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
)

type CourseResourceService struct {
	Dao    *repository.Dao
	Logger *log.Logger
}

func NewCourseResourceService(dao *repository.Dao, logger *log.Logger) *CourseResourceService {
	return &CourseResourceService{
		Dao:    dao,
		Logger: logger,
	}
}

func (c *CourseResourceService) InsertCourseResource(ctx context.Context, courseID uint64, title, content, attachmentURL string) error {
	now := time.Now()
	courseRecourse := &model.CourseResource{
		CourseID:      courseID,
		Title:         title,
		Content:       content,
		AttachMentURL: attachmentURL,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := courseRecourse.Insert(ctx, c.Dao.Storage.RDB); err != nil {
		c.Logger.Errorf(err, "insert courseRecourse %+v failed", courseRecourse)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CourseResourceService) ListCourseResource(ctx context.Context, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total           int
		courseResources []*model.CourseResource
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfCourseResourceByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountOfCourseResourceByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountOfCourseResourceByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			courseResources, err = model.QueryCourseResourcesByCourseID(ctx, c.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCourseResourcesByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCourseResourcesByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	records := make([]*CourseResourcesData, len(courseResources))
	for index, courseResource := range courseResources {
		records[index] = &CourseResourcesData{
			ID:            courseResource.ID,
			Title:         courseResource.Title,
			Content:       courseResource.Content,
			AttachMentURL: courseResource.AttachMentURL,
			CreatedAt:     courseResource.CreatedAt,
			UpdatedAt:     courseResource.UpdatedAt,
		}
	}

	return &PageResponse{
		Records:  courseResources,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseResourceService) DeleteCourseResource(ctx context.Context, courseResourceID uint64) error {
	if err := model.DeleteCourseResourceByID(ctx, c.Dao.Storage.RDB, courseResourceID); err != nil {
		c.Logger.Errorf(err, "delete courseResource by ID(%d) failed", courseResourceID)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CourseResourceService) getCourseResource(ctx context.Context, courseResourceID uint64) (*model.CourseResource, error) {
	courseResource, err := model.QueryCourseResourceByID(ctx, c.Dao.Storage.RDB, courseResourceID)
	if err != nil {
		return nil, err
	}
	return courseResource, nil
}

func (c *CourseResourceService) GetCourseResource(ctx context.Context, courseResourceID uint64) (*CourseResourcesData, error) {
	courseResource, err := c.getCourseResource(ctx, courseResourceID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("course resource is not found by ID(%d)", courseResourceID)
		return nil, errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "query course resource by ID(%d) failed", courseResourceID)
		return nil, errorx.InternalErr(err)
	}

	return &CourseResourcesData{
		ID:            courseResource.ID,
		Title:         courseResource.Title,
		Content:       courseResource.Content,
		AttachMentURL: courseResource.AttachMentURL,
		CreatedAt:     courseResource.CreatedAt,
		UpdatedAt:     courseResource.UpdatedAt,
	}, nil
}

func (c *CourseResourceService) UpdateCourseResource(ctx context.Context, courseResourceID uint64, title, content, attachmentURL string) error {
	courseResource, err := c.getCourseResource(ctx, courseResourceID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("course resource is not found by ID(%d)", courseResourceID)
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "query course resource by ID(%d) failed", courseResourceID)
		return errorx.InternalErr(err)
	}

	courseResource.Title = title
	courseResource.AttachMentURL = attachmentURL
	courseResource.Content = content

	if err := courseResource.Update(ctx, c.Dao.Storage.RDB); err != nil {
		c.Logger.Errorf(err, "update courseResource %+v failed", courseResource)
		return errorx.InternalErr(err)
	}
	return nil
}
