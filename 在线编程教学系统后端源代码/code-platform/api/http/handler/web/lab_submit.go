package web

import (
	"net/http"
	"net/url"
	"strings"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeUploadLabReport(c *gin.Context) {
	type updateReportRequest struct {
		ReportURL string `json:"reportURL"`
		LabID     uint64 `json:"labID"`
	}

	var req updateReportRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update report request")
		return
	}

	if req.LabID <= 0 {
		httpx.AbortBadParamsErr(c, "labID is invalid")
		return
	}

	studentID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthLabForStudent(ctx, c, srv, req.LabID, studentID) {
		return
	}

	switch err := srv.LabService.UpdateReport(ctx, req.LabID, studentID, req.ReportURL); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "record is not found by id")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeUpdateLabFinish(c *gin.Context) {
	type updateLabFinishRequest struct {
		LabID    uint64 `json:"labId"`
		IsFinish bool   `json:"isFinish"`
	}

	var req updateLabFinishRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update code finish")
		return
	}

	if req.LabID <= 0 {
		httpx.AbortBadParamsErr(c, "labID is invalid")
		return
	}

	studentID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthLabForStudent(ctx, c, srv, req.LabID, studentID) {
		return
	}

	switch err := srv.LabService.InsertCodeFinish(ctx, req.LabID, studentID, req.IsFinish); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "record is not found by ID")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeListLabSubmitsByID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthLabForTeacher(ctx, c, srv, labID, teacherID) {
			return
		}

		resp, err := srv.LabService.ListLabSubmitsByLabID(ctx, labID, (pageCurrent-1)*pageSize, pageSize)
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

func makeUpdateLabSubmitScore(c *gin.Context) {
	type updateLabSubmitScoreRequest struct {
		UserID uint64 `json:"userID"`
		LabID  uint64 `json:"labID"`
		Score  int32  `json:"score"`
	}

	var req updateLabSubmitScoreRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in updateLabSubmitScore request")
		return
	}

	if req.LabID <= 0 || req.UserID <= 0 || req.Score < 0 {
		httpx.AbortBadParamsErr(c, "params are invalid")
		return
	}

	if req.Score > 1<<31-1 {
		httpx.AbortBadParamsErr(c, "score is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthLabForTeacher(ctx, c, srv, req.LabID, teacherID) {
		return
	}

	switch err := srv.LabService.UpdateScore(ctx, req.UserID, req.LabID, req.Score); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "record is not found by ID")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeGetReportURL(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)
		userID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		url, err := srv.LabService.GetReportURL(ctx, labID, userID)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortGetParamsErr(c, "record is not found by ID")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url})))
	}
}

func makeUpdateLabComment(c *gin.Context) {
	type updateLabSubmitCommentRequest struct {
		Comment string `json:"comment"`
		UserID  uint64 `json:"userID"`
		LabID   uint64 `json:"labID"`
	}

	var req updateLabSubmitCommentRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update lab submit comment request")
		return
	}

	if req.LabID <= 0 || req.UserID <= 0 {
		httpx.AbortBadParamsErr(c, "ID is invalid")
		return
	}

	if strings.TrimSpace(req.Comment) == "" {
		httpx.AbortBadParamsErr(c, "comment shouldn't be empty")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthLabForTeacher(ctx, c, srv, req.LabID, teacherID) {
		return
	}

	switch err := srv.LabService.UpdateComment(ctx, req.UserID, req.LabID, req.Comment); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortGetParamsErr(c, "record is not found by id")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeGetCommentsByUserIDAndLabID(c *gin.Context) {
	type getCommentsByUserIDRequest struct {
		UserID uint64 `form:"stuID"`
		LabID  uint64 `form:"labID"`
	}

	var req getCommentsByUserIDRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in get comments request")
		return
	}

	if req.LabID <= 0 || req.UserID <= 0 {
		httpx.AbortBadParamsErr(c, "ID is invalid")
		return
	}

	ctx := c.Request.Context()
	comment, err := srv.LabService.GetCommentByUserIDAndLabID(ctx, req.UserID, req.LabID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortGetParamsErr(c, "record is not found by id")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"comment": comment})))
}

func makePlagiarismCheck(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthLabForTeacher(ctx, c, srv, labID, teacherID) {
			return
		}

		resp, err := srv.LabService.PlagiarismCheck(ctx, labID)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "code is not found by labId")
			return
		case errorx.ErrWrongCode:
			httpx.AbortBadParamsErr(c, "code is invalid")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		baseURL := "http://" + c.Request.Host + c.Request.URL.Path
		for _, v := range resp {
			v.URL = baseURL + "/" + v.URL
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeClickPlagiarismURL(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileName := c.Param("fileName")
		if strings.TrimSpace(fileName) == "" {
			httpx.AbortBadParamsErr(c, "fileName should not be empty string")
			return
		}

		dirName, ok := c.GetQuery("ts")
		if !ok {
			referer := c.Request.Referer()
			url, err := url.Parse(referer)
			if err != nil {
				httpx.AbortGetParamsErr(c, "url parse referrer failed because of %v", err)
				return
			}

			ts := url.Query().Get("ts")
			if ts == "" {
				httpx.AbortGetParamsErr(c, "cannot find ts param")
				return
			}

			currentURL := "http://" + c.Request.Host + c.Request.URL.Path + "?ts=" + ts
			c.Redirect(http.StatusMovedPermanently, currentURL)
			return
		}

		labID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthLabForTeacher(ctx, c, srv, labID, teacherID) {
			return
		}

		data, err := srv.LabService.ClickURL(ctx, labID, dirName, fileName)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			c.Status(http.StatusNotFound)
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Data(http.StatusOK, "Content-Type: text/html", data)
	}
}
