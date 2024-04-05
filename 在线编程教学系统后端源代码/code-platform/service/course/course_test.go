package course_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"testing"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/pkg/timex"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	. "code-platform/service/course"
	"code-platform/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHelper() (*storage.Storage, *CourseService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	courseService := NewCourseService(dao, log.Sub("course"))
	return testStorage, courseService
}

func TestListCourseInfosByTeacherID(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user")

	now := time.Now()
	teacher := &model.User{
		Role:      1,
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

	for _, c := range []struct {
		expectedError error
		label         string
		teacherID     uint64
	}{
		{label: "normal", teacherID: teacher.ID, expectedError: nil},
		{label: "not found", teacherID: 10, expectedError: errorx.ErrIsNotFound},
		{label: "canceled", teacherID: teacher.ID, expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := courseService.ListCourseInfosByTeacherID(ctx, c.teacherID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp.Records, 1)
		}
	}
}

func TestListCourseInfosByStudentID(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user", "arrange_course")

	now := time.Now()
	users := []*model.User{
		// 教师
		{Role: 1, Number: "1", CreatedAt: now, UpdatedAt: now},
		// 学生
		{Role: 0, Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	err := model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	courses := []*model.Course{
		{TeacherID: 1, CreatedAt: now, UpdatedAt: now},
		{TeacherID: 1, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertCourses(ctx, testStorage.RDB, courses)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{
			UserID:    2,
			CourseID:  1,
			IsPass:    true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			UserID:    2,
			CourseID:  2,
			IsPass:    false,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		studentID     uint64
	}{
		{label: "normal", studentID: 2, expectedError: nil},
		{label: "canceled", studentID: 2, expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := courseService.ListCourseInfosByStudentID(ctx, c.studentID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp.Records, 1, c.label)
		}
	}
}

func TestEnrollForStudent(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, c := range []struct {
		expectedError error
		label         string
		number        string
		name          string
		trueKey       sql.NullString
		secretKey     sql.NullString
		courseID      uint64
		userID        uint64
		isDuplicate   bool
	}{
		{
			label:         "normal 1",
			number:        "1",
			name:          "1",
			trueKey:       sql.NullString{String: "are you ok", Valid: true},
			secretKey:     sql.NullString{String: "are you ok", Valid: true},
			courseID:      1,
			userID:        1,
			expectedError: nil,
		},
		{
			label:         "normal 2",
			number:        "1",
			name:          "1",
			trueKey:       sql.NullString{Valid: false},
			secretKey:     sql.NullString{Valid: false},
			courseID:      1,
			userID:        1,
			expectedError: nil,
		},
		{
			label:         "user not found",
			number:        "1",
			name:          "1",
			trueKey:       sql.NullString{String: "are you ok", Valid: true},
			secretKey:     sql.NullString{String: "are you ok", Valid: true},
			courseID:      1,
			userID:        2,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "course not found",
			number:        "1",
			name:          "1",
			trueKey:       sql.NullString{String: "are you ok", Valid: true},
			secretKey:     sql.NullString{String: "are you ok", Valid: true},
			courseID:      2,
			userID:        1,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "auth failed",
			number:        "1",
			name:          "1",
			trueKey:       sql.NullString{String: "are you ok", Valid: true},
			secretKey:     sql.NullString{String: "are you ko", Valid: true},
			courseID:      1,
			userID:        1,
			expectedError: errorx.ErrFailToAuth,
		},
		{
			label:         "number not completed",
			name:          "1",
			trueKey:       sql.NullString{String: "are you ok", Valid: true},
			secretKey:     sql.NullString{String: "are you ok", Valid: true},
			courseID:      1,
			userID:        1,
			expectedError: errorx.ErrPersonalInfoNotComplete,
		},
		{
			label:         "name not completed",
			number:        "1",
			trueKey:       sql.NullString{String: "are you ok", Valid: true},
			secretKey:     sql.NullString{String: "are you ok", Valid: true},
			courseID:      1,
			userID:        1,
			expectedError: errorx.ErrPersonalInfoNotComplete,
		},
		{
			label:       "duplicate insert",
			number:      "1",
			name:        "1",
			trueKey:     sql.NullString{String: "are you ok", Valid: true},
			secretKey:   sql.NullString{String: "are you ok", Valid: true},
			courseID:    1,
			userID:      1,
			isDuplicate: true,
		},
		{
			label:         "canceled",
			number:        "1",
			name:          "1",
			trueKey:       sql.NullString{String: "are you ok", Valid: true},
			secretKey:     sql.NullString{String: "are you ok", Valid: true},
			courseID:      1,
			userID:        1,
			expectedError: context.Canceled,
		},
	} {
		testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user", "arrange_course")
		course := &model.Course{
			SecretKey: c.trueKey,
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := course.Insert(ctx, testStorage.RDB)
		require.NoError(t, err)

		user := &model.User{
			Name:      c.name,
			Number:    c.number,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = user.Insert(ctx, testStorage.RDB)
		require.NoError(t, err)

		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		err = courseService.EnrollForStudent(ctx, c.userID, c.courseID, c.secretKey)
		assert.Equalf(t, c.expectedError, err, c.label)
		if c.isDuplicate {
			err = courseService.EnrollForStudent(ctx, c.userID, c.courseID, c.secretKey)
			assert.Equalf(t, errorx.ErrMySQLDuplicateKey, err, c.label)
		}
	}
}

func TestListCourseInfosByCourseName(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user")
	teacher := &model.User{
		Role:      1,
		Name:      "教师",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := teacher.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	courses := []*model.Course{
		{TeacherID: teacher.ID, Name: "第一个课堂", Description: "描述", CreatedAt: now, UpdatedAt: now},
		{TeacherID: teacher.ID, Name: "第二个课堂", Description: "树苗", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertCourses(ctx, testStorage.RDB, courses)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError  error
		keyword        string
		expectedLength int
	}{
		{keyword: "一", expectedLength: 0, expectedError: nil},
		{keyword: "苗", expectedLength: 0, expectedError: nil},
		{keyword: "描述", expectedLength: 1, expectedError: nil},
		{keyword: "堂客", expectedLength: 0, expectedError: nil},
		{keyword: "课堂", expectedLength: 2, expectedError: nil},
		{keyword: "课堂", expectedLength: 2, expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := courseService.ListCourseInfosByCourseName(ctx, c.keyword, 0, 5)
		require.Equal(t, c.expectedError, err)
		if err != nil {
			continue
		}
		assert.Len(t, resp.Records, c.expectedLength, c.keyword)
		records := resp.Records.([]*CourseWithTeacherInfo)
		for _, record := range records {
			assert.Equal(t, "教师", record.TeacherName)
			assert.Equal(t, teacher.ID, record.TeacherID)
		}
	}
}

func TestListStudentsCheckedByCourseID(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user", "arrange_course")

	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{UserID: 2, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: false},
		// 另一条选课记录，以作干扰
		{UserID: 2, CourseID: 10, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		courseID      uint64
	}{
		{label: "normal", courseID: course.ID, expectedError: nil},
		{label: "canceled", courseID: course.ID, expectedError: context.Canceled},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := courseService.ListStudentsCheckedByCourseID(ctx, course.ID, 0, 5)
		require.Equal(t, c.expectedError, err)
		if err == nil {
			require.Len(t, resp.Records, 1)
		}
	}

}

func TestListCodingTimesByCourseID(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "coding_time", "user", "arrange_course", "lab")

	teacher := &model.User{
		Number:    "0",
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := teacher.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	course := &model.Course{
		TeacherID: 1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	lab := &model.Lab{
		CourseID:  course.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}
	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{UserID: 2, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}
	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	go func() {
		cancel()
	}()
	_, err = courseService.ListUserCodingTimesByCourseID(ctx, teacher.ID, course.ID, 0, 5)
	assert.Equal(t, context.Canceled, err)

	ctx = context.Background()
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	resp, err := courseService.ListUserCodingTimesByCourseID(ctx, 1, course.ID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 2)

	userInfos := resp.Records.([]*UserInfoWithCodingTime)
	// 此时编程时间数据应不存在
	for _, userInfo := range userInfos {
		require.Len(t, userInfo.CodingTimeInfos, 0)
	}

	todayMoring := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	todayNoon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
	yesterDay := now.Add(-time.Hour * 24)
	tomorrow := now.Add(time.Hour * 24)

	codingTimes := []*model.CodingTime{
		{LabID: lab.ID, UserID: 1, CreatedAt: yesterDay, CreatedAtDate: timex.StartOfDay(yesterDay), Duration: 18},
		{LabID: lab.ID, UserID: 1, CreatedAt: now, CreatedAtDate: timex.StartOfDay(now), Duration: 12},
		{LabID: lab.ID, UserID: 1, CreatedAt: todayMoring, CreatedAtDate: timex.StartOfDay(todayMoring), Duration: 10},
		{LabID: lab.ID, UserID: 2, CreatedAt: todayMoring, CreatedAtDate: timex.StartOfDay(todayMoring), Duration: 18},
		{LabID: lab.ID, UserID: 2, CreatedAt: todayNoon, CreatedAtDate: timex.StartOfDay(todayNoon), Duration: 12},
		{LabID: lab.ID, UserID: 2, CreatedAt: tomorrow, CreatedAtDate: timex.StartOfDay(tomorrow), Duration: 10},
	}

	err = model.BatchInsertCodingTimes(ctx, testStorage.RDB, codingTimes)
	require.NoError(t, err)

	resp, err = courseService.ListUserCodingTimesByCourseID(ctx, teacher.ID, course.ID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 2)

	yesterdayDate := yesterDay.Format("2006-01-02")
	todayDate := now.Format("2006-01-02")
	tomorrowDate := tomorrow.Format("2006-01-02")

	userInfos = resp.Records.([]*UserInfoWithCodingTime)
	for _, userInfo := range userInfos {
		switch userInfo.UserID {
		case 1:
			require.Len(t, userInfo.CodingTimeInfos, 2)
			for _, codingTimeInfo := range userInfo.CodingTimeInfos {
				switch codingTimeInfo.Date {
				case yesterdayDate:
					require.Equal(t, 18, codingTimeInfo.Duration)
				case todayDate:
					require.Equal(t, 22, codingTimeInfo.Duration)
				}
			}
		case 2:
			require.Len(t, userInfo.CodingTimeInfos, 2)
			for _, codingTimeInfo := range userInfo.CodingTimeInfos {
				switch codingTimeInfo.Date {
				case todayDate:
					require.Equal(t, 30, codingTimeInfo.Duration)
				case tomorrowDate:
					require.Equal(t, 10, codingTimeInfo.Duration)
				}
			}
		}
	}
}

func TestExportScoreCSV(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course", "user", "lab", "lab_submit")

	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
		{Number: "3", CreatedAt: now, UpdatedAt: now},
	}
	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{UserID: 2, CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
		{UserID: 3, CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
	}
	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	labs := []*model.Lab{
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	labSubmits := []*model.LabSubmit{
		{LabID: 1, UserID: 1, Score: sql.NullInt32{Valid: true, Int32: 50}, CreatedAt: now, UpdatedAt: now},
		{LabID: 2, UserID: 1, Score: sql.NullInt32{Valid: true, Int32: 68}, CreatedAt: now, UpdatedAt: now},
		{LabID: 1, UserID: 2, Score: sql.NullInt32{Valid: true, Int32: 92}, CreatedAt: now, UpdatedAt: now},
		{LabID: 2, UserID: 2, Score: sql.NullInt32{Valid: false}, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabSubmits(ctx, testStorage.RDB, labSubmits)
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
		resp, err := courseService.ExportScoreCSVByCourseID(ctx, course.ID)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			log.Info(string(resp))
		}
	}

}

func TestListAllCourses(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user")

	resp, err := courseService.ListAllCourses(ctx, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 0)

	teacher := &model.User{
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = teacher.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	courses := []*model.Course{
		{TeacherID: teacher.ID, CreatedAt: now, UpdatedAt: now},
		{TeacherID: teacher.ID, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertCourses(ctx, testStorage.RDB, courses)
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
		resp, err = courseService.ListAllCourses(ctx, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp.Records, len(courses))
		}
	}
}

func TestGetCourseInfoByUserIDAndCourseID(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user", "arrange_course")

	teacher := &model.User{
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

	// 用户ID插入后应为2，3
	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 2, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: false},
		{UserID: 3, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError    error
		label            string
		userID           uint64
		courseID         uint64
		expectedIsEnroll bool
	}{
		{
			label:            "normal is not enroll",
			userID:           2,
			courseID:         course.ID,
			expectedError:    nil,
			expectedIsEnroll: false,
		},
		{
			label:            "normal is enroll",
			userID:           3,
			courseID:         course.ID,
			expectedError:    nil,
			expectedIsEnroll: true,
		},
		{
			label:         "course not found",
			userID:        2,
			courseID:      10,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "canceled",
			userID:        2,
			courseID:      course.ID,
			expectedError: context.Canceled,
		},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := courseService.GetCourseInfoByUserIDAndCourseID(ctx, c.userID, c.courseID)
		assert.Equalf(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Equal(t, c.expectedIsEnroll, resp.IsEnroll, c.label)
		}
	}
}

func TestExportCSVTemplate(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	_, err := courseService.ExportCSVTemplate()
	require.NoError(t, err)
}

func TestImportCSVTemplate(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course", "user")

	buf := bytes.NewBuffer(nil)
	csvWriter := csv.NewWriter(buf)

	rows := [][]string{
		{"number", "realName", "college", "grade", "organization", "major"},
		{"1", "李红", "计算机学院", "2018", "4班", "计算机科学与技术"},
		{"2", "李明", "政治与行政学院", "2018", "3班", "政治与行政"},
		{"3", "李华", "心理学院", "2019", "4班", "心理学"},
		{"4", "张三", "光电子学院", "2020", "5班", "光电子"},
	}
	err := csvWriter.WriteAll(rows)
	require.NoError(t, err)
	csvWriter.Flush()

	now := time.Now()
	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
		{Number: "3", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{UserID: 2, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: false},
		// 省去3号，让其在表中被导入
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		f             func() error
		label         string
		filename      string
	}{
		{
			label:         "UnsupportFileType",
			filename:      "1.txt",
			f:             func() error { return nil },
			expectedError: errorx.ErrUnsupportFileType,
		},
		{
			label:         "normal",
			filename:      "1.csv",
			f:             func() error { return nil },
			expectedError: nil,
		},
		{
			label:    "PersonalInfoInvalid",
			filename: "1.csv",
			f: func() error {
				// 补充姓名不完整的情形
				buf.Reset()

				rows = [][]string{
					{"number", "realName", "college", "grade", "organization", "major"},
					{"5", "", "光电子学院", "2018", "4班", "光电子"},
				}
				err := csvWriter.WriteAll(rows)
				if err != nil {
					return err
				}
				csvWriter.Flush()
				return nil
			},
			expectedError: errorx.ErrPersonalInfoInvalid,
		},
	} {
		err := c.f()
		require.NoError(t, err)
		err = courseService.ImportCSVTemplate(ctx, c.filename, buf.Bytes(), course.ID)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestListStudentWaitForChecked(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course", "user")

	now := time.Now()
	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
		{Number: "3", CreatedAt: now, UpdatedAt: now},
		{Number: "4", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: course.ID, IsPass: false, CreatedAt: now, UpdatedAt: now},
		{UserID: 2, CourseID: course.ID, IsPass: false, CreatedAt: now, UpdatedAt: now},
		// 已经通过审核
		{UserID: 3, CourseID: course.ID, IsPass: true, CreatedAt: now, UpdatedAt: now},
		// 选了别的课程
		{UserID: 4, CourseID: 10, IsPass: true, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
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
		resp, err := courseService.ListStudentWaitForChecked(ctx, course.ID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp.Records, 2)
		}
	}

}

func TestCheckForStudents(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course", "lab_submit", "lab", "check_in_record", "check_in_detail")

	now := time.Now()
	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: course.ID, IsPass: false, CreatedAt: now, UpdatedAt: now},
		{UserID: 2, CourseID: course.ID, IsPass: false, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	checkInRecord := &model.CheckInRecord{
		CourseID:  course.ID,
		CreatedAt: now,
		DeadLine:  now,
	}
	err = checkInRecord.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	lab := &model.Lab{
		CourseID:  course.ID,
		CreatedAt: now,
		UpdatedAt: now,
		DeadLine:  sql.NullTime{Valid: true, Time: now},
	}

	err = lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	err = courseService.CheckForStudents(ctx, course.ID, []uint64{1}, true)
	require.NoError(t, err)

	err = courseService.CheckForStudents(ctx, course.ID, []uint64{2}, false)
	require.NoError(t, err)
}

func TestListCoursesScore(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "arrange_course", "user", "lab", "lab_submit", "check_in_record", "check_in_detail")

	now := time.Now()

	// 课程相关数据
	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	// 用户相关数据
	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	// 实验相关数据
	labs := []*model.Lab{
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	labSubmits := []*model.LabSubmit{
		{LabID: 1, UserID: 1, Score: sql.NullInt32{Valid: false}, CreatedAt: now, UpdatedAt: now},
		{LabID: 2, UserID: 1, Score: sql.NullInt32{Valid: true, Int32: 68}, CreatedAt: now, UpdatedAt: now},
		{LabID: 1, UserID: 2, Score: sql.NullInt32{Valid: true, Int32: 92}, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabSubmits(ctx, testStorage.RDB, labSubmits)
	require.NoError(t, err)

	// 签到相关数据
	checkInRecords := []*model.CheckInRecord{
		{CourseID: course.ID, CreatedAt: now, DeadLine: now},
		{CourseID: course.ID, CreatedAt: now, DeadLine: now},
	}

	err = model.BatchInsertCheckInRecords(ctx, testStorage.RDB, checkInRecords)
	require.NoError(t, err)

	checkInDetails := []*model.CheckInDetail{
		{RecordID: 1, UserID: 1, IsCheckIn: true, CreatedAt: now, UpdatedAt: now},
		{RecordID: 2, UserID: 1, IsCheckIn: true, CreatedAt: now, UpdatedAt: now},
		{RecordID: 1, UserID: 2, IsCheckIn: true, CreatedAt: now, UpdatedAt: now},
		// 用户2还未为第二次签到贡献签到记录
		{RecordID: 2, UserID: 2, IsCheckIn: false, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertCheckInDetails(ctx, testStorage.RDB, checkInDetails)
	require.NoError(t, err)

	// 选课表相关数据
	arrangeCourses := []*model.ArrangeCourse{
		{CourseID: course.ID, UserID: 1, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{CourseID: course.ID, UserID: 2, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	resp, err := courseService.ListCoursesScore(ctx, course.ID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 2)
	userInfos := resp.Records.([]*UserWithAverageScoreAndCheckInData)
	for _, userInfo := range userInfos {
		switch userInfo.ID {
		case 1:
			require.Equal(t, 68.0, userInfo.AvgScore)
			require.Equal(t, 2, userInfo.ShallCheckIn)
			require.Equal(t, 2, userInfo.ActualCheckIn)
		case 2:
			require.Equal(t, 92.0, userInfo.AvgScore)
			require.Equal(t, 2, userInfo.ShallCheckIn)
			require.Equal(t, 1, userInfo.ActualCheckIn)
		}
	}

	// 模拟补交实验后的情形
	labSubmit := &model.LabSubmit{
		LabID:     2,
		UserID:    2,
		Score:     sql.NullInt32{Valid: true, Int32: 93},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = labSubmit.Insert(ctx, testStorage.RDB)
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
		resp, err = courseService.ListCoursesScore(ctx, course.ID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err != nil {
			continue
		}
		require.Len(t, resp.Records, 2)
		userInfos = resp.Records.([]*UserWithAverageScoreAndCheckInData)
		for _, userInfo := range userInfos {
			switch userInfo.ID {
			case 2:
				require.Equal(t, 92.5, userInfo.AvgScore)
			}
		}
	}

}

func TestAddCourse(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "user")

	now := time.Now()

	teacher := &model.User{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := teacher.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	err = courseService.AddCourse(ctx, teacher.ID, "第一个课堂", "没有描述", false, "", "", 0, false)
	require.NoError(t, err)
}

func TestUpdateCourse(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course")

	now := time.Now()

	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

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
		err = courseService.UpdateCourse(ctx, c.courseID, "", "", "", "", 0, false)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestDeleteCourse(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course")
	const teacherID = 1

	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	err = courseService.DeleteCourse(ctx, course.ID, teacherID)
	require.NoError(t, err)

}

func TestQuitCourse(t *testing.T) {
	testStorage, courseService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	ctx := context.Background()

	testx.MustTruncateTable(ctx, testStorage.RDB, "arrange_course")

	arrangeCourses := []*model.ArrangeCourse{
		{CourseID: 1, UserID: 1, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{CourseID: 1, UserID: 2, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}

	err := model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	err = courseService.QuitCourse(ctx, 1, 1)
	require.NoError(t, err)
}
