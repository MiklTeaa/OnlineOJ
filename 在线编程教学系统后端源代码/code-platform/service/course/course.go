package course

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"time"
	"unicode/utf8"

	"code-platform/log"
	"code-platform/pkg/charsetx"
	"code-platform/pkg/errorx"
	"code-platform/pkg/parallelx"
	"code-platform/pkg/strconvx"
	"code-platform/pkg/transactionx"
	"code-platform/pkg/validx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/service/define"
	"code-platform/storage"

	"golang.org/x/crypto/bcrypt"
)

type CourseService struct {
	Dao    *repository.Dao
	Logger *log.Logger
}

func NewCourseService(dao *repository.Dao, logger *log.Logger) *CourseService {
	return &CourseService{
		Dao:    dao,
		Logger: logger,
	}
}

func (c *CourseService) courseInfosPacking(courses []*model.Course, teachersMap map[uint64]*model.User) []*CourseWithTeacherInfo {
	records := make([]*CourseWithTeacherInfo, len(courses))
	for index, course := range courses {
		var (
			teacherName   string
			teacherEmail  string
			teacherAvatar string
		)
		teacher := teachersMap[course.TeacherID]
		if teacher != nil {
			teacherName = teacher.Name
			teacherEmail = define.Number2Email(teacher.Number)
			teacherAvatar = teacher.Avatar
		}

		records[index] = &CourseWithTeacherInfo{
			CourseID:      course.ID,
			TeacherID:     course.TeacherID,
			TeacherName:   teacherName,
			TeacherEmail:  teacherEmail,
			TeacherAvatar: teacherAvatar,
			CourseName:    course.Name,
			CourseDes:     course.Description,
			PicURL:        course.PicURL,
			IsClose:       course.IsClosed,
			Language:      course.Language,
			NeedAudit:     course.NeedAudit,
			CreatedAt:     course.CreatedAt,
		}
	}
	return records
}

func (c *CourseService) convertToTeachersMap(ctx context.Context, courses []*model.Course) (map[uint64]*model.User, error) {
	userIDsMap := make(map[uint64]struct{}, len(courses))
	for _, course := range courses {
		userIDsMap[course.TeacherID] = struct{}{}
	}

	userIDs := make([]uint64, 0, len(userIDsMap))
	for userID := range userIDsMap {
		userIDs = append(userIDs, userID)
	}

	usersMap, err := model.QueryUserMapByIDs(ctx, c.Dao.Storage.RDB, userIDs)
	if err != nil {
		return nil, err
	}

	return usersMap, nil
}

