package comment

import (
	"time"

	"code-platform/service/define"
)

type (
	PageResponse = define.PageResponse
	PageInfo     = define.PageInfo
)

type SubCourseCommentResp struct {
	*CourseCommentResp
	ReplyUserName string `json:"reply_username"`
}

type CourseCommentResp struct {
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Username        string    `json:"username"`
	UserAvatar      string    `json:"user_avatar_url"`
	CommentText     string    `json:"comment_text"`
	CourseID        uint64    `json:"course_id"`
	UserID          uint64    `json:"user_id"`
	CourseCommentID uint64    `json:"course_comment_id"`
	ReplyUserID     uint64    `json:"reply_user_id"`
	Pid             uint64    `json:"pid"`
}

type CourseCommentEntity struct {
	Comment     *CourseCommentResp      `json:"comment"`
	SubComments []*SubCourseCommentResp `json:"reply_comments"`
}
type SubLabCommentResp struct {
	*LabCommentResp
	ReplyUserName string `json:"reply_username"`
}

type LabCommentResp struct {
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Username     string    `json:"username"`
	UserAvatar   string    `json:"user_avatar_url"`
	CommentText  string    `json:"comment_text"`
	LabID        uint64    `json:"lab_id"`
	UserID       uint64    `json:"user_id"`
	LabCommentID uint64    `json:"lab_comment_id"`
	ReplyUserID  uint64    `json:"reply_user_id"`
	Pid          uint64    `json:"pid"`
}

type LabCommentEntity struct {
	Comment     *LabCommentResp      `json:"comment"`
	SubComments []*SubLabCommentResp `json:"reply_comments"`
}
