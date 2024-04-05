package course

import (
	"time"

	"code-platform/service/define"
)

type (
	PageResponse = define.PageResponse
	PageInfo     = define.PageInfo
)

type (
	CodingTimeInfo                     = define.CodingTimeInfo
	UserWithAverageScoreAndCheckInData = define.UserWithAverageScoreAndCheckInData
)

var (
	batchToOuterUser = define.BatchToOuterUser
	toOuterUser      = define.ToOuterUser
)

type UserInfoWithCodingTime struct {
	Name            string            `json:"name"`
	Number          string            `json:"number"`
	CodingTimeInfos []*CodingTimeInfo `json:"coding_time"`
	UserID          uint64            `json:"user_id"`
}

type CourseWithTeacherInfo struct {
	CreatedAt     time.Time `json:"created_at"`
	CourseDes     string    `json:"course_des"`
	PicURL        string    `json:"pic_url"`
	TeacherName   string    `json:"teacher_name"`
	TeacherEmail  string    `json:"teacher_email"`
	CourseName    string    `json:"course_name"`
	TeacherAvatar string    `json:"teacher_avatar"`
	CourseID      uint64    `json:"course_id"`
	TeacherID     uint64    `json:"teacher_id"`
	IsClose       bool      `json:"is_close"`
	NeedAudit     bool      `json:"need_audit"`
	Language      int8      `json:"language"`
}

type CourseWithTeacherInfoAndIsEnroll struct {
	*CourseWithTeacherInfo
	IsEnroll bool `json:"is_enroll"`
}

type CourseInfo struct {
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	PictureURL        string    `json:"pic_url"`
	SecretKey         string    `json:"secret_key"`
	CourseName        string    `json:"name"`
	CourseDescription string    `json:"description"`
	CourseID          uint64    `json:"course_id"`
	IsClose           bool      `json:"is_close"`
	NeedAudit         bool      `json:"need_audit"`
	Language          int8      `json:"language"`
}
