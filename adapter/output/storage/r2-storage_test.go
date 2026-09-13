package storage

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3ClientStub struct {
	put    func(*s3.PutObjectInput) error
	get    func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	delete func(*s3.DeleteObjectInput) error
}

type trackedReadCloser struct {
	io.Reader
	reads  int
	closed bool
}

func (r *trackedReadCloser) Read(p []byte) (int, error) {
	r.reads++
	return r.Reader.Read(p)
}

func (r *trackedReadCloser) Close() error {
	r.closed = true
	return nil
}

func (s s3ClientStub) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, s.put(input)
}
func (s s3ClientStub) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return s.get(input)
}
func (s s3ClientStub) DeleteObject(_ context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, s.delete(input)
}

func TestR2StorageUsesPrivateBucketAndObjectKey(t *testing.T) {
	data := []byte("imagem")
	body := &trackedReadCloser{Reader: bytes.NewReader(data)}
	client := s3ClientStub{
		put: func(input *s3.PutObjectInput) error {
			if aws.ToString(input.Bucket) != "bucket" || aws.ToString(input.Key) != "7/7/FAZENDA/hash" {
				t.Fatalf("destino inesperado: bucket=%s key=%s", aws.ToString(input.Bucket), aws.ToString(input.Key))
			}
			if aws.ToString(input.ContentType) != "image/png" {
				t.Fatalf("content type=%s", aws.ToString(input.ContentType))
			}
			got, err := io.ReadAll(input.Body)
			if err != nil || !bytes.Equal(got, data) {
				t.Fatalf("body=%q err=%v", got, err)
			}
			return nil
		},
		get: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			if aws.ToString(input.Key) != "7/7/FAZENDA/hash" {
				t.Fatalf("key=%s", aws.ToString(input.Key))
			}
			return &s3.GetObjectOutput{Body: body}, nil
		},
		delete: func(input *s3.DeleteObjectInput) error {
			if aws.ToString(input.Key) != "7/7/FAZENDA/hash" {
				t.Fatalf("key=%s", aws.ToString(input.Key))
			}
			return nil
		},
	}
	storage := &R2Storage{client: client, bucket: "bucket"}
	if err := storage.Upload(context.Background(), "7/7/FAZENDA/hash", data, "image/png"); err != nil {
		t.Fatal(err)
	}
	reader, err := storage.Download(context.Background(), "7/7/FAZENDA/hash")
	if err != nil {
		t.Fatal(err)
	}
	if body.reads != 0 {
		t.Fatalf("download leu o objeto inteiro antes do streaming: %d leituras", body.reads)
	}
	got, err := io.ReadAll(reader)
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("data=%q err=%v", got, err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	if !body.closed {
		t.Fatal("corpo do R2 não foi fechado")
	}
	if err := storage.Delete(context.Background(), "7/7/FAZENDA/hash"); err != nil {
		t.Fatal(err)
	}
}

func TestNewR2StorageRequiresAllConfiguration(t *testing.T) {
	if _, err := NewR2Storage("", "access", "secret", "bucket"); err == nil {
		t.Fatal("configuração incompleta deveria falhar")
	}
}
