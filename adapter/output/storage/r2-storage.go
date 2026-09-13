package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Client interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type R2Storage struct {
	client s3Client
	bucket string
}

func NewR2Storage(baseURL, accessKey, secretKey, bucket string) (*R2Storage, error) {
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(accessKey) == "" ||
		strings.TrimSpace(secretKey) == "" || strings.TrimSpace(bucket) == "" {
		return nil, errors.New("configuração do R2 incompleta")
	}
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(strings.TrimRight(baseURL, "/"))
		options.UsePathStyle = true
	})
	return &R2Storage{client: client, bucket: bucket}, nil
}

func (r *R2Storage) Upload(ctx context.Context, objectKey string, data []byte, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(objectKey),
		Body: bytes.NewReader(data), ContentType: aws.String(contentType),
	})
	return err
}

func (r *R2Storage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (r *R2Storage) Delete(ctx context.Context, objectKey string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(objectKey),
	})
	return err
}
