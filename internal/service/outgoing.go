package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"traffic-backend/internal/model"
)

func PostFrontendTraffic(response model.FrontendTrafficResponse) error {
	url := os.Getenv("FRONTEND_API_URL")

	if url == "" {
		return fmt.Errorf("FRONTEND_API_URL is not configured")
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal frontend response: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create outgoing request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send outgoing request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(
			"outgoing API returned status %d",
			resp.StatusCode,
		)
	}

	return nil
}