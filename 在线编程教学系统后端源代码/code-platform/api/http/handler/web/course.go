package web

import (
	"database/sql"
	"net/http"
	"strings"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeGetCourseByID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		userID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		resp, err := srv.CourseService.GetCourseInfoByUserIDAndCourseID(ctx, userID, courseID)
		switch err {
		case nil:
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "Can't not find the course")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeListAllCourse(c *gin.Context) {
	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.CourseService.ListAllCourses(ctx, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeAddCourse(c *gin.Context) {
	type addCourseRequest struct {
		CourseName        string `json:"courseName"`
		CourseDescription string `json:"courseDes"`
		PicURL            string `json:"picurl"`
		SecretKey         string `json:"secretkey"`
		Language          int8   `json:"language"`
		NeedAudit         bool   `json:"need_audit"`
	}
	var req addCourseRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in add course request")
		return
	}

	userID := c.GetUint64(md.KeyUserID)
	// 去除密钥前后空格
	req.SecretKey = strings.TrimSpace(req.SecretKey)

	ctx := c.Request.Context()
	if err := srv.CourseService.AddCourse(ctx, userID, req.CourseName, req.CourseDescription, false, req.PicURL, req.SecretKey, req.Language, req.NeedAudit); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeUpdateCourse(c *gin.Context) {
	type updateCourseRequest struct {
		CourseName        string `json:"courseName"`
		CourseDescription string `json:"courseDes"`
		SecretKey         string `json:"secretkey"`
		PicURL            string `json:"picUrl"`
		CourseID          uint64 `json:"courseId"`
		Language          int8   `json:"language"`
		NeedAudit         bool   `json:"need_audit"`
	}

	var req updateCourseRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update course request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "course_id is invalid")
		return
	}

	// language 范围 [0, 2]
	if req.Language < 0 || req.Language > 2 {
		httpx.AbortBadParamsErr(c, "language is invalid")
		return
	}

	req.SecretKey = strings.TrimSpace(req.SecretKey)

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCourseForTeacher(ctx, c, srv, req.CourseID, teacherID) {
		return
	}

	err := srv.CourseService.UpdateCourse(ctx, req.CourseID, req.CourseName, req.CourseDescription, req.SecretKey, req.PicURL, req.Language, req.NeedAudit)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "course is not found by course_id")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeDeleteCourse(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		err := srv.CourseService.DeleteCourse(ctx, courseID, teacherID)
		switch err {
		case nil:
		case errorx.ErrFailToAuth:
			httpx.AbortFailToAuth(c, "Without enough privilege to delete this course")
			return
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "course is not found by course_id")
			return
		}

		c.Status(http.StatusOK)
	}
}

func makeListCoursesByTeacherID(c *gin.Context) {
	userID := c.GetUint64(md.KeyUserID)
	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.CourseService.ListCourseInfosByTeacherID(ctx, userID, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeListCoursesByStudentID(c *gin.Context) {
	userID := c.GetUint64(md.KeyUserID)
	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.CourseService.ListCourseInfosByStudentID(ctx, userID, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeListStudentsByCourseID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		resp, err := srv.CourseService.ListStudentsCheckedByCourseID(ctx, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeStudentEnrollCourse(c *gin.Context) {
	type enrollCourseRequest struct {
		SecretKey string `json:"secretKey"`
		CourseID  uint64 `json:"courseID"`
	}

	var req enrollCourseRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in enroll course request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "params is in valid")
		return
	}

	userID := c.GetUint64(md.KeyUserID)

	ctx := c.Request.Context()
	err := srv.CourseService.EnrollForStudent(ctx, userID, req.CourseID, sql.NullString{String: req.SecretKey, Valid: req.SecretKey != ""})
	switch err {
	case nil, errorx.ErrMySQLDuplicateKey:
	case errorx.ErrFailToAuth:
		httpx.AbortFailToAuth(c, "secret key is not correct")
		return
	case errorx.ErrPersonalInfoNotComplete:
		httpx.AbortForbidden(c)
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeStudentQuitCourse(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		studentID := c.GetUint64(md.KeyUserID)
		ctx := c.Request.Context()
		if err := srv.CourseService.QuitCourse(ctx, courseID, studentID); err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Status(http.StatusOK)
	}
}

func makeListCoursesByName(c *gin.Context) {
	name, ok := c.GetQuery("courseName")
	if !ok {
		httpx.AbortGetParamsErr(c, "Fail to get params in search course by name request for student")
		return
	}

	if strings.TrimSpace(name) == "" {
		httpx.AbortBadParamsErr(c, "name should't be empty str")
		return
	}

	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.CourseService.ListCourseInfosByCourseName(ctx, name, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeExportScoreCSV(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		data, err := srv.CourseService.ExportScoreCSVByCourseID(ctx, courseID)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Header("content-type", "application/csv")
		c.Header("content-disposition", "attachment;filename=成绩表.csv")
		c.Writer.Write(data)
		c.Writer.Flush()
		c.Status(http.StatusOK)
	}
}

func makeExportStudentCSVTemplate(c *gin.Context) {
	csvTemplate, err := srv.CourseService.ExportCSVTemplate()
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Header("content-type", "application/csv")
	c.Header("content-disposition", "attachment;filename=导入表模板.csv")
	c.Writer.Write(csvTemplate)
	c.Writer.Flush()
	c.Status(http.StatusOK)
}

func makeImportStudentCSV(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)

		mineHeader, err := c.FormFile("csv")
		if err != nil {
			httpx.AbortGetParamsErr(c, "Fail to get upload file")
			return
		}
		data, err := srv.FileService.MIMEHeaderToBytes(mineHeader)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		teacherID := c.GetUint64(md.KeyUserID)
		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		switch err := srv.CourseService.ImportCSVTemplate(ctx, mineHeader.Filename, data, courseID); err {
		case nil:
		case errorx.ErrUnsupportFileType:
			httpx.AbortBadParamsErr(c, "File type is not csv")
			return
		case errorx.ErrPersonalInfoInvalid:
			httpx.AbortBadParamsErr(c, "Some student data in file is invalid")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Status(http.StatusOK)
	}
}

func makeListStudentWaitingChecked(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		resp, err := srv.CourseService.ListStudentWaitForChecked(ctx, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makePermitStudents(c *gin.Context) {
	type checkStudentRequest struct {
		StudentIDs  []uint64 `json:"stuIDs"`
		CourseID    uint64   `json:"courseID"`
		IsPermitted bool     `json:"isPermitted"`
	}

	var req checkStudentRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in check student request")
		return
	}

	if req.CourseID <= 0 || len(req.StudentIDs) == 0 {
		httpx.AbortBadParamsErr(c, "params are invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCourseForTeacher(ctx, c, srv, req.CourseID, teacherID) {
		return
	}

	if err := srv.CourseService.CheckForStudents(ctx, req.CourseID, req.StudentIDs, req.IsPermitted); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeListCourseScores(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		resp, err := srv.CourseService.ListCoursesScore(ctx, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}
