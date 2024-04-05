package courseResource

import (
	"time"

	"code-platform/service/define"
)

type (
	PageResponse = define.PageResponse
	PageInfo     = define.PageInfo
)

type CourseResourcesData struct {
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	AttachMentURL string    `json:"attachment_url"`
	ID            uint64    `json:"course_recourse_id"` // TODO fix typo in json tag later
}
