package user

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"time"

	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/strconvx"
	"code-platform/repository/rdb/model"

	"golang.org/x/crypto/bcrypt"
)

func (u *UserService) ListAllUsers(ctx context.Context, offset, limit int) (*PageResponse, error) {
	var (
		total int
		users []*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfUsers(ctx, u.Dao.Storage.RDB)
			switch err {
			case nil:
			case context.Canceled:
				u.Logger.Debug("QueryTotalAmountCourses is canceled")
				return err
			default:
				u.Logger.Errorf(err, "QueryTotalAmountCourses failed")
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			users, err = model.QueryAllUsers(ctx, u.Dao.Storage.RDB, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				u.Logger.Debug("QueryAllUsers is canceled")
				return err
			default:
				u.Logger.Errorf(err, "QueryAllUsers is canceled")
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(u.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		PageInfo: &PageInfo{Total: total},
		Records:  BatchToOuterUser(users),
	}, nil
}

func (u *UserService) ExportCSVTemplate() ([]byte, error) {
	buf := bytes.NewBufferString("\xEF\xBB\xBF")
	writer := csv.NewWriter(buf)
	headLine := []string{"学号", "姓名", "学院", "年级", "班级", "专业"}

	if err := writer.Write(headLine); err != nil {
		u.Logger.Error(err, "csv.Writer write data failed")
		return nil, errorx.InternalErr(err)
	}
	writer.Flush()

	return buf.Bytes(), nil
}

func (u *UserService) DistributeAccount(ctx context.Context, number, name string, role uint16) (*OuterUser, error) {
	const initPassword = "123456"
	hashPassword, err := bcrypt.GenerateFromPassword(strconvx.StringToBytes(initPassword), bcrypt.DefaultCost)
	if err != nil {
		u.Logger.Errorf(err, "generate hash password for %q failed", initPassword)
		return nil, errorx.InternalErr(err)
	}

	_, err = model.QueryUserByNumber(ctx, u.Dao.Storage.RDB, number)
	switch err {
	case sql.ErrNoRows:
	case nil:
		u.Logger.Debugf("number %q is duplicate", number)
		return nil, errorx.ErrMySQLDuplicateKey
	default:
		u.Logger.Errorf(err, "Query User By Number %q failed", number)
		return nil, errorx.InternalErr(err)
	}

	now := time.Now()
	user := &model.User{
		Role:      role,
		Number:    number,
		Name:      name,
		Password:  string(hashPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := user.Insert(ctx, u.Dao.Storage.RDB); err != nil {
		if errorx.IsDuplicateMySQLError(err) {
			u.Logger.Debugf("user duplicate key error %v", err)
			return nil, errorx.ErrMySQLDuplicateKey
		}
		u.Logger.Errorf(err, "insert user %+v failed", user)
		return nil, errorx.InternalErr(err)
	}

	return ToOuterUser(user), nil
}

func (u *UserService) UpdateUserByAdmin(ctx context.Context, ID uint64, number, name, college, major, organization, avatar string, grade uint16, gender int8) error {
	var user *model.User
	tasks := []func() error{
		func() (err error) {
			user, err := model.QueryUserByNumber(ctx, u.Dao.Storage.RDB, number)
			switch err {
			case sql.ErrNoRows:
			case nil:
				if user.ID != ID {
					u.Logger.Debugf("number %q is duplicate", number)
					return errorx.ErrMySQLDuplicateKey
				}
			case context.Canceled:
				u.Logger.Debug("QueryUserByNumber is canceled")
				return err
			default:
				u.Logger.Errorf(err, "Query User By Number %q failed", number)
				return errorx.InternalErr(err)
			}

			return nil
		},
		func() (err error) {
			user, err = model.QueryUserByID(ctx, u.Dao.Storage.RDB, ID)
			switch err {
			case nil:
			case sql.ErrNoRows:
				u.Logger.Debugf("user is not found by ID(%d) failed", ID)
				return errorx.ErrIsNotFound
			case context.Canceled:
				u.Logger.Debug("QueryUserByID is canceled")
				return err
			default:
				u.Logger.Errorf(err, "query user by ID(%d) failed", ID)
				return errorx.InternalErr(err)
			}

			return nil
		},
	}

	if err := parallelx.Do(u.Logger, tasks...); err != nil {
		return err
	}

	user.Number = number
	user.Name = name
	user.Gender = gender
	user.College = college
	user.Grade = grade
	user.Major = major
	user.Organization = organization
	user.Avatar = avatar

	if err := user.Update(ctx, u.Dao.Storage.RDB); err != nil {
		if errorx.IsDuplicateMySQLError(err) {
			u.Logger.Debugf("user duplicate key error %v", err)
			return errorx.ErrMySQLDuplicateKey
		}
		u.Logger.Errorf(err, "update for user %+v failed", user)
		return errorx.InternalErr(err)
	}
	return nil
}

func (u *UserService) ResetPassordByAdmin(ctx context.Context, password string, ID uint64) error {
	user, err := model.QueryUserByID(ctx, u.Dao.Storage.RDB, ID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		u.Logger.Debugf("user is not found by id[%d]", ID)
		return errorx.ErrIsNotFound
	default:
		u.Logger.Errorf(err, "Query user by id[%d] failed", ID)
		return errorx.InternalErr(err)
	}

	hashPassword, err := bcrypt.GenerateFromPassword(strconvx.StringToBytes(password), bcrypt.DefaultCost)
	if err != nil {
		u.Logger.Error(err, "generate password by bcrypt failed")
		return errorx.InternalErr(err)
	}

	user.Password = strconvx.BytesToString(hashPassword)
	if err := user.Update(ctx, u.Dao.Storage.RDB); err != nil {
		u.Logger.Errorf(err, "update user by user_id[%d] failed", user.ID)
		return errorx.InternalErr(err)
	}

	return nil
}
