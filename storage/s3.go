package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

type S3Service struct {
	Bucket   string
	Uploader *manager.Uploader
	Client   *s3.Client
}

// NewS3Service creates a new S3 service
func NewS3Service(bucket string) *S3Service {

	if bucket == "" {
		log.Fatal().Msg("bucket cannot be empty")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load aws config")
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	_, err = client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find bucket")
	}

	return &S3Service{
		Bucket:   bucket,
		Uploader: uploader,
		Client:   client,
	}
}

// UploadFile uploads a file to the specified S3 bucket.
func (s *S3Service) UploadFile(key, contentType string, uploadFileReader io.Reader) error {

	putObjectInput := &s3.PutObjectInput{
		Bucket:             aws.String(s.Bucket),
		Key:                aws.String(key),
		Body:               uploadFileReader,
		ContentDisposition: aws.String("inline"),
		ContentType:        aws.String(contentType),
	}

	_, err := s.Uploader.Upload(context.TODO(), putObjectInput)
	if err != nil {
		return err
	}

	return nil
}

// UploadFile uploads a file to the specified S3 bucket.
func (s *S3Service) DeleteFile(key string) error {

	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.Client.DeleteObject(context.TODO(), deleteObjectInput)
	if err != nil {
		return err
	}

	return nil
}

// DoesKeyExist check that key is exist in bucket
func (s *S3Service) DoesKeyExist(key string) bool {

	headObjectInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.Client.HeadObject(context.Background(), headObjectInput)
	if err != nil {
		return false
	}

	return true
}
