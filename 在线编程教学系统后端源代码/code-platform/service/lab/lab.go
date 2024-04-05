package lab

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	idepb "code-platform/api/grpc/ide/pb"
	monacopb "code-platform/api/grpc/monaco/pb"
	"code-platform/api/grpc/plagiarismDetection/pb"
	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/strconvx"
	"code-platform/pkg/timex"
	"code-platform/pkg/transactionx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func NewPlagiarismDetectionClient() pb.PlagiarismDetectionClient {
	const port = "8086"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "localhost:"+port, grpc.WithInsecure(), grpc.WithBlock())

	switch err {
	case nil:
	case ctx.Err():
		log.Sub("grpc.PlagiarismDetection").Errorf(err, "dial to grpc server timeout")
		fallthrough
	default:
		panic(err)
	}
	return pb.NewPlagiarismDetectionClient(conn)
}

type LabService struct {
	Dao                       *repository.Dao
	Logger                    *log.Logger
	PlagiarismDetectionClient pb.PlagiarismDetectionClient
	IDEClient                 idepb.IDEServerServiceClient
	MonacoClient              monacopb.MonacoServerServiceClient
}

func NewLabService(
	dao *repository.Dao,
	logger *log.Logger,
	plagiarismDetectionClient pb.PlagiarismDetectionClient,
	ideClient idepb.IDEServerServiceClient,
	monacoClient monacopb.MonacoServerServiceClient,
) *LabService {
	return &LabService{
		Dao:                       dao,
		Logger:                    logger,
		PlagiarismDetectionClient: plagiarismDetectionClient,
		IDEClient:                 ideClient,
		MonacoClient:              monacoClient,
	}
}

