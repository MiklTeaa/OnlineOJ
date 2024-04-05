package admin

import (
	"time"

	xhttp "code-platform/api/http"
	"code-platform/api/http/md"

	"github.com/gin-gonic/gin"
)

var srv *xhttp.UnionService

func MakeAdminHandler(router gin.IRouter, s *xhttp.UnionService) {
	// init service
	srv = s

	router.Use(md.Timeout(15 * time.Second))
	router.Use(md.RestoreUserStat(srv), md.RequireAdmin(srv))

	router.POST("/monitor", md.Tracer("admin.makeGetGrafanaURL"), makeGetGrafanaURL)

	routerUser := router.Group("/user")
	{
		routerUser.POST("/import", md.Tracer("admin.user.makeImportStudentByCSV"), md.CheckFileHeader("csv"), makeImportStudentByCSV)
		routerUser.GET("/export/template", md.Tracer("admin.user.makeExportCSVTemplate"), makeExportCSVTemplate)
		routerUser.GET("", md.CheckPage, md.Tracer("admin.user.makeListAllUsers"), makeListAllUsers)

		routerUser.POST("/distribute", md.Tracer("admin.user.makeDistributeAccount"), makeDistributeAccount)

		routerUser.POST("/amend", md.Tracer("admin.user.makeAmendUser"), makeAmendUser)
		routerUser.POST("/password", md.Tracer("admin.user.makeResetPassword"), makeResetPassword)
	}

	routerCourse := router.Group("/course")
	{
		routerCourse.GET("", md.Tracer("admin.course.makeListAllCourses"), md.CheckPage, makeListAllCourses)
		routerCourse.POST("/amend", md.Tracer("admin.course.makeAmendCourse"), makeAmendCourse)
	}

	routerIDE := router.Group("/ide")
	{
		routerIDE.GET("", md.Tracer("admin.ide.makeListContainers"), md.CheckPage, makeListContainers)
		routerIDE.POST("/quit", md.Tracer("admin.ide.makeStopContainer"), makeStopContainer)
	}
}
