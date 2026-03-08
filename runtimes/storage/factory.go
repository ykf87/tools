package storage

import (
	"errors"
)

func New(cfg Config) (Storage, error) {

	switch cfg.Type {
	case "local":
		return newLocal(cfg)

	case "minio":
		return newMinio(cfg)
	}

	return nil, errors.New("unknown storage type")
}
