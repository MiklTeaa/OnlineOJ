package web

import (
	"net/http"
	"strings"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"
	"code-platform/pkg/stringx"

	"github.com/gin-gonic/gin"
)

func makeCheckNumber(c *gin.Context) {
	number, ok := c.GetQuery("number")
	if !ok {
		httpx.AbortGetParamsErr(c, "Fail to get number")
		return
	}

	if strings.TrimSpace(number) == "" {
		httpx.AbortBadParamsErr(c, "Number shouldn't not be empty")
		return
	}

	exists := false

	ctx := c.Request.Context()
	switch err := srv.UserService.CheckNumber(ctx, number); err {
	case errorx.ErrIsNotFound:
	case nil:
		exists = true
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(!exists)))
}

func makeSendVerificationCode(c *gin.Context) {
	type sendVerificationCodeRequest struct {
		Number string `json:"number"`
	}

	var req sendVerificationCodeRequest
	err := c.ShouldBindWith(&req, jsonx.SonicDecoder)
	if err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get number")
		return
	}

	if strings.TrimSpace(req.Number) == "" {
		httpx.AbortBadParamsErr(c, "number shouldn't be empty")
		return
	}

	ctx := c.Request.Context()
	switch err := srv.UserService.SendVerificationCode(ctx, req.Number); err {
	case nil:
	case errorx.ErrMailUserNotFound:
		httpx.AbortMailUserNotFound(c, "email user is not found")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeResetPassword(c *gin.Context) {
	type resetPasswordRequest struct {
		Number           string `json:"number"`
		Password         string `json:"password"`
		VerificationCode string `json:"verificationCode"`
	}

	var req resetPasswordRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in reset password request")
		return
	}

	if strings.TrimSpace(req.Number) == "" ||
		strings.TrimSpace(req.Password) == "" ||
		strings.TrimSpace(req.VerificationCode) == "" {
		httpx.AbortBadParamsErr(c, "params shouldn't be empty")
		return
	}

	if !stringx.IsLowerEqualThan(req.Number, 20) || !stringx.IsLowerEqualThan(req.Password, 200) {
		httpx.AbortInvalidLength(c, "number or password is too long")
		return
	}

	ctx := c.Request.Context()
	err := srv.UserService.ResetPassword(ctx, req.Number, req.Password, req.VerificationCode)
	switch err {
	case nil:
	case errorx.ErrFailToAuth:
		httpx.AbortFailToAuth(c, "VerficationCode is invalid")
		return
	case errorx.ErrIsNotFound:
		httpx.AbortNotFound(c, "user_id is invalid")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeGetOuterUserByToken(c *gin.Context) {
	userID := c.GetUint64(md.KeyUserID)

	ctx := c.Request.Context()
	outerUser, err := srv.UserService.GetOuterUserByUserID(ctx, userID)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(outerUser)))
}

func makeUpdateUser(c *gin.Context) {
	type updateRequest struct {
		RealName     string `json:"real_name"`
		Major        string `json:"major"`
		Organization string `json:"organization"`
		College      string `json:"college"`
		Grade        uint16 `json:"grade"`
		Gender       int8   `json:"gender"`
	}

	var req updateRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update user request")
		return
	}

	if !stringx.IsLowerEqualThan(req.Major, 50) {
		httpx.AbortInvalidLength(c, "major is too long")
		return
	}

	if !stringx.IsLowerEqualThan(req.Organization, 200) {
		httpx.AbortInvalidLength(c, "organization is too long")
		return
	}

	if !stringx.IsLowerEqualThan(req.RealName, 20) {
		httpx.AbortInvalidLength(c, "name is too long")
		return
	}

	if !stringx.IsLowerEqualThan(req.College, 50) {
		httpx.AbortInvalidLength(c, "college is too long")
		return
	}

	if req.Gender != 1 && req.Gender != 2 {
		req.Gender = 0
	}

	userID := c.GetUint64(md.KeyUserID)

	ctx := c.Request.Context()
	switch err := srv.UserService.UpdateUser(ctx, userID, req.RealName, req.College, req.Major, req.Organization, req.Grade, req.Gender); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "user_id is invalid")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeUpdateUserAvatar(c *gin.Context) {
	type updateUserAvatarRequest struct {
		AvatarURL string `json:"avatar_url"`
	}

	var req updateUserAvatarRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get update user avatar request")
		return
	}

	if !stringx.IsLowerEqualThan(req.AvatarURL, 200) {
		httpx.AbortBadParamsErr(c, "avatar_url is too long")
		return
	}

	userID := c.GetUint64(md.KeyUserID)

	ctx := c.Request.Context()
	switch err := srv.UserService.UpdateUserAvatar(ctx, userID, req.AvatarURL); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "user_id is invalid")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeRefreshUserToken(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
		c.Header("WWW-Authenticate", "Bearer")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token := parts[1]

	type refreshTokenRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req refreshTokenRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortBadParamsErr(c, "Fail to get params in refreshTokenRequest")
		return
	}

	// 不是空串
	if strings.TrimSpace(req.RefreshToken) == "" {
		httpx.AbortBadParamsErr(c, "refresh_token is empty")
		return
	}

	ctx := c.Request.Context()
	tokenRefreshed, newRefreshToken, err := srv.UserService.RefreshToken(ctx, token, req.RefreshToken)
	switch err {
	case nil:
	case errorx.ErrIsNotFound, errorx.ErrFailToAuth:
		httpx.AbortUnauthorized(c)
		return
	case errorx.ErrNotExpire:
		httpx.AbortBadParamsErr(c, "session is not expire")
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{
		"token":         tokenRefreshed,
		"refresh_token": newRefreshToken,
	})))
}
