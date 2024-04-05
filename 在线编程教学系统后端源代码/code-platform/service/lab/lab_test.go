package lab_test

import (
	"bytes"
	idepb "code-platform/api/grpc/ide/pb"
	"code-platform/api/grpc/plagiarismDetection/pb"
	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/service/ide"
	. "code-platform/service/lab"
	"code-platform/service/monaco"
	"code-platform/storage"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func testHelper() (*storage.Storage, *LabService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	labService := NewLabService(dao, log.Sub("lab"), NewPlagiarismDetectionClient(), ide.NewIDEClient(), monaco.NewMonacoClient())
	return testStorage, labService
}

func TestInsertLab(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	const courseID = 1

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "arrange_course", "lab", "lab_submit")
	now := time.Now()

	arrangeCourses := []*model.ArrangeCourse{
		{CourseID: courseID, UserID: 1, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{CourseID: courseID, UserID: 2, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}

	err := model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	err = labService.InsertLab(ctx, courseID, "", "", "", now.Add(time.Hour))
	require.NoError(t, err)
}

func TestListLabsByUserIDAndCourseID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "lab", "lab_submit")
	now := time.Now()

	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	labs := []*model.Lab{
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	labSubmits := []*model.LabSubmit{
		{LabID: 1, UserID: 1, CreatedAt: now, UpdatedAt: now, Score: sql.NullInt32{Valid: true, Int32: 92}},
		{LabID: 2, UserID: 1, CreatedAt: now, UpdatedAt: now, Score: sql.NullInt32{Valid: true, Int32: 90}},
		{LabID: 2, UserID: 2, CreatedAt: now, UpdatedAt: now, Score: sql.NullInt32{Valid: true, Int32: 85}},
	}

	err = model.BatchInsertLabSubmits(ctx, testStorage.RDB, labSubmits)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		courseID      uint64
		labAmount     int
	}{
		{
			label:         "normal 1",
			userID:        1,
			courseID:      1,
			labAmount:     2,
			expectedError: nil,
		},
		{
			label:         "normal 2",
			userID:        2,
			courseID:      1,
			labAmount:     1,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        1,
			courseID:      2,
			labAmount:     0,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "canceled",
			userID:        2,
			courseID:      1,
			labAmount:     1,
			expectedError: context.Canceled,
		},
	} {

		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := labService.ListLabsByUserIDAndCourseID(ctx, c.userID, c.courseID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Lenf(t, resp.Records, c.labAmount, c.label)
		}
	}
}