func (l *LabService) InsertLab(ctx context.Context, courseID uint64, title, content, attachmentURL string, deadLine time.Time) error {
	task := func(ctx context.Context, tx storage.RDBClient) error {
		now := time.Now()
		lab := &model.Lab{
			CourseID:      courseID,
			Title:         title,
			Content:       content,
			AttachMentURL: attachmentURL,
			DeadLine: sql.NullTime{
				Time:  deadLine,
				Valid: !deadLine.IsZero(),
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := lab.Insert(ctx, tx); err != nil {
			l.Logger.Errorf(err, "insert lab %+v failed", lab)
			return errorx.InternalErr(err)
		}

		userIDs, err := model.QueryUserIDsInArrangeCourseByCourseID(ctx, tx, courseID)
		if err != nil {
			l.Logger.Errorf(err, "query userIDs in arrange course by courseID(%d) failed", courseID)
			return errorx.InternalErr(err)
		}

		labSubmits := make([]*model.LabSubmit, len(userIDs))
		for index, userID := range userIDs {
			labSubmits[index] = &model.LabSubmit{
				LabID:     lab.ID,
				UserID:    userID,
				IsFinish:  false,
				CreatedAt: now,
				UpdatedAt: now,
			}
		}

		if err := model.BatchInsertLabSubmits(ctx, tx, labSubmits); err != nil {
			l.Logger.Errorf(err, "batch insert lab submits %+v failed", labSubmits)
			return errorx.InternalErr(err)
		}
		return nil
	}

	return transactionx.DoTransaction(ctx, l.Dao.Storage, l.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (l *LabService) labInfosToRecordsResponse(labInfos []*model.LabInfoByUserIDAndCourseID) []*LabInfo {
	records := make([]*LabInfo, len(labInfos))
	for index, info := range labInfos {
		records[index] = &LabInfo{
			LabID:         info.ID,
			CourseID:      info.CourseID,
			CourseName:    info.CourseName,
			Title:         info.Title,
			Content:       info.Content,
			IsFinish:      info.IsFinish,
			ReportURL:     info.ReportURL,
			Score:         info.Score.Int32,
			Comment:       info.Comment,
			AttachmentURL: info.AttachMentURL,
			DeadLine:      info.DeadLine.Time,
			CreatedAt:     info.CreatedAt,
			UpdatedAt:     info.UpdatedAt,
		}
	}
	return records
}

func (l *LabService) labsModelToService(labs []*model.Lab) []*Lab {
	results := make([]*Lab, len(labs))
	for i, lab := range labs {
		results[i] = &Lab{
			ID:            lab.ID,
			CourseID:      lab.CourseID,
			Title:         lab.Title,
			Content:       lab.Content,
			AttachMentURL: lab.AttachMentURL,
			DeadLine:      lab.DeadLine.Time,
			CreatedAt:     lab.CreatedAt,
		}
	}
	return results
}

func (l *LabService) ListLabsByUserIDAndCourseID(ctx context.Context, userID, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total    int
		course   *model.Course
		labInfos []*model.LabInfoByUserIDAndCourseID
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInLabByCourseID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryTotalAmountInLabByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryTotalAmountInLabByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			course, err = model.QueryCourseByID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryCourseByID is canceled")
				return err
			case sql.ErrNoRows:
				l.Logger.Debugf("course is not found by id(%d)", courseID)
				return errorx.ErrIsNotFound
			default:
				l.Logger.Errorf(err, "query course by id(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			labInfos, err = model.QueryLabSubmitInfosByUserIDAndCourseID(ctx, l.Dao.Storage.RDB, userID, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryLabSubmitInfosByUserIDAndCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryLabSubmitInfosByUserIDAndCourseID by userID(%d) and courseID(%d) failed", userID, courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}

	// 填充课程名字
	for index := range labInfos {
		labInfos[index].CourseName = course.Name
	}

	return &PageResponse{
		Records:  l.labInfosToRecordsResponse(labInfos),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (l *LabService) UpdateLab(ctx context.Context, labID uint64, title, content, attachmentURL string, deadLine time.Time) error {
	lab, err := model.QueryLabByID(ctx, l.Dao.Storage.RDB, labID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab is not found by id(%d)", labID)
		return errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "query lab by id(%d) failed", labID)
		return errorx.InternalErr(err)
	}

	lab.AttachMentURL = attachmentURL
	lab.Content = content
	lab.DeadLine = sql.NullTime{
		Time:  deadLine,
		Valid: !deadLine.IsZero(),
	}
	lab.Title = title
	if err := lab.Update(ctx, l.Dao.Storage.RDB); err != nil {
		l.Logger.Errorf(err, "update for lab %+v failed", lab)
		return errorx.InternalErr(err)
	}
	return nil
}

func (l *LabService) GetLab(ctx context.Context, labID uint64) (*LabInfo, error) {
	lab, err := model.QueryLabByID(ctx, l.Dao.Storage.RDB, labID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab is not found by id(%d)", labID)
		return nil, errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "query lab by id(%d) failed", labID)
		return nil, errorx.InternalErr(err)
	}
	return &LabInfo{
		LabID:         lab.ID,
		CourseID:      lab.CourseID,
		Title:         lab.Title,
		Content:       lab.Content,
		AttachmentURL: lab.AttachMentURL,
		DeadLine:      lab.DeadLine.Time,
		CreatedAt:     lab.CreatedAt,
		UpdatedAt:     lab.UpdatedAt,
	}, nil
}

func (l *LabService) DeleteLab(ctx context.Context, labID uint64) error {
	task := func(ctx context.Context, tx storage.RDBClient) error {
		if err := model.DeleteLabByID(ctx, tx, labID); err != nil {
			l.Logger.Errorf(err, "delete lab by id(%d) failed", labID)
			return errorx.InternalErr(err)
		}

		if err := model.DeleteLabSubmitsByLabID(ctx, tx, labID); err != nil {
			l.Logger.Errorf(err, "delete lab submits by labID(%d) failed", labID)
			return errorx.InternalErr(err)
		}
		return nil
	}
	return transactionx.DoTransaction(ctx, l.Dao.Storage, l.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (l *LabService) ListLabsByUserID(ctx context.Context, userID uint64, offset, limit int) (*PageResponse, error) {
	courseIDs, err := model.QueryCourseIDsInArrangeCourseByUserID(ctx, l.Dao.Storage.RDB, userID)
	if err != nil {
		l.Logger.Errorf(err, "QueryCourseIDsInArrangeCourseByUserID by userID(%d) failed", userID)
		return nil, errorx.InternalErr(err)
	}

	var (
		total      int
		labInfos   []*model.LabInfoByUserIDAndCourseID
		coursesMap map[uint64]*model.Course
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInLabByCourseIDs(ctx, l.Dao.Storage.RDB, courseIDs)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryTotalAmountInLabByCourseIDs is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryTotalAmountInLabByCourseIDs by courseIDs %v failed", courseIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			labInfos, err = model.QueryLabSubmitInfosByUserIDAndCourseIDs(ctx, l.Dao.Storage.RDB, userID, courseIDs, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryLabSubmitInfosByUserIDAndCourseIDs is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryLabSubmitInfosByUserIDAndCourseIDs by userID(%d) and courseIDs %v failed", userID, courseIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			coursesMap, err = model.QueryCourseMapsByIDs(ctx, l.Dao.Storage.RDB, courseIDs)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryCourseMapsByIDs is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryCourseMapsByIDs by courseIDs %v failed", courseIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}

	// 填充课程名字
	for index := range labInfos {
		labInfos[index].CourseName = coursesMap[labInfos[index].CourseID].Name
	}

	return &PageResponse{
		Records:  l.labInfosToRecordsResponse(labInfos),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (l *LabService) ListLabScoreByUserIDAndCourseID(ctx context.Context, userID, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total     int
		labScores []*LabScore
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInLabByCourseID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryTotalAmountInLabByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryTotalAmountInLabByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			infos, err := model.QueryLabSubmitInfosByUserIDAndCourseID(ctx, l.Dao.Storage.RDB, userID, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryLabSubmitInfosByUserIDAndCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryLabSubmitInfosByUserIDAndCourseID by userID(%d) and courseID(%d) failed", userID, courseID)
				return errorx.InternalErr(err)
			}

			labScores = make([]*LabScore, len(infos))
			for index, info := range infos {
				labScores[index] = &LabScore{
					LabID:     info.Lab.ID,
					Title:     info.Title,
					Score:     info.Score.Int32,
					CreatedAt: info.Lab.CreatedAt,
				}
			}
			return nil
		},
	}

	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  labScores,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (l *LabService) UpdateReport(ctx context.Context, labID, userID uint64, reportURL string) error {
	labSubmit, err := model.QueryLabSubmitByLabIDAndUserID(ctx, l.Dao.Storage.RDB, labID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab submit is not found by labID(%d) and userID(%d)", labID, userID)
		return errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "Query LabSubmit By LabID(%d) and UserID(%d) failed", labID, userID)
		return errorx.InternalErr(err)
	}

	labSubmit.ReportURL = reportURL
	if err := labSubmit.Update(ctx, l.Dao.Storage.RDB); err != nil {
		l.Logger.Errorf(err, "update for lab submit %+v failed", labSubmit)
		return errorx.InternalErr(err)
	}

	return nil
}

func (l *LabService) InsertCodeFinish(ctx context.Context, labID, userID uint64, isFinish bool) error {
	labSubmit, err := model.QueryLabSubmitByLabIDAndUserID(ctx, l.Dao.Storage.RDB, labID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab submit is not found by labID(%d) and userID(%d)", labID, userID)
		return errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "Query LabSubmit By LabID(%d) and UserID(%d) failed", labID, userID)
		return errorx.InternalErr(err)
	}

	if labSubmit.IsFinish == isFinish {
		return nil
	}

	labSubmit.IsFinish = isFinish
	if err := labSubmit.Update(ctx, l.Dao.Storage.RDB); err != nil {
		l.Logger.Errorf(err, "update for lab submit %+v failed", labSubmit)
		return errorx.InternalErr(err)
	}
	return nil
}

func (l *LabService) GetCommentByUserIDAndLabID(ctx context.Context, userID, labID uint64) (string, error) {
	labSubmit, err := model.QueryLabSubmitByLabIDAndUserID(ctx, l.Dao.Storage.RDB, labID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab submit is not found by labID(%d) and userID(%d)", labID, userID)
		return "", errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "Query LabSubmit By LabID(%d) and UserID(%d) failed", labID, userID)
		return "", errorx.InternalErr(err)
	}

	return labSubmit.Comment, nil
}

func (l *LabService) GetReportURL(ctx context.Context, userID, labID uint64) (string, error) {
	labSubmit, err := model.QueryLabSubmitByLabIDAndUserID(ctx, l.Dao.Storage.RDB, labID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab submit is not found by labID(%d) and userID(%d)", labID, userID)
		return "", errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "Query LabSubmit By LabID(%d) and UserID(%d) failed", labID, userID)
		return "", errorx.InternalErr(err)
	}

	return labSubmit.ReportURL, nil
}

func (l *LabService) UpdateScore(ctx context.Context, userID, labID uint64, score int32) error {
	labSubmit, err := model.QueryLabSubmitByLabIDAndUserID(ctx, l.Dao.Storage.RDB, labID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab submit is not found by labID(%d) and userID(%d)", labID, userID)
		return errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "Query LabSubmit By LabID(%d) and UserID(%d) failed", labID, userID)
		return errorx.InternalErr(err)
	}

	labSubmit.Score = sql.NullInt32{Valid: true, Int32: score}
	if err := labSubmit.Update(ctx, l.Dao.Storage.RDB); err != nil {
		l.Logger.Errorf(err, "update for lab submit %+v failed", labSubmit)
		return errorx.InternalErr(err)
	}
	return nil
}

func (l *LabService) UpdateComment(ctx context.Context, userID, labID uint64, comment string) error {

	labSubmit, err := model.QueryLabSubmitByLabIDAndUserID(ctx, l.Dao.Storage.RDB, labID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab submit is not found by labID(%d) and userID(%d)", labID, userID)
		return errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "Query LabSubmit By LabID(%d) and UserID(%d) failed", labID, userID)
		return errorx.InternalErr(err)
	}

	labSubmit.Comment = comment
	if err := labSubmit.Update(ctx, l.Dao.Storage.RDB); err != nil {
		l.Logger.Errorf(err, "update for lab submit %+v failed", labSubmit)
		return errorx.InternalErr(err)
	}
	return nil
}

func (l *LabService) ListLabsByCourseID(ctx context.Context, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total int
		labs  []*model.Lab
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInLabByCourseID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryTotalAmountInLabByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryTotalAmountInLabByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			labs, err = model.QueryLabsByCourseID(ctx, l.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryLabsByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "query labs by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  l.labsModelToService(labs),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (l *LabService) ListLabSubmitsByLabID(ctx context.Context, labID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total int
		infos []*model.LabSubmitInfoByLabID
	)

	tasks := []func() error{
		func() (err error) {
			courseID, err := model.QueryCourseIDByLabID(ctx, l.Dao.Storage.RDB, labID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryCourseIDByLabID is canceled")
				return err
			case sql.ErrNoRows:
				l.Logger.Debugf("courseID is not found by labID[%d]", labID)
				return errorx.ErrIsNotFound
			default:
				l.Logger.Errorf(err, "query lab by labID(%d) failed", labID)
				return errorx.InternalErr(err)
			}

			total, err = model.QueryTotalAmountOfArrangeCourseWithPassByCourseID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryTotalAmountOfArrangeCourseByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryTotalAmountOfArrangeCourseByCourseID by courseID[%d] failed", courseID)
				return errorx.InternalErr(err)
			}

			return nil
		},
		func() (err error) {
			infos, err = model.QueryLabSubmitInfosByLabID(ctx, l.Dao.Storage.RDB, labID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryLabSubmitInfosByLabID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryLabSubmitInfosByLabID by labID(%d) failed", labID)
				return errorx.InternalErr(err)
			}

			userIDs := make([]uint64, len(infos))
			for index, info := range infos {
				userIDs[index] = info.UserID
			}
			codingTimesMap, err := model.QueryCodingTimesMapByLabIDAndUserIDs(ctx, l.Dao.Storage.RDB, labID, userIDs)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryCodingTimesMapByLabIDAndUserIDs is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryCodingTimesMapByLabIDAndUserIDs by labID(%d) and userIDs %v failed", labID, userIDs)
				return errorx.InternalErr(err)
			}

			if len(codingTimesMap) != 0 {
				for index := range infos {
					infos[index].CodingTime = codingTimesMap[infos[index].UserID]
				}
			}
			return nil
		},
	}

	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}

	records := make([]*LabCodingTimeData, len(infos))
	for index, info := range infos {
		records[index] = &LabCodingTimeData{
			LabSubmitID: info.ID,
			LabID:       info.LabID,
			UserID:      info.UserID,
			UserName:    info.Name,
			Number:      info.Number,
			ReportURL:   info.ReportURL,
			IsFinish:    info.IsFinish,
			Score:       info.Score.Int32,
			Comment:     info.Comment,
			CodingTime:  info.CodingTime,
			CreatedTime: info.CreatedAt,
			UpdatedTime: info.UpdatedAt,
		}
	}

	return &PageResponse{
		Records:  records,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (l *LabService) QuickCheckCode(ctx context.Context, labID, studentID uint64) (*idepb.QuickViewCodeResponse_FileNode, error) {
	resp, err := l.IDEClient.QuickViewCode(ctx, &idepb.QuickViewCodeRequest{
		LabId:  labID,
		UserId: studentID,
	})
	if err != nil {
		l.Logger.Errorf(err, "QuickCheckCode for userID[%d] and studentID[%d] failed", labID, studentID)
		return nil, errorx.InternalErr(err)
	}

	return resp.RootNode, nil
}

func (l *LabService) PlagiarismCheck(ctx context.Context, labID uint64) ([]*PlagiarismCheckResponse, error) {
	courseID, err := model.QueryCourseIDByLabID(ctx, l.Dao.Storage.RDB, labID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab is not found by id(%d) failed", labID)
		return nil, errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "query lab by labID(%d) failed", labID)
		return nil, errorx.InternalErr(err)
	}

	var (
		resp     []*PlagiarismCheckResponse
		usersMap map[uint64]*model.User
	)

	tasks := []func() error{
		func() (err error) {
			course, err := model.QueryCourseByID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryCourseByID is canceled")
				return err
			case sql.ErrNoRows:
				l.Logger.Debugf("course is not found by courseID(%d) failed", courseID)
				return errorx.ErrIsNotFound
			default:
				l.Logger.Errorf(err, "query course by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			var lan pb.Language
			switch course.Language {
			case 0:
				lan = pb.Language_python3
			case 1:
				lan = pb.Language_cpp
			default:
				lan = pb.Language_java
			}

			jplagResp, err := l.PlagiarismDetectionClient.DuplicateCheck(ctx, &pb.DuplicateCheckRequest{
				LabID: labID,
				Lan:   lan,
			})
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("DuplicateCheck is canceled")
				return err
			default:
				// 不合理代码，无合法提交
				switch status.Code(err) {
				case codes.DataLoss:
					l.Logger.Debugf("DuplicateCheck for language[%d] and labID[%d] failed for error %s", lan, labID, err.Error())
					return errorx.ErrWrongCode
				case codes.NotFound:
					l.Logger.Debugf("DuplicateCheck for labID[%d] is not found", labID)
					return errorx.ErrIsNotFound
				}
				l.Logger.Errorf(err, "plagiarismCheck for labID(%d) failed", labID)
				return errorx.InternalErr(err)
			}

			if jplagResp.GetComparision() == nil {
				return nil
			}

			timeStampStr := jplagResp.GetTimeStamp()
			data, err := proto.Marshal(jplagResp.GetComparision())
			if err != nil {
				l.Logger.Errorf(err, "marshal for resp %v failed", jplagResp)
				return errorx.InternalErr(err)
			}
			timeStamp, err := strconv.ParseInt(timeStampStr, 10, 64)
			if err != nil {
				l.Logger.Errorf(err, "parse to int64 for %s failed", timeStampStr)
				return errorx.InternalErr(err)
			}

			detectionReport := &model.DetectionReport{
				LabID:     labID,
				Data:      data,
				CreatedAt: time.UnixMilli(timeStamp),
			}

			if err := detectionReport.Insert(ctx, l.Dao.Storage.RDB); err != nil {
				l.Logger.Errorf(err, "insert detection_report %+v failed")
				return errorx.InternalErr(err)
			}

			resp = comparisionToPlagiarismCheckResponse(jplagResp.GetComparision(), timeStampStr)
			return nil
		},
		func() (err error) {
			userIDs, err := model.QueryUserIDsInArrangeCourseByCourseID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryUserIDsInArrangeCourseByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryUserIDsInArrangeCourseByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}

			usersMap, err = model.QueryUserMapByIDs(ctx, l.Dao.Storage.RDB, userIDs)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryUserMapByIDs is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryUserMapByIDs by userIDs(%v) failed", userIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}
	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}
	return boxForPlagiarismCheckResponse(resp, usersMap), nil
}

