package checkin_test

import (
	"context"
	"testing"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	. "code-platform/service/checkin"
	"code-platform/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHelper() (*storage.Storage, *CheckInService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	checkInService := NewCheckInService(dao, log.Sub("checkin"))
	return testStorage, checkInService
}

func TestPrepareCheckIn(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course", "check_in_record", "check_in_detail")
	const teacherID = 1
	now := time.Now()
	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 2, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{UserID: 3, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	err = checkInService.PrepareCheckIn(ctx, teacherID, course.ID, "第一次签到", 120)
	require.NoError(t, err)
}

func TestStartCheckIn(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	const teacherID = 1
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course")

	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	arrangeCourse := &model.ArrangeCourse{
		UserID:    2,
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
		expireSeconds int
	}{
		{label: "normal", userID: 2, courseID: course.ID, expireSeconds: 120, expectedError: nil},
		{label: "redis timeout", userID: 2, courseID: course.ID, expireSeconds: 1, expectedError: errorx.ErrRedisKeyNil},
		// courseID 无效，因为先查Redis，所以导致 Redis key 找不到
		{label: "course is not found", userID: 2, courseID: 0, expireSeconds: 120, expectedError: errorx.ErrRedisKeyNil},
		{label: "check_in_detail is not found", userID: 1, courseID: course.ID, expireSeconds: 120, expectedError: errorx.ErrIsNotFound},
	} {
		testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "check_in_detail")
		testx.MustFlushDB(ctx, testStorage.Pool())

		err = checkInService.PrepareCheckIn(ctx, teacherID, course.ID, "第一次签到", c.expireSeconds)
		require.NoError(t, err)

		if c.expireSeconds == 1 {
			time.Sleep(time.Second)
		}
		err = checkInService.StartCheckIn(ctx, c.courseID, c.userID)
		assert.Equalf(t, c.expectedError, err, c.label)
	}
}

func TestUpdateCheckInDetail(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()

	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "check_in_detail")

	now := time.Now()
	const courseID = 1

	checkInRecord := &model.CheckInRecord{
		CourseID:  courseID,
		Name:      "第一次签到",
		CreatedAt: now,
		DeadLine:  now,
	}
	err := checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	const userID = 2
	checkInDetail := &model.CheckInDetail{
		RecordID:  checkInRecord.ID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = checkInDetail.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		recordID      uint64
	}{
		{label: "normal", userID: userID, recordID: 1, expectedError: nil},
		{label: "invalid record_id", userID: userID, recordID: 2, expectedError: errorx.ErrIsNotFound},
	} {
		err = checkInService.UpdateCheckInDetail(ctx, c.userID, c.recordID, true)
		assert.Equalf(t, c.expectedError, err, c.label)
	}
}

func TestDeleteCheckInDataByRecordID(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "course")

	now := time.Now()
	const teacherID = 1
	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)
	checkInRecord := &model.CheckInRecord{
		CourseID:  course.ID,
		Name:      "第一次签到",
		CreatedAt: now,
		DeadLine:  now,
	}
	err = checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	err = checkInService.DeleteCheckInDataByRecordID(ctx, checkInRecord.ID, teacherID)
	require.NoError(t, err)
}

func TestListRecordsByCourseID(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "check_in_detail")

	const (
		courseID = 1
		userID   = 1
	)

	now := time.Now()
	checkInRecord := &model.CheckInRecord{
		CourseID:  courseID,
		CreatedAt: now,
		DeadLine:  now,
	}
	err := checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	checkInDetail := &model.CheckInDetail{
		RecordID:  checkInRecord.ID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = checkInDetail.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	// 查询第一次签到记录
	for _, c := range []struct {
		expectedError error
		label         string
	}{
		{
			label:         "normal",
			expectedError: nil,
		},
		{
			label:         "canceled",
			expectedError: context.Canceled,
		},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
			resp, err := checkInService.ListRecordsByCourseID(ctx, courseID, 0, 5)
			require.Equal(t, c.expectedError, err, c.label)
			if err == nil {
				require.Len(t, resp.Records, 1)
				// 实际签到人数，此时应为0
				require.Equal(t, resp.Records.([]*CheckInData)[0].Actual, 0)
			}
		}
	}

	ctx = context.Background()
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	// 有人第一次签到
	checkInDetail.IsCheckIn = true
	err = checkInDetail.Update(ctx, testStorage.RDB)
	require.NoError(t, err)

	// 这次实际签到人数应为1
	resp, err := checkInService.ListRecordsByCourseID(ctx, courseID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 1)
	require.Equal(t, resp.Records.([]*CheckInData)[0].Actual, 1)

	// 新增一次签到，查询出来的签到数目应为2
	checkInRecord = &model.CheckInRecord{
		CourseID:  courseID,
		CreatedAt: now,
		DeadLine:  now,
	}
	err = checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	resp, err = checkInService.ListRecordsByCourseID(ctx, courseID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 2)
}

