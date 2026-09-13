package output

import (
	"context"
	"io"
)

type StoragePort interface {
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	Upload(ctx context.Context, objectKey string, data []byte, contentType string) error
	Delete(ctx context.Context, objectKey string) error
}
