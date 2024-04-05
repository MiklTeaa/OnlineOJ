package web

import (
	"net/http"

	"code-platform/api/http/md"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeListCodingTimeByUserID(c *gin.Context) {
	userID := c.GetUint64(md.KeyUserID)

	ctx := c.Request.Context()
	resp, err := srv.UserService.ListUserCodingTime(ctx, userID)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{
		"coding_time": resp,
	})))
}

func makeListCodingTimeByCourseID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		teacherID := c.GetUint64(md.KeyUserID)
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		resp, err := srv.CourseService.ListUserCodingTimesByCourseID(ctx, teacherID, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}
