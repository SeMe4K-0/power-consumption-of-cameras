package repository

import (
	"context"

	"awesomeProject/internal/app/ds"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	minio  *minio.Client
	bucket string
}

func New(dsn string, minioEndpoint, minioAccessKey, minioSecretKey, bucketName string, useSSL bool) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}


	err = db.AutoMigrate(
		&ds.User{},
		&ds.Cameras{},
		&ds.RequestCamerasCalculation{},
		&ds.CamerasCalculation{},
	)
	if err != nil {
		return nil, err
	}


	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}


	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}


	return &Repository{
		db:     db,
		minio:  minioClient,
		bucket: bucketName,
	}, nil
}
