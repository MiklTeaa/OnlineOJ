package ide_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"code-platform/api/grpc/ide/pb"
	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	. "code-platform/service/ide"
	"code-platform/service/ide/define"
	"code-platform/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHelper() (*storage.Storage, *IDEService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	RunSweaterOpt = false
	ideService := NewIDEService(dao, log.Sub("lab"), NewIDEClient())
	return testStorage, ideService
}

func TestOpenIDE(t *testing.T) {
	testStorage, ideService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "course", "lab")
	testx.MustFlushDB(ctx, testStorage.Pool())
	now := time.Now()

	user := &model.User{Role: 0, Number: "1", CreatedAt: now, UpdatedAt: now}

	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	courses := []*model.Course{
		{Language: 0, CreatedAt: now, UpdatedAt: now},
		{Language: 1, CreatedAt: now, UpdatedAt: now},
		{Language: 2, CreatedAt: now, UpdatedAt: now},
	}
	err = model.BatchInsertCourses(ctx, testStorage.RDB, courses)
	require.NoError(t, err)

	labs := []*model.Lab{
		{CourseID: 1, CreatedAt: now, UpdatedAt: now},
		{CourseID: 2, CreatedAt: now, UpdatedAt: now},
		{CourseID: 3, CreatedAt: now, UpdatedAt: now},
		{CourseID: 10, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		labID         uint64
	}{
		{label: "lab not found", labID: 0, expectedError: errorx.ErrIsNotFound},
		{label: "course not found", labID: 4, expectedError: errorx.ErrIsNotFound},
		{label: "python student", labID: 1, expectedError: nil},
		{label: "cpp student", labID: 2, expectedError: nil},
		{label: "java student", labID: 3, expectedError: nil},
	} {
		url, _, err := ideService.OpenIDE(ctx, c.labID, user.ID)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			fmt.Println(c.label, url)
		}
	}

	_, err = ideService.IDEClient.StopAllIDE(ctx, &pb.Empty{})
	require.NoError(t, err)
}

