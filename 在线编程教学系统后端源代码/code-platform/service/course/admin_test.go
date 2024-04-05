package course_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository/rdb/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAllCoursesByAdmin(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course")

	now := time.Now()
	courses := []*model.Course{
		{CreatedAt: now, UpdatedAt: now},
		{CreatedAt: now, UpdatedAt: now},
	}

	err := model.BatchInsertCourses(ctx, testStorage.RDB, courses)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError  error
		label          string
		expectedLength int
	}{
		{label: "normal", expectedLength: 2, expectedError: nil},
		{label: "canceled", expectedLength: 0, expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := courseService.ListAllCoursesByAdmin(ctx, 0, 10)
		assert.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Equal(t, c.expectedLength, resp.PageInfo.Total)
		}
	}
}

func TestUpdateCourseByAdmin(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course")

	now := time.Now()
	course := &model.Course{CreatedAt: now, UpdatedAt: now}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		courseID      uint64
	}{
		{label: "normal", courseID: course.ID, expectedError: nil},
		{label: "not found", courseID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		err = courseService.UpdateCourseByAdmin(ctx, c.courseID, "", "", "", sql.NullString{}, false, false)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}
