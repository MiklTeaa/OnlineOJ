package user

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"html/template"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"code-platform/log"
	"code-platform/pkg/charsetx"
	"code-platform/pkg/errorx"
	"code-platform/pkg/mailx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/randx"
	"code-platform/pkg/rediskey"
	"code-platform/pkg/strconvx"
	"code-platform/pkg/timex"
	"code-platform/pkg/transactionx"
	"code-platform/pkg/validx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/service/user/pb"
	"code-platform/storage"

	redigo "github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

type UserService struct {
	Dao    *repository.Dao
	Logger *log.Logger
}

func NewUserService(dao *repository.Dao, logger *log.Logger) *UserService {
	return &UserService{
		Dao:    dao,
		Logger: logger,
	}
}

var (
	tokenRedisKeyPool = &sync.Pool{
		New: func() interface{} {
			// 返回空key
			return rediskey.Newkey("")
		},
	}
)

func (u *UserService) UpdateUser(ctx context.Context, ID uint64, name, college, major, organization string, grade uint16, gender int8) error {
	user, err := model.QueryUserByID(ctx, u.Dao.Storage.RDB, ID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		u.Logger.Errorf(err, "user is not found by ID(%d) failed", ID)
		return errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "query user by ID(%d) failed", ID)
		return errorx.InternalErr(err)
	}

	user.Name = name
	user.Gender = gender
	user.Major = major
	user.Organization = organization
	user.College = college
	user.Grade = grade
	if err := user.Update(ctx, u.Dao.Storage.RDB); err != nil {
		u.Logger.Errorf(err, "update for user %+v failed", user)
		return errorx.InternalErr(err)
	}
	return nil
}

func (u *UserService) UpdateUserAvatar(ctx context.Context, userID uint64, avatar string) error {
	user, err := model.QueryUserByID(ctx, u.Dao.Storage.RDB, userID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		u.Logger.Errorf(err, "user is not found by ID(%d) failed", userID)
		return errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "query user by ID(%d) failed", userID)
		return errorx.InternalErr(err)
	}

	if user.Avatar == avatar {
		return nil
	}

	user.Avatar = avatar
	if err := user.Update(ctx, u.Dao.Storage.RDB); err != nil {
		u.Logger.Errorf(err, "update for user %+v failed", user)
		return errorx.InternalErr(err)
	}
	return nil
}

func (u *UserService) SendVerificationCode(ctx context.Context, number string) error {
	code, err := randx.NewRandCode(6)
	if err != nil {
		u.Logger.Error(err, "generate rand code failed")
		return errorx.InternalErr(err)
	}

	// 5分钟有效期
	key := rediskey.NewkeyFormat(VerificationCodeKeyPrefix, number).Pool(u.Dao.Storage.Pool())
	_, err = key.SetEXNX(ctx, code, int(VerificationCodeExpired.Seconds()))
	switch err {
	case nil:
	case redigo.ErrNil:
		// 验证码未过期，直接返回即可
		return nil
	default:
		u.Logger.Errorf(err, "setEXNX for key %q and value %v failed", key.String(), code)
		return errorx.InternalErr(err)
	}

	tmpl, err := template.New("welcome").Parse(WelcomeHtmlTemplate)
	if err != nil {
		u.Logger.Error(err, "parse welcome html file failed")
		return errorx.InternalErr(err)
	}

	buf := bytes.NewBuffer(nil)
	buf.Grow(len(WelcomeHtmlTemplate) + 28)
	err = tmpl.Execute(buf, &struct {
		Code   string
		Layout string
	}{
		Code:   code,
		Layout: time.Now().Format("2006 Jan 02 15:04:05"),
	})
	if err != nil {
		u.Logger.Error(err, "execute welcome html file layout failed")
		return errorx.InternalErr(err)
	}

	switch err := mailx.SendEmail(number, "code-platform注册邮件", buf.Bytes()); err {
	case nil:
	case errorx.ErrMailUserNotFound:
		u.Logger.Debugf("number %q is not found", number)
		return err
	default:
		u.Logger.Error(err, "send mail failed")
		return errorx.InternalErr(err)
	}
	return nil
}

