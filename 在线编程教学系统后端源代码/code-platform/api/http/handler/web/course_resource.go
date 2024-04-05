package web

import (
	"net/http"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeAddCourseResource(c *gin.Context) {
	type addCourseResourceRequest struct {
		Title         string `json:"title"`
		Content       string `json:"content"`
		AttachmentURL string `json:"attachmentURL"`
		CourseID      uint64 `json:"courseId"`
	}

	var req addCourseResourceRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in add course_resource request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortGetParamsErr(c, "courseId is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCourseForTeacher(ctx, c, srv, req.CourseID, teacherID) {
		return
	}

	if err := srv.CourseResourceService.InsertCourseResource(ctx, req.CourseID, req.Title, req.Content, req.AttachmentURL); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeUpdateResource(c *gin.Context) {
	type updateCourseResourceRequest struct {
		Title            string `json:"title"`
		Content          string `json:"content"`
		AttachmentURL    string `json:"attachmentURL"`
		CourseResourceID uint64 `json:"courseResourceID"`
	}

	var req updateCourseResourceRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update course_resource request")
		return
	}

	if req.CourseResourceID <= 0 {
		httpx.AbortGetParamsErr(c, "courseResourceID is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCourseResourceForTeacher(ctx, c, srv, req.CourseResourceID, teacherID) {
		return
	}

	if err := srv.CourseResourceService.UpdateCourseResource(ctx, req.CourseResourceID, req.Title, req.Content, req.AttachmentURL); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeListCourseResourcesByCourseID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		resp, err := srv.CourseResourceService.ListCourseResource(ctx, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeFindCourseResourceByID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseResourceID := c.GetUint64(tag)

		ctx := c.Request.Context()
		resp, err := srv.CourseResourceService.GetCourseResource(ctx, courseResourceID)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "courseResource is not found by id")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeDeleteCourseResource(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseResourceID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthCourseResourceForTeacher(ctx, c, srv, courseResourceID, teacherID) {
			return
		}

		switch err := srv.CourseResourceService.DeleteCourseResource(ctx, courseResourceID); err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "courseResource is not found by id")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Status(http.StatusOK)
	}
}
