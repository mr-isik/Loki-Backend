package nodes

import (
	"context"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type WebhookNode struct{}

func (n *WebhookNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	// WebhookNode is usually a trigger. When executed (e.g. manually or by the system passing initial data),
	// it just passes the data through.

	// In a real scenario, rawData might contain the webhook payload.
	// We'll just parse it as generic map if possible, or pass as is.

	// Since we don't know the structure, we just pass it to OutputData.
	// If rawData is JSON, we could try to unmarshal it, but for now let's assume rawData IS the payload
	// or we just return it as "payload".

	// Let's try to unmarshal to map[string]interface{} to be nicer, but fallback to string.

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output",
		Log:             "Webhook triggered",
		OutputData: map[string]interface{}{
			"payload": string(rawData), // Simple pass-through for now
		},
	}, nil
}
