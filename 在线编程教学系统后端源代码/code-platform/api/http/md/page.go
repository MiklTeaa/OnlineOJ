package md

import (
	"code-platform/pkg/httpx"

	"github.com/gin-gonic/gin"
)

type pageRequest struct {
	PageCurrent int `form:"pageCurrent"`
	PageSize    int `form:"pageSize"`
}

func CheckPage(c *gin.Context) {
	var req pageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in page request")
		return
	}
	if req.PageSize <= 0 || req.PageCurrent <= 0 {
		httpx.AbortBadParamsErr(c, "page params are invalid")
		return
	}

	if req.PageSize > 200 {
		httpx.AbortBadParamsErr(c, "pageSize is too big")
		return
	}

	c.Set(KeyPageCurrent, req.PageCurrent)
	c.Set(KeyPageSize, req.PageSize)

}
