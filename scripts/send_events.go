package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const API_URL = "http://localhost:8080/api/v1/traffic/events"

type TrafficEvent struct {
	EventID   string          `json:"event_id"`
	BusID     string          `json:"bus_id"`
	Timestamp string `json:"timestamp"`
	Location  json.RawMessage `json:"location"`
	Frame     json.RawMessage `json:"frame"`
	Traffic   json.RawMessage `json:"traffic"`
	Vehicles  json.RawMessage `json:"vehicles"`
	Source    json.RawMessage `json:"source"`
}

func main() {

	// --------------------------------------------------
	// 1. Read the original 150-event JSON
	// --------------------------------------------------

	data, err := os.ReadFile("test_data/traffic_events.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	var events []TrafficEvent

	err = json.Unmarshal(data, &events)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	fmt.Printf("Loaded %d events\n", len(events))

	// --------------------------------------------------
	// 2. Use today's date in IST
	// --------------------------------------------------
     
	ist, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println("Error loading IST timezone:", err)
		return
	}

	now := time.Now().In(ist)

	today := now.Format("2006-01-02")

	fmt.Println("Sending events with date:", today)
	fmt.Println()

	// --------------------------------------------------
	// 3. HTTP client
	// --------------------------------------------------

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	successful := 0
	failed := 0

	// --------------------------------------------------
	// 4. Send events one by one
	// --------------------------------------------------

	for i := range events {

		event := &events[i]
        
		// ----------------------------------------------
		// Convert timestamp to today's date
		// while keeping the original time.
		// ----------------------------------------------
        event.EventID = fmt.Sprintf("EVT_Run_%06d_%d", i+1, time.Now().Unix())

		originalTime, err := time.Parse(
    "2006-01-02T15:04:05",
    event.Timestamp,
)

if err != nil {
    fmt.Printf(
        "[%d/%d] %s -> TIMESTAMP ERROR: %v\n",
        i+1,
        len(events),
        event.EventID,
        err,
    )

    failed++
    continue
}

eventTime := time.Date(
    now.Year(),
    now.Month(),
    now.Day(),
    originalTime.Hour(),
    originalTime.Minute(),
    originalTime.Second(),
    0,
    ist,
)

event.Timestamp = eventTime.Format(time.RFC3339)

		// ----------------------------------------------
		// Convert event back to JSON
		// ----------------------------------------------

		jsonData, err := json.Marshal(event)
		if err != nil {
			fmt.Printf(
				"[%d/%d] %s -> JSON ERROR: %v\n",
				i+1,
				len(events),
				event.EventID,
				err,
			)

			failed++
			continue
		}

		// ----------------------------------------------
		// Create POST request
		// ----------------------------------------------

		req, err := http.NewRequest(
			http.MethodPost,
			API_URL,
			bytes.NewBuffer(jsonData),
		)

		if err != nil {
			fmt.Printf(
				"[%d/%d] %s -> REQUEST ERROR: %v\n",
				i+1,
				len(events),
				event.EventID,
				err,
			)

			failed++
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		// ----------------------------------------------
		// Send request
		// ----------------------------------------------

		resp, err := client.Do(req)

		if err != nil {
			fmt.Printf(
				"[%d/%d] %s -> FAILED: %v\n",
				i+1,
				len(events),
				event.EventID,
				err,
			)

			failed++
			continue
		}

		// Read backend response
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			fmt.Printf(
				"[%d/%d] %s -> RESPONSE ERROR: %v\n",
				i+1,
				len(events),
				event.EventID,
				readErr,
			)

			failed++
			continue
		}

		// ----------------------------------------------
		// Check response
		// ----------------------------------------------

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {

			successful++

			fmt.Printf(
				"[%d/%d] %s -> SUCCESS | timestamp=%s\n",
				i+1,
				len(events),
				event.EventID,
				event.Timestamp,
			)

		} else {

			failed++

			fmt.Printf(
				"[%d/%d] %s -> FAILED (%d): %s\n",
				i+1,
				len(events),
				event.EventID,
				resp.StatusCode,
				string(body),
			)
		}

		// ----------------------------------------------
		// Small delay between events
		// ----------------------------------------------

		time.Sleep(200 * time.Millisecond)
	}

	// --------------------------------------------------
	// 5. Final summary
	// --------------------------------------------------

	fmt.Println()
	fmt.Println("================================")
	fmt.Println("        SENDING COMPLETE")
	fmt.Println("================================")
	fmt.Println("Total events :", len(events))
	fmt.Println("Successful   :", successful)
	fmt.Println("Failed       :", failed)
	fmt.Println("Date used    :", today)
	fmt.Println("================================")
}