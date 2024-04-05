package model

import (
	"database/sql"
	"time"
)

type LabWithUserData struct {
	*Lab
	UserID uint64        `db:"user_id"`
	Score  sql.NullInt32 `db:"score"`
}

type CheckInDetailWithUserData struct {
	*User
	RecordID  uint64 `db:"record_id"`
	IsCheckIn bool   `db:"is_check_in"`
}

type CheckInRecordWithIsCheckInStatus struct {
	*CheckInRecord
	IsCheckIn bool `db:"is_check_in"`
}

type CourseCommentWithUserInfo struct {
	*CourseComment
	Name   string `db:"name"`
	Avatar string `db:"avatar"`
}

type LabInfoByUserIDAndCourseID struct {
	*Lab
	CourseName string
	ReportURL  string        `db:"report_url"`
	Comment    string        `db:"comment"`
	Score      sql.NullInt32 `db:"score"`
	IsFinish   bool          `db:"is_finish"`
}

type LabSubmitInfoByLabID struct {
	*LabSubmit
	Name       string `db:"name"`
	Number     string `db:"number"`
	CodingTime uint64
}

type CodingTimeInfo struct {
	Date     time.Time `db:"created_at_date"`
	Duration int       `db:"duration"`
}

type CodingTimeInfoWithUserID struct {
	*CodingTimeInfo
	UserID uint64 `db:"user_id"`
}

type DetectionReportIDWithCreatedAt struct {
	CreatedAt time.Time `db:"created_at"`
	ID        uint64    `db:"id"`
}
