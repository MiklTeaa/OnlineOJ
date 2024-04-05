package web

import (
	"net/http"
	"strings"

	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeExecCode(c *gin.Context) {
	type execCodeRequest struct {
		Code     string `json:"code"`
		Language int8   `json:"language"`
	}

	var req execCodeRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get data from exec code request")
		return
	}

	if req.Language < 0 || req.Language > 2 {
		httpx.AbortBadParamsErr(c, "labId is invalid")
		return
	}

	if strings.TrimSpace(req.Code) == "" {
		httpx.AbortBadParamsErr(c, "code is empty")
		return
	}

	type execCodeResponse struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      uint8  `json:"status"`
	}

	var response *execCodeResponse

	ctx := c.Request.Context()
	dockerResp, err := srv.MonacoService.ExecCode(ctx, req.Language, req.Code)
	switch err {
	case nil:
		response = &execCodeResponse{
			Status:      0,
			Title:       "执行成功",
			Description: dockerResp,
		}
	case errorx.ErrWrongCode:
		response = &execCodeResponse{
			Status:      1,
			Title:       "执行出错",
			Description: dockerResp,
		}
	case errorx.ErrContextCancel:
		response = &execCodeResponse{
			Status: 2,
			Title:  "超出时间限制",
		}
	case errorx.ErrOOMKilled:
		response = &execCodeResponse{
			Status: 3,
			Title:  "超出内存限制",
		}
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(response))
}
