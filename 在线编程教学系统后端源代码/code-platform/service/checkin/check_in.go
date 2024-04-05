package checkin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/rediskey"
	"code-platform/pkg/transactionx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/storage"

	redigo "github.com/gomodule/redigo/redis"
)

type CheckInService struct {
	Dao    *repository.Dao
	Logger *log.Logger
}

func NewCheckInService(dao *repository.Dao, logger *log.Logger) *CheckInService {
	return &CheckInService{
		Dao:    dao,
		Logger: logger,
	}
}

const (
	checkInRedisKeyPrefix = "sci-%d"
)

func (c *CheckInService) PrepareCheckIn(ctx context.Context, teacherID, courseID uint64, name string, seconds int) error {
	userIDs, err := model.QueryUserIDsInArrangeCourseByCourseID(ctx, c.Dao.Storage.RDB, courseID)
	if err != nil {
		c.Logger.Errorf(err, "query userIDs by courseID(%d) failed", courseID)
		return errorx.InternalErr(err)
	}

	task := func(ctx context.Context, rdbClient storage.RDBClient) error {
		now := time.Now()
		checkInRecord := &model.CheckInRecord{
			CourseID:  courseID,
			Name:      name,
			DeadLine:  now.Add(time.Duration(seconds) * time.Second),
			CreatedAt: now,
		}
		if err := checkInRecord.Insert(ctx, rdbClient); err != nil {
			c.Logger.Errorf(err, "insert checkInRecord %+v failed", checkInRecord)
			return errorx.InternalErr(err)
		}

		details := make([]*model.CheckInDetail, len(userIDs))
		for index, userID := range userIDs {
			details[index] = &model.CheckInDetail{
				RecordID:  checkInRecord.ID,
				UserID:    userID,
				IsCheckIn: false,
				CreatedAt: now,
				UpdatedAt: now,
			}
		}
		if err := model.BatchInsertCheckInDetails(ctx, rdbClient, details); err != nil {
			c.Logger.Errorf(err, "batch insert checkInDetails %+v failed", details)
			return errorx.InternalErr(err)
		}

		key := rediskey.NewkeyFormat(checkInRedisKeyPrefix, courseID).Pool(c.Dao.Storage.Pool())
		if _, err := key.SetEX(ctx, checkInRecord.ID, seconds); err != nil {
			c.Logger.Errorf(err, "SETEX for key %q and value %v failed", key.String(), checkInRecord.ID)
			return errorx.InternalErr(err)
		}
		return nil
	}

	return transactionx.DoTransaction(ctx, c.Dao.Storage, c.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (c *CheckInService) ListRecordsByCourseID(ctx context.Context, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		checkRecords        []*model.CheckInRecord
		total               int
		totalStudents       int
		recordIDToActualMap map[uint64]int
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInCheckInRecordByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCheckInRecordsByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountInCheckInRecordByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			checkRecords, err = model.QueryCheckInRecordsByCourseID(ctx, c.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCheckInRecordsByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCheckInRecordsByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			totalStudents, err = model.QueryTotalAmountOfArrangeCourseWithPassByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountOfArrangeCourseByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountOfArrangeCourseByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			recordIDToActualMap, err = model.QueryCheckInRecordIDToAmountMapByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCheckInRecordIDToAmountMapByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCheckInRecordIDToAmountMapByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	records := make([]*CheckInData, len(checkRecords))
	for index, record := range checkRecords {
		records[index] = &CheckInData{
			ID:        record.ID,
			CourseID:  record.CourseID,
			Name:      record.Name,
			Actual:    recordIDToActualMap[record.ID],
			Total:     totalStudents,
			CreatedAt: record.CreatedAt,
		}
	}

	return &PageResponse{
		PageInfo: &PageInfo{Total: total},
		Records:  records,
	}, nil
}

func (c *CheckInService) ListCheckInDetailsByRecordID(ctx context.Context, recordID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total int
		infos []*model.CheckInDetailWithUserData
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInCheckInDetailByCheckRecordID(ctx, c.Dao.Storage.RDB, recordID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountInCheckInDetailByCheckRecordID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountInCheckInDetailByCheckRecordID by recordID(%d) failed", recordID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			infos, err = model.QueryCheckInAboutUser(ctx, c.Dao.Storage.RDB, recordID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCheckInAboutUser is canceled")
				return err
			default:
				c.Logger.Errorf(err, "queryCheckInDataAboutUser by recordID(%d) failed", recordID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	records := make([]*CheckInWithUserData, len(infos))
	for index, info := range infos {
		records[index] = &CheckInWithUserData{
			UserID:          info.ID,
			Number:          info.Number,
			Name:            info.Name,
			Organization:    info.Organization,
			CheckinRecordID: info.RecordID,
			IsCheckIn:       info.IsCheckIn,
		}
	}

	return &PageResponse{
		PageInfo: &PageInfo{Total: total},
		Records:  records,
	}, nil
}

func (c *CheckInService) ExportCheckInRecordsCSV(ctx context.Context, courseID uint64) ([]byte, error) {
	var (
		recordNames        []string
		userIDToDetailsMap map[uint64][]*model.CheckInDetail
		users              []*model.User
	)
	tasks := []func() error{
		func() (err error) {
			recordNames, err = model.QueryCheckInRecordNamesByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			if err != nil {
				c.Logger.Errorf(err, "Query check in record names by courseID[%d] failed", courseID)
				return errorx.InternalErr(err)
			}
			/*
				fake error
				该子任务大概率是最快的，如果无数据，则直接取消其他子任务
			*/
			if len(recordNames) == 0 {
				return errorx.ErrIsNotFound
			}
			return nil
		},
		func() (err error) {
			userIDToDetailsMap, err = model.QueryUserIDToCheckInDetailsMapByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			if err != nil {
				c.Logger.Errorf(err, "QueryUserIDToCheckInDetailsMap by courseID[%d] failed", courseID)
				return errorx.InternalErr(err)
			}

			// fake error 同上
			if len(userIDToDetailsMap) == 0 {
				return errorx.ErrIsNotFound
			}

			userIDs := make([]uint64, 0, len(userIDToDetailsMap))
			for userID := range userIDToDetailsMap {
				userIDs = append(userIDs, userID)
			}

			users, err = model.QueryUsersByIDs(ctx, c.Dao.Storage.RDB, userIDs)
			if err != nil {
				c.Logger.Errorf(err, "QueryUsers by userIDs %v failed", userIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	buf := bytes.NewBufferString("\xEF\xBB\xBF")
	writer := csv.NewWriter(buf)
	headLine := make([]string, 0, len(recordNames)+3)
	headLine = append(headLine, "学号", "姓名")
	headLine = append(headLine, "出勤率")

	switch err := parallelx.Do(c.Logger, tasks...); err {
	case nil:
	case errorx.ErrIsNotFound:
		if err := writer.Write(headLine); err != nil {
			c.Logger.Errorf(err, "csv writer write data %+v failed")
			return nil, errorx.InternalErr(err)
		}
		writer.Flush()
		return buf.Bytes(), nil
	default:
		return nil, err
	}

	headLine = append(headLine, recordNames...)
	if err := writer.Write(headLine); err != nil {
		c.Logger.Errorf(err, "csv writer write data %+v failed")
		return nil, errorx.InternalErr(err)
	}

	rows := make([][]string, len(users))
	for index, user := range users {
		row := make([]string, 0, len(recordNames)+3)
		row = append(row, user.Number, user.Name)
		var count float64 = 0
		for _, v := range userIDToDetailsMap[user.ID] {
			if v.IsCheckIn {
				row = append(row, "√")
				count++
			} else {
				row = append(row, " ")
			}
		}
		row = append(row, fmt.Sprintf("%.2f", count/float64(len(recordNames))))
		rows[index] = row
	}

	if err := writer.WriteAll(rows); err != nil {
		c.Logger.Errorf(err, "csvWriter write failed")
		return nil, errorx.InternalErr(err)
	}

	writer.Flush()
	return buf.Bytes(), nil
}

func (c *CheckInService) StartCheckIn(ctx context.Context, courseID, userID uint64) error {
	key := rediskey.NewkeyFormat(checkInRedisKeyPrefix, courseID).Pool(c.Dao.Storage.Pool())
	value, err := key.Get(ctx)
	switch err {
	case nil:
	case redigo.ErrNil:
		// 已经过期或根本不存在
		c.Logger.Debugf("redisKey %q is time out or invalid", key.String())
		return errorx.ErrRedisKeyNil
	default:
		c.Logger.Errorf(err, "GET %q in redis failed", key.String())
		return errorx.InternalErr(err)
	}

	recordID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		c.Logger.Errorf(err, "parse string %q to uint64 failed", value)
		return errorx.InternalErr(err)
	}

	detail, err := model.QueryCheckInDetailByRecordIDAndUserID(ctx, c.Dao.Storage.RDB, recordID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("checkInDetails are not found by recordID(%d) and userID(%d)", recordID, userID)
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "QueryCheckInDetailByRecordIDAndUserID failed by recordID(%d) and userID(%d)", recordID, userID)
		return errorx.InternalErr(err)
	}

	// 防止不必要的再次签到
	if detail.IsCheckIn {
		c.Logger.Debugf("checkIn again for courseID(%d) and userID(%d)", courseID, userID)
		return nil
	}

	detail.IsCheckIn = true
	if err := detail.Update(ctx, c.Dao.Storage.RDB); err != nil {
		c.Logger.Errorf(err, "update checkInDetail %+v failed", detail)
		return errorx.InternalErr(err)
	}

	return nil
}

func (c *CheckInService) ListCheckInDetailsByUserIDAndCourseID(ctx context.Context, courseID, userID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total int
		infos []*model.CheckInRecordWithIsCheckInStatus
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountInCheckInRecordByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountInCheckInRecordByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountInCheckInRecordByCourseID failed by courseID(%d)", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			infos, err = model.QueryCheckInRecordWithIsCheckInStatusByCourseIDAndUserID(ctx, c.Dao.Storage.RDB, courseID, userID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCheckInRecordWithIsCheckInStatusByCourseIDAndUserID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCheckInRecordWithIsCheckInStatusByCourseIDAndUserID by courseID(%d) and userID(%d) failed", courseID, userID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	records := make([]*CheckInRecordWithDetailData, len(infos))
	for index, info := range infos {
		records[index] = &CheckInRecordWithDetailData{
			CheckinRecordID: info.ID,
			Name:            info.Name,
			IsCheckIn:       info.IsCheckIn,
			CreatedAt:       info.CreatedAt,
		}
	}
	return &PageResponse{
		PageInfo: &PageInfo{Total: total},
		Records:  records,
	}, nil
}

func (c *CheckInService) DeleteCheckInDataByRecordID(ctx context.Context, recordID, userID uint64) error {
	if err := model.DeleteCheckInRecordByID(ctx, c.Dao.Storage.RDB, recordID); err != nil {
		c.Logger.Errorf(err, "delete checkInRecord by recordID(%d) failed", recordID)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CheckInService) UpdateCheckInDetail(ctx context.Context, userID, recordID uint64, isCheckIn bool) error {

	detail, err := model.QueryCheckInDetailByRecordIDAndUserID(ctx, c.Dao.Storage.RDB, recordID, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("CheckInDetail in not found by recordID(%d) and userID(%d)", recordID, userID)
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "QueryCheckInDetailByRecordIDAndUserID by recordID(%d) and userID(%d) failed", recordID, userID)
		return errorx.InternalErr(err)
	}

	if detail.IsCheckIn == isCheckIn {
		return nil
	}

	detail.IsCheckIn = isCheckIn
	if err := detail.Update(ctx, c.Dao.Storage.RDB); err != nil {
		c.Logger.Errorf(err, "update checkInDetail for %+v failed", detail)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CheckInService) courseIDsToCoursesAndCheckInDetailsMap(
	ctx context.Context,
	userID uint64,
	courseIDs []uint64,
	records []*model.CheckInRecord,
) (
	coursesMap map[uint64]*model.Course,
	isCheckInMap map[uint64]bool,
	err error,
) {
	if len(courseIDs) == 0 || len(records) == 0 {
		return nil, nil, nil
	}

	tasks := []func() error{
		func() (err error) {
			coursesMap, err = model.QueryCourseMapsByIDs(ctx, c.Dao.Storage.RDB, courseIDs)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCourseMapsByIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCourseMapsByIDs by courseIDs %v failed", courseIDs)
				return err
			}
			return nil
		},
		func() (err error) {
			recordIDs := make([]uint64, len(records))
			for i, record := range records {
				recordIDs[i] = record.ID
			}

			isCheckInMap, err = model.QueryIsCheckInMapByUserIDAndRecordIDs(ctx, c.Dao.Storage.RDB, userID, recordIDs)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryIsCheckInMapByUserIDAndRecordIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCheckInDetailsMapByUserIDAndRecordIDs by userID %d and recordIDs %v", userID, recordIDs)
				return err
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, nil, err
	}

	return coursesMap, isCheckInMap, nil
}

func (c *CheckInService) ListRecentUserCheckIn(ctx context.Context, userID uint64) ([]*CheckInRecordPersonalData, error) {

	courseIDs, err := model.QueryCourseIDsInArrangeCourseByUserID(ctx, c.Dao.Storage.RDB, userID)
	if err != nil {
		c.Logger.Errorf(err, "QueryCourseIDsInArrangeCourseByUserID by userID(%d) failed", userID)
		return nil, errorx.InternalErr(err)
	}

	records, err := model.QueryCheckInRecordsByCourseIDsWithoutTimeout(ctx, c.Dao.Storage.RDB, courseIDs)
	if err != nil {
		c.Logger.Errorf(err, "QueryCheckInRecordsByCourseIDsWithoutTimeout by courseIDs %v failed", courseIDs)
		return nil, errorx.InternalErr(err)
	}

	coursesMap, isCheckInMap, err := c.courseIDsToCoursesAndCheckInDetailsMap(ctx, userID, courseIDs, records)
	if err != nil {
		return nil, errorx.InternalErr(err)
	}

	personalData := make([]*CheckInRecordPersonalData, 0, len(records))
	for _, record := range records {
		// 未截止但已签到的，直接跳过不返回
		if isFinish := isCheckInMap[record.ID]; isFinish {
			continue
		}

		personalData = append(personalData, &CheckInRecordPersonalData{
			ID:         record.ID,
			CourseID:   record.CourseID,
			Name:       record.Name,
			CourseName: coursesMap[record.CourseID].Name,
			CreatedAt:  record.CreatedAt,
			DeadLine:   record.DeadLine,
		})
	}

	return personalData[:len(personalData):len(personalData)], nil
}

func (c *CheckInService) ListUserCheckIn(ctx context.Context, userID uint64, offset, limit int) (*PageResponse, error) {

	courseIDs, err := model.QueryCourseIDsInArrangeCourseByUserID(ctx, c.Dao.Storage.RDB, userID)
	if err != nil {
		c.Logger.Errorf(err, "QueryCourseIDsInArrangeCourseByUserID by userID(%d) failed", userID)
		return nil, errorx.InternalErr(err)
	}

	var (
		total        int
		records      []*model.CheckInRecord
		coursesMap   map[uint64]*model.Course
		isCheckInMap map[uint64]bool
	)

	tasks := []func() error{
		func() (err error) {
			records, err = model.QueryCheckInRecordsByCourseIDs(ctx, c.Dao.Storage.RDB, courseIDs, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCheckInRecordsByCourseIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCheckInRecordsByCourseIDs by courseIDs %v failed", courseIDs)
				return errorx.InternalErr(err)
			}

			coursesMap, isCheckInMap, err = c.courseIDsToCoursesAndCheckInDetailsMap(ctx, userID, courseIDs, records)
			if err != nil {
				return errorx.InternalErr(err)
			}

			return nil
		},
		func() (err error) {
			total, err = model.QueryTotalAmountInCheckInRecordByCourseIDs(ctx, c.Dao.Storage.RDB, courseIDs)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountInCheckInRecordByCourseIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountInCheckInRecordByCourseIDs by courseIDs %v failed", courseIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	personalData := make([]*CheckInRecordPersonalData, len(records))
	for i, record := range records {
		isFinish := isCheckInMap[record.ID]
		personalData[i] = &CheckInRecordPersonalData{
			ID:         record.ID,
			Name:       record.Name,
			CourseName: coursesMap[record.CourseID].Name,
			CreatedAt:  record.CreatedAt,
			DeadLine:   record.DeadLine,
			IsFinish:   &isFinish,
		}
	}

	return &PageResponse{
		Records:  personalData,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CheckInService) GetCourseIDByRecordID(ctx context.Context, recordID uint64) (uint64, error) {
	courseID, err := model.QueryCourseIDByRecordID(ctx, c.Dao.Storage.RDB, recordID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return 0, errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "record is not found by id[%d]", recordID)
		return 0, errorx.InternalErr(err)
	}
	return courseID, nil
}
