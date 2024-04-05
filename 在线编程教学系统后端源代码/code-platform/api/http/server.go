package http

import (
	"code-platform/log"
	"code-platform/repository"
	"code-platform/service/checkin"
	"code-platform/service/comment"
	"code-platform/service/course"
	"code-platform/service/courseResource"
	"code-platform/service/file"
	"code-platform/service/ide"
	"code-platform/service/lab"
	"code-platform/service/monaco"
	"code-platform/service/user"
)

type UnionService struct {
	CheckInService        *checkin.CheckInService
	CommentService        *comment.CommentService
	UserService           *user.UserService
	LabService            *lab.LabService
	CourseResourceService *courseResource.CourseResourceService
	CourseService         *course.CourseService
	FileService           *file.FileService
	IDEService            *ide.IDEService
	MonacoService         *monaco.MonacoService
}

func NewUnionService() *UnionService {
	dao := repository.NewDao()
	serviceLogger := log.Sub("service")
	ideClient := ide.NewIDEClient()
	monacoClient := monaco.NewMonacoClient()
	return &UnionService{
		CheckInService:        checkin.NewCheckInService(dao, serviceLogger.Sub("checkIn")),
		CommentService:        comment.NewCommentService(dao, serviceLogger.Sub("comment")),
		UserService:           user.NewUserService(dao, serviceLogger.Sub("user")),
		LabService:            lab.NewLabService(dao, serviceLogger.Sub("lab"), lab.NewPlagiarismDetectionClient(), ideClient, monacoClient),
		CourseResourceService: courseResource.NewCourseResourceService(dao, serviceLogger.Sub("course_resource")),
		CourseService:         course.NewCourseService(dao, serviceLogger.Sub("course")),
		FileService:           file.NewFileService(dao, serviceLogger.Sub("file")),
		IDEService:            ide.NewIDEService(dao, serviceLogger.Sub("ide"), ideClient),
		MonacoService:         monaco.NewMonacoService(dao, log.Sub("monaco"), monacoClient),
	}
}
