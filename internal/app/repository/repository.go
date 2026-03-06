package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"it-maintenance-backend/internal/app/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

const cameraSelectColumns = `
	id, name, description, status, image_url, price, power, type, resolution, night_vision, created_at, updated_at
`

const orderSelectColumns = `
	id, status, created_at, creator_id, formation_date, completion_date, moderator_id, client_name, project_name, calculated_field
`

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) GetDB() *gorm.DB {
	return r.db
}

func scanCameraRow(rows *sql.Rows) (models.Camera, error) {
	var camera models.Camera
	var imageURL sql.NullString

	err := rows.Scan(
		&camera.ID,
		&camera.Name,
		&camera.Description,
		&camera.Status,
		&imageURL,
		&camera.Price,
		&camera.Power,
		&camera.Type,
		&camera.Resolution,
		&camera.NightVision,
		&camera.CreatedAt,
		&camera.UpdatedAt,
	)
	if err != nil {
		return models.Camera{}, err
	}

	if imageURL.Valid {
		camera.ImageURL = &imageURL.String
	}

	return camera, nil
}

func scanOrderRow(rows *sql.Rows) (models.SurveillanceOrder, error) {
	var order models.SurveillanceOrder
	var formationDate sql.NullTime
	var completionDate sql.NullTime
	var moderatorID sql.NullInt64

	err := rows.Scan(
		&order.ID,
		&order.Status,
		&order.CreatedAt,
		&order.CreatorID,
		&formationDate,
		&completionDate,
		&moderatorID,
		&order.ClientName,
		&order.ProjectName,
		&order.CalculatedField,
	)
	if err != nil {
		return models.SurveillanceOrder{}, err
	}

	if formationDate.Valid {
		order.FormationDate = &formationDate.Time
	}
	if completionDate.Valid {
		order.CompletionDate = &completionDate.Time
	}
	if moderatorID.Valid {
		value := uint(moderatorID.Int64)
		order.ModeratorID = &value
	}

	return order, nil
}

func (r *Repository) GetCameras() ([]models.Camera, error) {
	var cameras []models.Camera

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			fmt.Sprintf("DECLARE cameras_cursor NO SCROLL CURSOR FOR SELECT %s FROM public.cameras ORDER BY id", cameraSelectColumns),
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE cameras_cursor")

		rows, err := tx.Raw("FETCH ALL FROM cameras_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			camera, scanErr := scanCameraRow(rows)
			if scanErr != nil {
				return scanErr
			}
			cameras = append(cameras, camera)
		}
		return rows.Err()
	})

	return cameras, err
}

func (r *Repository) GetCameraByID(id uint) (models.Camera, error) {
	var camera models.Camera
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			fmt.Sprintf("DECLARE camera_by_id_cursor NO SCROLL CURSOR FOR SELECT %s FROM public.cameras WHERE id = $1", cameraSelectColumns),
			id,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE camera_by_id_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM camera_by_id_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if !rows.Next() {
			return gorm.ErrRecordNotFound
		}

		camera, err = scanCameraRow(rows)
		if err != nil {
			return err
		}

		return rows.Err()
	})

	return camera, err
}

func (r *Repository) GetCamerasBySearch(query string) ([]models.Camera, error) {
	var cameras []models.Camera
	searchQuery := "%" + query + "%"

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			fmt.Sprintf("DECLARE camera_search_cursor NO SCROLL CURSOR FOR SELECT %s FROM public.cameras WHERE (name ILIKE $1 OR description ILIKE $2 OR type ILIKE $3) ORDER BY id", cameraSelectColumns),
			searchQuery,
			searchQuery,
			searchQuery,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE camera_search_cursor")

		rows, err := tx.Raw("FETCH ALL FROM camera_search_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			camera, scanErr := scanCameraRow(rows)
			if scanErr != nil {
				return scanErr
			}
			cameras = append(cameras, camera)
		}
		return rows.Err()
	})

	return cameras, err
}

func (r *Repository) HasDraftOrder(userID uint) bool {
	count := int64(0)
	_ = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DECLARE draft_order_count_cursor NO SCROLL CURSOR FOR SELECT COUNT(*) FROM surveillance_orders WHERE creator_id = $1 AND status = $2",
			userID,
			models.OrderStatusDraft,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE draft_order_count_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM draft_order_count_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return err
			}
		}

		return rows.Err()
	})

	return count > 0
}