func (u *UserService) ResetPassword(ctx context.Context, number, password, inputCode string) error {
	key := rediskey.NewkeyFormat(VerificationCodeKeyPrefix, number).Pool(u.Dao.Storage.Pool())
	code, err := key.Get(ctx)
	switch err {
	case nil:
	case redigo.ErrNil:
		return errorx.ErrFailToAuth
	default:
		u.Logger.Errorf(err, "checkVerCode for number %q and code %q failed", number, inputCode)
		return errorx.InternalErr(err)
	}

	if code != inputCode {
		return errorx.ErrFailToAuth
	}

	hashPassword, err := bcrypt.GenerateFromPassword(strconvx.StringToBytes(password), bcrypt.DefaultCost)
	if err != nil {
		u.Logger.Errorf(err, "generate hash password for %q failed", password)
		return errorx.InternalErr(err)
	}

	user, err := model.QueryUserByNumber(ctx, u.Dao.Storage.RDB, number)
	switch err {
	case nil:
	case sql.ErrNoRows:
		u.Logger.Debugf("user is not found by number %q", number)
		return errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "query user by number %q failed")
		return errorx.InternalErr(err)
	}

	user.Password = string(hashPassword)
	if err := user.Update(ctx, u.Dao.Storage.RDB); err != nil {
		u.Logger.Errorf(err, "update for user %+v failed", user)
		return errorx.InternalErr(err)
	}
	return nil
}

func (u *UserService) GetUserByID(ctx context.Context, ID uint64) (*model.User, error) {
	user, err := model.QueryUserByID(ctx, u.Dao.Storage.RDB, ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) GetOuterUserByUserID(ctx context.Context, ID uint64) (*OuterUser, error) {
	user, err := u.GetUserByID(ctx, ID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		u.Logger.Debugf("user is not found by ID(%d)", ID)
		return nil, errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "query user by ID(%d) failed", ID)
		return nil, errorx.InternalErr(err)
	}

	return ToOuterUser(user), nil
}

func (u *UserService) ListUserCodingTime(ctx context.Context, userID uint64) ([]*CodingTimeInfo, error) {
	codingTimes, err := model.QueryCodingTimeInfosByUserIDInNaturalYear(ctx, u.Dao.Storage.RDB, userID)
	if err != nil {
		u.Logger.Errorf(err, "query codingTimes by userID(%d) failed", userID)
		return nil, errorx.InternalErr(err)
	}
	codingTimeInfos := make([]*CodingTimeInfo, len(codingTimes))
	for i, codingTime := range codingTimes {
		codingTimeInfos[i] = &CodingTimeInfo{
			Date:     codingTime.Date.In(timex.ShanghaiLocation).Format("2006-01-02"),
			Duration: codingTime.Duration,
		}
	}
	return codingTimeInfos, nil
}

func (u *UserService) CheckNumber(ctx context.Context, number string) error {
	_, err := model.QueryUserByNumber(ctx, u.Dao.Storage.RDB, number)
	switch err {
	case nil:
	case sql.ErrNoRows:
		return errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "query user by number %q failed", number)
		return errorx.InternalErr(err)
	}
	return nil
}

func (u *UserService) Login(ctx context.Context, number, password string) (*LoginResponse, error) {
	user, err := model.QueryUserByNumber(ctx, u.Dao.Storage.RDB, number)
	switch err {
	case nil:
	case sql.ErrNoRows:
		u.Logger.Debugf("user is not found by number %q", number)
		return nil, errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "query user by number %q failed", number)
		return nil, errorx.InternalErr(err)
	}

	// 比对密码
	if err := bcrypt.CompareHashAndPassword(strconvx.StringToBytes(user.Password), strconvx.StringToBytes(password)); err != nil {
		return nil, errorx.ErrFailToAuth
	}

	token, err := randx.NewRandCode(TokenLength)
	if err != nil {
		u.Logger.Error(err, "generate rand code failed")
		return nil, errorx.InternalErr(err)
	}

	now := time.Now()
	session := &model.Session{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: now,
		ExpireAt:  now.Add(accessTokenDuration),
	}

	// generate refresh_token
	refToken, err := randx.NewRandCode(RefreshTokenLength)
	if err != nil {
		u.Logger.Error(err, "generate rand code failed")
		return nil, errorx.InternalErr(err)
	}
	refreshToken := &model.RefreshToken{
		UserID:    user.ID,
		Token:     refToken,
		CreatedAt: now,
		ExpireAt:  now.Add(refreshTokenDuration),
	}

	task := func(ctx context.Context, tx storage.RDBClient) error {
		if err := session.Insert(ctx, tx); err != nil {
			u.Logger.Errorf(err, "insert session %+v failed", session)
			return errorx.InternalErr(err)
		}
		if err := refreshToken.Insert(ctx, tx); err != nil {
			u.Logger.Errorf(err, "insert refresh_token %+v failed", refreshToken)
			return errorx.InternalErr(err)
		}
		return nil
	}

	if err := transactionx.DoTransaction(ctx, u.Dao.Storage, u.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted}); err != nil {
		return nil, err
	}

	// 异步任务，补充 token 到 redis 缓存池
	u.backFilledToken(token, &pb.TokenStat{UserId: user.ID, Role: uint32(user.Role)}, session.ExpireAt, 5*time.Second)

	return &LoginResponse{
		Role:         user.Role,
		Token:        token,
		RefreshToken: refToken,
	}, nil
}

