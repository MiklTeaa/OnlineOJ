package ide

import (
	"context"
	"sync"
	"time"

	"code-platform/api/grpc/ide/pb"
	"code-platform/pkg/errorx"
	"code-platform/pkg/rediskey"
	"code-platform/service/ide/define"

	redigo "github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
)

var (
	heartBeatPoolForStudent = &sync.Pool{
		New: func() interface{} {
			return rediskey.Newkey("")
		},
	}

	heartBeatPoolForTeacher = &sync.Pool{
		New: func() interface{} {
			return rediskey.Newkey("")
		},
	}
)

func (i *IDEService) HeartBeatForStudent(ctx context.Context, labID, studentID uint64) error {
	key := heartBeatPoolForStudent.Get().(*rediskey.EntityKey).
		Pool(i.Dao.Storage.Pool()).
		Replace(define.HeartBeatTagFormatForStudent, labID, studentID)
	defer func() {
		key.Clear()
		heartBeatPoolForStudent.Put(key)
	}()

	value, err := i.refreshHeartBeatStat(ctx, key)
	if err != nil {
		return err
	}
	if _, err := key.Set(ctx, value); err != nil {
		i.Logger.Errorf(err, "set for key %q with value %v failed", key.String(), value)
		return errorx.InternalErr(err)
	}
	return nil
}

func (i *IDEService) HeartBeatForTeacher(ctx context.Context, labID, studentID, teacherID uint64) error {
	key := heartBeatPoolForTeacher.Get().(*rediskey.EntityKey).
		Pool(i.Dao.Storage.Pool()).
		Replace(define.HeartBeatTagFormatForTeacher, labID, studentID, teacherID)
	defer func() {
		key.Clear()
		heartBeatPoolForTeacher.Put(key)
	}()

	_, err := key.SetEX(ctx, 0, int(define.HeartBeatDuration/time.Second))
	if err != nil {
		i.Logger.Errorf(err, "set ex for key %q with duration %v failed", key.String(), define.HeartBeatDuration)
		return errorx.InternalErr(err)
	}
	return nil
}

func (i *IDEService) refreshHeartBeatStat(ctx context.Context, key *rediskey.EntityKey) ([]byte, error) {
	value, err := key.GetBytes(ctx)
	switch err {
	case nil:
	case redigo.ErrNil:
		return nil, errorx.ErrRedisKeyNil
	default:
		i.Logger.Errorf(err, "get redis key %q failed", key.String())
		return nil, errorx.InternalErr(err)
	}
	var stat pb.HeartBeatStat
	if err := proto.Unmarshal(value, &stat); err != nil {
		i.Logger.Errorf(err, "proto unmarshal %v for heart beat stat failed", value)
		return nil, errorx.InternalErr(err)
	}
	stat.LastVisitedAt = time.Now().Unix()
	value, err = proto.Marshal(&stat)
	if err != nil {
		i.Logger.Errorf(err, "proto marshal for value %v failed", value)
		return nil, errorx.InternalErr(err)
	}
	return value, nil
}

func (i *IDEService) HeartBeatWhenStartingForStudent(ctx context.Context, labID, studentID uint64) error {
	key := rediskey.NewkeyFormat(define.HeartBeatTagFormatForStudent, labID, studentID).Pool(i.Dao.Storage.Pool())
	now := time.Now()
	stat := &pb.HeartBeatStat{
		CreatedAt:     now.Unix(),
		LastVisitedAt: now.Unix(),
	}
	value, err := proto.Marshal(stat)
	if err != nil {
		i.Logger.Errorf(err, "proto marshal for heartbeatstat %+v failed", stat)
		return errorx.InternalErr(err)
	}

	// set nx 防止并发问题
	switch _, err := key.SetNX(ctx, value); err {
	case nil:
	case redigo.ErrNil:
		i.Logger.Debugf("set nx failed for ide key %q because it exists", key.String())
	default:
		i.Logger.Errorf(err, "set nx for key %q failed", key.String())
		return errorx.InternalErr(err)
	}
	return nil
}
