package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"traffic-backend/internal/model"
)

func FetchTrafficEvents(ctx context.Context) ([]model.TrafficEvent, error) {
	url := os.Getenv("INCOMING_API_URL")

	if url == "" {
		return nil, fmt.Errorf("INCOMING_API_URL is not configured")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call incoming API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"incoming API returned status %d",
			resp.StatusCode,
		)
	}

	var events []model.TrafficEvent

	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode incoming API response: %w", err)
	}

	return events, nil
}