func (r *Repository) GetCurrentOrder(userID uint) (models.SurveillanceOrder, error) {
	var order models.SurveillanceOrder
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			fmt.Sprintf("DECLARE current_order_cursor NO SCROLL CURSOR FOR SELECT %s FROM surveillance_orders WHERE creator_id = $1 AND status = $2 ORDER BY id LIMIT 1", orderSelectColumns),
			userID,
			models.OrderStatusDraft,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE current_order_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM current_order_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if !rows.Next() {
			return gorm.ErrRecordNotFound
		}

		order, err = scanOrderRow(rows)
		if err != nil {
			return err
		}

		return rows.Err()
	})
	if err != nil {
		return models.SurveillanceOrder{}, err
	}

	orderCameras, err := r.GetOrderCameras(order.ID)
	if err != nil {
		return models.SurveillanceOrder{}, err
	}
	order.OrderCameras = orderCameras

	return order, nil
}

func (r *Repository) CreateOrder(userID uint, clientName, projectName string) (models.SurveillanceOrder, error) {
	order := models.SurveillanceOrder{}
	now := time.Now()

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`INSERT INTO surveillance_orders (status, created_at, creator_id, client_name, project_name, calculated_field)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			models.OrderStatusDraft,
			now,
			userID,
			clientName,
			projectName,
			0,
		).Error; err != nil {
			return err
		}

		if err := tx.Exec(
			fmt.Sprintf("DECLARE created_order_cursor NO SCROLL CURSOR FOR SELECT %s FROM surveillance_orders WHERE creator_id = $1 AND status = $2 ORDER BY id DESC LIMIT 1", orderSelectColumns),
			userID,
			models.OrderStatusDraft,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE created_order_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM created_order_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if !rows.Next() {
			return gorm.ErrRecordNotFound
		}

		order, err = scanOrderRow(rows)
		if err != nil {
			return err
		}
		return rows.Err()
	})

	return order, err
}

func (r *Repository) AddCameraToOrder(orderID, cameraID uint, quantity int, comment, other string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		existingQuantity := 0
		hasRecord := false

		if err := tx.Exec(
			"DECLARE order_camera_exists_cursor NO SCROLL CURSOR FOR SELECT quantity FROM order_cameras WHERE order_id = $1 AND camera_id = $2",
			orderID,
			cameraID,
		).Error; err != nil {
			return err
		}

		rows, err := tx.Raw("FETCH NEXT FROM order_camera_exists_cursor").Rows()
		if err != nil {
			tx.Exec("CLOSE order_camera_exists_cursor")
			return err
		}

		if rows.Next() {
			hasRecord = true
			if err := rows.Scan(&existingQuantity); err != nil {
				rows.Close()
				tx.Exec("CLOSE order_camera_exists_cursor")
				return err
			}
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			tx.Exec("CLOSE order_camera_exists_cursor")
			return err
		}
		rows.Close()
		tx.Exec("CLOSE order_camera_exists_cursor")

		if !hasRecord {
			return tx.Exec(
				`INSERT INTO order_cameras (order_id, camera_id, quantity, order_num, is_main, comment, other)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				orderID,
				cameraID,
				quantity,
				1,
				false,
				comment,
				other,
			).Error
		}

		return tx.Exec(
			"UPDATE order_cameras SET quantity = $1, comment = $2, other = $3 WHERE order_id = $4 AND camera_id = $5",
			existingQuantity+quantity,
			comment,
			other,
			orderID,
			cameraID,
		).Error
	})
}

func (r *Repository) GetOrderFormData(userID uint) (models.OrderFormData, error) {
	order, err := r.GetCurrentOrder(userID)
	if err != nil {
		return models.OrderFormData{}, err
	}

	var availableCameras []models.Camera
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			fmt.Sprintf("DECLARE active_cameras_cursor NO SCROLL CURSOR FOR SELECT %s FROM cameras WHERE status = $1 ORDER BY id", cameraSelectColumns),
			models.CameraStatusActive,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE active_cameras_cursor")

		rows, err := tx.Raw("FETCH ALL FROM active_cameras_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			camera, scanErr := scanCameraRow(rows)
			if scanErr != nil {
				return scanErr
			}
			availableCameras = append(availableCameras, camera)
		}
		return rows.Err()
	})
	if err != nil {
		return models.OrderFormData{}, err
	}

	return models.OrderFormData{
		Order:            order,
		AvailableCameras: availableCameras,
		OrderCameras:     order.OrderCameras,
	}, nil
}

func (r *Repository) GetOrdersCount(userID uint) int64 {
	count := int64(0)
	_ = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DECLARE orders_count_cursor NO SCROLL CURSOR FOR SELECT COUNT(*) FROM surveillance_orders WHERE creator_id = $1",
			userID,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE orders_count_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM orders_count_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return err
			}
		}
		return rows.Err()
	})

	return count
}

