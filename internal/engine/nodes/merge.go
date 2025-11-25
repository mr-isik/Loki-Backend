package nodes

import (
	"context"

	"github.com/mr-isik/loki-backend/internal/domain"
)

type MergeNode struct{}

func (n *MergeNode) Execute(ctx context.Context, rawData []byte) (*domain.NodeResult, error) {
	// MergeNode simply passes execution through.
	// The engine handles the fact that multiple nodes point to this one.
	// We just return success.

	return &domain.NodeResult{
		Status:          "completed",
		TriggeredHandle: "output",
		Log:             "Merge point reached",
		OutputData:      map[string]interface{}{"merged": true},
	}, nil
}
