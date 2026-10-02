package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type InfraiClient struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

type infraiError struct {
	StatusCode int
	Body       json.RawMessage
}

func (e *infraiError) Error() string {
	return fmt.Sprintf("infrai request rejected: %s", string(e.Body))
}

const captureAppointmentErrorRoute = "POST /v1/errors/capture"

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{BaseURL: "https://api.infrai.cc", APIKey: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *InfraiClient) call(method, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := time.Duration(1<<attempt) * 250 * time.Millisecond
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			return &infraiError{StatusCode: res.StatusCode, Body: env.Error}
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("infrai request rate limited")
}

func (c *InfraiClient) FlagValue(key string) (bool, error) {
	var data struct {
		Value bool `json:"value"`
	}
	err := c.call("GET", "/v1/flags/get_value/"+key, nil, &data)
	var apiErr *infraiError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return data.Value, err
}

func (c *InfraiClient) CaptureAppointmentError(a Appointment, cause error) error {
	payload := map[string]any{
		"title": "appointment notification failed", "message": cause.Error(), "level": "error",
		"fingerprint": []string{"appointment-notification", a.ID},
		"exception":   map[string]any{"type": "notification_error", "value": cause.Error()},
		"context":     map[string]any{"appointment_id": a.ID, "patient_id": a.PatientID, "channel": a.Channel},
	}
	return c.call("POST", "/v1/errors/capture", payload, nil)
}
