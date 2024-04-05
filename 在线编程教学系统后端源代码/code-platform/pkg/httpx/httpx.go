package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AbortInternalErr(c *gin.Context) {
	c.Abort()
	c.String(http.StatusInternalServerError, "服务器开小差了")
}

func AbortGetParamsErr(c *gin.Context, format string, values ...interface{}) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		NewErrCode(ErrCodeFailToGetParams, format, values...),
	)
}

func AbortBadParamsErr(c *gin.Context, format string, values ...interface{}) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		NewErrCode(ErrCodeBadParams, format, values...),
	)
}

func AbortFailToAuth(c *gin.Context, format string, values ...interface{}) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		NewErrCode(ErrCodeFailToAuth, format, values...),
	)
}

func AbortUnauthorized(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

func AbortForbidden(c *gin.Context) {
	c.AbortWithStatus(http.StatusForbidden)
}

func AbortInvalidLength(c *gin.Context, format string, values ...interface{}) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		NewErrCode(ErrCodeInvalidLength, format, values...),
	)
}

func AbortUnsupportFileType(c *gin.Context, format string, values ...interface{}) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		NewErrCode(ErrCodeUnsupportFileType, format, values...),
	)
}

func AbortMailUserNotFound(c *gin.Context, format string, values ...interface{}) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		NewErrCode(ErrCodeNotFound, format, values...),
	)
}

func AbortNotFound(c *gin.Context, format string, values ...interface{}) {
	c.AbortWithStatusJSON(
		http.StatusBadRequest,
		NewErrCode(ErrCodeNotFound, format, values...),
	)
}
