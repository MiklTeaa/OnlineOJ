package md

import (
	"net/http"
	"strconv"
	"strings"

	"code-platform/pkg/stringx"

	"github.com/gin-gonic/gin"
)

type corsConfig struct {
	allowOrigins     []string
	allowMethods     []string
	allowHeaders     []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int
}

func (config *corsConfig) build() gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin := c.GetHeader("Origin"); origin != "" {
			if stringx.SliceContains(config.allowOrigins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				if !stringx.SliceContainsFold(config.allowOrigins, origin) {
					c.AbortWithStatus(http.StatusNoContent)
					return
				}
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		if config.allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// preflight request
		if c.Request.Method == http.MethodOptions && c.Request.Header.Get("Access-Control-Request-Method") != "" {
			method := c.GetHeader("Access-Control-Request-Method")
			if !stringx.SliceContains(config.allowMethods, strings.ToUpper(method)) {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Header("Access-Control-Allow-Methods", method)

			if requestHeadersString := c.GetHeader("Access-Control-Request-Headers"); requestHeadersString != "" {
				requestHeaders := strings.Split(requestHeadersString, ",")

				for _, requestHeader := range requestHeaders {
					if !stringx.SliceContainsFold(config.allowHeaders, strings.TrimSpace(requestHeader)) {
						c.AbortWithStatus(http.StatusNoContent)
						return
					}
				}
				c.Header("Access-Control-Allow-Headers", requestHeadersString)
			}

			if len(config.exposedHeaders) != 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(config.exposedHeaders, ", "))
			}

			if config.maxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(config.maxAge))
			}

			if config.allowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			c.AbortWithStatus(http.StatusNoContent)
		}
	}
}

func CORS() gin.HandlerFunc {
	config := &corsConfig{
		allowOrigins:     []string{"http://localhost:3600", "http://127.0.0.1:3600", "http://175.178.37.132:3600"},
		allowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		allowHeaders:     []string{"Origin", "Content-Type", "Accept", "User-Agent", "Cookie", "Authorization", "X-Auth-Token", "X-Requested-With"},
		exposedHeaders:   nil,
		allowCredentials: true,
		maxAge:           60 * 60,
	}
	return config.build()
}
