package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type HttpRequestNode struct{}

type httpData struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

func (n *HttpRequestNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	var data httpData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	if data.URL == "" {
		return domain.NodeResult{
			Status:     "failed",
			Log:        "URL is required",
			OutputData: map[string]interface{}{"error": "URL is required"},
		}, fmt.Errorf("URL is required")
	}

	var bodyReader io.Reader
	if data.Body != nil {
		jsonBody, err := json.Marshal(data.Body)
		if err != nil {
			return domain.NodeResult{
				Status:     "failed",
				Log:        fmt.Sprintf("Failed to marshal body: %v", err),
				OutputData: map[string]interface{}{"error": err.Error()},
			}, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, data.Method, data.URL, bodyReader)
	if err != nil {
		return domain.NodeResult{
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
		return domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Request failed: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil // Return nil error to allow workflow to continue on error path if needed
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var responseBody interface{}
	// Try to parse as JSON, otherwise keep as string
	if err := json.Unmarshal(body, &responseBody); err != nil {
		responseBody = string(body)
	}

	return domain.NodeResult{
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
