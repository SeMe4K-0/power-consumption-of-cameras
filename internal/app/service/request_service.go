package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"

	"github.com/sirupsen/logrus"
)

// AsyncTokenValue - значение псевдо-токена, которым обмениваются сервисы.
const AsyncTokenValue = "a3f8b2c9d1e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1"

// AsyncServiceURL - URL асинхронного сервиса (Django).
const AsyncServiceURL = "http://localhost:8001/calculate-consumption"

type RequestService struct {
	repo *repository.Repository
}

func NewRequestService(repo *repository.Repository) *RequestService {
	return &RequestService{repo: repo}
}

func (s *RequestService) CompleteRequest(id uint, approve bool, moderatorID uint) error {
	status := ds.RequestStatusRejected
	if approve {
		status = ds.RequestStatusCompleted
	}

	if err := s.repo.UpdateRequestStatus(id, status, &moderatorID); err != nil {
		return err
	}

	if approve && status == ds.RequestStatusCompleted {
		go s.callAsyncService(id)
	}

	return nil
}

func (s *RequestService) ApplyAsyncResult(requestID uint, results []CalculationResult) error {
	for _, result := range results {
		if err := s.repo.UpdateCalculationMonthlyCost(result.CalculationID, result.MonthlyCost); err != nil {
			return err
		}
	}
	return nil
}

type CalculationResult struct {
	CalculationID uint     `json:"calculation_id"`
	MonthlyCost   *float64 `json:"monthly_cost"`
}

func (s *RequestService) callAsyncService(requestID uint) {
	calculations, err := s.repo.GetCalculationsForRequest(requestID)
	if err != nil {
		logrus.Errorf("failed to get calculations for async service: %v", err)
		return
	}

	calcData := make([]map[string]interface{}, 0, len(calculations))
	for _, calc := range calculations {
		calcData = append(calcData, map[string]interface{}{
			"id":    calc.ID,
			"power": calc.Power,
		})
	}

	body := map[string]interface{}{
		"request_id":  requestID,
		"calculations": calcData,
	}

	data, err := json.Marshal(body)
	if err != nil {
		logrus.Errorf("failed to marshal async body: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, AsyncServiceURL, bytes.NewReader(data))
	if err != nil {
		logrus.Errorf("failed to create async request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("failed to call async service: %v", err)
		return
	}
	_ = resp.Body.Close()
}
