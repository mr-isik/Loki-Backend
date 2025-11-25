package nodes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type SlackNode struct{}

type slackData struct {
	WebhookURL string `json:"webhook_url"`
	Message    string `json:"message"`
	Channel    string `json:"channel"` // Optional, if webhook supports it
}

func (n *SlackNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	var data slackData
	if err := json.Unmarshal(rawData, &data); err != nil {
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to parse input: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}

	if data.WebhookURL == "" {
		return domain.NodeResult{
			Status:     "failed",
			Log:        "Webhook URL is required",
			OutputData: map[string]interface{}{"error": "Webhook URL is required"},
		}, fmt.Errorf("webhook URL is required")
	}

	payload := map[string]interface{}{
		"text": data.Message,
	}
	if data.Channel != "" {
		payload["channel"] = data.Channel
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", data.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return domain.NodeResult{
			Status:     "failed",
			Log:        fmt.Sprintf("Failed to create request: %v", err),
			OutputData: map[string]interface{}{"error": err.Error()},
		}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Failed to send Slack message: %v", err),
			OutputData:      map[string]interface{}{"error": err.Error()},
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return domain.NodeResult{
			Status:          "failed",
			TriggeredHandle: "output_error",
			Log:             fmt.Sprintf("Slack API returned status: %d", resp.StatusCode),
			OutputData:      map[string]interface{}{"status": resp.StatusCode},
		}, nil
	}

	return domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output_success",
		Log:             "Slack message sent",
		OutputData:      map[string]interface{}{"sent": true},
	}, nil
}
