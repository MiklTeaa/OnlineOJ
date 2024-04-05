package web

import (
	"net/http"

	"code-platform/api/http/md"
	"code-platform/pkg/httpx"
	"code-platform/pkg/jsonx"

	"github.com/gin-gonic/gin"
)

func makeUploadPicture(tag string, widthTag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader := md.GetFileHeader(c, tag)
		width := c.GetInt(widthTag)

		file, err := srv.FileService.MIMEHeaderToFile(fileHeader)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}
		defer file.Close()

		contentType := c.GetString(md.KeyContentType)
		ext := c.GetString(md.KeyExtName)

		ctx := c.Request.Context()
		userID := c.GetUint64(md.KeyUserID)
		url, err := srv.FileService.UploadPicture(ctx, userID, contentType, width, fileHeader.Size, fileHeader.Filename, ext, file)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url})))

	}
}

func makeUploadReport(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader := md.GetFileHeader(c, tag)

		file, err := srv.FileService.MIMEHeaderToFile(fileHeader)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}
		defer file.Close()

		ext := c.GetString(md.KeyExtName)

		ctx := c.Request.Context()
		userID := c.GetUint64(md.KeyUserID)
		url, err := srv.FileService.UploadPDF(ctx, userID, fileHeader.Size, fileHeader.Filename, ext, file)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url})))
	}
}

func makeUploadVideo(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader := md.GetFileHeader(c, tag)

		file, err := srv.FileService.MIMEHeaderToFile(fileHeader)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}
		defer file.Close()

		contentType := c.GetString(md.KeyContentType)
		ext := c.GetString(md.KeyExtName)

		ctx := c.Request.Context()
		userID := c.GetUint64(md.KeyUserID)
		url, err := srv.FileService.UploadVideo(ctx, userID, fileHeader.Size, contentType, fileHeader.Filename, ext, file)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url})))
	}
}

func makeUploadAttachment(tag string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader := md.GetFileHeader(c, tag)

		file, err := srv.FileService.MIMEHeaderToFile(fileHeader)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}
		defer file.Close()

		ext := c.GetString(md.KeyExtName)

		ctx := c.Request.Context()
		userID := c.GetUint64(md.KeyUserID)
		url, err := srv.FileService.UploadAttachment(ctx, userID, fileHeader.Size, fileHeader.Filename, ext, file)
		if err != nil {
			httpx.AbortInternalErr(c)
			return
		}

		c.Render(http.StatusOK, jsonx.NewSonicEncoder(httpx.NewJSONResponse(gin.H{"url": url})))
	}
}
