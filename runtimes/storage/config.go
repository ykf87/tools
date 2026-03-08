package storage

import "context"

type Config struct {
	Type string

	Ctx context.Context

	// local
	LocalPath string
	LocalURL  string

	// minio
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	URL       string
}