func (c *CourseService) ListCourseInfosByTeacherID(ctx context.Context, teacherID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total   int
		teacher *model.User
		courses []*model.Course
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountCoursesByTeacherID(ctx, c.Dao.Storage.RDB, teacherID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountCoursesByTeacherID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountCoursesByTeacherID by teacherID(%d) failed", teacherID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			teacher, err = model.QueryUserByID(ctx, c.Dao.Storage.RDB, teacherID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUserByID is canceled")
				return err
			case sql.ErrNoRows:
				c.Logger.Debugf("user is not found by userID(%d)", teacherID)
				return errorx.ErrIsNotFound
			default:
				c.Logger.Errorf(err, "QueryUserByID(%d) failed", teacherID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			courses, err = model.QueryCoursesByTeacherID(ctx, c.Dao.Storage.RDB, teacherID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCoursesByTeacherID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCoursesByTeacherID(%d) failed", teacherID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  c.courseInfosPacking(courses, map[uint64]*model.User{teacherID: teacher}),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) ListCourseInfosByStudentID(ctx context.Context, studentID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total       int
		courses     []*model.Course
		teachersMap map[uint64]*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfArrangeCourseByStudentIDWithPass(ctx, c.Dao.Storage.RDB, studentID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountOfArrangeCourseByStudentIDWithPass is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountOfArrangeCourseByStudentIDWIthPass by studentID(%d) failed", studentID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			courses, err = model.QueryCoursesByStudentID(ctx, c.Dao.Storage.RDB, studentID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCoursesByStudentID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCourses by studentID(%d) failed", studentID)
				return errorx.InternalErr(err)
			}

			teachersMap, err = c.convertToTeachersMap(ctx, courses)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("convertToTeachersMap is canceled")
				return err
			default:
				c.Logger.Errorf(err, "convertToTeachersMap by courses %v failed", courses)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  c.courseInfosPacking(courses, teachersMap),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) ListCourseInfosByCourseName(ctx context.Context, keyword string, offset, limit int) (*PageResponse, error) {
	var (
		total       int
		courses     []*model.Course
		teachersMap map[uint64]*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfCourseByFuzzyCourseName(ctx, c.Dao.Storage.RDB, keyword)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountOfCourseByFuzzyCourseName is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountOfCourseByFuzzyCourseName by keyword %q failed", keyword)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			courses, err = model.QueryCoursesByName(ctx, c.Dao.Storage.RDB, keyword, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCoursesByName is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCourses by keyword %q failed")
				return errorx.InternalErr(err)
			}

			teachersMap, err = c.convertToTeachersMap(ctx, courses)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("convertToTeachersMap is canceled")
				return err
			default:
				c.Logger.Errorf(err, "convertToTeachersMap by courses %v failed", courses)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  c.courseInfosPacking(courses, teachersMap),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) EnrollForStudent(ctx context.Context, studentID, courseID uint64, secretKey sql.NullString) error {
	var course *model.Course
	tasks := []func() error{
		func() (err error) {
			student, err := model.QueryUserByID(ctx, c.Dao.Storage.RDB, studentID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUserByID is canceled")
				return err
			case sql.ErrNoRows:
				c.Logger.Debugf("user is not found by ID(%d)", studentID)
				return errorx.ErrIsNotFound
			default:
				c.Logger.Errorf(err, "query user by ID(%d) failed", studentID)
				return errorx.InternalErr(err)
			}

			if student.Number == "" || student.Name == "" {
				return errorx.ErrPersonalInfoNotComplete
			}

			return nil
		},
		func() (err error) {
			course, err = model.QueryCourseByID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCourseByID is canceled")
				return err
			case sql.ErrNoRows:
				c.Logger.Debugf("course is not found by courseID(%d)", courseID)
				return errorx.ErrIsNotFound
			default:
				c.Logger.Errorf(err, "query course by ID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}

			if course.SecretKey.Valid != secretKey.Valid || course.SecretKey.String != secretKey.String {
				return errorx.ErrFailToAuth
			}

			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return err
	}

	now := time.Now()
	arrangeCourse := &model.ArrangeCourse{
		UserID:    studentID,
		CourseID:  courseID,
		IsPass:    !course.NeedAudit,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := arrangeCourse.Insert(ctx, c.Dao.Storage.RDB); err != nil {
		if errorx.IsDuplicateMySQLError(err) {
			c.Logger.Debugf("insert failed duplicate key for %v", err)
			return errorx.ErrMySQLDuplicateKey
		}
		c.Logger.Errorf(err, "insert arrange_course failed for %+v", arrangeCourse)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CourseService) ListStudentsCheckedByCourseID(ctx context.Context, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total int
		users []*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfArrangeCourseWithPassByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountOfArrangeCourseWithPassByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountOfArrangeCourseWithPassByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			users, err = model.QueryUsersInArrangeCourseWithPassByCourseID(ctx, c.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUsersInArrangeCourseWithPassByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryUsersInArrangeCourseWithPassByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  batchToOuterUser(users),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) ListUserCodingTimesByCourseID(ctx context.Context, userID, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total int
		users []*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfArrangeCourseWithPassByCourseID(ctx, c.Dao.Storage.RDB, courseID)
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
			users, err = model.QueryUsersByCourseID(ctx, c.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUsersByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "query users by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}
	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	userInfoWithCodingTimes := make([]*UserInfoWithCodingTime, len(users))
	userIDs := make([]uint64, len(users))
	for index, user := range users {
		userIDs[index] = user.ID
		userInfoWithCodingTimes[index] = &UserInfoWithCodingTime{
			UserID: user.ID,
			Name:   user.Name,
			Number: user.Number,
		}
	}

	codingTimeInfos, err := model.QueryCodingTimesByCourseIDAndUserIDs(ctx, c.Dao.Storage.RDB, courseID, userIDs)
	if err != nil {
		c.Logger.Errorf(err, "query codingTimes by courseID(%d) and userIDs(%v) failed", courseID, userIDs)
		return nil, errorx.InternalErr(err)
	}

	if codingTimeInfos == nil {
		return &PageResponse{
			Records:  userInfoWithCodingTimes,
			PageInfo: &PageInfo{Total: total},
		}, nil
	}

	userIDToCodingInfosMap := make(map[uint64][]*CodingTimeInfo)
	for _, info := range codingTimeInfos {
		userIDToCodingInfosMap[info.UserID] = append(userIDToCodingInfosMap[info.UserID], &CodingTimeInfo{
			Date:     info.Date.Format("2006-01-02"),
			Duration: info.Duration,
		})
	}

	for i, info := range userInfoWithCodingTimes {
		userInfoWithCodingTimes[i].CodingTimeInfos = userIDToCodingInfosMap[info.UserID]
	}

	return &PageResponse{
		Records:  userInfoWithCodingTimes,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) ExportScoreCSVByCourseID(ctx context.Context, courseID uint64) ([]byte, error) {
	var (
		users            []*model.User
		labWithUserDatas []*model.LabWithUserData
	)

	tasks := []func() error{
		func() (err error) {
			users, err = model.QueryAllUsersInArrangeCourseWithPassByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUsersInArrangeCourseByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryUsersInArrangeCourseByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			labWithUserDatas, err = model.QueryLabWithUserLabDatasByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryLabWithUserLabDatasByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryLabWithUserLabDatasByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	// reorder by userID
	sort.Slice(users, func(i, j int) bool {
		return users[i].ID < users[j].ID
	})

	return getCSVData(initUserIDTolabWithUserDataMap(labWithUserDatas), users)
}

func initUserIDTolabWithUserDataMap(labWithUserDatas []*model.LabWithUserData) map[uint64][]*model.LabWithUserData {
	m := make(map[uint64][]*model.LabWithUserData)
	for _, labWithUserData := range labWithUserDatas {
		m[labWithUserData.UserID] = append(m[labWithUserData.UserID], labWithUserData)
	}
	return m
}

func getOrderedHeadLine(userIDTolabWithUserDatasMap map[uint64][]*model.LabWithUserData) []string {
	headLine := append(make([]string, 0, 3+len(userIDTolabWithUserDatasMap)), "学号", "姓名")
	labIDToTitleMap := make(map[uint64]string)
	for _, labWithUserDatas := range userIDTolabWithUserDatasMap {
		for _, labWithUserData := range labWithUserDatas {
			if _, ok := labIDToTitleMap[labWithUserData.ID]; !ok {
				labIDToTitleMap[labWithUserData.ID] = labWithUserData.Title
			}
		}
	}

	labIDs := make([]uint64, 0, len(labIDToTitleMap))
	for labID := range labIDToTitleMap {
		labIDs = append(labIDs, labID)
	}

	// reorder
	sort.Slice(labIDs, func(i, j int) bool {
		return labIDs[i] < labIDs[j]
	})

	for _, labID := range labIDs {
		headLine = append(headLine, labIDToTitleMap[labID])
	}
	headLine = append(headLine, "平均分")
	return headLine
}

func getCSVData(userIDTolabWithUserDatasMap map[uint64][]*model.LabWithUserData, users []*model.User) ([]byte, error) {
	buf := bytes.NewBufferString("\xEF\xBB\xBF")
	writer := csv.NewWriter(buf)

	if err := writer.Write(getOrderedHeadLine(userIDTolabWithUserDatasMap)); err != nil {
		return nil, err
	}

	rows := make([][]string, 0, len(users))
	for _, user := range users {
		labSubmitInfos, ok := userIDTolabWithUserDatasMap[user.ID]
		if !ok {
			continue
		}

		// order by lab_id asc
		sort.Slice(labSubmitInfos, func(i, j int) bool {
			return labSubmitInfos[i].Lab.ID < labSubmitInfos[j].Lab.ID
		})

		totalScore := 0
		total := 0
		row := append(make([]string, 0, 3+len(labSubmitInfos)), user.Number, user.Name)
		for _, labSubmitInfo := range labSubmitInfos {
			if labSubmitInfo.Score.Valid {
				total++
				totalScore += int(labSubmitInfo.Score.Int32)
				row = append(row, fmt.Sprintf("%d", labSubmitInfo.Score.Int32))
			} else {
				row = append(row, "NULL")
			}
		}

		if total != 0 {
			averageScore := float64(totalScore) / float64(total)
			row = append(row, fmt.Sprintf("%.2f", averageScore))
		} else {
			row = append(row, "NULL")
		}
		rows = append(rows, row)
	}

	if err := writer.WriteAll(rows); err != nil {
		return nil, err
	}

	writer.Flush()
	return buf.Bytes(), nil
}

func (c *CourseService) ListAllCourses(ctx context.Context, offset, limit int) (*PageResponse, error) {
	var (
		total       int
		courses     []*model.Course
		teachersMap map[uint64]*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountCourses(ctx, c.Dao.Storage.RDB)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountCourses is canceled")
				return err
			default:
				c.Logger.Error(err, "QueryTotalAmountCourses failed")
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			courses, err = model.QueryCourses(ctx, c.Dao.Storage.RDB, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCourses is canceled")
				return err
			default:
				c.Logger.Error(err, "QueryCourseWithTeacherInfos failed")
				return errorx.InternalErr(err)
			}

			teachersMap, err = c.convertToTeachersMap(ctx, courses)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("convertToTeachersMap is canceled")
				return err
			default:
				c.Logger.Errorf(err, "convertToTeachersMap by courses %v failed", courses)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  c.courseInfosPacking(courses, teachersMap),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) GetCourseInfoByUserIDAndCourseID(ctx context.Context, userID, courseID uint64) (*CourseWithTeacherInfoAndIsEnroll, error) {
	var (
		isEnroll bool = false
		course   *model.Course
		teacher  *model.User
	)

	tasks := []func() error{
		func() (err error) {
			course, err = model.QueryCourseByID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryCourseByID is canceled")
				return err
			case sql.ErrNoRows:
				c.Logger.Debugf("course is not found by courseID(%d)", courseID)
				return errorx.ErrIsNotFound
			default:
				c.Logger.Errorf(err, "GetCourse by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}

			teacher, err = model.QueryUserByID(ctx, c.Dao.Storage.RDB, course.TeacherID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUserByID is canceled")
				return err
			case sql.ErrNoRows:
				c.Logger.Debugf("user is not found by userID(%d)", course.TeacherID)
				return errorx.ErrIsNotFound
			default:
				c.Logger.Errorf(err, "QueryUserByID(%d) failed", course.TeacherID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			err = model.QueryArrangeCourseExistsByCourseIDAndUserID(ctx, c.Dao.Storage.RDB, courseID, userID)
			switch err {
			case sql.ErrNoRows:
			case nil:
				isEnroll = true
			case context.Canceled:
				c.Logger.Debug("QueryArrangeCourseByCourseIDAndUserID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryArrangeCourseByCourseIDAndUserID by courseID(%d) and userID(%d) failed", courseID, userID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &CourseWithTeacherInfoAndIsEnroll{
		CourseWithTeacherInfo: &CourseWithTeacherInfo{
			CourseID:      courseID,
			TeacherID:     course.TeacherID,
			TeacherName:   teacher.Name,
			TeacherAvatar: teacher.Avatar,
			CourseName:    course.Name,
			CourseDes:     course.Description,
			PicURL:        course.PicURL,
			IsClose:       course.IsClosed,
			Language:      course.Language,
			NeedAudit:     course.NeedAudit,
			CreatedAt:     course.CreatedAt,
		},
		IsEnroll: isEnroll,
	}, nil
}

func (c *CourseService) ExportCSVTemplate() ([]byte, error) {
	buf := bytes.NewBufferString("\xEF\xBB\xBF")
	writer := csv.NewWriter(buf)
	headLine := []string{"学号", "姓名", "学院", "年级", "班级", "专业"}

	if err := writer.Write(headLine); err != nil {
		c.Logger.Error(err, "csv.Writer write data failed")
		return nil, errorx.InternalErr(err)
	}
	writer.Flush()

	return buf.Bytes(), nil
}

func (c *CourseService) ImportCSVTemplate(ctx context.Context, fileName string, data []byte, courseID uint64) error {
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
		c.Logger.Error(err, "csv.Reader readall data failed")
		return errorx.InternalErr(err)
	}

	task := func(ctx context.Context, tx storage.RDBClient) error {
		type userInfo struct {
			name         string
			college      string
			organization string
			major        string
			grade        uint16
		}

		var (
			rowsLength     = len(csvRows) - 1
			usersNumberMap = make(map[string]*userInfo, rowsLength)
			numbers        = make([]string, 0, rowsLength)
			userIDs        = make([]uint64, 0, rowsLength)

			arrangeCourseIDsNeededToUpdate []uint64
			arrangeCourses                 []*model.ArrangeCourse
			users                          []*model.User
		)

		for index, row := range csvRows {
			if index == 0 {
				continue
			}

			number, name, college, gradeStr, organization, major := row[0], row[1], row[2], row[3], row[4], row[5]
			if !validx.CheckUserData(number, name, organization) {
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

		numberExistsMap, err := model.QueryUserNumberToUserIDPointerMapByNumber(ctx, tx, numbers)
		if err != nil {
			c.Logger.Errorf(err, "QueryUserNumberExistMapByNumber by numbers %v failed", numbers)
			return errorx.InternalErr(err)
		}

		now := time.Now()
		// 默认密码
		hashPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		if err != nil {
			c.Logger.Error(err, "generate hash password failed")
			return errorx.InternalErr(err)
		}

		hashPasswordStr := strconvx.BytesToString(hashPassword)

		for number, userIDPointer := range numberExistsMap {
			// 存在记录
			if userIDPointer != nil {
				userID := *userIDPointer
				userIDs = append(userIDs, userID)
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

		// 先将用户导入
		err = model.BatchInsertUsers(ctx, tx, users)
		if err != nil {
			c.Logger.Errorf(err, "batch insert users %v failed", users)
			return errorx.InternalErr(err)
		}

		// 补充刚插入的用户ID
		for _, user := range users {
			userIDs = append(userIDs, user.ID)
		}

		userIDsExistInArrangeCourseMap, err := model.QueryArrangeCourseExistMapByCourseIDAndUserIDs(ctx, tx, courseID, userIDs)
		if err != nil {
			return errorx.InternalErr(err)
		}

		for userID, arrangeCourse := range userIDsExistInArrangeCourseMap {
			if arrangeCourse == nil {
				arrangeCourses = append(arrangeCourses, &model.ArrangeCourse{
					UserID:    userID,
					CourseID:  courseID,
					IsPass:    true,
					CreatedAt: now,
					UpdatedAt: now,
				})
				continue
			}

			// 未通过则记录下ID，届时更新为通过
			if !arrangeCourse.IsPass {
				arrangeCourseIDsNeededToUpdate = append(arrangeCourseIDsNeededToUpdate, arrangeCourse.ID)
			}
		}

		if err := model.BatchInsertArrangeCourses(ctx, tx, arrangeCourses); err != nil {
			c.Logger.Errorf(err, "batch insert arrangeCourses %+v failed", arrangeCourses)
			return errorx.InternalErr(err)
		}

		if err := model.BatchUpdateIsPassInArrangeCourseByIDs(ctx, tx, arrangeCourseIDsNeededToUpdate, true); err != nil {
			c.Logger.Errorf(err, "batch update is_pass for arrange_course IDs %v failed", arrangeCourseIDsNeededToUpdate)
			return errorx.InternalErr(err)
		}
		return nil
	}
	return transactionx.DoTransaction(ctx, c.Dao.Storage, c.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (c *CourseService) ListStudentWaitForChecked(ctx context.Context, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total int
		users []*model.User
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfArrangeCourseWithoutPassByCourseID(ctx, c.Dao.Storage.RDB, courseID)
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
			users, err = model.QueryUsersInArrangeCourseWithoutPassByCourseID(ctx, c.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUsersInArrangeCourseByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryUsersInArrangeCourseByCourseID by courseID(%d) failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	return &PageResponse{
		Records:  batchToOuterUser(users),
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) CheckForStudents(ctx context.Context, courseID uint64, userIDs []uint64, permit bool) error {
	if len(userIDs) == 0 {
		return nil
	}

	if !permit {
		if err := model.BatchDeleteArrangeCourseByCourseIDAndUserIDs(ctx, c.Dao.Storage.RDB, courseID, userIDs); err != nil {
			c.Logger.Errorf(err, "batch delete in arrangeCourse failed by courseID(%d) and userIDs %v", courseID, userIDs)
			return errorx.InternalErr(err)
		}
		return nil
	}

	var (
		labIDs           []uint64
		checkInRecordIDs []uint64
	)

	tasks := []func() error{
		func() (err error) {
			labIDs, err = model.QueryLabIDsByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			if err != nil {
				c.Logger.Errorf(err, "query labIDs by courseID[%d] failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			checkInRecordIDs, err = model.QueryCheckInRecordIDsByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			if err != nil {
				c.Logger.Errorf(err, "query checkInRecordIDs by courseID[%d] failed", courseID)
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return err
	}

	now := time.Now()
	labSubmits := make([]*model.LabSubmit, 0, len(labIDs)*len(userIDs))
	checkInDetails := make([]*model.CheckInDetail, 0, len(checkInRecordIDs)*len(userIDs))
	for _, userID := range userIDs {
		for _, labID := range labIDs {
			labSubmits = append(labSubmits, &model.LabSubmit{
				LabID:     labID,
				UserID:    userID,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}

		for _, recordID := range checkInRecordIDs {
			checkInDetails = append(checkInDetails, &model.CheckInDetail{
				RecordID:  recordID,
				UserID:    userID,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
	}

	task := func(ctx context.Context, tx storage.RDBClient) error {
		if err := model.BatchUpdateIsPassInArrangeCourseByCourseIDAndUserIDs(ctx, tx, courseID, userIDs); err != nil {
			c.Logger.Errorf(err, "batch update in arrangeCourse failed by courseID(%d) and userIDs %v", courseID, userIDs)
			return errorx.InternalErr(err)
		}
		if err := model.BatchInsertLabSubmits(ctx, tx, labSubmits); err != nil {
			c.Logger.Errorf(err, "BatchInsertLabSubmits for %+v failed", labSubmits)
			return errorx.InternalErr(err)
		}
		if err := model.BatchInsertCheckInDetails(ctx, tx, checkInDetails); err != nil {
			c.Logger.Errorf(err, "BatchInsertLabSubmits for %+v failed", labSubmits)
			return errorx.InternalErr(err)
		}
		return nil
	}
	return transactionx.DoTransaction(ctx, c.Dao.Storage, c.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
}

func (c *CourseService) getScoreMapAndCheckInMapByCourseIDAndUserIDs(
	ctx context.Context,
	courseID uint64,
	userIDs []uint64,
) (
	userIDToAVGScoreMap map[uint64]float64,
	actualCheckInMap map[uint64]int,
	err error,
) {

	tasks := []func() error{
		func() (err error) {
			userIDToAVGScoreMap, err = model.QueryAverageScoreMapByCourseIDAndUserIDs(ctx, c.Dao.Storage.RDB, courseID, userIDs)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryAverageScoreMapByCourseIDAndUserIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryAverageScoreMapByCourseIDAndUserIDs failed by courseID(%d) and userIDs %v", courseID, userIDs)
				return err
			}
			return nil
		},
		func() (err error) {
			actualCheckInMap, err = model.QuerySuccessCheckInAmountMapByCourseIDAndUserIDs(ctx, c.Dao.Storage.RDB, courseID, userIDs)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QuerySuccessCheckInAmountMapByCourseIDAndUserIDs is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryCheckInAmountMapByCourseIDAndUserIDs by courseID(%d) and userIDs %v failed", courseID, userIDs)
				return err
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, nil, err
	}

	return userIDToAVGScoreMap, actualCheckInMap, nil
}

func (c *CourseService) ListCoursesScore(ctx context.Context, courseID uint64, offset, limit int) (*PageResponse, error) {
	var (
		total               int
		shallCheckIn        int
		userIDToAVGScoreMap map[uint64]float64
		actualCheckInMap    map[uint64]int
		userInfos           []*UserWithAverageScoreAndCheckInData
	)

	tasks := []func() error{
		func() (err error) {
			total, err = model.QueryTotalAmountOfArrangeCourseWithPassByCourseID(ctx, c.Dao.Storage.RDB, courseID)
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
			users, err := model.QueryUsersInArrangeCourseWithPassByCourseID(ctx, c.Dao.Storage.RDB, courseID, offset, limit)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryUsersInArrangeCourseByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryUsersInArrangeCourseByCourseID by courseID(%d) failed")
				return errorx.InternalErr(err)
			}
			userInfos = make([]*UserWithAverageScoreAndCheckInData, len(users))
			userIDs := make([]uint64, len(users))
			for index, user := range users {
				userIDs[index] = user.ID
				userInfos[index] = &UserWithAverageScoreAndCheckInData{
					OuterUser: toOuterUser(user),
				}
			}

			userIDToAVGScoreMap, actualCheckInMap, err = c.getScoreMapAndCheckInMapByCourseIDAndUserIDs(ctx, courseID, userIDs)
			if err != nil {
				return errorx.InternalErr(err)
			}
			return nil
		},
		func() (err error) {
			shallCheckIn, err = model.QueryTotalAmountInCheckInRecordByCourseID(ctx, c.Dao.Storage.RDB, courseID)
			switch err {
			case nil:
			case context.Canceled:
				c.Logger.Debug("QueryTotalAmountInCheckInRecordByCourseID is canceled")
				return err
			default:
				c.Logger.Errorf(err, "QueryTotalAmountInCheckInRecordByCourseID by courseID(%d) failed")
				return errorx.InternalErr(err)
			}
			return nil
		},
	}

	if err := parallelx.Do(c.Logger, tasks...); err != nil {
		return nil, err
	}

	for _, userInfo := range userInfos {
		userInfo.AvgScore = userIDToAVGScoreMap[userInfo.OuterUser.ID]
		userInfo.ActualCheckIn = actualCheckInMap[userInfo.OuterUser.ID]
		userInfo.ShallCheckIn = shallCheckIn
	}

	return &PageResponse{
		Records:  userInfos,
		PageInfo: &PageInfo{Total: total},
	}, nil
}

func (c *CourseService) AddCourse(ctx context.Context, teacherID uint64, courseName, courseDescription string, isClosed bool, picURL, secretKey string, language int8, needAudit bool) error {
	now := time.Now()
	course := &model.Course{
		TeacherID:   teacherID,
		Name:        courseName,
		Description: courseDescription,
		PicURL:      picURL,
		SecretKey: sql.NullString{
			String: secretKey,
			Valid:  secretKey != "",
		},
		NeedAudit: needAudit,
		IsClosed:  isClosed,
		Language:  language,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := course.Insert(ctx, c.Dao.Storage.RDB); err != nil {
		c.Logger.Errorf(err, "insert course %+v failed", course)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CourseService) UpdateCourse(ctx context.Context, courseID uint64, name, description, secretKey, picURL string, language int8, needAudit bool) error {
	course, err := model.QueryCourseByID(ctx, c.Dao.Storage.RDB, courseID)
	switch err {
	case nil:
	case sql.ErrNoRows:
		c.Logger.Debugf("course is not found by courseID(%d)", courseID)
		return errorx.ErrIsNotFound
	default:
		c.Logger.Errorf(err, "QueryCourseByID by courseID(%d) failed", courseID)
		return errorx.InternalErr(err)
	}

	course.Name = name
	course.Description = description
	course.SecretKey = sql.NullString{
		String: secretKey,
		Valid:  secretKey != "",
	}
	course.PicURL = picURL
	course.Language = language
	course.NeedAudit = needAudit

	if err := course.Update(ctx, c.Dao.Storage.RDB); err != nil {
		c.Logger.Errorf(err, "update course for %+v failed", course)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CourseService) DeleteCourse(ctx context.Context, courseID, teacherID uint64) error {
	if err := model.DeleteCourseByID(ctx, c.Dao.Storage.RDB, courseID); err != nil {
		c.Logger.Errorf(err, "delete course by courseID(%d) failed", courseID)
		return errorx.InternalErr(err)
	}
	return nil
}

func (c *CourseService) QuitCourse(ctx context.Context, courseID uint64, studentID uint64) error {
	if err := model.DeleteArrangeCourseByCourseIDAndUserID(ctx, c.Dao.Storage.RDB, courseID, studentID); err != nil {
		c.Logger.Errorf(err, "batch delete arrange course by courseID(%d) and userID(%d) failed", courseID, studentID)
		return errorx.InternalErr(err)
	}
	return nil
}
