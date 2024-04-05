package web

import (
	"net/http"
	"strings"

	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"
	"code-platform/pkg/stringx"

	"github.com/gin-gonic/gin"
)

func makeLoginHandler(c *gin.Context) {
	type loginRequest struct {
		Number   string `json:"username"`
		Password string `json:"password"`
	}

	var req loginRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in login request")
		return
	}

	if strings.TrimSpace(req.Number) == "" || strings.TrimSpace(req.Password) == "" {
		httpx.AbortBadParamsErr(c, "params shouldn't be empty")
		return
	}

	if !stringx.IsLowerEqualThan(req.Number, 50) || !stringx.IsLowerEqualThan(req.Password, 200) {
		httpx.AbortInvalidLength(c, "number or password is too long")
		return
	}

	ctx := c.Request.Context()
	loginResp, err := srv.UserService.Login(ctx, req.Number, req.Password)
	switch err {
	case nil:
	case errorx.ErrIsNotFound, errorx.ErrFailToAuth:
		httpx.AbortFailToAuth(c, "number or password is incorrect")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(loginResp)))
}