func TestListCheckInDetailsByRecordID(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "check_in_detail")

	const (
		courseID = 1
		userID   = 1
	)

	now := time.Now()
	checkInRecord := &model.CheckInRecord{
		CourseID:  courseID,
		CreatedAt: now,
		DeadLine:  now,
	}
	err := checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	checkInDetail := &model.CheckInDetail{
		RecordID:  checkInRecord.ID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = checkInDetail.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	go func() {
		cancel()
	}()
	_, err = checkInService.ListCheckInDetailsByRecordID(ctx, 1, 0, 5)
	require.Equal(t, context.Canceled, err)

	ctx = context.Background()
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	listCheckInDetailsByRecordID := func() []*CheckInWithUserData {
		resp, err := checkInService.ListCheckInDetailsByRecordID(ctx, 1, 0, 5)
		require.NoError(t, err)

		// 三个用户的签到数据
		require.Len(t, resp.Records, 1)

		return resp.Records.([]*CheckInWithUserData)
	}

	records := listCheckInDetailsByRecordID()

	for _, record := range records {
		require.False(t, record.IsCheckIn)
	}

	checkInDetail.IsCheckIn = true
	err = checkInDetail.Update(ctx, testStorage.RDB)
	require.NoError(t, err)

	records = listCheckInDetailsByRecordID()

	for _, record := range records {
		require.True(t, record.IsCheckIn)
	}
}

func TestListCheckInDetailsByUserIDAndCourseID(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "check_in_detail")

	const (
		courseID = 1
		userID   = 1
	)

	now := time.Now()
	checkInRecord := &model.CheckInRecord{
		CourseID:  courseID,
		CreatedAt: now,
		DeadLine:  now,
	}
	err := checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	checkInDetail := &model.CheckInDetail{
		RecordID:  checkInRecord.ID,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = checkInDetail.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError  error
		label          string
		userID         uint64
		expectedLength int
	}{
		{label: "normal", userID: 1, expectedLength: 1, expectedError: nil},
		{label: "not found", userID: 2, expectedLength: 0, expectedError: nil},
		{label: "canceled", userID: 1, expectedLength: 1, expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := checkInService.ListCheckInDetailsByUserIDAndCourseID(ctx, courseID, c.userID, 0, 5)
		assert.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Lenf(t, resp.Records, c.expectedLength, c.label)
		}
	}

}