func TestUpdateLab(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "lab")
	now := time.Now()
	const courseID = 1

	lab := &model.Lab{CourseID: courseID, CreatedAt: now, UpdatedAt: now}
	err := lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		labID         uint64
	}{
		{label: "normal", labID: lab.ID, expectedError: nil},
		{label: "not found", labID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		err := labService.UpdateLab(ctx, c.labID, "", "", "", time.Time{})
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestGetLab(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "lab")
	now := time.Now()
	const courseID = 1

	lab := &model.Lab{CourseID: courseID, CreatedAt: now, UpdatedAt: now}
	err := lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		labID         uint64
	}{
		{
			label:         "normal",
			labID:         lab.ID,
			expectedError: nil,
		},
		{
			label:         "not found",
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		_, err := labService.GetLab(ctx, c.labID)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestDeleteLab(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "lab", "lab_submit")
	now := time.Now()
	const courseID = 1

	lab := &model.Lab{CourseID: courseID, CreatedAt: now, UpdatedAt: now}
	err := lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	labSubmits := []*model.LabSubmit{
		{LabID: lab.ID, UserID: 1, CreatedAt: now, UpdatedAt: now},
		{LabID: lab.ID, UserID: 2, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabSubmits(ctx, testStorage.RDB, labSubmits)
	require.NoError(t, err)

	err = labService.DeleteLab(ctx, lab.ID)
	require.NoError(t, err)
}

func TestListLabsByUserID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "lab", "lab_submit", "course", "arrange_course")
	now := time.Now()

	const userID = 1

	course := &model.Course{CreatedAt: now, UpdatedAt: now}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	arrangeCourse := &model.ArrangeCourse{CourseID: course.ID, UserID: userID, CreatedAt: now, UpdatedAt: now, IsPass: true}

	err = arrangeCourse.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	labs := []*model.Lab{
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
		{CourseID: course.ID, CreatedAt: now, UpdatedAt: now},
		{CourseID: 10, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	labSubmits := []*model.LabSubmit{
		{LabID: 1, UserID: userID, CreatedAt: now, UpdatedAt: now},
		{LabID: 1, UserID: 10, CreatedAt: now, UpdatedAt: now},
		{LabID: 2, UserID: userID, CreatedAt: now, UpdatedAt: now},
		{LabID: 3, UserID: userID, CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertLabSubmits(ctx, testStorage.RDB, labSubmits)
	require.NoError(t, err)

	resp, err := labService.ListLabsByUserID(ctx, userID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 2)
}

func TestListLabScoreByUserIDAndCourseID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	testx.MustTruncateTable(ctx, testStorage.RDB, "lab", "lab_submit")

	now := time.Now()
	const (
		courseID = 1
		userID   = 1
	)

	labs := []*model.Lab{
		{CourseID: courseID, CreatedAt: now, UpdatedAt: now},
		{CourseID: courseID, CreatedAt: now, UpdatedAt: now},
	}

	err := model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	labSubmits := []*model.LabSubmit{
		{LabID: 1, UserID: userID, CreatedAt: now, UpdatedAt: now},
		{LabID: 2, UserID: 10, CreatedAt: now, UpdatedAt: now},
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
		resp, err := labService.ListLabScoreByUserIDAndCourseID(ctx, userID, courseID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp.Records, 1)
		}
	}
}

func TestUpdateReport(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "lab", "lab_submit")
	now := time.Now()

	const (
		userID = 1
		labID  = 1
	)
	labSubmit := &model.LabSubmit{LabID: labID, UserID: userID, CreatedAt: now, UpdatedAt: now}
	err := labSubmit.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		labID         uint64
	}{
		{label: "normal", labID: labID, expectedError: nil},
		{label: "not found", labID: 10, expectedError: errorx.ErrIsNotFound},
	} {
		err := labService.UpdateReport(ctx, c.labID, userID, "")
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestInsertCodeFinish(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()

	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "lab_submit")

	labSubmits := []*model.LabSubmit{
		{UserID: 1, LabID: 1, IsFinish: true, CreatedAt: now, UpdatedAt: now},
		{UserID: 1, LabID: 2, IsFinish: false, CreatedAt: now, UpdatedAt: now},
	}

	err := model.BatchInsertLabSubmits(ctx, testStorage.RDB, labSubmits)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		labID         uint64
	}{
		{
			label:         "normal",
			userID:        1,
			labID:         1,
			expectedError: nil,
		},
		{
			label:         "normal with update",
			userID:        1,
			labID:         2,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        2,
			labID:         2,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		err := labService.InsertCodeFinish(ctx, c.labID, c.userID, true)
		assert.Equal(t, c.expectedError, err, c.labID)
	}

}

func TestGetCommentByUserIDAndLabID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "lab_submit")

	const (
		userID  = 1
		labID   = 1
		content = "hello"
	)

	labSubmit := model.LabSubmit{
		UserID:    userID,
		LabID:     labID,
		Comment:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := labSubmit.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		labID         uint64
	}{
		{
			label:         "normal",
			userID:        userID,
			labID:         labID,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        userID,
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		resp, err := labService.GetCommentByUserIDAndLabID(ctx, c.userID, c.labID)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Equal(t, content, resp)
		}
	}
}

func TestGetReportURL(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "lab_submit")

	const (
		userID  = 1
		labID   = 1
		content = "hello"
	)

	labSubmit := model.LabSubmit{
		UserID:    userID,
		LabID:     labID,
		ReportURL: content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := labSubmit.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		labID         uint64
	}{
		{
			label:         "normal",
			userID:        userID,
			labID:         labID,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        userID,
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		resp, err := labService.GetReportURL(ctx, c.userID, c.labID)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Equal(t, content, resp)
		}
	}
}

