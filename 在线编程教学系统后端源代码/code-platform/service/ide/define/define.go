package define

import (
	"fmt"
	"time"

	"code-platform/config"
	"code-platform/service/define"
)

const ContainerNamePrefix = "mytheia-"

const (
	HeartBeatTagFormatPrefixForStudent = "hbs:"
	// HeartBeatTagFormatForTeacher labID:studentID:teacherID
	HeartBeatTagFormatForTeacher = "hbt:%d:%d:%d"
)

// HeartBeatTagFormatForStudent labID:studentID
var HeartBeatTagFormatForStudent = HeartBeatTagFormatPrefixForStudent + "%d:%d"

var InitBasePath = define.InitBasePath()

func GetContainerNameForStudent(labID, studentID uint64) string {
	return fmt.Sprintf(ContainerNamePrefix+"%d-%d", labID, studentID)
}

func GetContainerNameForTeacher(labID, studentID, teacherID uint64) string {
	return fmt.Sprintf(ContainerNamePrefix+"%d-%d-%d", labID, studentID, teacherID)
}

func GetImageName(language int8) string {
	languageMap := config.Theia.GetStringMapString("imageName")
	switch language {
	case 0:
		return languageMap["python3"]
	case 1:
		return languageMap["cpp"]
	default:
		return languageMap["java"]
	}
}

type TeacherInfo struct {
	TeacherName string `json:"teacher_name"`
	TeacherID   uint64 `json:"teacher_id"`
}

type ContainerInfo struct {
	CreatedAt   time.Time    `json:"created_at"`
	TeacherInfo *TeacherInfo `json:"teacher_info,omitempty"`
	ContainerID string       `json:"container_id"`
	Size        string       `json:"size"`
	CourseName  string       `json:"course_name"`
	Port        uint16       `json:"port"`
	StudentName string       `json:"student_name"`
	LabName     string       `json:"lab_name"`
	CPUPerc     string       `json:"cpu_perc"`
	MemUsage    string       `json:"mem_usage"`
	LabID       uint64       `json:"lab_id"`
	CourseID    uint64       `json:"course_id"`
	StudentID   uint64       `json:"student_id"`
	LabHasEnd   bool         `json:"lab_is_end"`
}

type (
	PageResponse = define.PageResponse
	PageInfo     = define.PageInfo
)

// 假设前端 30 秒发送轮询，则 64 秒至少有两次机会可以命中
// 64 秒内无命中，则自动删除key
const HeartBeatDuration = 64 * time.Second
