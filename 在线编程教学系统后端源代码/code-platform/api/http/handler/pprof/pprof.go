package pprof

import (
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

const token = "yY5B8XGBk9vCPffkrOFAx0HPKTxNRGn1iT4f"

func requireToken(c *gin.Context) {
	if c.Query("token") != token {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
}

func MakeMonitorHandler(engine *gin.Engine) {
	router := engine.Group("/pprof")
	router.Use(requireToken)
	pprof.RouteRegister(router, "")
}
