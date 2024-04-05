package ide

import (
	"context"
	"database/sql"
	"time"

	"code-platform/api/grpc/ide/pb"
	"code-platform/config"
	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/service/ide/define"
	"code-platform/service/ide/monitor"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var RunSweaterOpt = true

func NewIDEClient() pb.IDEServerServiceClient {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	port := config.IDEServer.GetString("port")
	conn, err := grpc.DialContext(ctx, "localhost:"+port, grpc.WithInsecure(), grpc.WithBlock())

	switch err {
	case nil:
	case ctx.Err():
		log.Sub("grpc.checkCode").Errorf(err, "dial to grpc server timeout")
		fallthrough
	default:
		panic(err)
	}

	return pb.NewIDEServerServiceClient(conn)
}

type IDEService struct {
	Dao       *repository.Dao
	Logger    *log.Logger
	IDEClient pb.IDEServerServiceClient
}

func NewIDEService(dao *repository.Dao, logger *log.Logger, ideClient pb.IDEServerServiceClient) *IDEService {
	ideService := &IDEService{
		Dao:       dao,
		Logger:    logger,
		IDEClient: ideClient,
	}
	if RunSweaterOpt {
		// 启动Worker，清理过期容器
		monitor.HeartBeatSweaping(dao.Storage, logger, ideClient)
	}
	return ideService
}

func (i *IDEService) OpenIDE(ctx context.Context, labID, studentID uint64) (port uint32, token string, err error) {
	lab, err := model.QueryLabByID(ctx, i.Dao.Storage.RDB, labID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		i.Logger.Debugf("lab is not found by id[%d]", labID)
		return 0, "", errorx.ErrIsNotFound
	default:
		i.Logger.Errorf(err, "Query lab by id[%d] failed", labID)
		return 0, "", errorx.InternalErr(err)
	}

	course, err := model.QueryCourseByID(ctx, i.Dao.Storage.RDB, lab.CourseID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		i.Logger.Debugf("Query Course By ID[%d] failed", lab.CourseID)
		return 0, "", errorx.ErrIsNotFound
	default:
		i.Logger.Errorf(err, "Query course by id[%d] failed", lab.CourseID)
		return 0, "", errorx.InternalErr(err)
	}

	var canEdit bool
	if lab.DeadLine.Valid && time.Since(lab.DeadLine.Time) > 0 {
		canEdit = false
	} else {
		canEdit = true
	}

	resp, err := i.IDEClient.GetIDEForStudent(ctx, &pb.GetIDEForStudentRequest{
		LabId:     labID,
		StudentId: studentID,
		Language:  uint32(course.Language),
		CanEdit:   canEdit,
	})
	if err != nil {
		i.Logger.Errorf(err, "get IDE failed with labID[%d] and studentID[%d]", labID, studentID)
		return 0, "", errorx.InternalErr(err)
	}

	if !resp.IsReused {
		// 第一次启动前手动 heart beat 一次
		if err := i.HeartBeatWhenStartingForStudent(ctx, labID, studentID); err != nil {
			return 0, "", err
		}
	}
	return resp.Port, resp.Token, nil
}

func (i *IDEService) CheckCode(ctx context.Context, labID, studentID, teacherID uint64) (url uint32, token string, err error) {
	lab, err := model.QueryLabByID(ctx, i.Dao.Storage.RDB, labID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		i.Logger.Debugf("lab is not found by id[%d]", labID)
		return 0, "", errorx.ErrIsNotFound
	default:
		i.Logger.Errorf(err, "Query lab by id[%d] failed", labID)
		return 0, "", errorx.InternalErr(err)
	}

	course, err := model.QueryCourseByID(ctx, i.Dao.Storage.RDB, lab.CourseID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		i.Logger.Debugf("Query Course By ID[%d] failed", lab.CourseID)
		return 0, "", errorx.ErrIsNotFound
	default:
		i.Logger.Errorf(err, "Query course by id[%d] failed", lab.CourseID)
		return 0, "", errorx.InternalErr(err)
	}

	if course.TeacherID != teacherID {
		i.Logger.Debugf("teacher[%d] want to see the lab of teacher[%d]", teacherID, course.TeacherID)
		return 0, "", errorx.ErrFailToAuth
	}

	resp, err := i.IDEClient.GetIDEForTeacher(ctx, &pb.GetIDEForTeacherRequest{
		LabId:     labID,
		StudentId: studentID,
		TeacherId: teacherID,
		Language:  uint32(course.Language),
	})
	if err != nil {
		i.Logger.Errorf(err, "get IDE failed with labID[%d] and studentID[%d] and teacherID[%d]", labID, studentID, teacherID)
		return 0, "", errorx.InternalErr(err)
	}

	// 启动前手动 heart beat 一次
	if err := i.HeartBeatForTeacher(ctx, labID, studentID, teacherID); err != nil {
		i.Logger.Errorf(err, "heart beat for teacher failed after starting ide")
		return 0, "", errorx.InternalErr(err)
	}
	return resp.Port, resp.Token, nil
}

func (i *IDEService) ListContainers(ctx context.Context, offset, limit int, order pb.OrderType, isReverse bool) (*define.PageResponse, error) {
	resp, err := i.IDEClient.GetContainers(ctx, &pb.GetContainersRequest{
		Offset:    uint32(offset),
		Limit:     uint32(limit),
		Order:     order,
		IsReverse: isReverse,
	})
	if err != nil {
		if status.Code(err) == codes.Canceled {
			i.Logger.Debug("GetContainers is canceled")
			return nil, context.Canceled
		}
		i.Logger.Error(err, "get containers failed")
		return nil, errorx.InternalErr(err)
	}

	containerInfos := make([]*define.ContainerInfo, len(resp.ContainerInfos))

	labIDs := make([]uint64, len(resp.ContainerInfos))
	studentIDs := make([]uint64, len(resp.ContainerInfos))
	teacherIDs := make([]uint64, 0, len(resp.ContainerInfos))

	for index, info := range resp.ContainerInfos {
		var teacherInfo *define.TeacherInfo
		if info.TeacherInfo != nil {
			teacherInfo = &define.TeacherInfo{
				TeacherID: info.TeacherInfo.TeacherId,
			}
			teacherIDs = append(teacherIDs, info.TeacherInfo.TeacherId)
		}

		containerInfos[index] = &define.ContainerInfo{
			ContainerID: info.ContainerId,
			LabID:       info.LabId,
			StudentID:   info.StudentId,
			Size:        info.Size,
			TeacherInfo: teacherInfo,
			CreatedAt:   time.Unix(info.CreatedAt, 0),
			Port:        uint16(info.Port),
			CPUPerc:     info.CpuPerc,
			MemUsage:    info.MemoryUsage,
		}
		labIDs[index] = info.LabId
		studentIDs[index] = info.StudentId
	}

	// reduce capacity
	teacherIDs = teacherIDs[:len(teacherIDs):len(teacherIDs)]

	var (
		labIDsMap          map[uint64]*model.Lab
		courseIDsMap       map[uint64]*model.Course
		labIDToCourseIDMap map[uint64]uint64

		studentsMap map[uint64]*model.User
		teachersMap map[uint64]*model.User
	)

	tasks := []func() error{
		// 查询实验与
		func() (err error) {
			subTasks := []func() error{
				func() (err error) {
					labIDsMap, err = model.QueryLabMapsByIDs(ctx, i.Dao.Storage.RDB, labIDs)
					switch err {
					case nil:
					case context.Canceled:
						i.Logger.Debug("QueryLabMapsByIDs is canceled")
						return err
					default:
						i.Logger.Errorf(err, "QueryLabMapsBy ids %v failed", labIDs)
						return errorx.InternalErr(err)
					}
					return nil
				},
				func() (err error) {
					var courseIDs []uint64
					labIDToCourseIDMap, courseIDs, err = model.QueryLabIDToCourseIDMapByLabIDs(ctx, i.Dao.Storage.RDB, labIDs)
					switch err {
					case nil:
					case context.Canceled:
						i.Logger.Debug("QueryLabIDToCourseIDMapByLabIDs is canceled")
						return err
					default:
						i.Logger.Errorf(err, "QueryLabIDToCourseIDMap by labIDs %v failed", labIDs)
						return errorx.InternalErr(err)
					}

					courseIDsMap, err = model.QueryCourseMapsByIDs(ctx, i.Dao.Storage.RDB, courseIDs)
					switch err {
					case nil:
					case context.Canceled:
						i.Logger.Debug("QueryCourseMapsByIDs is canceled")
						return err
					default:
						i.Logger.Errorf(err, "QueryCourseMapsByIDs by courseIDs %v failed", courseIDs)
						return errorx.InternalErr(err)
					}
					return nil
				},
			}

			if err := parallelx.Do(i.Logger, subTasks...); err != nil {
				return err
			}
			return nil
		},
		// 查询用户信息
		func() (err error) {
			studentsMap, err = model.QueryUserMapByIDs(ctx, i.Dao.Storage.RDB, studentIDs)
			switch err {
			case nil:
			case context.Canceled:
				i.Logger.Debug("QueryUserMapByIDs is canceled")
				return err
			default:
				i.Logger.Errorf(err, "QueryUserMap by IDs %v failed", studentIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			teachersMap, err = model.QueryUserMapByIDs(ctx, i.Dao.Storage.RDB, teacherIDs)
			switch err {
			case nil:
			case context.Canceled:
				i.Logger.Debug("QueryUserMapByIDs is canceled")
				return err
			default:
				i.Logger.Errorf(err, "QueryUserMapByIDs by ids %v failed", teacherIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(i.Logger, tasks...); err != nil {
		return nil, err
	}

	type labInfo struct {
		LabName    string
		CourseName string
		CourseID   uint64
		HasEnd     bool
	}

	var labInfos = make(map[uint64]*labInfo)
	for _, labID := range labIDs {
		lab := labIDsMap[labID]
		courseID := labIDToCourseIDMap[labID]
		course := courseIDsMap[courseID]
		labInfos[labID] = &labInfo{
			LabName:    lab.Title,
			CourseID:   courseID,
			CourseName: course.Name,
			HasEnd:     lab.DeadLine.Valid && time.Since(lab.DeadLine.Time) >= 0,
		}
	}

	for index, containerInfo := range containerInfos {
		// for user and course
		containerInfos[index].LabName = labInfos[containerInfo.LabID].LabName
		containerInfos[index].CourseID = labInfos[containerInfo.LabID].CourseID
		containerInfos[index].CourseName = labInfos[containerInfo.LabID].CourseName
		containerInfos[index].LabHasEnd = labInfos[containerInfo.LabID].HasEnd

		// for user
		containerInfos[index].StudentName = studentsMap[containerInfo.StudentID].Name
		if containerInfos[index].TeacherInfo != nil {
			containerInfos[index].TeacherInfo.TeacherName = teachersMap[containerInfo.TeacherInfo.TeacherID].Name
		}
	}

	return &define.PageResponse{
		Records: containerInfos,
		PageInfo: &define.PageInfo{
			Total: int(resp.Total),
		},
	}, nil
}

func (i *IDEService) StopContainer(ctx context.Context, containerID string) error {
	_, err := i.IDEClient.StopContainer(ctx, &pb.StopContainerRequest{ContainerId: containerID})
	if err == nil {
		return nil
	}
	switch {
	case status.Code(err) == codes.Aborted:
		i.Logger.Debugf("%q is not found in containers", containerID)
		return errorx.ErrWrongCode
	default:
		i.Logger.Errorf(err, "StopContainer for containerID %q failed", containerID)
		return errorx.InternalErr(err)
	}
}
