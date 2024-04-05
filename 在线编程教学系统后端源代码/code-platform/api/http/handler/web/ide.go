package web

import (
	"fmt"
	"net/http"

	"code-platform/api/http/md"
	"code-platform/config"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeOpenIDE(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthLabForStudent(ctx, c, srv, labID, teacherID) {
			return
		}

		port, token, err := srv.IDEService.OpenIDE(ctx, labID, teacherID)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "userID is invalid")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		host := config.Theia.GetString("dockerHost")
		c.SetCookie("token", token, 0, "/", "", false, true)
		url := fmt.Sprintf("http://%s:%d", host, port)
		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url, "token": token})))
	}
}

func makeCheckCode(c *gin.Context) {
	type checkCodeRequest struct {
		StudentID uint64 `json:"stuID"`
		LabID     uint64 `json:"labID"`
	}

	var req checkCodeRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in getCheckCode request")
		return
	}

	if req.LabID <= 0 || req.StudentID <= 0 {
		httpx.AbortBadParamsErr(c, "id is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	port, token, err := srv.IDEService.CheckCode(ctx, req.LabID, req.StudentID, teacherID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "fail to find ide by params")
		return
	case errorx.ErrFailToAuth:
		httpx.AbortForbidden(c)
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	host := config.Theia.GetString("dockerHost")
	c.SetCookie("token", token, 0, "/", "", false, true)
	url := fmt.Sprintf("http://%s:%d", host, port)

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url, "token": token})))
}

func makeHeartBeatForStudent(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		labID := c.GetUint64(tag)
		studentID := c.GetUint64(md.KeyUserID)
		ctx := c.Request.Context()
		if !md.AuthLabForStudent(ctx, c, srv, labID, studentID) {
			return
		}

		switch err := srv.IDEService.HeartBeatForStudent(ctx, labID, studentID); err {
		case nil:
		case errorx.ErrRedisKeyNil:
			httpx.AbortNotFound(c, "ide is closed")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}
		c.Status(http.StatusOK)
	}
}

func makeHeartBeatForTeacher(c *gin.Context) {
	type heartBeatForTeacherRequest struct {
		LabID     uint64 `json:"labid"`
		StudentID uint64 `json:"stuid"`
	}

	var req heartBeatForTeacherRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Failed to get heart beat request")
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

	if err := srv.IDEService.HeartBeatForTeacher(ctx, req.LabID, req.StudentID, teacherID); err != nil {
		httpx.AbortInternalErr(c)
		return
	}
	c.Status(http.StatusOK)
}
