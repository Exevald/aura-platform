package objectstore

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"filestorage/internal/filestorage/app"
	"filestorage/internal/filestorage/config"
)

type store struct {
	client *minio.Client
	bucket string
}

func New(ctx context.Context, c config.S3) (app.ObjectStore, error) {
	client, e := minio.New(c.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""), Secure: c.Secure})
	if e != nil {
		return nil, e
	}
	exists, e := client.BucketExists(ctx, c.Bucket)
	if e != nil {
		return nil, e
	}
	if !exists {
		if e = client.MakeBucket(ctx, c.Bucket, minio.MakeBucketOptions{}); e != nil {
			return nil, e
		}
	}
	return &store{client, c.Bucket}, nil
}

func (s *store) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	info, e := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return info.Size, e
}

func (s *store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	object, e := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if e != nil {
		return nil, e
	}
	if _, e = object.Stat(); e != nil {
		_ = object.Close()
		return nil, e
	}
	return object, nil
}

func (s *store) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