func comparisionToPlagiarismCheckResponse(
	comparision *pb.DuplicateCheckResponse_DuplicateCheckResponseValue,
	timeStampStr string,
) []*PlagiarismCheckResponse {
	resp := make([]*PlagiarismCheckResponse, len(comparision.GetComparisions()))
	b := &strings.Builder{}
	const connectSymbol = "?ts="
	for i, v := range comparision.Comparisions {
		b.Reset()
		b.Grow(len(v.GetHtmlFileName()) + len(timeStampStr) + len(connectSymbol))
		b.WriteString(v.GetHtmlFileName())
		b.WriteString(connectSymbol)
		b.WriteString(timeStampStr)

		resp[i] = &PlagiarismCheckResponse{
			UserID1:    v.GetUserId(),
			UserID2:    v.GetAnotherUserId(),
			Similarity: fmt.Sprintf("%.2f", float64(v.GetSimilarity()/100)),
			URL:        b.String(),
		}
	}
	return resp
}

func boxForPlagiarismCheckResponse(resp []*PlagiarismCheckResponse, usersMap map[uint64]*model.User) []*PlagiarismCheckResponse {
	for _, v := range resp {
		user1 := usersMap[v.UserID1]
		user2 := usersMap[v.UserID2]
		if user1 != nil {
			v.Num1 = user1.Number
			v.RealName1 = user1.Name
		}
		if user2 != nil {
			v.Num2 = user2.Number
			v.RealName2 = user2.Name
		}
	}
	return resp
}

