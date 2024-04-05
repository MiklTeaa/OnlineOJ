package user_test

import (
	"context"
	"testing"
	"time"

	"code-platform/pkg/errorx"
	"code-platform/pkg/testx"
	"code-platform/repository/rdb/model"
	. "code-platform/service/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAllUsers(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")

	now := time.Now()
	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}
	err := model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	resp, err := userService.ListAllUsers(ctx, 0, 10)
	require.NoError(t, err)

	require.Equal(t, 2, resp.PageInfo.Total)
	records := resp.Records.([]*OuterUser)
	require.Len(t, records, 2)
}

func TestExportCSVTemplate(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	_, err := userService.ExportCSVTemplate()
	require.NoError(t, err)
}

func TestDistributeAccount(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")

	now := time.Now()
	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}
	err := model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		number        string
	}{
		{
			label:         "normal",
			number:        "3",
			expectedError: nil,
		},
		{
			label:         "number duplicate",
			number:        "2",
			expectedError: errorx.ErrMySQLDuplicateKey,
		},
	} {
		_, err := userService.DistributeAccount(ctx, c.number, "", 0)
		require.Equal(t, err, err)
	}
}

func TestUpdateUserByAdmin(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")

	now := time.Now()
	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
	}
	err := model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		number        string
		id            uint64
	}{
		{
			label:         "normal",
			id:            1,
			number:        "1",
			expectedError: nil,
		},
		{
			label:         "number duplicate",
			id:            2,
			number:        "1",
			expectedError: errorx.ErrMySQLDuplicateKey,
		},
		{
			label:         "not found",
			id:            10,
			number:        "10",
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		err := userService.UpdateUserByAdmin(ctx, c.id, c.number, "", "", "", "", "", 0, 0)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestResetPasswordByAdmin(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()
	ctx := context.Background()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")
	now := time.Now()
	user := &model.User{CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		id            uint64
	}{
		{expectedError: nil, label: "normal", id: user.ID},
		{expectedError: errorx.ErrIsNotFound, label: "not found", id: 10},
	} {
		err := userService.ResetPassordByAdmin(ctx, "123456", c.id)
		require.Equal(t, c.expectedError, err)
	}
}
