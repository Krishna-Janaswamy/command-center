package virtualization

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3BlobStore struct {
	client *s3.Client
	bucket string
}

func NewConfiguredBlobStore(ctx context.Context) BlobStore {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return NewMemoryBlobStore()
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return NewMemoryBlobStore()
	}
	endpoint := os.Getenv("AWS_ENDPOINT")
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		if endpoint != "" {
			options.BaseEndpoint = aws.String(endpoint)
			options.UsePathStyle = true
		}
	})
	return &S3BlobStore{client: client, bucket: bucket}
}

func (s *S3BlobStore) Put(key, value string) error {
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), Body: bytes.NewBufferString(value)})
	return err
}

func (s *S3BlobStore) Get(key string) (string, error) {
	result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return "", err
	}
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	return string(data), err
}
