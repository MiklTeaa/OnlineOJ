package course_test

import (
	"context"
	"testing"
	"time"

	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository/rdb/model"

	"github.com/stretchr/testify/require"
)

func TestAuthAddComment(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course")

	const (
		teacherID = 1
		studentID = 2
	)

	course := &model.Course{
		TeacherID: teacherID,
		Language:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	arrangeCourse := &model.ArrangeCourse{
		UserID:    studentID,
		CourseID:  course.ID,
		CreatedAt: now,
		UpdatedAt: now,
		IsPass:    true,
	}

	err = arrangeCourse.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		courseID      uint64
	}{
		{label: "op by teacher", userID: teacherID, courseID: course.ID},
		{label: "op by student", userID: studentID, courseID: course.ID},
		{label: "invalid course id", userID: teacherID, courseID: 10, expectedError: errorx.ErrIsNotFound},
		{label: "no auth user id", userID: 10, courseID: course.ID, expectedError: errorx.ErrFailToAuth},
		{label: "canceled", userID: 10, courseID: course.ID, expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		err := courseService.AuthAddComment(ctx, c.userID, c.courseID)
		require.Equal(t, c.expectedError, err, c.label)
	}
}

func TestAuthCourseForTeacher(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course")

	const teacherID = 1
	now := time.Now()
	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		courseID      uint64
		userID        uint64
	}{
		{label: "normal", courseID: course.ID, userID: teacherID},
		{label: "course not found", courseID: 10, userID: teacherID, expectedError: errorx.ErrIsNotFound},
		{label: "teacher no auth", courseID: course.ID, userID: 10, expectedError: errorx.ErrFailToAuth},
	} {
		err := courseService.AuthCourseForTeacher(ctx, c.courseID, c.userID)
		require.Equal(t, c.expectedError, err, c.label)
	}
}

func TestQueryWhetherStudentInCourse(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "arrange_course")

	const (
		courseID  = 1
		studentID = 1
	)
	now := time.Now()
	arrangeCourse := &model.ArrangeCourse{
		CourseID:  courseID,
		UserID:    studentID,
		IsPass:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := arrangeCourse.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		courseID      uint64
		studentID     uint64
	}{
		{label: "normal", courseID: courseID, studentID: studentID, expectedError: nil},
		{label: "not found", courseID: courseID, studentID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		err := courseService.QueryWhetherStudentInCourse(ctx, c.courseID, c.studentID)
		require.Equal(t, c.expectedError, err)
	}
}
