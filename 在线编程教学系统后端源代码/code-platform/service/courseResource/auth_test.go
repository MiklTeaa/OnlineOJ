package courseResource_test

import (
	"context"
	"testing"
	"time"

	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository/rdb/model"

	"github.com/stretchr/testify/require"
)

func TestAuthCourseResourceForTeacher(t *testing.T) {
	testStorage, courseResourceService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "course_resource")
	const teacherID = 1
	now := time.Now()

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

	for _, c := range []struct {
		expectedError    error
		label            string
		courseResourceID uint64
	}{
		{label: "normal", courseResourceID: courseResource.ID},
		{label: "not found", courseResourceID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		_, err := courseResourceService.GetCourseIDByCourseResourceID(ctx, c.courseResourceID)
		require.Equal(t, c.expectedError, err, c.label)
	}
}