func TestExportCheckInRecordsCSV(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "arrange_course", "course", "check_in_record", "check_in_detail")

	now := time.Now()
	teacher := &model.User{
		Number:    "1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := teacher.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	course := &model.Course{
		TeacherID: teacher.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	users := []*model.User{
		{Number: "2", CreatedAt: now, UpdatedAt: now},
		{Number: "3", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	var arrangeCourses []*model.ArrangeCourse
	for _, user := range users {
		arrangeCourses = append(arrangeCourses, &model.ArrangeCourse{
			UserID:    user.ID,
			CourseID:  course.ID,
			IsPass:    true,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	// 没有任何一次签到
	_, err = checkInService.ExportCheckInRecordsCSV(ctx, course.ID)
	require.NoError(t, err)

	// 进行一次签到
	err = checkInService.PrepareCheckIn(ctx, teacher.ID, course.ID, "第一次签到", 120)
	require.NoError(t, err)

	_, err = checkInService.ExportCheckInRecordsCSV(ctx, course.ID)
	require.NoError(t, err)

	// 签到一次
	err = checkInService.StartCheckIn(ctx, course.ID, 2)
	require.NoError(t, err)

	_, err = checkInService.ExportCheckInRecordsCSV(ctx, course.ID)
	require.NoError(t, err)
}

func TestListRecentUserCheckIn(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "check_in_detail", "course", "arrange_course")

	now := time.Now()
	course := &model.Course{CreatedAt: now, UpdatedAt: now}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	const userID = 1
	arrangeCourse := &model.ArrangeCourse{
		CourseID:  course.ID,
		UserID:    userID,
		IsPass:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = arrangeCourse.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	checkInRecords := []*model.CheckInRecord{
		{CourseID: course.ID, Name: "第一次签到", DeadLine: now.Add(-time.Minute), CreatedAt: now},
		{CourseID: course.ID, Name: "第二次签到", DeadLine: now.Add(time.Hour), CreatedAt: now},
	}

	err = model.BatchInsertCheckInRecords(ctx, testStorage.RDB, checkInRecords)
	require.NoError(t, err)

	checkInDetails := []*model.CheckInDetail{
		{RecordID: 1, UserID: userID, IsCheckIn: false, CreatedAt: now, UpdatedAt: now},
		{RecordID: 2, UserID: userID, IsCheckIn: false, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertCheckInDetails(ctx, testStorage.RDB, checkInDetails)
	require.NoError(t, err)

	resp, err := checkInService.ListRecentUserCheckIn(ctx, userID)
	require.NoError(t, err)
	require.Len(t, resp, 1)

	// 未到期，但完成了签到，理应也被过滤掉
	checkInDetail, err := model.QueryCheckInDetailByRecordIDAndUserID(ctx, testStorage.RDB, 2, userID)
	require.NoError(t, err)
	checkInDetail.IsCheckIn = true
	err = checkInDetail.Update(ctx, testStorage.RDB)
	require.NoError(t, err)

	resp, err = checkInService.ListRecentUserCheckIn(ctx, userID)
	require.NoError(t, err)
	require.Len(t, resp, 0)
}

func TestListUserCheckIn(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record", "check_in_detail", "arrange_course", "course")

	now := time.Now()

	course := &model.Course{CreatedAt: now, UpdatedAt: now}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	const userID = 1
	arrangeCourse := &model.ArrangeCourse{
		CourseID:  course.ID,
		UserID:    userID,
		IsPass:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = arrangeCourse.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	checkInRecords := []*model.CheckInRecord{
		{CourseID: course.ID, Name: "第一次签到", DeadLine: now.Add(-time.Minute), CreatedAt: now},
		{CourseID: course.ID, Name: "第二次签到", DeadLine: now.Add(time.Hour), CreatedAt: now},
	}

	err = model.BatchInsertCheckInRecords(ctx, testStorage.RDB, checkInRecords)
	require.NoError(t, err)

	checkInDetails := []*model.CheckInDetail{
		{RecordID: 1, UserID: userID, IsCheckIn: false, CreatedAt: now, UpdatedAt: now},
		{RecordID: 2, UserID: userID, IsCheckIn: true, CreatedAt: now, UpdatedAt: now},
	}
	err = model.BatchInsertCheckInDetails(ctx, testStorage.RDB, checkInDetails)
	require.NoError(t, err)

	resp, err := checkInService.ListUserCheckIn(ctx, userID, 0, 10)
	require.NoError(t, err)
	require.Len(t, resp.Records, 2)
	require.Equal(t, resp.PageInfo.Total, 2)

	records := resp.Records.([]*CheckInRecordPersonalData)
	for _, record := range records {
		switch record.ID {
		case 1:
			assert.False(t, *record.IsFinish)
		case 2:
			assert.True(t, *record.IsFinish)
		}
	}
}

func TestGetCourseIDByRecordID(t *testing.T) {
	testStorage, checkInService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "check_in_record")
	checkInRecord := &model.CheckInRecord{
		CreatedAt: now,
		DeadLine:  now,
	}
	err := checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		recordID      uint64
	}{
		{label: "normal", recordID: checkInRecord.ID, expectedError: nil},
		{label: "not found", recordID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		_, err := checkInService.GetCourseIDByRecordID(ctx, c.recordID)
		require.Equal(t, c.expectedError, err, c.label)
	}
}
