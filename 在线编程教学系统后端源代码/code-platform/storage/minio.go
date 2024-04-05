package storage

import (
	"context"
	"fmt"

	"code-platform/config"

	"github.com/bytedance/sonic"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	*minio.Client
	pictureBucketName    string
	reportBucketName     string
	attachmentBucketName string
	videoBucketName      string
	policyReadOnly       string
	policyWriteOnly      string
	policyReadWrite      string
	urlFormat            string
}

func MustInitMinioClient() MinioClient {
	endPoint := config.Minio.GetString("endPoint")
	accessKeyID := config.Minio.GetString("accessKeyID")
	secretAccessKey := config.Minio.GetString("secretAccessKey")
	urlPrefix := config.Minio.GetString("urlPrefix")

	client, err := minio.New(endPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}

	bucketNames := config.Minio.GetStringMapString("bucketName")
	minioClient := MinioClient{
		Client:               client,
		pictureBucketName:    bucketNames["picture"],
		reportBucketName:     bucketNames["report"],
		attachmentBucketName: bucketNames["attachment"],
		videoBucketName:      bucketNames["video"],
		policyReadOnly:       newPolicyToJSON(newPolicyReadOnly()),
		policyWriteOnly:      newPolicyToJSON(newPolicyWriteOnly()),
		policyReadWrite:      newPolicyToJSON(newPolicyReadWrite()),
		urlFormat:            urlPrefix + "/%s/%s",
	}

	ctx := context.TODO()
	for _, name := range []string{
		minioClient.pictureBucketName,
		minioClient.reportBucketName,
		minioClient.attachmentBucketName,
		minioClient.videoBucketName,
	} {
		if err := minioClient.newBucket(ctx, name, minioClient.policyReadOnly); err != nil {
			panic(err)
		}
	}

	return minioClient
}

func (m *MinioClient) newBucket(ctx context.Context, bucketName string, policy string) error {

	const region = "cn-south-1"

	exists, err := m.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		if err := m.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: region}); err != nil {
			return err
		}
		if err := m.SetBucketPolicy(ctx, bucketName, fmt.Sprintf(policy, bucketName, bucketName)); err != nil {
			return err
		}
	}
	return nil
}

func (m *MinioClient) PictureBucketName() string {
	return m.pictureBucketName
}

func (m *MinioClient) AttachmentBucketName() string {
	return m.attachmentBucketName
}

func (m *MinioClient) ReportBucketName() string {
	return m.reportBucketName
}

func (m *MinioClient) VideoBucketName() string {
	return m.videoBucketName
}

func (m *MinioClient) URLFormat() string {
	// add proto
	return "http://" + m.urlFormat
}

const minioVersion = "2012-10-17"

type principal struct {
	AWS []string `json:"AWS"`
}

type statement struct {
	Principal *principal `json:"Principal,omitempty"`
	Effect    string     `json:"Effect,omitempty"`
	Resource  string     `json:"Resource,omitempty"`
	Action    []string   `json:"Action,omitempty"`
}

type policy struct {
	Version    string      `json:"Version"`
	Statements []statement `json:"Statement"`
}

func newPolicyReadOnly() *policy {
	return &policy{
		Version: minioVersion,
		Statements: []statement{
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetBucketLocation",
					"s3:ListBucket",
				},
				Resource:  "arn:aws:s3:::%s",
				Principal: &principal{AWS: []string{"*"}},
			},
			{
				Effect:    "Allow",
				Action:    []string{"s3:GetObject"},
				Resource:  "arn:aws:s3:::%s/*",
				Principal: &principal{AWS: []string{"*"}},
			},
		},
	}
}

func newPolicyWriteOnly() *policy {
	return &policy{
		Version: minioVersion,
		Statements: []statement{
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetBucketLocation",
					"s3:ListBucketMultipartUploads",
				},
				Resource:  "arn:aws:s3:::%s",
				Principal: &principal{AWS: []string{"*"}},
			},
			{
				Effect: "Allow",
				Action: []string{
					"s3:AbortMultipartUpload",
					"s3:DeletePic",
					"s3:ListMultipartUploadParts",
					"s3:PutObject",
				},
				Resource:  "arn:aws:s3:::%s/*",
				Principal: &principal{AWS: []string{"*"}},
			},
		},
	}
}

func newPolicyToJSON(p *policy) string {
	data, err := sonic.Marshal(&p)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func newPolicyReadWrite() *policy {
	return &policy{
		Version: minioVersion,
		Statements: []statement{
			{
				Effect: "Allow",
				Action: []string{
					"s3:GetBucketLocation",
					"s3:ListBucketMultipartUploads",
					"s3.ListBucket",
				},
				Resource:  "arn:aws:s3:::%s",
				Principal: &principal{AWS: []string{"*"}},
			},
			{
				Effect: "Allow",
				Action: []string{
					"s3:AbortMultipartUpload",
					"s3:DeletePic",
					"s3:GetObject",
					"s3:ListMultipartUploadParts",
					"s3:PutObject",
				},
				Resource:  "arn:aws:s3:::%s/*",
				Principal: &principal{AWS: []string{"*"}},
			},
		},
	}
}
