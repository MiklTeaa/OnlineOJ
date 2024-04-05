package admin

import (
	"net/http"

	"code-platform/config"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeGetGrafanaURL(c *gin.Context) {
	host := config.Theia.GetString("dockerHost")
	url := "http://" + host + ":3000" + "/dashboards"
	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url})))
}
