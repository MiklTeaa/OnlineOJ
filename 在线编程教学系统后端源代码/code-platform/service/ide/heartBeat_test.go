package ide_test

import (
	"context"
	"testing"

	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"

	"github.com/stretchr/testify/require"
)

func TestHeartBeatWhenStartingForStudent(t *testing.T) {
	testStorage, ideService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustFlushDB(ctx, testStorage.Pool())
	const (
		labID     = 1
		studentID = 1
	)

	for i := 0; i < 3; i++ {
		err := ideService.HeartBeatWhenStartingForStudent(ctx, labID, studentID)
		require.NoError(t, err)
	}
}

func TestHeartBeatForStudent(t *testing.T) {
	testStorage, ideService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustFlushDB(ctx, testStorage.Pool())
	const (
		labID     = 1
		studentID = 1
	)
	err := ideService.HeartBeatForStudent(ctx, labID, studentID)
	require.Equal(t, errorx.ErrRedisKeyNil, err)

	err = ideService.HeartBeatWhenStartingForStudent(ctx, labID, studentID)
	require.NoError(t, err)

	err = ideService.HeartBeatForStudent(ctx, labID, studentID)
	require.NoError(t, err)
}

func TestHeartBeatForTeacher(t *testing.T) {
	testStorage, ideService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustFlushDB(ctx, testStorage.Pool())
	const (
		labID     = 1
		teacherID = 1
		studentID = 2
	)
	err := ideService.HeartBeatForTeacher(ctx, labID, studentID, teacherID)
	require.NoError(t, err)
}
