package md

import (
	"errors"
	"net/http"
	"strings"

	xhttp "code-platform/api/http"
	"code-platform/pkg/errorx"

	"github.com/gin-gonic/gin"
)

// 与数据库字段严格对应，注意一定不要随意调换顺序，保持递增即可
// 需要添加角色在最下面加
const (
	// RoleStudent 0
	RoleStudent = iota
	// RoleTeacher 1
	RoleTeacher
	// RoleAdmin 2
	RoleAdmin
	// Add in it
)

func RestoreUserStat(srv *xhttp.UnionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token := parts[1]
		ctx := c.Request.Context()
		userID, userRole, err := srv.UserService.ParseUserStatFromToken(ctx, token)
		switch err {
		case nil:
		case errorx.ErrIsNotFound, errorx.ErrFailToAuth:
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		default:
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Set(KeyUserID, userID)
		c.Set(KeyUserRole, userRole)
	}
}

func getUserRole(c *gin.Context, s *xhttp.UnionService) (uint16, error) {
	userRoleIFace, exists := c.Get(KeyUserRole)
	if !exists {
		return 0, errors.New("userRole is not exists")
	}
	return userRoleIFace.(uint16), nil
}

func RequireAdmin(s *xhttp.UnionService) gin.HandlerFunc {
	return requireRole(s, RoleAdmin)
}

func RequireTeacher(s *xhttp.UnionService) gin.HandlerFunc {
	return requireRole(s, RoleTeacher)
}

func RequireStudent(s *xhttp.UnionService) gin.HandlerFunc {
	return requireRole(s, RoleStudent)
}

func requireRole(s *xhttp.UnionService, role uint16) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, err := getUserRole(c, s)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if userRole != role {
			var tips string
			switch role {
			case 0:
				tips = "you aren't a student"
			case 1:
				tips = "you aren't a teacher"
			case 2:
				tips = "you aren't an admin"
			}
			c.Abort()
			c.String(http.StatusForbidden, tips)
			return
		}
	}
}
