package handler

import (
	"net/http"
	"os"
	"time"

	xhttp "code-platform/api/http"
	"code-platform/api/http/handler/admin"
	"code-platform/api/http/handler/pprof"
	"code-platform/api/http/handler/web"
	"code-platform/api/http/md"
	"code-platform/log"

	"github.com/gin-gonic/gin"
)

var srv *xhttp.UnionService

func Init() {
	// init service
	srv = xhttp.NewUnionService()

	gin.DisableBindValidation()
	engine := gin.New()

	var logFile *os.File
	if logFilePath := os.Getenv("LOG_PATH"); logFilePath == "" {
		engine.Use(gin.Logger(), gin.Recovery())
	} else {
		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Error(err, "open log file failed")
			return
		}
		defer logFile.Close()
		engine.Use(gin.LoggerWithWriter(logFile), gin.RecoveryWithWriter(logFile))
	}
	// pprof 监控
	pprof.MakeMonitorHandler(engine)
	// 跨域中间件
	engine.Use(md.CORS())
	// 注册路由
	makeAllHandler(engine)

	server := &http.Server{
		Addr:        ":8081",
		Handler:     engine,
		IdleTimeout: 120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

}

func makeAllHandler(router gin.IRouter) {
	web.MakeWebHandler(router.Group("/web"), srv)
	admin.MakeAdminHandler(router.Group("/admin"), srv)
}
