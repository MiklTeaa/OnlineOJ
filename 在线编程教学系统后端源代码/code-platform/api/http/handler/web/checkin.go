package web

import (
	"net/http"

	"code-platform/api/http/md"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeTeacherPrepareCheckIn(c *gin.Context) {
	type prepareCheckInRequest struct {
		Name     string `json:"name"`
		CourseID uint64 `json:"courseId"`
		Duration int    `json:"duration"`
	}

	var req prepareCheckInRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in prepare checkIn request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "courseID is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCourseForTeacher(ctx, c, srv, req.CourseID, teacherID) {
		return
	}

	if err := srv.CheckInService.PrepareCheckIn(ctx, teacherID, req.CourseID, req.Name, req.Duration); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeStudentStartCheckIn(c *gin.Context) {
	type startCheckInRequest struct {
		CourseID uint64 `json:"courseId"`
	}
	var req startCheckInRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in start checkIn request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "courseID is invalid")
		return
	}

	userID := c.GetUint64(md.KeyUserID)

	ctx := c.Request.Context()
	switch err := srv.CheckInService.StartCheckIn(ctx, req.CourseID, userID); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "courseID or studentID is invalid to find a record")
		return
	case errorx.ErrRedisKeyNil:
		httpx.AbortFailToAuth(c, "ddl")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeListCheckInRecordsByCourseID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		resp, err := srv.CheckInService.ListRecordsByCourseID(ctx, courseID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeListCheckInDetailsByRecordID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		recordID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		if !md.AuthCheckInForTeacher(ctx, c, srv, recordID, teacherID) {
			return
		}

		resp, err := srv.CheckInService.ListCheckInDetailsByRecordID(ctx, recordID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeListCheckInDetailsByUserIDAndCourseID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		userID := c.GetUint64(md.KeyUserID)
		pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

		ctx := c.Request.Context()
		resp, err := srv.CheckInService.ListCheckInDetailsByUserIDAndCourseID(ctx, courseID, userID, (pageCurrent-1)*pageSize, pageSize)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
	}
}

func makeDeleteCheckInData(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint64(md.KeyUserID)
		recordID := c.GetUint64(tag)

		ctx := c.Request.Context()
		if !md.AuthCheckInForTeacher(ctx, c, srv, recordID, userID) {
			return
		}

		switch err := srv.CheckInService.DeleteCheckInDataByRecordID(ctx, recordID, userID); err {
		case nil:
		case errorx.ErrFailToAuth:
			httpx.AbortFailToAuth(c, "you don't have privilege to do it")
			return
		case errorx.ErrIsNotFound:
			httpx.AbortBadParamsErr(c, "record is not found by recordID")
			return
		default:
			httpx.AbortInternalErr(c)
			return
		}

		c.Status(http.StatusOK)
	}
}

func makeUpdateCheckInDetail(c *gin.Context) {
	type updateCheckInDetailRequest struct {
		StudentID       uint64 `json:"stuID"`
		CheckInRecordID uint64 `json:"checkInRecordID"`
		IsCheckIn       bool   `json:"isCheckIN"`
	}

	var req updateCheckInDetailRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "Fail to get params in update checkInDetail request")
		return
	}

	if req.StudentID <= 0 || req.CheckInRecordID <= 0 {
		httpx.AbortBadParamsErr(c, "ID is invalid")
		return
	}

	teacherID := c.GetUint64(md.KeyUserID)
	ctx := c.Request.Context()
	if !md.AuthCheckInForTeacher(ctx, c, srv, req.CheckInRecordID, teacherID) {
		return
	}

	switch err := srv.CheckInService.UpdateCheckInDetail(ctx, req.StudentID, req.CheckInRecordID, req.IsCheckIn); err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "record is not found by userID and recordID")
		return
	default:
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}

func makeExportCheckInRecords(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(tag)
		teacherID := c.GetUint64(md.KeyUserID)

		ctx := c.Request.Context()
		if !md.AuthCourseForTeacher(ctx, c, srv, courseID, teacherID) {
			return
		}

		data, err := srv.CheckInService.ExportCheckInRecordsCSV(ctx, courseID)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Header("content-type", "application/csv")
		c.Header("content-disposition", "attachment;filename=签到表.csv")
		c.Writer.Write(data)
		c.Writer.Flush()
		c.Status(http.StatusOK)
	}
}

func makeListUserRecentCheckIn(c *gin.Context) {
	userID := c.GetUint64(md.KeyUserID)

	ctx := c.Request.Context()
	resp, err := srv.CheckInService.ListRecentUserCheckIn(ctx, userID)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeListUserAllCheckIn(c *gin.Context) {
	userID := c.GetUint64(md.KeyUserID)
	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.CheckInService.ListUserCheckIn(ctx, userID, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}
