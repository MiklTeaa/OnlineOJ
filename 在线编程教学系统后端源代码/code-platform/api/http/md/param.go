package md

import (
	"strconv"
	"strings"

	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func parseIDToUint64AndCheckPositive(c *gin.Context, IDParam string) (uint64, bool) {
	ID, err := strconv.ParseUint(IDParam, 10, 64)
	if err != nil {
		httpx.AbortBadParamsErr(c, "ID is not a number")
		return 0, false
	}
	if ID <= 0 {
		httpx.AbortBadParamsErr(c, "ID is negative")
		return 0, false
	}
	return ID, true
}

func CheckParamID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		IDParam := c.Param(tag)
		if strings.TrimSpace(IDParam) == "" {
			httpx.AbortGetParamsErr(c, "ID shouldn't not be empty")
			return
		}

		ID, ok := parseIDToUint64AndCheckPositive(c, IDParam)
		if !ok {
			return
		}

		c.Set(tag, ID)
	}
}

func CheckQueryID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		IDParam, ok := c.GetQuery(tag)
		if !ok {
			httpx.AbortGetParamsErr(c, "Get ID failed")
			return
		}
		if strings.TrimSpace(IDParam) == "" {
			httpx.AbortBadParamsErr(c, "ID shouldn't not be empty")
			return
		}

		ID, ok := parseIDToUint64AndCheckPositive(c, IDParam)
		if !ok {
			return
		}

		c.Set(tag, ID)
	}
}

func CheckFormID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		IDParam, ok := c.GetPostForm(tag)
		if !ok {
			httpx.AbortGetParamsErr(c, "Get ID failed")
			return
		}
		if strings.TrimSpace(IDParam) == "" {
			httpx.AbortBadParamsErr(c, "ID shouldn't not be empty")
			return
		}

		ID, ok := parseIDToUint64AndCheckPositive(c, IDParam)
		if !ok {
			return
		}

		c.Set(tag, ID)
	}
}

func CheckJSONID(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		m := make(map[string]uint64, 1)
		if err := c.ShouldBindWith(&m, jsonx.SonicDecoder); err != nil {
			httpx.AbortGetParamsErr(c, "Get ID failed")
			return
		}

		var (
			ID uint64
			ok bool
		)
		for k, v := range m {
			if strings.EqualFold(k, tag) {
				ID = v
				ok = true
				break
			}
		}
		if !ok {
			httpx.AbortGetParamsErr(c, "Get ID failed")
			return
		}

		if ID <= 0 {
			httpx.AbortBadParamsErr(c, "ID is negative")
			return
		}

		c.Set(tag, ID)
	}
}

func CheckFormInt(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dataParam, ok := c.GetPostForm(tag)
		if !ok {
			httpx.AbortGetParamsErr(c, "Get data failed")
			return
		}
		if strings.TrimSpace(dataParam) == "" {
			httpx.AbortBadParamsErr(c, "data shouldn't not be empty")
			return
		}

		data, err := strconv.Atoi(dataParam)
		if err != nil {
			httpx.AbortBadParamsErr(c, "%s is not a number", dataParam)
			return
		}

		c.Set(tag, data)
	}
}
