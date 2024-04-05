package web

import (
	"net/http"
	"time"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeAddLab(c *gin.Context) {
	type addLabRequest struct {
		DeadLine      time.Time `json:"deadLine"`
		Title         string    `json:"title"`
		Content       string    `json:"content"`
		AttachmentURL string    `json:"attachmentURL"`
		CourseID      uint64    `json:"courseID"`
	}
	var req addLabRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in add lab request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "courseID is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCourseForTeacher(ctx, c, srv, req.CourseID, teacherID) {
		return
	}
	if err := srv.LabService.InsertLab(ctx, req.CourseID, req.Title, req.Content, req.AttachmentURL, req.DeadLine); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeUpdateLab(c *gin.Context) {
	type updateLabRequest struct {
		DeadLine      time.Time `json:"deadLine"`
		Title         string    `json:"title"`
		Content       string    `json:"content"`
		AttachmentURL string    `json:"attachmentURL"`
		LabID         uint64    `json:"labId"`
	}

	var req updateLabRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update lab request")
		return
	}

	if req.LabID <= 0 {
		httpx.AbortBadParamsErr(c, "labID is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthLabForTeacher(ctx, c, srv, req.LabID, teacherID) {
		return
	}
	if err := srv.LabService.UpdateLab(ctx, req.LabID, req.Title, req.Content, req.AttachmentURL, req.DeadLine); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeListLabsByUserIDAndCourseID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		userID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		resp, err := srv.LabService.ListLabsByUserIDAndCourseID(ctx, userID, courseID, (pageCurrent-1)*pageSize, pageSize)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "record is not found by courseID")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeListLabsByCourseID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		resp, err := srv.LabService.ListLabsByCourseID(ctx, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}
		c.Render(http.StatusOK, jsonx.NewSonicEncoder(resp))
	}
}

func makeGetLabByID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)

		ctx := c.Request.Context()
		resp, err := srv.LabService.GetLab(ctx, labID)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "record is not found by labID")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeDeleteLabByID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)

		ctx := c.Request.Context()
		teacherID := c.GetUint64(md.KeyUserID)
		if !md.AuthLabForTeacher(ctx, c, srv, labID, teacherID) {
			return
		}
		if err := srv.LabService.DeleteLab(ctx, labID); err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Status(http.StatusOK)
	}
}

func makeGetLabByStudentID(c *gin.Context) {
	userID := c.GetUint64(md.KeyUserID)
	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.LabService.ListLabsByUserID(ctx, userID, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeListLabsScoreByStudentIDAndCourseID(c *gin.Context) {
	type listLabScoreRequest struct {
		CourseID  uint64 `form:"courseID"`
		StudentID uint64 `form:"stuID"`
	}

	var req listLabScoreRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in list lab score request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "courseID is invalid")
		return
	}

	if req.StudentID == 0 {
		req.StudentID = c.GetUint64(md.KeyUserID)
	}

	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.LabService.ListLabScoreByUserIDAndCourseID(ctx, req.StudentID, req.CourseID, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeGetTreeNode(c *gin.Context) {
	type checkCodeRequest struct {
		StudentID uint64 `form:"stuID"`
		LabID     uint64 `form:"labID"`
	}

	var req checkCodeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in gettreeNode request")
		return
	}

	if req.LabID <= 0 || req.StudentID <= 0 {
		httpx.AbortBadParamsErr(c, "id is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthLabForTeacher(ctx, c, srv, req.LabID, teacherID) {
		return
	}

	resp, err := srv.LabService.QuickCheckCode(ctx, req.LabID, req.StudentID)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeListHistoryDetectionReports(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)
		ctx := c.Request.Context()
		if !md.AuthLabForTeacher(ctx, c, srv, labID, teacherID) {
			return
		}
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)
		resp, err := srv.LabService.ListDetectionReportsByLabID(ctx, labID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}
		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeGetDetectionReport(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reportID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		resp, err := srv.LabService.ViewPerviousDetection(ctx, reportID, teacherID, c.Request.Host)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "report_id is invalid")
			return
		case errorx.ErrFailToAuth:
			httpx.AbortForbidden(c)
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}
		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}
