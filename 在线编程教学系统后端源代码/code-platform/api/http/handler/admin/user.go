package admin

import (
	"net/http"
	"strings"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeImportStudentByCSV(c *gin.Context) {
	fileHeader := md.GetFileHeader(c, "csv")

	data, err := srv.FileService.MIMEHeaderToBytes(fileHeader)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	ctx := c.Request.Context()
	switch err := srv.UserService.ImportStudentByCSV(ctx, fileHeader.Filename, data); err {
	case nil:
	case errorx.ErrUnsupportFileType:
		httpx.AbortBadParamsErr(c, "File type is not csv")
		return
	case errorx.ErrPersonalInfoInvalid:
		httpx.AbortBadParamsErr(c, "Some student data in file is invalid")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeExportCSVTemplate(c *gin.Context) {

	resp, err := srv.UserService.ExportCSVTemplate()
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Header("content-type", "application/csv")
	c.Header("content-disposition", "attachment;filename=用户导入表模板.csv")
	c.Writer.Write(resp)
	c.Writer.Flush()
	c.Status(http.StatusOK)
}

func makeListAllUsers(c *gin.Context) {

	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.UserService.ListAllUsers(ctx, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeDistributeAccount(c *gin.Context) {
	type distributeAccountRequest struct {
		Number string `json:"number"`
		Name   string `json:"name"`
		Role   uint16 `json:"role"`
	}

	var req distributeAccountRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get param in distribute in account")
		return
	}

	if req.Role != 0 && req.Role != 1 {
		httpx.AbortBadParamsErr(c, "role is invalid")
		return
	}

	if strings.TrimSpace(req.Number) == "" {
		httpx.AbortBadParamsErr(c, "number can't be empty")
		return
	}

	ctx := c.Request.Context()
	resp, err := srv.UserService.DistributeAccount(ctx, req.Number, req.Name, req.Role)
	switch err {
	case nil:
	case errorx.ErrMySQLDuplicateKey:
		httpx.AbortBadParamsErr(c, "number is duplicate")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeAmendUser(c *gin.Context) {
	type amendUserRequest struct {
		Major        string `json:"major"`
		Number       string `json:"number"`
		Name         string `json:"name"`
		College      string `json:"college"`
		Organization string `json:"organization"`
		Avatar       string `json:"avatar_url"`
		UserID       uint64 `json:"user_id"`
		Grade        uint16 `json:"grade"`
		Gender       int8   `json:"gender"`
	}

	var req amendUserRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in amendUserRequest")
		return
	}

	if req.UserID <= 0 {
		httpx.AbortBadParamsErr(c, "user_id is invalid")
		return
	}

	if req.Gender != 1 && req.Gender != 2 {
		req.Gender = 0
	}

	if strings.TrimSpace(req.Number) == "" {
		httpx.AbortBadParamsErr(c, "number can't be empty")
		return
	}

	ctx := c.Request.Context()
	err := srv.UserService.UpdateUserByAdmin(ctx, req.UserID, req.Number, req.Name, req.College, req.Major, req.Organization, req.Avatar, req.Grade, req.Gender)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "user is not found by user_id")
		return
	case errorx.ErrMySQLDuplicateKey:
		httpx.AbortBadParamsErr(c, "number is duplicate")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeResetPassword(c *gin.Context) {
	type resetPasswordRequest struct {
		UserID   uint64 `json:"user_id"`
		Password string `json:"password"`
	}

	var req resetPasswordRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in amendUserRequest")
		return
	}

	if req.UserID <= 0 {
		httpx.AbortBadParamsErr(c, "user_id is invalid")
		return
	}

	if strings.TrimSpace(req.Password) == "" {
		httpx.AbortBadParamsErr(c, "password shouldn't be empty")
		return
	}

	ctx := c.Request.Context()
	switch err := srv.UserService.ResetPassordByAdmin(ctx, req.Password, req.UserID); err {
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
