package admin

import (
	"database/sql"
	"net/http"

	"code-platform/api/http/md"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeListAllCourses(c *gin.Context) {
	pageCurrent, pageSize := c.GetInt(md.KeyPageCurrent), c.GetInt(md.KeyPageSize)

	ctx := c.Request.Context()
	resp, err := srv.CourseService.ListAllCoursesByAdmin(ctx, (pageCurrent-1)*pageSize, pageSize)
	if err != nil {
		httpx.AbortInternalErr(c)
		return
	}
	c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(resp)))
}

func makeAmendCourse(c *gin.Context) {
	type updateCourseRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		PicURL      string `json:"pic_url"`
		SecretKey   string `json:"secret_key"`
		CourseID    uint64 `json:"course_id"`
		IsClosed    bool   `json:"is_closed"`
		NeedAudit   bool   `json:"need_audit"`
	}

	var req updateCourseRequest
	if err := c.ShouldBindWith(&req, jsonx.SonicDecoder); err != nil {
		httpx.AbortGetParamsErr(c, "fail to get params for update course request")
		return
	}

	if req.CourseID <= 0 {
		httpx.AbortBadParamsErr(c, "course_id is invalid")
		return
	}

	ctx := c.Request.Context()
	if err := srv.CourseService.UpdateCourseByAdmin(ctx, req.CourseID, req.Name, req.Description, req.PicURL, sql.NullString{Valid: req.SecretKey != "", String: req.SecretKey}, req.IsClosed, req.NeedAudit); err != nil {
		httpx.AbortInternalErr(c)
		return
	}

	c.Status(http.StatusOK)
}
