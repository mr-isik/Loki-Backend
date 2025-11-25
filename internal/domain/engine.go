package domain

import (
	"context"
)

type NodeResult struct {
	Status          string
	OutputData      map[string]interface{}
	TriggeredHandle string
	Log             string
}

type INodeExecutor interface {
	Execute(ctx context.Context, nodeData []byte) (*NodeResult, error)
}
