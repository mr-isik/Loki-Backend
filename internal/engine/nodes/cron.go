package nodes

import (
	"context"
	"time"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type CronNode struct{}

func (n *CronNode) Execute(ctx context.Context, rawData []byte) (domain.NodeResult, error) {
	// CronNode is a trigger. It usually passes the current time.

	return domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output",
		Log:             "Cron triggered",
		OutputData: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}, nil
}
