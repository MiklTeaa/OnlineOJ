package web

import (
	"time"

	xhttp "code-platform/api/http"
	"code-platform/api/http/md"

	"github.com/gin-gonic/gin"
)

var srv *xhttp.UnionService

func MakeWebHandler(router gin.IRouter, s *xhttp.UnionService) {
	// init service
	srv = s

	// ping
	router.GET("/ping", md.Tracer("web.ping"), func(c *gin.Context) {
		c.String(200, "ok")
	})

	router.Use(md.Timeout(10 * time.Second))

	router.POST("/login", md.Tracer("web.makeLoginHandler"), makeLoginHandler)

	router.POST("/refresh", md.Tracer("web.makeRefreshUserToken"), makeRefreshUserToken)

	routerUser := router.Group("/user")
	{
		routerUser.GET("/number", md.Tracer("web.user.makeCheckNumber"), makeCheckNumber)
		routerUser.POST("/verificationCode", md.Tracer("web.user.makeSendVerificationCode"), makeSendVerificationCode)
		routerUser.PUT("/password", md.Tracer("web.user.makeResetPassword"), makeResetPassword)

		// 需要token
		routerUser.Use(md.RestoreUserStat(srv))

		routerUser.GET("", md.Tracer("web.user.makeGetOuterUserByToken"), makeGetOuterUserByToken)

		routerUser.PUT("", md.Tracer("web.user.makeUpdateUser"), makeUpdateUser)
		routerUser.PUT("/avatar", md.Tracer("web.user.makeUpdateUserAvatar"), makeUpdateUserAvatar)
	}

	// 需要token
	router.Use(md.RestoreUserStat(srv))

	routerComment := router.Group("/comment")
	{
		routerComment.POST("/course", md.Tracer("web.comment.makeAddCourseComment"), makeAddCourseComment)
		routerComment.GET("/course",
			md.Tracer("web.comment.makeGetCourseComments"), md.CheckPage, md.CheckQueryID("courseId"), md.AuthCourse(srv, "courseId"),
			makeGetCourseComments("courseId"),
		)
		routerComment.DELETE("/course", md.Tracer("web.comment.makeDeleteCourseComment"), md.CheckJSONID("commentId"), makeDeleteCourseComment("commentId"))
	}

	routerCodingTime := router.Group("/coding_time")
	{
		routerCodingTime.GET("", md.Tracer("web.coding_time.makeListCodingTimeByUserID"), md.RequireStudent(srv), makeListCodingTimeByUserID)
		routerCodingTime.GET("/:courseID",
			md.Tracer("web.coding_time.makeListCodingTimeByCourseID"), md.CheckPage, md.CheckParamID("courseID"), md.RequireTeacher(srv),
			makeListCodingTimeByCourseID("courseID"),
		)
	}

	routerCheckIn := router.Group("/checkin")
	{
		routerCheckIn.POST("/check", md.Tracer("web.checkin.makeStudentStartCheckIn"), md.RequireStudent(srv), makeStudentStartCheckIn)
		routerCheckIn.GET("/record/user",
			md.Tracer("web.checkin.makeListCheckInDetailsByUserIDAndCourseID"), md.CheckPage, md.CheckQueryID("courseId"), md.RequireStudent(srv),
			makeListCheckInDetailsByUserIDAndCourseID("courseId"),
		)

		routerCheckIn.POST("/start", md.Tracer("web.checkin.makeTeacherPrepareCheckIn"), md.RequireTeacher(srv), makeTeacherPrepareCheckIn)
		routerCheckIn.GET("/records",
			md.Tracer("web.checkin.makeListCheckInRecordsByCourseID"), md.CheckPage, md.CheckQueryID("courseId"), md.RequireTeacher(srv),
			makeListCheckInRecordsByCourseID("courseId"),
		)
		routerCheckIn.GET("/details",
			md.Tracer("web.checkin.makeListCheckInDetailsByRecordID"), md.CheckQueryID("checkInRecordId"), md.RequireTeacher(srv),
			makeListCheckInDetailsByRecordID("checkInRecordId"),
		)
		routerCheckIn.DELETE("/record", md.Tracer("web.checkin.makeDeleteCheckInData"), md.CheckQueryID("checkInRecordId"), md.RequireTeacher(srv), makeDeleteCheckInData("checkInRecordId"))
		routerCheckIn.PUT("/detail", md.Tracer("web.checkin.makeUpdateCheckInDetail"), md.RequireTeacher(srv), makeUpdateCheckInDetail)
		routerCheckIn.GET("/export/:courseID", md.Tracer("web.checkin.makeExportCheckInRecords"), md.CheckParamID("courseID"), md.RequireTeacher(srv), makeExportCheckInRecords("courseID"))
		routerCheckIn.GET("/recent", md.Tracer("web.checkin.makeListUserRecentCheckIn"), md.RequireStudent(srv), makeListUserRecentCheckIn)
		routerCheckIn.GET("/self", md.Tracer("web.checkin.makeListUserAllCheckIn"), md.CheckPage, md.RequireStudent(srv), makeListUserAllCheckIn)
	}

	routerCourse := router.Group("/course")
	{
		// teacher or student
		routerCourse.GET("", md.Tracer("web.course.makeListAllCourse"), md.CheckPage, makeListAllCourse)
		routerCourse.GET("/search", md.Tracer("web.course.makeListCoursesByName"), md.CheckPage, makeListCoursesByName)

		// teacher
		routerCourse.GET("/student/:courseID",
			md.Tracer("web.course.makeListStudentsByCourseID"), md.CheckPage, md.CheckParamID("courseID"), md.RequireTeacher(srv),
			makeListStudentsByCourseID("courseID"),
		)
		routerCourse.GET("/setup", md.Tracer("web.course.makeListCoursesByTeacherID"), md.CheckPage, md.RequireTeacher(srv), makeListCoursesByTeacherID)
		routerCourse.GET("/coding_time/:courseID",
			md.Tracer("web.course.makeListCodingTimeByCourseID"), md.CheckPage, md.CheckParamID("courseID"), md.RequireTeacher(srv),
			makeListCodingTimeByCourseID("courseID"),
		)
		routerCourse.GET("/student/examine/:courseID",
			md.Tracer("web.course.makeListStudentWaitingChecked"), md.CheckPage, md.CheckParamID("courseID"), md.RequireTeacher(srv),
			makeListStudentWaitingChecked("courseID"),
		)
		routerCourse.GET("/score/export/:courseID", md.Tracer("web.course.makeExportScoreCSV"), md.CheckParamID("courseID"), md.RequireTeacher(srv), makeExportScoreCSV("courseID"))
		routerCourse.GET("/student/export/template", md.Tracer("web.course.makeExportStudentCSVTemplate"), md.RequireTeacher(srv), makeExportStudentCSVTemplate)
		routerCourse.GET("/score", md.Tracer("web.course.makeListCourseScores"), md.CheckPage, md.CheckQueryID("courseId"), md.RequireTeacher(srv), makeListCourseScores("courseId"))
		routerCourse.POST("/examine", md.Tracer("web.course.makePermitStudents"), md.RequireTeacher(srv), makePermitStudents)
		routerCourse.POST("", md.Tracer("web.course.makeAddCourse"), md.RequireTeacher(srv), makeAddCourse)
		routerCourse.POST("/student/import", md.Tracer("web.course.makeImportStudentCSV"), md.CheckFormID("courseId"), md.RequireTeacher(srv), makeImportStudentCSV("courseId"))
		routerCourse.PUT("", md.Tracer("web.course.makeUpdateCourse"), md.RequireTeacher(srv), makeUpdateCourse)
		routerCourse.DELETE("/:courseID", md.Tracer("web.course.makeDeleteCourse"), md.CheckParamID("courseID"), md.RequireTeacher(srv), makeDeleteCourse("courseID"))

		// student
		routerCourse.GET("/study", md.Tracer("web.course.makeListCoursesByStudentID"), md.CheckPage, md.RequireStudent(srv), makeListCoursesByStudentID)
		routerCourse.POST("/attend", md.Tracer("web.course.makeStudentEnrollCourse"), md.RequireStudent(srv), makeStudentEnrollCourse)
		routerCourse.DELETE("/quit", md.Tracer("web.course.makeStudentQuitCourse"), md.RequireStudent(srv), md.CheckJSONID("courseid"), makeStudentQuitCourse("courseid"))
		routerCourse.GET("/:courseID", md.Tracer("web.course.makeGetCourseByID"), md.CheckParamID("courseID"), makeGetCourseByID("courseID"))

		routerCourseResource := routerCourse.Group("/resource")
		{
			routerCourseResource.GET("",
				md.Tracer("web.course.resource.makeListCourseResourcesByCourseID"), md.CheckPage, md.CheckQueryID("courseId"), md.AuthCourse(srv, "courseId"),
				makeListCourseResourcesByCourseID("courseId"),
			)
			routerCourseResource.GET("/:courseResourceID",
				md.Tracer("web.course.resource.makeFindCourseResourceByID"), md.CheckPage, md.CheckParamID("courseResourceID"), md.AuthCourseResource(srv, "courseResourceID"),
				makeFindCourseResourceByID("courseResourceID"),
			)

			// teacher
			routerCourseResource.POST("", md.Tracer("web.course.resource.makeAddCourseResource"), md.RequireTeacher(srv), makeAddCourseResource)
			routerCourseResource.PUT("", md.Tracer("web.course.resource.makeUpdateResource"), md.RequireTeacher(srv), makeUpdateResource)
			routerCourseResource.DELETE("/:courseResourceID",
				md.Tracer("web.course.resource.makeDeleteCourseResource"), md.CheckParamID("courseResourceID"), md.RequireTeacher(srv),
				makeDeleteCourseResource("courseResourceID"),
			)
		}
	}

	routerUpload := router.Group("/upload")
	{
		routerUpload.POST("/pic",
			md.Tracer("web.upload.makeUploadPicture"), md.CheckFormInt("width"), md.CheckFileHeader("pic"), md.CheckFileExt("pic", []string{"gif", "png", "jpg"}), md.SetImageType(),
			makeUploadPicture("pic", "width"),
		)
		routerUpload.POST("/video",
			md.Tracer("web.upload.makeUploadVideo"), md.CheckFileHeader("video"), md.CheckFileExt("video", []string{"mp4", "avi"}), md.SetVideoType(),
			makeUploadVideo("video"),
		)
		routerUpload.POST("/attachments",
			md.Tracer("web.upload.makeUploadAttachment"), md.CheckFileHeader("attachment"), md.CheckFileExt("attachment", []string{"pdf", "docx", "doc", "txt", "rar", "zip", "ppt", "pptx", "csv", "md"}),
			makeUploadAttachment("attachment"),
		)
		routerUpload.POST("/pdf",
			md.Tracer("web.upload.makeUploadReport"), md.CheckFileHeader("pdf"), md.CheckFileExt("pdf", []string{"pdf", "docx", "doc", "txt"}),
			makeUploadReport("pdf"),
		)
	}

	routerIDE := router.Group("/ide")
	{
		routerIDE.POST("", md.Tracer("web.ide.makeOpenIDE"), md.CheckJSONID("labId"), md.RequireStudent(srv), makeOpenIDE("labId"))
		routerIDE.POST("/heartbeat", md.Tracer("web.ide.makeHeartBeatForStudent"), md.CheckJSONID("labid"), md.RequireStudent(srv), makeHeartBeatForStudent("labid"))
		routerIDE.POST("/heartbeat/teacher", md.Tracer("web.ide.makeHeartBeatForTeacher"), md.RequireTeacher(srv), makeHeartBeatForTeacher)
	}

	routerLab := router.Group("/lab")
	{
		routerLab.GET("", md.Tracer("web.lab.makeListLabsByCourseID"), md.CheckPage, md.CheckQueryID("courseId"), makeListLabsByCourseID("courseId"))
		routerLab.GET("/score", md.Tracer("web.lab.makeListLabsScoreByStudentIDAndCourseID"), md.CheckPage, makeListLabsScoreByStudentIDAndCourseID)
		routerLab.GET("/:labID", md.Tracer("web.lab.makeGetLabByID"), md.CheckParamID("labID"), makeGetLabByID("labID"))

		// teacher only
		routerLab.POST("", md.Tracer("web.lab.makeAddLab"), md.RequireTeacher(srv), makeAddLab)
		routerLab.PUT("", md.Tracer("web.lab.makeUpdateLab"), md.RequireTeacher(srv), makeUpdateLab)
		routerLab.DELETE("", md.Tracer("web.lab.makeDeleteLabByID"), md.CheckJSONID("labId"), md.RequireTeacher(srv), makeDeleteLabByID("labId"))
		routerLab.GET("/check_code/quick", md.Tracer("web.lab.makeGetTreeNode"), md.RequireTeacher(srv), makeGetTreeNode)
		routerLab.POST("/check_code", md.Tracer("web.lab.makeCheckCode"), md.RequireTeacher(srv), makeCheckCode)
		routerLab.GET("/plagiarism_history/:labid",
			md.Tracer("web.lab.makeListHistoryDetectionReports"), md.CheckPage, md.CheckParamID("labid"), md.RequireTeacher(srv),
			makeListHistoryDetectionReports("labid"),
		)
		routerLab.GET("/plagiarism_view/:reportid", md.Tracer("web.lab.makeGetDetectionReport"), md.CheckParamID("reportid"), md.RequireTeacher(srv), makeGetDetectionReport("reportid"))

		// student only
		routerLab.GET("/details",
			md.Tracer("web.lab.makeListLabsByUserIDAndCourseID"), md.CheckPage, md.CheckQueryID("courseId"), md.RequireStudent(srv),
			makeListLabsByUserIDAndCourseID("courseId"),
		)
		routerLab.GET("/student", md.Tracer("web.lab.makeGetLabByStudentID"), md.CheckPage, md.RequireStudent(srv), makeGetLabByStudentID)

		routerLabSumit := routerLab.Group("/summit")
		{
			routerLabSumit.GET("/comment", md.Tracer("web.lab.summit.makeGetCommentsByUserIDAndLabID"), makeGetCommentsByUserIDAndLabID)

			routerLabSumit.GET("/:labID",
				md.Tracer("web.lab.summit.makeListLabSubmitsByID"), md.CheckPage, md.CheckParamID("labID"), md.RequireTeacher(srv),
				makeListLabSubmitsByID("labID"),
			)
			routerLabSumit.GET("/plagiarism/:labID/:fileName",
				md.Tracer("web.lab.summit.makeClickPlagiarismURL"), md.CheckParamID("labID"), md.RequireTeacher(srv),
				makeClickPlagiarismURL("labID"),
			)
			routerLabSumit.GET("/plagiarism/:labID", md.Tracer("web.lab.summit.makePlagiarismCheck"), md.CheckParamID("labID"), md.RequireTeacher(srv), makePlagiarismCheck("labID"))
			routerLabSumit.PUT("/comment", md.Tracer("web.lab.summit.makeUpdateLabComment"), md.RequireTeacher(srv), makeUpdateLabComment)
			routerLabSumit.PUT("/score", md.Tracer("web.lab.summit.makeUpdateLabSubmitScore"), md.RequireTeacher(srv), makeUpdateLabSubmitScore)

			routerLabSumit.GET("/report", md.Tracer("web.lab.summit.makeGetReportURL"), md.CheckQueryID("labId"), md.RequireStudent(srv), makeGetReportURL("labId"))
			routerLabSumit.POST("/code", md.Tracer("web.lab.summit.makeUpdateLabFinish"), md.RequireStudent(srv), makeUpdateLabFinish)
			routerLabSumit.POST("/report", md.Tracer("web.lab.summit.makeUploadLabReport"), md.RequireStudent(srv), makeUploadLabReport)
		}
	}

	routerMonaco := router.Group("/monaco")
	{
		routerMonaco.POST("/exec", md.Tracer("web.monaco.makeExecCode"), makeExecCode)
	}

}
