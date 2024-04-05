package checkin

import (
	"time"

	"code-platform/service/define"
)

type (
	PageResponse = define.PageResponse
	PageInfo     = define.PageInfo
)

type CheckInData struct {
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	ID        uint64    `json:"checkin_record_id"`
	CourseID  uint64    `json:"course_id"`
	Actual    int       `json:"actual"`
	Total     int       `json:"total"`
}

type CheckInWithUserData struct {
	Number          string `json:"num"`
	Name            string `json:"real_name"`
	Organization    string `json:"organization"`
	UserID          uint64 `json:"user_id"`
	CheckinRecordID uint64 `json:"checkin_record_id"`
	IsCheckIn       bool   `json:"is_check_in"`
}

type CheckInRecordWithDetailData struct {
	CreatedAt       time.Time `json:"created_at"`
	Name            string    `json:"name"`
	CheckinRecordID uint64    `json:"checkin_record_id"`
	IsCheckIn       bool      `json:"is_checkin"`
}

type CheckInRecordPersonalData struct {
	CreatedAt  time.Time `json:"created_at"`
	DeadLine   time.Time `json:"dead_line"`
	IsFinish   *bool     `json:"is_finish,omitempty"`
	Name       string    `json:"name"`
	CourseName string    `json:"course_name"`
	CourseID   uint64    `json:"courseId"`
	ID         uint64    `json:"checkin_record_id"`
}