func TestCheckCode(t *testing.T) {
	testStorage, ideService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "course", "lab")
	testx.MustFlushDB(ctx, testStorage.Pool())
	now := time.Now()
	users := []*model.User{
		// students
		{Role: 0, Number: "1", CreatedAt: now, UpdatedAt: now},
		// teacher
		{Role: 1, Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	const (
		studentID = 1
		teacherID = 2
	)

	err := model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	courses := []*model.Course{
		{Language: 0, TeacherID: teacherID, CreatedAt: now, UpdatedAt: now},
		{Language: 1, TeacherID: teacherID, CreatedAt: now, UpdatedAt: now},
		{Language: 2, TeacherID: teacherID, CreatedAt: now, UpdatedAt: now},
	}
	err = model.BatchInsertCourses(ctx, testStorage.RDB, courses)
	require.NoError(t, err)

	labs := []*model.Lab{
		{CourseID: 1, CreatedAt: now, UpdatedAt: now},
		{CourseID: 2, CreatedAt: now, UpdatedAt: now},
		{CourseID: 3, CreatedAt: now, UpdatedAt: now},
		{CourseID: 10, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		labID         uint64
		teacherID     uint64
	}{
		{label: "lab not found", labID: 0, teacherID: teacherID, expectedError: errorx.ErrIsNotFound},
		{label: "course not found", labID: 4, teacherID: teacherID, expectedError: errorx.ErrIsNotFound},
		{label: "no auth", labID: 1, teacherID: 10, expectedError: errorx.ErrFailToAuth},
		{label: "python teacher", labID: 1, teacherID: teacherID, expectedError: nil},
		{label: "cpp teacher", labID: 2, teacherID: teacherID, expectedError: nil},
		{label: "java teacher", labID: 3, teacherID: teacherID, expectedError: nil},
	} {
		url, _, err := ideService.CheckCode(ctx, c.labID, studentID, c.teacherID)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			fmt.Println(c.label, url)
		}
	}

	_, err = ideService.IDEClient.StopAllIDE(ctx, &pb.Empty{})
	require.NoError(t, err)
}

func TestListContainers(t *testing.T) {
	testStorage, ideService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "course", "lab")
	testx.MustFlushDB(ctx, testStorage.Pool())

	now := time.Now()

	users := []*model.User{
		{Role: 0, Name: "学生", Number: "1", CreatedAt: now, UpdatedAt: now},
		{Role: 1, Name: "教师", Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	ideService.IDEClient.StopAllIDE(ctx, &pb.Empty{})

	err := model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	const (
		studentID = 1
		teacherID = 2
	)

	course := &model.Course{Language: 0, Name: "Python", TeacherID: teacherID, CreatedAt: now, UpdatedAt: now}
	err = course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	lab := &model.Lab{CourseID: 1, Title: "Python 实验", CreatedAt: now, UpdatedAt: now}
	err = lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	url, _, err := ideService.OpenIDE(ctx, lab.ID, studentID)
	require.NoError(t, err)
	fmt.Println(url)

	url, _, err = ideService.CheckCode(ctx, lab.ID, studentID, teacherID)
	require.NoError(t, err)
	fmt.Println(url)

	for _, c := range []struct {
		expectedError error
		label         string
		order         pb.OrderType
		isReverse     bool
	}{
		{label: "by created_time", order: pb.OrderType_byTime, expectedError: nil, isReverse: false},
		{label: "by disksize", order: pb.OrderType_byDiskSize, expectedError: nil},
		{label: "by cpu perc", order: pb.OrderType_byCPU, expectedError: nil},
		{label: "by mem usage", order: pb.OrderType_byMemory, expectedError: nil},
		{label: "by created_time reverse", order: pb.OrderType_byTime, expectedError: nil, isReverse: true},
		{label: "by disksize reverse", order: pb.OrderType_byDiskSize, expectedError: nil, isReverse: true},
		{label: "by cpu perc reverse", order: pb.OrderType_byCPU, expectedError: nil, isReverse: true},
		{label: "by mem usage reverse", order: pb.OrderType_byMemory, expectedError: nil, isReverse: true},
		{label: "canceled", expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}

		resp, err := ideService.ListContainers(ctx, 0, 10, c.order, c.isReverse)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp.Records, 2, c.label)
			require.Equal(t, 2, resp.PageInfo.Total, c.label)
		}
	}

	ctx = context.Background()
	_, err = ideService.IDEClient.StopAllIDE(ctx, &pb.Empty{})
	require.NoError(t, err)
}

func TestStopContainer(t *testing.T) {
	testStorage, ideService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "course", "lab")
	now := time.Now()

	users := []*model.User{
		{Role: 0, Name: "学生", Number: "1", CreatedAt: now, UpdatedAt: now},
		{Role: 1, Name: "教师", Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	ideService.IDEClient.StopAllIDE(ctx, &pb.Empty{})

	err := model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	const (
		studentID = 1
		teacherID = 2
	)

	course := &model.Course{Language: 0, Name: "Python", TeacherID: teacherID, CreatedAt: now, UpdatedAt: now}
	err = course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	lab := &model.Lab{CourseID: course.ID, Title: "Python 实验", CreatedAt: now, UpdatedAt: now}
	err = lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	_, _, err = ideService.CheckCode(ctx, lab.ID, studentID, teacherID)
	require.NoError(t, err)

	resp, err := ideService.ListContainers(ctx, 0, 10, pb.OrderType_byTime, false)
	require.NoError(t, err)
	require.Len(t, resp.Records, 1)

	records := resp.Records.([]*define.ContainerInfo)

	for _, c := range []struct {
		expectedError error
		label         string
		containerID   string
	}{
		{label: "normal", containerID: records[0].ContainerID, expectedError: nil},
		{label: "not found", containerID: "1111", expectedError: errorx.ErrWrongCode},
	} {
		err = ideService.StopContainer(ctx, c.containerID)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}
