package monitor

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"code-platform/api/grpc/ide/pb"
	"code-platform/log"
	"code-platform/monitor"
	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/rediskey"
	"code-platform/pkg/slicex"
	"code-platform/pkg/timex"
	"code-platform/pkg/transactionx"
	"code-platform/repository/rdb/model"
	"code-platform/service/ide/define"
	"code-platform/storage"

	"google.golang.org/protobuf/proto"
)

var heartBeatTagRegexp = regexp.MustCompile("^" + define.HeartBeatTagFormatPrefixForStudent + `(\d+):(\d+)$`)

func HeartBeatSweaping(storage *storage.Storage, logger *log.Logger, ideClient pb.IDEServerServiceClient) {
	parentCtx := context.TODO()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				logger.Errorf(err, "heart beat sweaper panic")
			}
			// 5分钟后恢复重试
			time.Sleep(time.Minute * 5)
			HeartBeatSweaping(storage, logger, ideClient)
		}()

		for {
			// 定义最大超时时间200秒
			startTime := time.Now()
			ctx, cancel := context.WithTimeout(parentCtx, 200*time.Second)
			resp, err := ideClient.GetContainerNames(ctx, &pb.Empty{})
			if err != nil {
				logger.Error(err, "GetContainerNames failed")
				cancel()
				break
				// 等待下一次再重试
			}
			keys := make([]interface{}, len(resp.Infos))
			containerNames := make([]string, len(resp.Infos))
			labIDs := make([]uint64, len(resp.Infos))
			for index, info := range resp.Infos {
				// not teacher
				if info.TeacherInfo == nil {
					keys[index] = fmt.Sprintf(define.HeartBeatTagFormatForStudent, info.LabId, info.StudentId)
					containerNames[index] = define.GetContainerNameForStudent(info.LabId, info.StudentId)
					labIDs[index] = info.LabId
				} else {
					keys[index] = fmt.Sprintf(define.HeartBeatTagFormatForTeacher, info.LabId, info.StudentId, info.TeacherInfo.TeacherId)
					containerNames[index] = define.GetContainerNameForTeacher(info.LabId, info.StudentId, info.TeacherInfo.TeacherId)
				}
			}
			// 检查key与清扫容器任务
			if len(keys) != 0 {
				newSweaterTask(ctx, storage, logger, ideClient, keys, containerNames, slicex.DistinctUint64Slice(labIDs))
			}
			cancel()
			// 间隔周期为300秒
			monitor.IDESweaterCollector.Set(float64(time.Since(startTime) / time.Microsecond))
			time.Sleep(time.Second * 300)
		}
	}()
}

func calculateCodingTime(createdAt, endAt time.Time, labID, studentID uint64) []*model.CodingTime {
	createdAt = createdAt.In(timex.ShanghaiLocation)
	endAt = endAt.In(timex.ShanghaiLocation)
	var codingTimes []*model.CodingTime
	if endAt.Day() == createdAt.Day() {
		codingTimes = append(codingTimes, &model.CodingTime{
			LabID:         labID,
			UserID:        studentID,
			Duration:      uint32(endAt.Sub(createdAt).Minutes()),
			CreatedAt:     createdAt,
			CreatedAtDate: timex.StartOfDay(createdAt),
		})
	} else {
		// 不在同一天
		// createdAt 到 当天晚上十二点
		codingTimes = append(codingTimes, &model.CodingTime{
			LabID:         labID,
			UserID:        studentID,
			Duration:      uint32(timex.EndOfDay(createdAt).Sub(createdAt).Minutes()),
			CreatedAt:     createdAt,
			CreatedAtDate: timex.StartOfDay(createdAt),
		})

		for t := createdAt; t.Day() != endAt.Day(); t = t.Add(24 * time.Hour) {
			c := timex.StartOfDay(t)
			codingTimes = append(codingTimes, &model.CodingTime{
				LabID:         labID,
				UserID:        studentID,
				Duration:      24 * 60,
				CreatedAt:     c,
				CreatedAtDate: c,
			})
		}

		// 当日凌晨 12 点到 endAt
		s := timex.StartOfDay(endAt)
		codingTimes = append(codingTimes, &model.CodingTime{
			LabID:         labID,
			UserID:        studentID,
			Duration:      uint32(endAt.Sub(s).Minutes()),
			CreatedAt:     s,
			CreatedAtDate: s,
		})
	}
	return codingTimes
}

