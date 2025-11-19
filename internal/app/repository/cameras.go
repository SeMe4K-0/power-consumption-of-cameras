package repository

import (
	"awesomeProject/internal/app/ds"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func (r *Repository) GetCameras(cameraName string) ([]ds.Cameras, error) {
	var cameras []ds.Cameras
	query := r.db.Where("is_deleted = ?", false)

	if cameraName != "" {
		query = query.Where("name ILIKE ?", "%"+cameraName+"%")
	}

	err := query.Find(&cameras).Error
	return cameras, err
}

func (r *Repository) GetCamera(id uint) (ds.Cameras, error) {
	var camera ds.Cameras
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&camera).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Cameras{}, errors.New("camera not found")
		}
		return ds.Cameras{}, err
	}
	return camera, nil
}

func (r *Repository) CreateCamera(camera ds.Cameras) (ds.Cameras, error) {
	err := r.db.Create(&camera).Error
	return camera, err
}

func (r *Repository) UpdateCamera(id uint, camera ds.Cameras) error {
	tx := r.db.Model(&ds.Cameras{}).Where("id = ? AND is_deleted = ?", id, false).Updates(camera)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("camera not found")
	}
	return nil
}

func (r *Repository) DeleteCamera(id uint) error {
	tx := r.db.Model(&ds.Cameras{}).Where("id = ?", id).Update("is_deleted", true)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("camera not found")
	}
	return nil
}

func (r *Repository) UpdateCameraImage(id uint, imagePath string) error {
	tx := r.db.Model(&ds.Cameras{}).Where("id = ? AND is_deleted = ?", id, false).Update("image", imagePath)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("camera not found")
	}
	return nil
}

func (r *Repository) UploadFileToMinIO(ctx context.Context, fileName string, fileReader io.Reader, fileSize int64, contentType string) error {
	// Проверяем доступность MinIO
	_, err := r.minio.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("MinIO is not accessible: %v", err)
	}

	// Загружаем файл
	_, err = r.minio.PutObject(ctx, r.bucket, fileName, fileReader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (r *Repository) DeleteFileFromMinIO(ctx context.Context, fileName string) error {
	err := r.minio.RemoveObject(ctx, r.bucket, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}
	return nil
}
