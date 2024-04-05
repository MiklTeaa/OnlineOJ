package comment

import (
	"context"
	"database/sql"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/transactionx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/storage"
)

type CommentService struct {
	Dao    *repository.Dao
	Logger *log.Logger
}

func NewCommentService(dao *repository.Dao, logger *log.Logger) *CommentService {
	return &CommentService{
		Dao:    dao,
		Logger: logger,
	}
}

func (c *CommentService) ListCourseCommentsByCourseID(ctx context.Context, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total          int
		comments       []*model.CourseCommentWithUserInfo
		subCommentsMap map[uint64][]*model.CourseCommentWithUserInfo
		IDToUserMap    map[uint64]*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInCourseParentCommentByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountInCourseParentCommentByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountInCourseParentCommentByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			comments, err = model.QueryCourseParentCommentWithUserInfosByCourseID(ctx, c.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCourseParentCommentWithUserInfosByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "queryCourseParentComment by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}

			parentCommentIDs := make([]uint64, len(comments))
			for index, comment := range comments {
				parentCommentIDs[index] = comment.ID
			}

			var replyUserIDs []uint64
			subCommentsMap, replyUserIDs, err = model.QuerySubCourseCommentsMapByParentCommentIDs(ctx, c.Dao.Storage.RDB, parentCommentIDs)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QuerySubCourseCommentsMapByParentCommentIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "query sub comments by parentCommentIDs %v failed", parentCommentIDs)
				return errorx.InternalErr(err)
			}

			IDToUserMap, err = model.QueryUserMapByIDs(ctx, c.Dao.Storage.RDB, replyUserIDs)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUserMapByIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "query users by userIDs %v failed", replyUserIDs)
				return errorx.InternalErr(err)
			}

			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	records := make([]*CourseCommentEntity, len(comments))
	for index, comment := range comments {
		subComments := subCommentsMap[comment.ID]
		subCommentWithUserInfos := make([]*SubCourseCommentResp, len(subComments))
		for index, subComment := range subComments {
			var replyUserName string
			if replyUser := IDToUserMap[subComment.ReplyUserID]; replyUser != nil {
				replyUserName = replyUser.Name
			}
			subCommentWithUserInfos[index] = &SubCourseCommentResp{
				CourseCommentResp: &CourseCommentResp{
					CourseCommentID: subComment.ID,
					CourseID:        subComment.CourseID,
					CommentText:     subComment.CommentText,
					Pid:             subComment.PID,
					UserID:          subComment.UserID,
					UserAvatar:      subComment.Avatar,
					Username:        subComment.Name,
					ReplyUserID:     subComment.ReplyUserID,
					CreatedAt:       subComment.CreatedAt,
					UpdatedAt:       subComment.UpdatedAt,
				},
				ReplyUserName: replyUserName,
			}
		}
		records[index] = &CourseCommentEntity{
			Comment: &CourseCommentResp{
				CourseCommentID: comment.ID,
				CourseID:        comment.CourseID,
				CommentText:     comment.CommentText,
				Pid:             comment.PID,
				UserID:          comment.UserID,
				UserAvatar:      comment.Avatar,
				Username:        comment.Name,
				ReplyUserID:     comment.ReplyUserID,
				CreatedAt:       comment.CreatedAt,
				UpdatedAt:       comment.UpdatedAt,
			},
			SubComments: subCommentWithUserInfos,
		}
	}

	return &PageResponse{
		Records:  records,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CommentService) InsertCourseComment(ctx context.Context, content string, courseID, pid, userID, replyCommentID uint64) error {
	now := time.Now()
	courseComment := &model.CourseComment{
		UserID:      userID,
		CourseID:    courseID,
		PID:         pid,
		CommentText: content,
		ReplyUserID: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 非主评论
	if replyCommentID != 0 {
		anotherComment, err := model.QueryCourseCommentByID(ctx, c.Dao.Storage.RDB, replyCommentID)
		switch err {
		case nil:
		case sql.ErrNoRows:
			c.Logger.Debugf("courseComment is not found by id(%d)", replyCommentID)
			return errorx.ErrIsNotFound
		default:
			c.Logger.Errorf(err, "queryCourseComment by commentID(%d) failed", replyCommentID)
			return errorx.InternalErr(err)
		}

		courseComment.ReplyUserID = anotherComment.UserID
		if anotherComment.IsParent() {
			courseComment.PID = replyCommentID
		} else {
			courseComment.PID = anotherComment.PID
		}
	}

	if err := courseComment.Insert(ctx, c.Dao.Storage.RDB); err != nil {
		c.Logger.Errorf(err, "insert courseComment %+v failed", courseComment)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CommentService) DeleteCourseComment(ctx context.Context, commentID, userID uint64) (err error) {
	comment, err := model.QueryCourseCommentByID(ctx, c.Dao.Storage.RDB, commentID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("courseComment is not found by commentID(%d)", commentID)
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "QueryCourseCommentByID by commentID(%d) failed", commentID)
		return errorx.InternalErr(err)
	}

	// 尝试删除非自己产生的评论
	if comment.UserID != userID {
		c.Logger.Debugf("fail to auth when try to delete courseComment for userID(%d) by userID(%d)", comment.UserID, userID)
		return errorx.ErrFailToAuth
	}

	task := func(ctx context.Context, tx storage.RDBClient) error {
		if err := model.DeleteCourseCommentByID(ctx, tx, commentID); err != nil {
			c.Logger.Errorf(err, "delete courseComment by ID(%d) failed", commentID)
			return errorx.InternalErr(err)
		}

		// 删除其所有子评论
		if comment.IsParent() {
			if err := model.DeleteSubCourseCommentsByParentCommentID(ctx, tx, comment.ID); err != nil {
				c.Logger.Errorf(err, "delete sub courseComments by parentCommentID(%d)", comment.PID)
				return errorx.InternalErr(err)
			}
		}
		return nil
	}

	return transactionx.DoTransaction(ctx, c.Dao.Storage, c.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}
