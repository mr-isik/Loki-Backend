package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type HttpRequestNode struct{}

type httpData struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"query_params"`
	Timeout     int               `json:"timeout"`
	Body        interface{}       `json:"body"`
}

func (n *HttpRequestNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	var data httpData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	if data.URL == "" {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        "URL is required",
			OutputData: map[string]interface{}{"error": "URL is required"},
		}, fmt.Errorf("URL is required")
	}

	reqURL := data.URL
	if len(data.QueryParams) > 0 {
		u, err := url.Parse(reqURL)
		if err != nil {
			return &domain.NodeResult{
				Status:     "failed",
				Log:        fmt.Sprintf("Failed to parse URL: %v", err),
				OutputData: map[string]interface{}{"error": err.Error()},
			}, err
		}
		q := u.Query()
		for k, v := range data.QueryParams {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	var bodyReader io.Reader
	if data.Body != nil {
		jsonBody, err := json.Marshal(data.Body)
		if err != nil {
			return &domain.NodeResult{
				Status:     "failed",
				Log:        fmt.Sprintf("Failed to marshal body: %v", err),
				OutputData: map[string]interface{}{"error": err.Error()},
			}, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	if data.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(data.Timeout)*time.Second)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, data.Method, reqURL, bodyReader)
	if err != nil {
		return &domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to create request: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	// Add headers if provided
	for k, v := range data.Headers {
		req.Header.Set(k, v)
	}

	// Default to JSON content type if body is present and not set
	if data.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Request failed: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var responseBody interface{}
	// Try to parse as JSON, otherwise keep as string
	if err := json.Unmarshal(body, &responseBody); err != nil {
		responseBody = string(body)
	}

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             fmt.Sprintf("Request to %s completed with status %d", data.URL, resp.StatusCode),
		OutputData: map[string]interface{}{
			"status":  resp.StatusCode,
			"body":    responseBody,
			"headers": resp.Header,
		},
	}, nil
}
