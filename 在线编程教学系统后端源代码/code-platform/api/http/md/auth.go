package md

import (
	"context"

	xhttp "code-platform/api/http"
	"code-platform/pkg/errorx"
	"code-platform/pkg/httpx"

	"github.com/gin-gonic/gin"
)

func AuthCourseForTeacher(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, courseID, teacherID uint64) bool {
	err := srv.CourseService.AuthCourseForTeacher(ctx, courseID, teacherID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "id is invalid")
		return false
	case errorx.ErrFailToAuth:
		httpx.AbortForbidden(c)
		return false
	default:
		httpx.AbortInternalErr(c)
		return false
	}
	return true
}

func AuthCourseForStudent(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, courseID, studentID uint64) bool {
	err := srv.CourseService.QueryWhetherStudentInCourse(ctx, courseID, studentID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortForbidden(c)
		return false
	default:
		httpx.AbortInternalErr(c)
		return false
	}
	return true
}

func AuthCourse(srv *xhttp.UnionService, courseTag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.GetUint64(courseTag)
		userID := c.GetUint64(KeyUserID)
		ctx := c.Request.Context()
		AuthCourseForStudentAndTeacher(ctx, c, srv, courseID, userID)
	}
}

func AuthCheckInForTeacher(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, recordID, teacherID uint64) bool {
	courseID, err := srv.CheckInService.GetCourseIDByRecordID(ctx, recordID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "id is invalid")
		return false
	default:
		httpx.AbortInternalErr(c)
		return false
	}
	return AuthCourseForTeacher(ctx, c, srv, courseID, teacherID)
}

func AuthCourseResourceForTeacher(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, courseResourceID, teacherID uint64) bool {
	courseID, err := srv.CourseResourceService.GetCourseIDByCourseResourceID(ctx, courseResourceID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "id is invalid")
		return false
	default:
		httpx.AbortInternalErr(c)
		return false
	}
	return AuthCourseForTeacher(ctx, c, srv, courseID, teacherID)
}

func AuthCourseResourceForstudent(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, courseResourceID, studentID uint64) bool {
	courseID, err := srv.CourseResourceService.GetCourseIDByCourseResourceID(ctx, courseResourceID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "id is invalid")
		return false
	default:
		httpx.AbortInternalErr(c)
		return false
	}
	return AuthCourseForStudent(ctx, c, srv, courseID, studentID)
}

func AuthCourseForStudentAndTeacher(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, courseID uint64, userID uint64) bool {
	role, err := getUserRole(c, srv)
	if err != nil {
		httpx.AbortInternalErr(c)
		return false
	}
	switch role {
	case RoleStudent:
		return AuthCourseForStudent(ctx, c, srv, courseID, userID)
	case RoleTeacher:
		return AuthCourseForTeacher(ctx, c, srv, courseID, userID)
	default:
		httpx.AbortForbidden(c)
		return false
	}
}

func AuthCourseResourceForStudentAndTeacher(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, courseResourceID uint64, userID uint64) bool {
	role, err := getUserRole(c, srv)
	if err != nil {
		httpx.AbortInternalErr(c)
		return false
	}
	switch role {
	case RoleStudent:
		return AuthCourseResourceForstudent(ctx, c, srv, courseResourceID, userID)
	case RoleTeacher:
		return AuthCourseResourceForTeacher(ctx, c, srv, courseResourceID, userID)
	default:
		httpx.AbortForbidden(c)
		return false
	}
}

func AuthCourseResource(srv *xhttp.UnionService, courseResourceTag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseResourceID := c.GetUint64(courseResourceTag)
		userID := c.GetUint64(KeyUserID)
		ctx := c.Request.Context()
		AuthCourseResourceForStudentAndTeacher(ctx, c, srv, courseResourceID, userID)
	}
}

func AuthLabForTeacher(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, labID, teacherID uint64) bool {
	courseID, err := srv.LabService.GetCourseIDByLabID(ctx, labID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortBadParamsErr(c, "id is invalid")
		return false
	default:
		httpx.AbortInternalErr(c)
		return false
	}
	return AuthCourseForTeacher(ctx, c, srv, courseID, teacherID)
}

func AuthLabForStudent(ctx context.Context, c *gin.Context, srv *xhttp.UnionService, labID, studentID uint64) bool {
	courseID, err := srv.LabService.GetCourseIDByLabID(ctx, labID)
	switch err {
	case nil:
	case errorx.ErrIsNotFound:
		httpx.AbortForbidden(c)
		return false
	default:
		httpx.AbortInternalErr(c)
		return false
	}
	return AuthCourseForStudent(ctx, c, srv, courseID, studentID)
}
