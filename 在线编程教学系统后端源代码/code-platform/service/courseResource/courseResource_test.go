package courseResource_test

import (
	"context"
	"testing"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	. "code-platform/service/courseResource"
	"code-platform/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHelper() (*storage.Storage, *CourseResourceService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	courseResourceService := NewCourseResourceService(dao, log.Sub("lab"))
	return testStorage, courseResourceService
}

func TestInsertCourseResource(t *testing.T) {
	testStorage, courseResourceService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "course_resource")
	now := time.Now()
	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	err = courseResourceService.InsertCourseResource(ctx, course.ID, "", "", "")
	require.NoError(t, err)
}

func TestListCourseResource(t *testing.T) {
	testStorage, courseResourceService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	const courseID = 1

	testx.MustTruncateTable(ctx, testStorage.RDB, "course_resource")

	courseResources := []*model.CourseResource{
		{CourseID: courseID, CreatedAt: now, UpdatedAt: now},
		{CourseID: courseID, CreatedAt: now, UpdatedAt: now},
	}

	err := model.BatchInsertCourseResources(ctx, testStorage.RDB, courseResources)
	require.NoError(t, err)
	for _, c := range []struct {
		expectedError error
		label         string
	}{
		{label: "normal", expectedError: nil},
		{label: "canceled", expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := courseResourceService.ListCourseResource(ctx, courseID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp.Records, 2)
		}
	}

}

func TestDeleteCourseResource(t *testing.T) {
	testStorage, courseResourceService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "course_resource")
	now := time.Now()
	const teacherID = 1

	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	courseResource := &model.CourseResource{
		CourseID:  course.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = courseResource.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)
	err = courseResourceService.DeleteCourseResource(ctx, courseResource.ID)
	require.NoError(t, err)
}

func TestGetCourseResource(t *testing.T) {
	testStorage, courseResourceService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course_resource")
	now := time.Now()
	const courseID = 1

	courseResource := &model.CourseResource{
		CourseID:  courseID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := courseResource.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError    error
		label            string
		courseResourceID uint64
	}{
		{label: "normal", courseResourceID: 1, expectedError: nil},
		{label: "not found", courseResourceID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		_, err := courseResourceService.GetCourseResource(ctx, c.courseResourceID)
		assert.Equalf(t, c.expectedError, err, c.label)
	}
}

func TestUpdateCourseResource(t *testing.T) {
	testStorage, courseResourceService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course_resource")
	now := time.Now()
	const courseID = 1

	courseResource := &model.CourseResource{
		CourseID:  courseID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := courseResource.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError    error
		label            string
		courseResourceID uint64
	}{
		{label: "normal", courseResourceID: 1, expectedError: nil},
		{label: "not found", courseResourceID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		err := courseResourceService.UpdateCourseResource(ctx, c.courseResourceID, "", "", "")
		assert.Equalf(t, c.expectedError, err, c.label)
	}
}
