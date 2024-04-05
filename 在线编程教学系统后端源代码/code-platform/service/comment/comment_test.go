package comment_test

import (
	"context"
	"testing"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	. "code-platform/service/comment"
	"code-platform/storage"

	"github.com/stretchr/testify/require"
)

func testHelper() (*storage.Storage, *CommentService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	commentService := NewCommentService(dao, log.Sub("comment"))
	return testStorage, commentService
}

func TestListCourseCommentsByCourseID(t *testing.T) {
	testStorage, commentService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course_comment", "course", "user")

	now := time.Now()
	user := &model.User{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	go func() {
		cancel()
	}()
	_, err = commentService.ListCourseCommentsByCourseID(ctx, course.ID, 0, 5)
	require.Equal(t, context.Canceled, err)

	ctx = context.Background()
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	resp, err := commentService.ListCourseCommentsByCourseID(ctx, course.ID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 0)

	courseComment := &model.CourseComment{
		UserID:      1,
		CourseID:    course.ID,
		CommentText: "父评论",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = courseComment.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	resp, err = commentService.ListCourseCommentsByCourseID(ctx, course.ID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 1)

	// 3条子评论
	subCourseComments := make([]*model.CourseComment, 3)
	for index := range subCourseComments {
		subCourseComments[index] = &model.CourseComment{
			UserID:      1,
			CourseID:    course.ID,
			PID:         courseComment.ID,
			ReplyUserID: 1,
			CommentText: "子评论",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	err = model.BatchInsertCourseComments(ctx, testStorage.RDB, subCourseComments)
	require.NoError(t, err)

	resp, err = commentService.ListCourseCommentsByCourseID(ctx, course.ID, 0, 5)
	require.NoError(t, err)
	require.Len(t, resp.Records, 1)

	records := resp.Records.([]*CourseCommentEntity)
	require.Len(t, records[0].SubComments, 3)
}

func TestInsertCourseComment(t *testing.T) {
	testStorage, commentService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course", "course_comment")

	now := time.Now()
	course := &model.Course{
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := course.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	err = commentService.InsertCourseComment(ctx, "第一条父评论", course.ID, 0, 1, 0)
	require.NoError(t, err)

	err = commentService.InsertCourseComment(ctx, "第一条子评论", course.ID, 1, 2, 1)
	require.NoError(t, err)

	err = commentService.InsertCourseComment(ctx, "第二条子评论", course.ID, 1, 3, 2)
	require.NoError(t, err)

	err = commentService.InsertCourseComment(ctx, "错误的子评论", course.ID, 1, 2, 10)
	require.Equal(t, errorx.ErrIsNotFound, err)
}

func TestDeleteCourseComment(t *testing.T) {
	testStorage, commentService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()

	testx.MustTruncateTable(ctx, testStorage.RDB, "course_comment")

	now := time.Now()
	// 该评论是用户1产生
	courseComment := &model.CourseComment{
		UserID:      1,
		CommentText: "第一条父评论",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := courseComment.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	subCourseComment := &model.CourseComment{
		UserID:      2,
		PID:         1,
		ReplyUserID: 1,
		CommentText: "第一条子评论",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err = subCourseComment.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	err = commentService.DeleteCourseComment(ctx, 0, 1)
	require.Equal(t, errorx.ErrIsNotFound, err)

	// 用户2尝试删除
	err = commentService.DeleteCourseComment(ctx, 1, 2)
	require.Equal(t, errorx.ErrFailToAuth, err)

	err = commentService.DeleteCourseComment(ctx, 1, 1)
	require.NoError(t, err)
}
