package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
)

type MinIOStore struct {
	client *minio.Client
	bucket string
}

type MinIOConfig struct {
	Client *minio.Client
	Bucket string
}

func NewMinIOStore(cfg MinIOConfig) (*MinIOStore, error) {
	ctx := context.Background()

	exists, err := cfg.Client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := cfg.Client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinIOStore{
		client: cfg.Client,
		bucket: cfg.Bucket,
	}, nil
}
