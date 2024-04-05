package user_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"code-platform/config"
	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/mailx"
	"code-platform/pkg/rediskey"
	"code-platform/pkg/strconvx"
	"code-platform/pkg/testx"
	"code-platform/pkg/timex"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	. "code-platform/service/user"
	"code-platform/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func testHelper() (*storage.Storage, *UserService) {
	testStorage := testx.NewStorage()
	dao := &repository.Dao{Storage: testStorage}
	userService := NewUserService(dao, log.Sub("user"))
	return testStorage, userService
}

func TestSendVerificationCode(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	var m mailx.MailConfig
	err := config.Mail.Unmarshal(&m)
	require.NoError(t, err)
	var email = m.Email

	ctx := context.Background()
	testx.MustFlushDB(ctx, testStorage.Pool())

	err = userService.SendVerificationCode(ctx, email)
	require.NoError(t, err)
}

func TestQueryUserByID(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")

	user := &model.User{CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	_, err = userService.GetUserByID(ctx, user.ID)
	require.NoError(t, err)
}

func TestGetOuterUserByUserID(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")

	user := &model.User{CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
	}{
		{
			label:         "normal",
			userID:        user.ID,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		_, err := userService.GetOuterUserByUserID(ctx, c.userID)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestUpdateUser(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")

	user := &model.User{CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
	}{
		{
			label:         "normal",
			userID:        user.ID,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		err := userService.UpdateUser(ctx, c.userID, user.Name, user.College, user.Major, user.Organization, user.Grade, user.Gender)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestUpdateAvatar(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")

	user := &model.User{CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		userID        uint64
	}{
		{
			label:         "normal",
			userID:        user.ID,
			expectedError: nil,
		},
		{
			label:         "not found",
			userID:        10,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		err := userService.UpdateUserAvatar(ctx, c.userID, user.Avatar)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestResetPassword(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")
	testx.MustFlushDB(ctx, testStorage.Pool())
	const number = "lgb"
	const verificationCode = "123456"

	user := &model.User{Number: number, CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	key := rediskey.NewkeyFormat(VerificationCodeKeyPrefix, number)
	_, err = key.Pool(testStorage.Pool()).SetEXNX(ctx, verificationCode, int(VerificationCodeExpired.Seconds()))
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		number        string
		code          string
	}{
		{
			label:         "normal",
			number:        number,
			code:          verificationCode,
			expectedError: nil,
		},
		// number错误，则找不到key，认证失败
		{
			label:         "not found",
			number:        "1234",
			code:          verificationCode,
			expectedError: errorx.ErrFailToAuth,
		},
		{
			label:         "auth failed",
			number:        number,
			code:          "are you ok?",
			expectedError: errorx.ErrFailToAuth,
		},
	} {
		err := userService.ResetPassword(ctx, c.number, "are you ok", c.code)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestListUserCodingTime(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "coding_time")

	user := &model.User{CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	yesterday := now.Add(-time.Hour * 24)
	tomorrow := now.Add(time.Hour * 24)

	codingTimes := []*model.CodingTime{
		{UserID: user.ID, CreatedAt: yesterday, CreatedAtDate: timex.StartOfDay(yesterday), Duration: 60},
		{UserID: user.ID, CreatedAt: now, CreatedAtDate: timex.StartOfDay(now), Duration: 100},
		{UserID: user.ID, CreatedAt: now, CreatedAtDate: timex.StartOfDay(now), Duration: 60},
		{UserID: user.ID, CreatedAt: tomorrow, CreatedAtDate: timex.StartOfDay(tomorrow), Duration: 100},
	}

	err = model.BatchInsertCodingTimes(ctx, testStorage.RDB, codingTimes)
	require.NoError(t, err)

	_, err = userService.ListUserCodingTime(ctx, user.ID)
	require.NoError(t, err)
}

func TestCheckNumber(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user")
	const number = "lgbgbl@lgb.com"

	user := &model.User{Number: number, CreatedAt: now, UpdatedAt: now}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		number        string
	}{
		{
			label:         "normal",
			number:        number,
			expectedError: nil,
		},
		{
			label:         "not found",
			number:        "123456@163.com",
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		err := userService.CheckNumber(ctx, c.number)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestLogin(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "session", "refresh_token")
	testx.MustFlushDB(ctx, testStorage.LRUPool())
	const (
		number   = "2018"
		password = "123456"
	)

	hashPassword, err := bcrypt.GenerateFromPassword(strconvx.StringToBytes(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &model.User{Number: number, Password: string(hashPassword), CreatedAt: now, UpdatedAt: now}
	err = user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		number        string
		password      string
	}{
		{
			label:         "normal",
			number:        number,
			password:      password,
			expectedError: nil,
		},
		{
			label:         "not found",
			number:        "123456",
			password:      password,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "auth failed",
			number:        number,
			password:      "are you ok?",
			expectedError: errorx.ErrFailToAuth,
		},
	} {
		resp, err := userService.Login(ctx, c.number, c.password)
		assert.Equal(t, c.expectedError, err, c.label)
		if err == nil {
			assert.Lenf(t, resp.Token, TokenLength, c.label)
		}
	}
}

func TestParseUserStatFromToken(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	now := time.Now()

	testx.MustTruncateTable(ctx, testStorage.RDB, "user", "session")
	testx.MustFlushDB(ctx, testStorage.LRUPool())

	const (
		token = "abcdefghijklmnopqrstuvwxyz0123456789"

		fakeToken = "111111111111111111111111111111111111"
	)

	user := &model.User{
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := user.Insert(ctx, testStorage.RDB)
	require.NoError(t, err)

	session := &model.Session{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: now,
		ExpireAt:  now.Add(time.Minute),
	}

	expiredSession := &model.Session{
		UserID:    user.ID,
		Token:     token + "1",
		CreatedAt: now.Add(-time.Hour * 24),
		ExpireAt:  now.Add(-time.Hour * 23),
	}

	err = model.BatchInsertSessions(ctx, testStorage.RDB, []*model.Session{session, expiredSession})
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		label         string
		token         string
	}{
		{
			label:         "normal",
			token:         token,
			expectedError: nil,
		},
		{
			label:         "invalid length",
			token:         "token",
			expectedError: errorx.ErrFailToAuth,
		},
		{
			label:         "not found",
			token:         fakeToken,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "expired_at",
			token:         token + "1",
			expectedError: errorx.ErrFailToAuth,
		},
	} {
		_, _, err := userService.ParseUserStatFromToken(ctx, c.token)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestImportStudentByCSV(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	ctx := context.Background()
	testx.MustTruncateTable(ctx, testStorage.RDB, "user")
	now := time.Now()

	rows := [][]string{
		{"number", "realName", "college", "grade", "organization", "major"},
		{"1", "李红", "计算机学院", "2018", "4班", "计算机科学与技术"},
		{"2", "李明", "政治与行政学院", "2018", "3班", "政治与行政"},
		{"3", "李华", "心理学院", "2019", "4班", "心理学"},
		{"4", "张三", "光电子学院", "2020", "5班", "光电子"},
	}

	buf := bytes.NewBuffer(nil)
	csvWriter := csv.NewWriter(buf)
	err := csvWriter.WriteAll(rows)
	require.NoError(t, err)
	csvWriter.Flush()

	users := []*model.User{
		{Number: "1", CreatedAt: now, UpdatedAt: now},
		{Number: "2", CreatedAt: now, UpdatedAt: now},
		{Number: "3", CreatedAt: now, UpdatedAt: now},
	}
	err = model.BatchInsertUsers(ctx, testStorage.RDB, users)
	require.NoError(t, err)

	for _, c := range []struct {
		expectedError error
		f             func() error
		label         string
		filename      string
	}{
		{
			label:         "UnsupportFileType",
			filename:      "1.txt",
			f:             func() error { return nil },
			expectedError: errorx.ErrUnsupportFileType,
		},
		{
			label:         "normal",
			filename:      "1.csv",
			f:             func() error { return nil },
			expectedError: nil,
		},
		{
			label:    "PersonalInfoInvalid",
			filename: "1.csv",
			f: func() error {
				// 补充姓名不完整的情形
				buf.Reset()

				rows = [][]string{
					{"number", "realName", "college", "grade", "organization", "major"},
					{"5", "", "光电子学院", "2018", "4班", "光电子"},
				}
				err := csvWriter.WriteAll(rows)
				if err != nil {
					return err
				}
				csvWriter.Flush()
				return nil
			},
			expectedError: errorx.ErrPersonalInfoInvalid,
		},
	} {
		err := c.f()
		require.NoError(t, err)
		err = userService.ImportStudentByCSV(ctx, c.filename, buf.Bytes())
		assert.Equal(t, c.expectedError, err, c.label)
	}
}

func TestRefreshToken(t *testing.T) {
	testStorage, userService := testHelper()
	defer testStorage.Close()

	now := time.Now()
	const userID = 1

	b := &strings.Builder{}
	getToken := func(base byte, length int) string {
		b.Reset()
		for i := 0; i < length; i++ {
			b.WriteByte(base)
		}
		return b.String()
	}

	tokenNotExpire := getToken('1', TokenLength)

	sessionNotExpire := &model.Session{
		UserID:    userID,
		Token:     tokenNotExpire,
		CreatedAt: now,
		ExpireAt:  now.Add(time.Hour),
	}

	tokenExpire := getToken('2', TokenLength)

	sessionExpire := &model.Session{
		UserID:    userID,
		Token:     tokenExpire,
		CreatedAt: now.Add(-24 * time.Hour),
		ExpireAt:  now.Add(-23 * time.Hour),
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	refreshTokenNotExpire := getToken('3', RefreshTokenLength)

	refreshSessionNotExpire := &model.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenNotExpire,
		CreatedAt: now,
		ExpireAt:  now.Add(time.Hour),
	}

	refreshTokenExpire := getToken('5', RefreshTokenLength)

	refreshSessionExpire := &model.RefreshToken{
		UserID:    userID,
		Token:     refreshTokenExpire,
		CreatedAt: now.Add(-24 * time.Hour),
		ExpireAt:  now.Add(-23 * time.Hour),
	}
	refreshTokenWrontID := getToken('6', RefreshTokenLength)

	refreshSessionWrongID := &model.RefreshToken{
		UserID:    10,
		Token:     refreshTokenWrontID,
		CreatedAt: now,
		ExpireAt:  now.Add(time.Hour),
	}
	tokenNotFound := getToken('7', TokenLength)

	refreshTokenNotFound := getToken('8', RefreshTokenLength)

	for _, c := range []struct {
		expectedError error
		label         string
		token         string
		refreshToken  string
	}{
		{
			label:         "normal",
			token:         tokenExpire,
			refreshToken:  refreshTokenNotExpire,
			expectedError: nil,
		},
		{
			label:         "invalid token length",
			token:         "tokenExpire",
			refreshToken:  refreshTokenNotExpire,
			expectedError: errorx.ErrFailToAuth,
		},
		{
			label:         "invalid token length",
			token:         tokenExpire,
			refreshToken:  "refreshTokenNotExpire",
			expectedError: errorx.ErrFailToAuth,
		},
		{
			label:         "token not expire",
			token:         tokenNotExpire,
			refreshToken:  refreshTokenNotExpire,
			expectedError: errorx.ErrNotExpire,
		},
		{
			label:         "refresh_token expire",
			token:         tokenExpire,
			refreshToken:  refreshTokenExpire,
			expectedError: errorx.ErrFailToAuth,
		},
		{
			label:         "refresh_token wrong user id",
			token:         tokenExpire,
			refreshToken:  refreshTokenWrontID,
			expectedError: errorx.ErrFailToAuth,
		},
		{
			label:         "token not found",
			token:         tokenNotFound,
			refreshToken:  refreshTokenNotExpire,
			expectedError: errorx.ErrIsNotFound,
		},
		{
			label:         "refresh_token not found",
			token:         tokenExpire,
			refreshToken:  refreshTokenNotFound,
			expectedError: errorx.ErrIsNotFound,
		},
	} {
		testx.MustTruncateTable(ctx, testStorage.RDB, "session", "refresh_token")
		testx.MustFlushDB(ctx, testStorage.Pool())
		err := model.BatchInsertSessions(ctx, testStorage.RDB, []*model.Session{sessionExpire, sessionNotExpire})
		require.NoError(t, err)

		err = model.BatchInsertRefreshTokens(ctx, testStorage.RDB, []*model.RefreshToken{refreshSessionExpire, refreshSessionNotExpire, refreshSessionWrongID})
		require.NoError(t, err)

		_, _, err = userService.RefreshToken(ctx, c.token, c.refreshToken)
		assert.Equal(t, c.expectedError, err, c.label)
	}
}