func (l *LabService) ClickURL(ctx context.Context, labID uint64, dirName string, fileName string) ([]byte, error) {
	resp, err := l.PlagiarismDetectionClient.ViewReport(ctx, &pb.ViewReportRequest{
		LabId:        labID,
		TimeStamp:    dirName,
		HtmlFileName: fileName,
	})

	switch status.Code(err) {
	case codes.OK:
	case codes.NotFound:
		l.Logger.Debugf("htmlFilePath by labID[%d] 、timeStamp[%s] and fileName[%s] is not found", labID, dirName, fileName)
		return nil, errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "get report by labID[%d] 、timeStamp[%s] and fileName[%s] failed", labID, dirName, fileName)
		return nil, errorx.InternalErr(err)
	}

	return strconvx.StringToBytes(resp.HtmlFileContent), nil
}

func (l *LabService) GetCourseIDByLabID(ctx context.Context, labID uint64) (uint64, error) {
	courseID, err := model.QueryCourseIDByLabID(ctx, l.Dao.Storage.RDB, labID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return 0, errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "Failed to GetCourseID by lab_id[%d]", labID)
		return 0, errorx.InternalErr(err)
	}
	return courseID, nil
}

func (l *LabService) ListDetectionReportsByLabID(ctx context.Context, labID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total            int
		detectionReports []*model.DetectionReportIDWithCreatedAt
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountDetectionReportByLabID(ctx, l.Dao.Storage.RDB, labID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryTotalAmountDetectionReportByLabID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryTotalAmountDetectionReport by labID[%d] failed", labID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			detectionReports, err = model.QueryDetectionReportsByLabID(ctx, l.Dao.Storage.RDB, labID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryDetectionReportsByLabID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryDetectionReport by labID[%d] failed", labID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}

	resp := make([]*DetectionReportResponse, len(detectionReports))
	for i, detectionReport := range detectionReports {
		resp[i] = &DetectionReportResponse{
			ID:        detectionReport.ID,
			CreatedAt: detectionReport.CreatedAt.In(timex.ShanghaiLocation).Format("2006-01-02 15:04:05"),
		}
	}

	return &PageResponse{
		Records:  resp,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (l *LabService) ViewPerviousDetection(ctx context.Context, detectionReportID, teacherID uint64, host string) ([]*PlagiarismCheckResponse, error) {
	detectionReport, err := model.QueryDetectionReportByID(ctx, l.Dao.Storage.RDB, detectionReportID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return nil, errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "QueryDetectionReportData by ID[%d] failed", detectionReportID)
		return nil, errorx.InternalErr(err)
	}
	var v pb.DuplicateCheckResponse_DuplicateCheckResponseValue
	if err := proto.Unmarshal(detectionReport.Data, &v); err != nil {
		l.Logger.Errorf(err, "unmarshal for %v failed", detectionReport.Data)
		return nil, errorx.InternalErr(err)
	}

	resp := comparisionToPlagiarismCheckResponse(&v, strconv.FormatInt(detectionReport.CreatedAt.UnixMilli(), 10))
	courseID, err := model.QueryCourseIDByLabID(ctx, l.Dao.Storage.RDB, detectionReport.LabID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		l.Logger.Debugf("lab is not found by id(%d) failed", detectionReport.LabID)
		return nil, errorx.ErrIsNotFound
	default:
		l.Logger.Errorf(err, "query lab by labID(%d) failed", detectionReport.LabID)
		return nil, errorx.InternalErr(err)
	}

	var usersMap map[uint64]*model.User
	tasks := []func() error{
		func() (err error) {
			teacherIDQueried, err := model.QueryCourseTeacherIDByCourseID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryCourseTeacherIDByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryCourseTeacherID by courseID[%d] failed", courseID)
				return errorx.InternalErr(err)
			}

			if teacherIDQueried != teacherID {
				return errorx.ErrFailToAuth
			}
			return nil
		},
		func() (err error) {
			userIDs, err := model.QueryUserIDsInArrangeCourseByCourseID(ctx, l.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryUserIDsInArrangeCourseByCourseID is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryUserIDsInArrangeCourseByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			usersMap, err = model.QueryUserMapByIDs(ctx, l.Dao.Storage.RDB, userIDs)
			switch err {
			case nil:
			case context.Canceled:
				l.Logger.Debug("QueryUserMapByIDs is canceled")
				return err
			default:
				l.Logger.Errorf(err, "QueryUserMapByIDs by userIDs(%v) failed", userIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(l.Logger, tasks...); err != nil {
		return nil, err
	}
	resp = boxForPlagiarismCheckResponse(resp, usersMap)
	baseURL := "http://" + host + fmt.Sprintf("/web/lab/summit/plagiarism/%d/", detectionReport.LabID)
	for _, v := range resp {
		v.URL = baseURL + v.URL
	}
	return resp, nil
}