func TestUpdateScore(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "lab_submit")

	const (
		userID = 1
		labID  = 1
	)

	labSubmit := model.LabSubmit{
		UserID:    userID,
		LabID:     labID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := labSubmit.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		labID         uint64
	}{
		{
			label:         "normal",
			userID:        userID,
			labID:         labID,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        userID,
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		err := labService.UpdateScore(ctx, c.userID, c.labID, 100)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestUpdateComment(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "lab_submit")

	const (
		userID = 1
		labID  = 1
	)

	labSubmit := model.LabSubmit{
		UserID:    userID,
		LabID:     labID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := labSubmit.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
		labID         uint64
	}{
		{
			label:         "normal",
			userID:        userID,
			labID:         labID,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        userID,
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		err := labService.UpdateComment(ctx, c.userID, c.labID, "h")
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestListLabsByCourseID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "lab")

	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	labs := []*model.Lab{
		{
			CourseID:  course.ID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			CourseID:  course.ID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err = model.BatchInsertLabs(ctx, testStorage.RDB, labs)
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
		resp, err := labService.ListLabsByCourseID(ctx, course.ID, 0, 10)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Equal(t, 2, resp.PageInfo.Total)
			require.Len(t, resp.Records, 2)
		}
	}
}

func TestListLabSubmitsByLabID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "lab", "lab_submit", "user")

	const courseID = 1

	labs := []*model.Lab{
		{CourseID: courseID, CreatedAt: now, UpdatedAt: now},
		{CourseID: courseID, CreatedAt: now, UpdatedAt: now},
	}

	err := model.BatchInsertLabs(ctx, testStorage.RDB, labs)
	require.NoError(t, err)

	labSubmits := []*model.LabSubmit{
		{UserID: 1, LabID: 1, CreatedAt: now, UpdatedAt: now},
		{UserID: 2, LabID: 1, CreatedAt: now, UpdatedAt: now},
		{UserID: 1, LabID: 2, CreatedAt: now, UpdatedAt: now},
	}

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	err = model.BatchInsertLabSubmits(ctx, testStorage.RDB, labSubmits)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError  error
		label          string
		labID          uint64
		expectedLength int
	}{
		{
			label:          "normal 1",
			labID:          1,
			expectedError:  nil,
			expectedLength: 2,
		},
		{
			label:          "normal 2",
			labID:          2,
			expectedError:  nil,
			expectedLength: 1,
		},
		{
			label:          "not found",
			labID:          3,
			expectedError:  errorx.ErrIsNotFound,
			expectedLength: 2,
		},
		{
			label:          "canceled",
			labID:          2,
			expectedError:  context.Canceled,
			expectedLength: 1,
		},
	} {
		if c.expectedError == context.Canceled {
			go func() {
				cancel()
			}()
		}
		resp, err := labService.ListLabSubmitsByLabID(ctx, c.labID, 0, 5)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Lenf(t, resp.Records, c.expectedLength, c.label)
		}
	}
}

func TestQuickCheckCode(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	_, err := labService.IDEClient.GenerateTestFileForViewCode(ctx, &idepb.Empty{})
	require.NoError(t, err)

	resp, err := labService.QuickCheckCode(ctx, 0, 0)
	require.NoError(t, err)
	log.Infof("%+v", resp)

	_, err = labService.IDEClient.RemoveGenerateTestFileForViewCode(ctx, &idepb.Empty{})
	require.NoError(t, err)
}