func newSweaterTask(
	ctx context.Context,
	st *storage.Storage,
	logger *log.Logger,
	ideClient pb.IDEServerServiceClient,
	keys []interface{},
	containerNames []string,
	labIDs []uint64,
) {
	var (
		resp             [][]byte
		labIsDeadlineMap map[uint64]time.Time
	)
	tasks := []func() error{
		func() (err error) {
			key := rediskey.NewEmptyKey().Pool(st.Pool())
			resp, err = key.MGet(ctx, keys...)
			switch err {
			case nil:
			case context.Canceled:
				logger.Debug("MGet keys is canceled")
				return err
			default:
				logger.Errorf(err, "mget keys %v failed", keys)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			labIsDeadlineMap, err = model.QueryLabIDToDeadlineMapAfterDeadline(ctx, st.RDB, labIDs)
			switch err {
			case nil:
			case context.Canceled:
				logger.Debug("QueryLabIsDeadlineMap is canceled")
				return err
			default:
				logger.Errorf(err, "QueryLabIsDeadlineMap by labIDs %v failed", labIDs)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}
	if err := parallelx.Do(logger, tasks...); err != nil {
		return
	}

	containersNeedToStop := make([]string, 0, len(containerNames))
	codingTimes := make([]*model.CodingTime, 0, len(containerNames))
	keysNeedToDel := make([]interface{}, 0, len(containerNames))

	ignoreMap := make(map[int]struct{})
	for index, v := range resp {
		// key 已被删除
		if len(v) == 0 {
			ignoreMap[index] = struct{}{}
			containersNeedToStop = append(containersNeedToStop, containerNames[index])
			continue
		}
		// 教师key，无法 unmarshal，故直接忽略
		if bytes.Equal(v, []byte("0")) {
			ignoreMap[index] = struct{}{}
			continue
		}

		key := keys[index].(string)
		IDs := heartBeatTagRegexp.FindStringSubmatch(key)
		labID, err := strconv.ParseUint(IDs[1], 10, 64)
		// 避免后续数据不一致性，先直接退出返回
		if err != nil {
			logger.Errorf(err, "parse uint for %v failed", IDs[1])
			return
		}

		deadline, ok := labIsDeadlineMap[labID]
		if !ok {
			continue
		}

		var stat pb.HeartBeatStat
		if err := proto.Unmarshal(v, &stat); err != nil {
			logger.Errorf(err, "proto unmarshal data %v for heartbeatstat failed", v)
			return
		}
		createdAt := time.Unix(stat.CreatedAt, 0)
		if deadline.Before(createdAt) {
			continue
		}

		// 实验过期但IDE仍存在，则可认为该容器应被销毁
		containersNeedToStop = append(containersNeedToStop, containerNames[index])
		ignoreMap[index] = struct{}{}

		keysNeedToDel = append(keysNeedToDel, keys[index])

		// 处理 coding_time 数据
		if duration := uint32(deadline.Sub(createdAt).Minutes()); duration != 0 {
			studentID, err := strconv.ParseUint(IDs[2], 10, 64)
			if err != nil {
				logger.Errorf(err, "parse uint for %v failed", IDs[2])
				return
			}
			codingTimes = append(codingTimes, calculateCodingTime(createdAt, deadline, labID, studentID)...)
		}
	}

	for index, v := range resp {
		if _, ok := ignoreMap[index]; ok {
			continue
		}
		var stat pb.HeartBeatStat
		if err := proto.Unmarshal(v, &stat); err != nil {
			logger.Errorf(err, "proto unmarshal data %v for heartbeatstat failed", v)
			continue
		}
		if time.Since(time.Unix(stat.LastVisitedAt, 0)) > define.HeartBeatDuration {
			createdAt := time.Unix(stat.CreatedAt, 0)
			lastVisitedAt := time.Unix(stat.LastVisitedAt, 0)
			if duration := uint32(lastVisitedAt.Sub(createdAt).Minutes()); duration != 0 {
				key := keys[index].(string)
				IDs := heartBeatTagRegexp.FindStringSubmatch(key)
				labID, err := strconv.ParseUint(IDs[1], 10, 64)
				if err != nil {
					logger.Errorf(err, "parse uint for %v failed", IDs[1])
					continue
				}
				studentID, err := strconv.ParseUint(IDs[2], 10, 64)
				if err != nil {
					logger.Errorf(err, "parse uint for %v failed", IDs[2])
					continue
				}
				codingTimes = append(codingTimes, calculateCodingTime(createdAt, lastVisitedAt, labID, studentID)...)
			}
			containersNeedToStop = append(containersNeedToStop, containerNames[index])
			keysNeedToDel = append(keysNeedToDel, keys[index])
		}
	}

	if len(containersNeedToStop) == 0 {
		return
	}

	transactionx.DoTransaction(ctx, st, logger, func(ctx context.Context, tx storage.RDBClient) (err error) {
		if err := model.BatchInsertCodingTimes(ctx, tx, codingTimes); err != nil {
			logger.Errorf(err, "batch insert coding_time %+v failed", codingTimes)
			return err
		}
		// 关闭指定docker容器
		if _, err = ideClient.RemoveContainer(ctx, &pb.RemoveContainerRequest{ContainerNames: containersNeedToStop}); err != nil {
			logger.Errorf(err, "remove container %v failed", containersNeedToStop)
			return err
		}
		if len(keysNeedToDel) != 0 {
			emptyKey := rediskey.NewEmptyKey().Pool(st.Pool())
			if _, err := emptyKey.Del(ctx, keysNeedToDel...); err != nil {
				logger.Errorf(err, "delete redis key %v failed", keysNeedToDel)
				return err
			}
		}
		return nil
	},
		&sql.TxOptions{Isolation: sql.LevelReadCommitted},
	)
}
