package define

import (
	"code-platform/repository/rdb/model"
)

type CodingTimeInfo struct {
	Date     string `json:"date"`
	Duration int    `json:"time"`
}

type UserWithAverageScoreAndCheckInData struct {
	*OuterUser
	AvgScore      float64 `json:"avg_score"`
	ShallCheckIn  int     `json:"shall_check_in"`
	ActualCheckIn int     `json:"act_check_in"`
}

type OuterUser struct {
	College      string `json:"college"`
	Email        string `json:"email"`
	Number       string `json:"num"`
	Name         string `json:"real_name"`
	Avatar       string `json:"avatar_url"`
	Major        string `json:"major"`
	Organization string `json:"organization"`
	ID           uint64 `json:"user_id"`
	Role         uint16 `json:"role"`
	Grade        uint16 `json:"grade"`
	Gender       int8   `json:"gender"`
}

func ToOuterUser(u *model.User) *OuterUser {
	return &OuterUser{
		ID:           u.ID,
		Role:         u.Role,
		Email:        Number2Email(u.Number),
		Number:       u.Number,
		Name:         u.Name,
		Avatar:       u.Avatar,
		Gender:       u.Gender,
		College:      u.College,
		Grade:        u.Grade,
		Major:        u.Major,
		Organization: u.Organization,
	}
}

func BatchToOuterUser(users []*model.User) []*OuterUser {
	if len(users) == 0 {
		return nil
	}

	outerUsers := make([]*OuterUser, len(users))
	for i, user := range users {
		outerUsers[i] = ToOuterUser(user)
	}
	return outerUsers
}

func Number2Email(number string) string {
	return number + "@m.scnu.edu.cn"
}