func TestPlagiarismCheck(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "lab", "arrange_course", "course", "detection_report")
	const teacherID = 0
	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
		Language:  0,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	lab := &model.Lab{ID: 0, CourseID: course.ID, CreatedAt: now, UpdatedAt: now}
	err = lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)
	err = lab.UpdateIDToZero(ctx, testStorage.RDB)
	require.NoError(t, err)

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}

	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	fakeLab := &model.Lab{
		CourseID:  10,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = fakeLab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: 1, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{UserID: 2, CourseID: 1, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}

	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	fakePythonBuf := bytes.NewBufferString(`
print("hello world")
	`)
	pythonBuf := bytes.NewBufferString(`
	print("hello world")
	print("world hello")
`)
	cppBuf := bytes.NewBufferString(`
#include <iostream>
using namespace std;
int main(){
	cout<<"hello world"<<endl;
}
	`)
	javaBuf := bytes.NewBufferString(`public class Solution{
		public static void main(String ...args) {
			System.out.println("hello world");
		}
	}`)

	for _, c := range []struct {
		expectedError error
		buf           *bytes.Buffer
		f             func() error
		label         string
		labID         uint64
		language      pb.Language
	}{

		{
			label:         "fake python",
			language:      pb.Language_python3,
			buf:           fakePythonBuf,
			expectedError: errorx.ErrWrongCode,
		},
		{
			label:         "normal python",
			language:      pb.Language_python3,
			buf:           pythonBuf,
			expectedError: nil,
		},
		{
			label:         "normal cpp",
			language:      pb.Language_cpp,
			buf:           cppBuf,
			expectedError: nil,
			f: func() error {
				course.Language = 1
				return course.Update(ctx, testStorage.RDB)
			},
		},

		{
			label:    "normal java",
			language: pb.Language_java,
			buf:      javaBuf,
			f: func() error {
				course.Language = 2
				return course.Update(ctx, testStorage.RDB)
			},
			expectedError: nil,
		},
		{
			label:         "not found",
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "course is not found",
			labID:         fakeLab.ID,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		if c.f != nil {
			err := c.f()
			require.NoError(t, err)
		}
		if c.buf != nil {
			for j := 1; j <= 2; j++ {
				_, err := labService.PlagiarismDetectionClient.GenerateTestFilesForDuplicateCheck(ctx, &pb.GenerateTestFilesForDuplicateCheckRequest{
					CodeContent: c.buf.String(),
					Lan:         c.language,
				})
				require.NoError(t, err)
			}
		}
		_, err := labService.PlagiarismCheck(ctx, c.labID)
		require.Equal(t, c.expectedError, err, c.label)

		if err != nil {
			continue
		}
		_, err = labService.PlagiarismDetectionClient.RemoveTestFilesForDuplicateCheck(ctx, &pb.Empty{})
		require.NoError(t, err)
	}
}

func TestClickURL(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()

	dirName := "12345"
	fileName := "match.html"

	_, err := labService.PlagiarismDetectionClient.GenerateTestHTMLFileForViewReport(ctx, &pb.GenerateTestHTMLFileForViewReportRequest{
		TimeStamp:    dirName,
		HtmlFileName: fileName,
	})
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		labID         uint64
	}{
		{
			label:         "normal",
			labID:         0,
			expectedError: nil,
		},
		{
			label:         "not found",
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		_, err := labService.ClickURL(ctx, c.labID, dirName, fileName)
		assert.Equal(t, c.expectedError, err, c.label)
	}

	_, err = labService.PlagiarismDetectionClient.RemoveTestHTMLFileForViewReport(ctx, &pb.Empty{})
	require.NoError(t, err)
}

func TestGetCourseIDByLabID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "lab")

	now := time.Now()
	const courseID = 1
	lab := &model.Lab{
		CourseID:  courseID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := lab.Insert(ctx, labService.Dao.Storage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		labID         uint64
	}{
		{
			label:         "normal",
			labID:         lab.ID,
			expectedError: nil,
		}, {
			label:         "not found",
			labID:         10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		_, err = labService.GetCourseIDByLabID(ctx, c.labID)
		require.Equal(t, c.expectedError, err, c.label)
	}
}
func TestListDetectionReportsByLabID(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "detection_report")

	const labID = 1
	now := time.Now()
	detectionReports := []*model.DetectionReport{
		{LabID: labID, Data: []byte(""), CreatedAt: now},
		{LabID: labID, Data: []byte(""), CreatedAt: now.Add(time.Second)},
		{LabID: 10, Data: []byte(""), CreatedAt: now},
	}
	err := model.BatchInsertDetectionReports(ctx, testStorage.RDB, detectionReports)
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
		resp, err := labService.ListDetectionReportsByLabID(ctx, labID, 0, 10)
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Equal(t, 2, resp.PageInfo.Total, c.label)
		}
	}
}

func TestViewPerviousDetection(t *testing.T) {
	testStorage, labService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "detection_report", "user", "arrange_course", "lab", "course")
	now := time.Now()
	const teacherID = 1
	course := &model.Course{
		TeacherID: teacherID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)
	lab := &model.Lab{CourseID: course.ID, CreatedAt: now, UpdatedAt: now}
	err = lab.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	arrangeCourses := []*model.ArrangeCourse{
		{UserID: 1, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
		{UserID: 2, CourseID: course.ID, CreatedAt: now, UpdatedAt: now, IsPass: true},
	}
	err = model.BatchInsertArrangeCourses(ctx, testStorage.RDB, arrangeCourses)
	require.NoError(t, err)

	var comparision = pb.DuplicateCheckResponse_DuplicateCheckResponseValue{
		Comparisions: []*pb.DuplicateCheckResponse_DuplicateCheckResponseValue_Comparsion{{
			UserId:        1,
			AnotherUserId: 2,
		}},
	}
	data, err := proto.Marshal(&comparision)
	require.NoError(t, err)
	detectionReport := &model.DetectionReport{LabID: lab.ID, Data: data, CreatedAt: now}
	err = detectionReport.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError     error
		label             string
		detectionReportID uint64
		teacherID         uint64
	}{
		{label: "normal", detectionReportID: detectionReport.ID, teacherID: teacherID, expectedError: nil},
		{label: "not found", detectionReportID: 10, teacherID: teacherID, expectedError: errorx.ErrIsNotFound},
		{label: "no auth", detectionReportID: detectionReport.ID, teacherID: 10, expectedError: errorx.ErrFailToAuth},
	} {
		resp, err := labService.ViewPerviousDetection(ctx, c.detectionReportID, c.teacherID, "")
		require.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			require.Len(t, resp, 1, c.label)
		}
	}
}