func (r *Repository) GetCurrentOrderServicesCount(userID uint) int64 {
	orderID := r.GetFirstOrderID(userID)
	if orderID == 0 {
		return 0
	}
	return r.GetOrderServicesCount(orderID)
}

func (r *Repository) GetOrderServicesCount(orderID uint) int64 {
	count := int64(0)
	_ = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DECLARE order_services_count_cursor NO SCROLL CURSOR FOR SELECT COUNT(*) FROM order_cameras WHERE order_id = $1",
			orderID,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE order_services_count_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM order_services_count_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if rows.Next() {
			if err := rows.Scan(&count); err != nil {
				return err
			}
		}
		return rows.Err()
	})

	return count
}

func (r *Repository) CheckOrderAccess(orderID uint, userID uint) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DECLARE order_access_cursor NO SCROLL CURSOR FOR SELECT id FROM surveillance_orders WHERE id = $1 AND creator_id = $2",
			orderID,
			userID,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE order_access_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM order_access_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if !rows.Next() {
			return gorm.ErrRecordNotFound
		}
		return rows.Err()
	})
	if err != nil {
		return err
	}

	servicesCount := r.GetOrderServicesCount(orderID)
	if servicesCount == 0 {
		return fmt.Errorf("заявка пуста")
	}

	return nil
}

func (r *Repository) GetFirstOrderID(userID uint) uint {
	orderID := uint(0)
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			"DECLARE first_order_cursor NO SCROLL CURSOR FOR SELECT id FROM surveillance_orders WHERE creator_id = $1 AND status = $2 ORDER BY id LIMIT 1",
			userID,
			models.OrderStatusDraft,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE first_order_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM first_order_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if !rows.Next() {
			return gorm.ErrRecordNotFound
		}

		return rows.Scan(&orderID)
	})
	if err != nil {
		return 0
	}
	return orderID
}

func (r *Repository) GetOrderByID(id uint) (models.SurveillanceOrder, error) {
	var order models.SurveillanceOrder
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			fmt.Sprintf("DECLARE order_by_id_cursor NO SCROLL CURSOR FOR SELECT %s FROM surveillance_orders WHERE id = $1", orderSelectColumns),
			id,
		).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE order_by_id_cursor")

		rows, err := tx.Raw("FETCH NEXT FROM order_by_id_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		if !rows.Next() {
			return gorm.ErrRecordNotFound
		}

		order, err = scanOrderRow(rows)
		if err != nil {
			return err
		}
		return rows.Err()
	})

	if err != nil {
		return models.SurveillanceOrder{}, err
	}

	orderCameras, err := r.GetOrderCameras(order.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.SurveillanceOrder{}, err
	}
	order.OrderCameras = orderCameras
	return order, nil
}

func (r *Repository) GetOrderCameras(orderID uint) ([]models.OrderCamera, error) {
	var orderCameras []models.OrderCamera

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DECLARE order_cameras_cursor NO SCROLL CURSOR FOR
			SELECT
				oc.order_id, oc.camera_id, oc.quantity, oc.order_num, oc.is_main, oc.comment, oc.other,
				c.id, c.name, c.description, c.status, c.image_url, c.price, c.power, c.type, c.resolution, c.night_vision, c.created_at, c.updated_at
			FROM order_cameras oc
			JOIN cameras c ON oc.camera_id = c.id
			WHERE oc.order_id = $1
			ORDER BY oc.order_num
		`, orderID).Error; err != nil {
			return err
		}
		defer tx.Exec("CLOSE order_cameras_cursor")

		rows, err := tx.Raw("FETCH ALL FROM order_cameras_cursor").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var oc models.OrderCamera
			var c models.Camera
			var imageURL sql.NullString
			var comment sql.NullString
			var other sql.NullString

			err := rows.Scan(
				&oc.OrderID, &oc.CameraID, &oc.Quantity, &oc.OrderNum, &oc.IsMain, &comment, &other,
				&c.ID, &c.Name, &c.Description, &c.Status, &imageURL, &c.Price, &c.Power, &c.Type, &c.Resolution, &c.NightVision, &c.CreatedAt, &c.UpdatedAt,
			)
			if err != nil {
				return err
			}

			if imageURL.Valid {
				c.ImageURL = &imageURL.String
			}
			if comment.Valid {
				oc.Comment = comment.String
			}
			if other.Valid {
				oc.Other = other.String
			}

			oc.Camera = c
			orderCameras = append(orderCameras, oc)
		}

		return rows.Err()
	})

	return orderCameras, err
}

func (r *Repository) DeleteOrder(orderID uint) error {
	result := r.db.Exec("UPDATE surveillance_orders SET status = ?, completion_date = ? WHERE id = ?", models.OrderStatusDeleted, time.Now(), orderID)
	return result.Error
}
