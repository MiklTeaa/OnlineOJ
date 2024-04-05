package file

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code-platform/log"
	"code-platform/pkg/errorx"
	"code-platform/pkg/imagex"
	"code-platform/pkg/transactionx"
	"code-platform/repository"
	"code-platform/repository/rdb/model"
	"code-platform/storage"

	"github.com/minio/minio-go/v7"
)

type FileService struct {
	Dao    *repository.Dao
	Logger *log.Logger
}

func NewFileService(dao *repository.Dao, logger *log.Logger) *FileService {
	return &FileService{
		Dao:    dao,
		Logger: logger,
	}
}

func getUploadFileName(fileName string) string {
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return fileName + "-" + strconv.FormatInt(time.Now().UnixMicro(), 10)
}

func (f *FileService) uploading(ctx context.Context, userID uint64, reader io.Reader, size int64, bucketName string, contentType, fileName, ext string) (url string, err error) {
	minioClient := f.Dao.Storage.Minio
	uploadName := getUploadFileName(fileName)
	objectName := uploadName + ext
	task := func(ctx context.Context, tx storage.RDBClient) error {
		uploadObj := &model.Upload{
			UserID:     userID,
			BucketName: bucketName,
			ObjectName: objectName,
			CreatedAt:  time.Now(),
		}

		if err := uploadObj.Insert(ctx, tx); err != nil {
			return err
		}

		if _, err := minioClient.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
			return err
		}
		return nil
	}

	if err := transactionx.DoTransaction(ctx, f.Dao.Storage, f.Logger, task, &sql.TxOptions{Isolation: sql.LevelReadCommitted}); err != nil {
		return "", err
	}

	url = fmt.Sprintf(minioClient.URLFormat(), bucketName, objectName)
	return url, nil
}

func (f *FileService) UploadPicture(ctx context.Context, userID uint64, contentType string, width int, fileSize int64, fileName, ext string, file multipart.File) (string, error) {
	var imageType imagex.ImageType
	switch contentType {
	case "image/gif":
		imageType = imagex.GIF
	case "image/png":
		imageType = imagex.PNG
	case "image/jpg":
		imageType = imagex.JPEG
	}

	buf, size, err := imagex.Resize(file, width, 0, imageType)
	if err != nil {
		f.Logger.Error(err, "resize image failed")
		return "", errorx.InternalErr(err)
	}

	url, err := f.uploading(ctx, userID, buf, size, f.Dao.Storage.Minio.PictureBucketName(), contentType, fileName, ext)
	if err != nil {
		f.Logger.Errorf(err, "uploading to bucket %q failed", f.Dao.Storage.Minio.PictureBucketName())
		return "", errorx.InternalErr(err)
	}

	return url, nil
}

func (f *FileService) UploadPDF(ctx context.Context, userID uint64, fileSize int64, fileName, ext string, file multipart.File) (string, error) {
	const contentType = "application/pdf"
	url, err := f.uploading(ctx, userID, file, fileSize, f.Dao.Storage.Minio.ReportBucketName(), contentType, fileName, ext)
	if err != nil {
		f.Logger.Errorf(err, "uploading to bucket %q failed", f.Dao.Storage.Minio.ReportBucketName())
		return "", errorx.InternalErr(err)
	}
	return url, nil
}

func (f *FileService) UploadAttachment(ctx context.Context, userID uint64, fileSize int64, fileName, ext string, file multipart.File) (string, error) {
	const contentType = "application/octet-stream"
	url, err := f.uploading(ctx, userID, file, fileSize, f.Dao.Storage.Minio.AttachmentBucketName(), contentType, fileName, ext)
	if err != nil {
		f.Logger.Errorf(err, "uploading to bucket %q failed", f.Dao.Storage.Minio.AttachmentBucketName())
		return "", errorx.InternalErr(err)
	}
	return url, nil
}

func (f *FileService) UploadVideo(ctx context.Context, userID uint64, fileSize int64, contentType, fileName, ext string, file multipart.File) (string, error) {
	url, err := f.uploading(ctx, userID, file, fileSize, f.Dao.Storage.Minio.VideoBucketName(), contentType, fileName, ext)
	if err != nil {
		f.Logger.Errorf(err, "uploading to bucket %q failed", f.Dao.Storage.Minio.VideoBucketName())
		return "", errorx.InternalErr(err)
	}
	return url, nil
}

func (f *FileService) MIMEHeaderToBytes(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		f.Logger.Errorf(err, "open file via FileHeader %+v failed", fileHeader)
		return nil, errorx.InternalErr(err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		f.Logger.Error(err, "read all data failed")
		return nil, errorx.InternalErr(errors.New("ReadAllData failed"))
	}
	return data, nil
}

func (f *FileService) MIMEHeaderToFile(fileHeader *multipart.FileHeader) (multipart.File, error) {
	file, err := fileHeader.Open()
	if err != nil {
		f.Logger.Errorf(err, "open file via FileHeader %+v failed", fileHeader)
		return nil, errorx.InternalErr(err)
	}
	return file, nil
}