func tokenToUserStat(value []byte) (userID uint64, role uint16, err error) {
	var stat pb.TokenStat
	if err := proto.Unmarshal(value, &stat); err != nil {
		return 0, 0, err
	}
	return stat.UserId, uint16(stat.Role), nil
}

func (u *UserService) ParseUserStatFromToken(ctx context.Context, token string) (userID uint64, role uint16, err error) {
	if len(token) != TokenLength {
		return 0, 0, errorx.ErrFailToAuth
	}

	// 从池子中取
	key := tokenRedisKeyPool.Get().(*rediskey.EntityKey).Pool(u.Dao.Storage.LRUPool()).Replace(token)
	defer func() {
		// 清空key后，放回池中
		key.Clear()
		tokenRedisKeyPool.Put(key)
	}()
	value, err := key.GetBytes(ctx)
	switch err {
	case redigo.ErrNil:
	case nil:
		return tokenToUserStat(value)
	default:
		u.Logger.Errorf(err, "get key %q from redis lru pool failed", key.String())
		// 不直接 return, fallback 至 mysql 保护逻辑
	}

	session, err := model.QuerySessionByToken(ctx, u.Dao.Storage.RDB, token)
	switch err {
	case nil:
	case sql.ErrNoRows:
		u.Logger.Debugf("session is not found by token %q", token)
		return 0, 0, errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "query session by token %q failed", token)
		return 0, 0, errorx.InternalErr(err)
	}

	// key 已经过期
	if time.Since(session.ExpireAt) >= 0 {
		return 0, 0, errorx.ErrFailToAuth
	}

	user, err := model.QueryUserByID(ctx, u.Dao.Storage.RDB, session.UserID)
	if err != nil {
		u.Logger.Errorf(err, "query user by id[%d] failed", session.UserID)
		return 0, 0, errorx.InternalErr(err)
	}

	/*
		未过期，但 key 因太久不用被 lru 算法淘汰
		开启异步任务, 回填 key 至 redis
	*/

	u.backFilledToken(token, &pb.TokenStat{UserId: user.ID, Role: uint32(user.Role)}, session.ExpireAt, 5*time.Second)

	return session.UserID, user.Role, nil
}

