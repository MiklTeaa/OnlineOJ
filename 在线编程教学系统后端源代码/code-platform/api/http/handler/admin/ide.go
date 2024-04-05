package admin

import (
	"net/http"
	"strings"

	idepb "code-platform/api/grpc/ide/pb"
	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeListContainers(c *gin.Context) {
	type listContainersRequest struct {
		Order     idepb.OrderType `form:"order"`
		IsReverse bool            `form:"isReverse"`
	}

	var req listContainersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get request")
		return
	}

	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)
	ctx := c.Request.Context()
	resp, err := srv.IDEService.ListContainers(ctx, (pageCurrent-1)*pageSize, pageSize, req.Order, req.IsReverse)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}
	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeStopContainer(c *gin.Context) {
	type stopContainerRequest struct {
		ContainerID string `json:"containerID"`
	}

	var req stopContainerRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Get ContainerID failed")
		return
	}

	if req.ContainerID == "" || strings.Contains(req.ContainerID, " ") {
		httpx.AbortBadParamsErr(c, "containerID is invalid")
		return
	}

	ctx := c.Request.Context()
	err := srv.IDEService.StopContainer(ctx, req.ContainerID)
	switch err {
	case nil:
	case errorx.ErrWrongCode:
		httpx.AbortBadParamsErr(c, "container is not found by containerID")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}
