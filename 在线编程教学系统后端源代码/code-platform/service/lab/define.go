package lab

import (
	"time"

	"code-platform/service/define"
)

type (
	PageResponse = define.PageResponse
	PageInfo     = define.PageInfo
)

type LabInfo struct {
	DeadLine      time.Time `json:"dead_line"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ReportURL     string    `json:"report_url"`
	Content       string    `json:"content"`
	Title         string    `json:"title"`
	CourseName    string    `json:"course_name"`
	AttachmentURL string    `json:"attachment_url"`
	Comment       string    `json:"comment"`
	CourseID      uint64    `json:"course_id"`
	LabID         uint64    `json:"lab_id"`
	Score         int32     `json:"score"`
	IsFinish      bool      `json:"is_finish"`
}

type LabScore struct {
	CreatedAt time.Time `json:"create_time"`
	Title     string    `json:"lab_title"`
	LabID     uint64    `json:"lab_id"`
	Score     int32     `json:"score"`
}

type LabCodingTimeData struct {
	CreatedTime time.Time `json:"CreatedAt"`
	UpdatedTime time.Time `json:"UpdatedAt"`
	Comment     string    `json:"Comment"`
	UserName    string    `json:"Name"`
	Number      string    `json:"Number"`
	ReportURL   string    `json:"ReportURL"`
	LabSubmitID uint64    `json:"ID"`
	LabID       uint64    `json:"LabID"`
	UserID      uint64    `json:"UserID"`
	CodingTime  uint64    `json:"CodingTime"`
	Score       int32     `json:"Score"`
	IsFinish    bool      `json:"IsFinish"`
}

type PlagiarismCheckResponse struct {
	URL        string `json:"url"`
	Similarity string `json:"similarity"`
	RealName1  string `json:"real_name_1"`
	RealName2  string `json:"real_name_2"`
	Num1       string `json:"num_1"`
	Num2       string `json:"num_2"`
	UserID1    uint64 `json:"user_id_1"`
	UserID2    uint64 `json:"user_id_2"`
}

type Lab struct {
	CreatedAt     time.Time `json:"created_at"`
	DeadLine      time.Time `json:"dead_line"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	AttachMentURL string    `json:"attachment_url"`
	CourseID      uint64    `json:"course_id"`
	ID            uint64    `json:"id"`
}

type DetectionReportResponse struct {
	CreatedAt string `json:"created_at"`
	ID        uint64 `json:"id"`
}
