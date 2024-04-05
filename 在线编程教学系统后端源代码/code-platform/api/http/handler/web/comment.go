package web

import (
	"net/http"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"
	"code-platform/pkg/stringx"

	"github.com/gin-gonic/gin"
)

func makeAddCourseComment(c *gin.Context) {
	type addCourseCommentRequest struct {
		Text           string `json:"CommentText"`
		CourseID       uint64 `json:"CourseId"`
		PID            uint64 `json:"Pid"`
		ReplyCommentID uint64 `json:"ReplyId"`
	}

	var req addCourseCommentRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in add course comment request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "id is invalid")
		return
	}

	if !stringx.IsLowerEqualThan(req.Text, 120) {
		httpx.AbortBadParamsErr(c, "comment text is too long")
		return
	}

	userID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCourseForStudentAndTeacher(ctx, c, srv, req.CourseID, userID) {
		return
	}

	err := srv.CourseService.AuthAddComment(ctx, userID, req.CourseID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "course id is invalid")
		return
	case errorx.ErrFailToAuth:
		httpx.AbortForbidden(c)
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	switch err := srv.CommentService.InsertCourseComment(ctx, req.Text, req.CourseID, req.PID, userID, req.ReplyCommentID); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortForbidden(c)
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeGetCourseComments(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		resp, err := srv.CommentService.ListCourseCommentsByCourseID(ctx, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeDeleteCourseComment(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID := c.GetUint64(tag)
		userID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		switch err := srv.CommentService.DeleteCourseComment(ctx, commentID, userID); err {
		case nil:
		case errorx.ErrFailToAuth:
			httpx.AbortUnauthorized(c)
			return
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "comment is not found")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Status(http.StatusOK)
	}
}