func (u *UserService) ImportStudentByCSV(ctx context.Context, fileName string, data []byte) (err error) {
	if ext := filepath.Ext(fileName); ext != ".csv" {
		return errorx.ErrUnsupportFileType
	}

	var reader *csv.Reader
	if !utf8.Valid(data) {
		reader = csv.NewReader(charsetx.GBKToUTF8(bytes.NewReader(data)))
	} else {
		reader = csv.NewReader(bytes.NewReader(data))
	}

	csvRows, err := reader.ReadAll()
	if err != nil {
		u.Logger.Error(err, "csvReader readAll data from file failed")
		return errorx.InternalErr(err)
	}

	type userInfo struct {
		name         string
		college      string
		organization string
		major        string
		grade        uint16
	}

	var (
		rowsLength = len(csvRows) - 1

		usersNumberMap = make(map[string]*userInfo, rowsLength)
		numbers        = make([]string, 0, rowsLength)

		users []*model.User
	)

	for index, row := range csvRows {
		if index == 0 {
			continue
		}

		number, name, college, gradeStr, organization, major := row[0], row[1], row[2], row[3], row[4], row[5]
		if !validx.CheckUserData(number, name, organization) {
			u.Logger.Debug("personal info is invalid")
			return errorx.ErrPersonalInfoInvalid
		}

		grade, err := strconv.Atoi(gradeStr)
		if err != nil {
			return errorx.ErrPersonalInfoInvalid
		}
		usersNumberMap[number] = &userInfo{
			name:         name,
			college:      college,
			grade:        uint16(grade),
			organization: organization,
			major:        major,
		}

		numbers = append(numbers, number)
	}

	numberExistsMap, err := model.QueryUserNumberToUserIDPointerMapByNumber(ctx, u.Dao.Storage.RDB, numbers)
	if err != nil {
		u.Logger.Errorf(err, "QueryUserNumberExistMapByNumber by numbers %v failed", numbers)
		return errorx.InternalErr(err)
	}

	now := time.Now()

	// 默认密码
	hashPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		u.Logger.Error(err, "generate hash password failed")
		return errorx.InternalErr(err)
	}

	hashPasswordStr := strconvx.BytesToString(hashPassword)

	for number, userIDPointer := range numberExistsMap {
		// 存在记录
		if userIDPointer != nil {
			continue
		}

		userInfo := usersNumberMap[number]
		users = append(users, &model.User{
			Number:       number,
			Name:         userInfo.name,
			Password:     hashPasswordStr,
			College:      userInfo.college,
			Grade:        userInfo.grade,
			Major:        userInfo.major,
			Organization: userInfo.organization,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	err = model.BatchInsertUsers(ctx, u.Dao.Storage.RDB, users)
	if err != nil {
		u.Logger.Errorf(err, "batch insert users %v failed", users)
		return errorx.InternalErr(err)
	}

	return nil
}

func (u *UserService) RefreshToken(ctx context.Context, token, refreshToken string) (newToken string, newRefreshToken string, err error) {
	if len(token) != TokenLength || len(refreshToken) != RefreshTokenLength {
		return "", "", errorx.ErrFailToAuth
	}

	var (
		session         *model.Session
		refreshTokenObj *model.RefreshToken
	)

	tasks := []func() error{
		func() (err error) {
			session, err = model.QuerySessionByToken(ctx, u.Dao.Storage.RDB, token)
			switch err {
			case nil:
			case context.Canceled:
				u.Logger.Debug("QuerySessionByToken is canceled")
				return err
			case sql.ErrNoRows:
				u.Logger.Debugf("session is not found by token %q", token)
				return errorx.ErrIsNotFound
			default:
				u.Logger.Errorf(err, "query session by token %q failed", token)
				return errorx.InternalErr(err)
			}

			// session 还没到期，那就不刷新了
			if time.Until(session.ExpireAt) > 0 {
				u.Logger.Debugf("token %q is not expire", token)
				return errorx.ErrNotExpire
			}

			return nil
		},
		func() (err error) {
			refreshTokenObj, err = model.QueryRefreshTokenByToken(ctx, u.Dao.Storage.RDB, refreshToken)
			switch err {
			case nil:
			case context.Canceled:
				u.Logger.Debug("QueryRefreshTokenByToken is canceled")
				return err
			case sql.ErrNoRows:
				u.Logger.Debugf("refresh_token is not found by token %q", refreshToken)
				return errorx.ErrIsNotFound
			default:
				u.Logger.Errorf(err, "query refresh_token by token %q failed", refreshToken)
				return errorx.InternalErr(err)
			}

			/*
				refresh_token 数据较少，大概率这个任务查询速度会快些，因此紧接着判断是否到期即可
				如果refresh_token 已经过期：
			*/
			if time.Since(refreshTokenObj.ExpireAt) >= 0 {
				return errorx.ErrFailToAuth
			}

			return nil
		},
	}

	if err := parallelx.Do(u.Logger, tasks...); err != nil {
		return "", "", err
	}

	// userID 无法对应
	if session.UserID != refreshTokenObj.UserID {
		return "", "", errorx.ErrFailToAuth
	}

	newToken, newRefreshToken, err = u.refreshToken(ctx, session.UserID, refreshTokenObj)
	if err != nil {
		return "", "", err
	}
	return newToken, newRefreshToken, nil
}

func (u *UserService) refreshToken(ctx context.Context, userID uint64, refreshTokenObj *model.RefreshToken) (newToken, newRefreshToken string, err error) {
	newToken, err = randx.NewRandCode(TokenLength)
	if err != nil {
		u.Logger.Error(err, "generate rand code failed")
		return "", "", errorx.InternalErr(err)
	}

	newRefreshToken, err = randx.NewRandCode(RefreshTokenLength)
	if err != nil {
		u.Logger.Error(err, "generate rand code failed")
		return "", "", errorx.InternalErr(err)
	}

	user, err := model.QueryUserByID(ctx, u.Dao.Storage.RDB, userID)
	if err != nil {
		u.Logger.Errorf(err, "query user by id[%d] failed", userID)
		return "", "", errorx.InternalErr(err)
	}

	now := time.Now()
	newSession := &model.Session{
		UserID:    userID,
		Token:     newToken,
		CreatedAt: now,
		ExpireAt:  now.Add(accessTokenDuration),
	}

	task := func(ctx context.Context, tx storage.RDBClient) error {
		// 新 token 插入生成
		if err := newSession.Insert(ctx, tx); err != nil {
			u.Logger.Errorf(err, "insert session %+v failed", newSession)
			return errorx.InternalErr(err)
		}
		// 旧 refresh_token 失效
		refreshTokenObj.ExpireAt = now
		if err := refreshTokenObj.Update(ctx, tx); err != nil {
			u.Logger.Errorf(err, "update refresh_token %+v failed", refreshTokenObj)
			return errorx.InternalErr(err)
		}

		// 新 refresh_token 生成
		newRefreshTokenObj := &model.RefreshToken{
			UserID:    userID,
			Token:     newRefreshToken,
			ExpireAt:  now.Add(refreshTokenDuration),
			CreatedAt: now,
		}
		if err := newRefreshTokenObj.Insert(ctx, tx); err != nil {
			u.Logger.Errorf(err, "insert refresh_token %+v failed", newRefreshTokenObj)
			return errorx.InternalErr(err)
		}

		// 同步回填至缓存，防止下次请求直接穿透Redis
		stat := &pb.TokenStat{
			UserId: userID,
			Role:   uint32(user.Role),
		}
		value, err := proto.Marshal(stat)
		if err != nil {
			u.Logger.Errorf(err, "proto marshal for %+v failed", stat)
			return errorx.InternalErr(err)
		}

		key := rediskey.Newkey(newToken).Pool(u.Dao.Storage.LRUPool())
		duration := time.Until(newSession.ExpireAt)
		if _, err := key.SetEX(ctx, value, int(duration/time.Second)); err != nil {
			u.Logger.Errorf(err, "setex value %+v for key %q failed with duration %d seconds", value, key.String(), duration/time.Second)
		}

		return nil
	}

	if err := transactionx.DoTransaction(ctx, u.Dao.Storage, u.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted}); err != nil {
		return "", "", err
	}
	return newToken, newRefreshToken, nil
}

// 回填 token 至缓存。异步任务，无返回值
func (u *UserService) backFilledToken(token string, userStat *pb.TokenStat, expireAt time.Time, duration time.Duration) {
	parallelx.DoAsyncWithTimeOut(context.TODO(), duration, u.Logger, func(ctx context.Context) (err error) {
		// proto marshal
		value, err := proto.Marshal(userStat)
		if err != nil {
			u.Logger.Errorf(err, "proto marshal for %+v failed", userStat)
			return err
		}
		// 异步启动，需再计算一次时间
		duration := time.Until(expireAt)
		// 少于 1 s 则不执行
		if duration < time.Second {
			u.Logger.Debugf("token %q is ready to expire, so don't backFilledToken to redis server", token)
			return nil
		}
		key := rediskey.Newkey(token).Pool(u.Dao.Storage.LRUPool())
		if _, err := key.SetEX(ctx, value, int(duration/time.Second)); err != nil {
			u.Logger.Errorf(err, "setex value %v for key %q failed with duration %d seconds", value, key.String(), duration/time.Second)
			return err
		}
		return nil
	})
}
