package md

import (
	"mime/multipart"
	"path/filepath"
	"strings"

	"code-platform/log"
	"code-platform/pkg/httpx"
	"code-platform/pkg/stringx"

	"github.com/gin-gonic/gin"
)

func CheckFileHeader(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader, err := c.FormFile(tag)
		if err != nil {
			log.Debugf("failed to form file by tag %q for error %v", tag, err)
			httpx.AbortGetParamsErr(c, "Fail to get file in upload attachment request")
			return
		}

		c.Set(tag, fileHeader)
	}
}

func GetFileHeader(c *gin.Context, tag string) *multipart.FileHeader {
	fileHeader, _ := c.Get(tag)
	return fileHeader.(*multipart.FileHeader)
}

// CheckFileExt should be called after calling CheckFileHeader
func CheckFileExt(tag string, types []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader := GetFileHeader(c, tag)

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !stringx.SliceContains(types, ext[1:]) {
			httpx.AbortUnsupportFileType(c, "unsupported file type")
			return
		}
		c.Set(KeyExtName, ext)
	}
}

// SetImageType should be called after calling CheckFileExt
func SetImageType() gin.HandlerFunc {
	return setContentType("image/")
}

// SetVideoType should be called after calling CheckFileExt
func SetVideoType() gin.HandlerFunc {
	return setContentType("video/")
}

func setContentType(contentTypePrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ext := c.GetString(KeyExtName)
		c.Set(KeyContentType, contentTypePrefix+ext[1:])
	}
}
