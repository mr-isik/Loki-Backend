package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type WorkflowEngine struct {
	Nodes      map[uuid.UUID]domain.WorkflowNode
	Edges      []domain.WorkflowEdge
	RunID      uuid.UUID
	LogRepo    domain.NodeRunLogRepository
	RunRepo    domain.WorkflowRunRepository
	WorkflowID uuid.UUID

	nodeOutputs map[uuid.UUID]map[string]interface{}
	mu          sync.RWMutex
}

func NewWorkflowEngine(
	nodes []domain.WorkflowNode,
	edges []domain.WorkflowEdge,
	runID uuid.UUID,
	workflowID uuid.UUID,
	logRepo domain.NodeRunLogRepository,
	runRepo domain.WorkflowRunRepository,
) *WorkflowEngine {
	nodeMap := make(map[uuid.UUID]domain.WorkflowNode)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	return &WorkflowEngine{
		Nodes:       nodeMap,
		Edges:       edges,
		RunID:       runID,
		WorkflowID:  workflowID,
		LogRepo:     logRepo,
		RunRepo:     runRepo,
		nodeOutputs: make(map[uuid.UUID]map[string]interface{}),
	}
}

func (e *WorkflowEngine) Execute(ctx context.Context) error {
	if err := e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusRunning, nil); err != nil {
		return fmt.Errorf("failed to start run: %w", err)
	}
	startNodes := e.findStartNodes()
	if len(startNodes) == 0 {
		return e.failRun(ctx, "No start nodes found")
	}

	queue := make([]uuid.UUID, 0, len(e.Nodes))
	queue = append(queue, startNodes...)

	visited := make(map[uuid.UUID]bool)

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		if visited[nodeID] {
			continue
		}
		visited[nodeID] = true

		triggeredHandle, err := e.processNode(ctx, nodeID)
		if err != nil {
			e.failRun(ctx, fmt.Sprintf("Node %s failed: %v", nodeID, err))
			return err
		}

		nextNodes := e.findNextNodes(nodeID, triggeredHandle)
		queue = append(queue, nextNodes...)
	}

	now := time.Now()
	if err := e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusCompleted, &now); err != nil {
		return fmt.Errorf("failed to complete run: %w", err)
	}

	return nil
}

// processNode executes a single node.
func (e *WorkflowEngine) processNode(ctx context.Context, nodeID uuid.UUID) (string, error) {
	node, exists := e.Nodes[nodeID]
	if !exists {
		return "", fmt.Errorf("node %s not found", nodeID)
	}

	// 1. Create Log Entry (Pending)
	logEntry, err := e.LogRepo.Create(ctx, &domain.CreateNodeRunLogRequest{
		RunID:  e.RunID,
		NodeID: nodeID,
		Status: domain.NodeRunLogStatusRunning,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create log: %w", err)
	}

	inputData := make(map[string]interface{})

	for k, v := range node.Data {
		inputData[k] = v
	}

	incomingEdges := e.getIncomingEdges(nodeID)
	inputsFromUpstream := make(map[string]interface{})

	e.mu.RLock()
	for _, edge := range incomingEdges {
		sourceOutput, ok := e.nodeOutputs[edge.SourceNodeID]
		if ok {
			// We can map specific outputs to specific inputs if the Edge has that info.
			// For now, we merge the whole output map or use the SourceHandle.
			// A common pattern: inputs[edge.TargetHandle] = sourceOutput[edge.SourceHandle]
			if val, valOk := sourceOutput[edge.SourceHandle]; valOk {
				inputsFromUpstream[edge.TargetHandle] = val
			} else {
				// Fallback: if source output is just a value, or we want to pass everything
				inputsFromUpstream[edge.TargetHandle] = sourceOutput
			}
		}
	}
	e.mu.RUnlock()

	inputData["input"] = inputsFromUpstream

	typeVal, ok := node.Data["type"]
	if !ok {
		return "", fmt.Errorf("node type not found in data for node %s", nodeID)
	}
	nodeType, ok := typeVal.(string)
	if !ok {
		return "", fmt.Errorf("invalid node type format for node %s", nodeID)
	}

	executor, err := NewNodeExecutor(nodeType)
	if err != nil {
		e.updateLog(ctx, logEntry.ID, domain.NodeRunLogStatusFailed, "", err.Error())
		return "", err
	}

	// 4. Execute
	jsonData, _ := json.Marshal(inputData)
	result, err := executor.Execute(ctx, jsonData)

	if err != nil {
		e.updateLog(ctx, logEntry.ID, domain.NodeRunLogStatusFailed, "", err.Error())
		return "", err
	}

	// 5. Save Output and Log
	e.mu.Lock()
	e.nodeOutputs[nodeID] = result.OutputData
	e.mu.Unlock()

	status := domain.NodeRunLogStatusCompleted
	if result.Status == "failed" {
		status = domain.NodeRunLogStatusFailed
	}

	if err := e.updateLog(ctx, logEntry.ID, status, result.Log, ""); err != nil {
		// Just log error, don't fail flow
		fmt.Printf("failed to update log: %v\n", err)
	}

	if result.Status == "failed" {
		return "", fmt.Errorf("node execution failed")
	}

	return result.TriggeredHandle, nil
}

// Helper methods

func (e *WorkflowEngine) findStartNodes() []uuid.UUID {
	// Nodes with no incoming edges
	incoming := make(map[uuid.UUID]bool)
	for _, edge := range e.Edges {
		incoming[edge.TargetNodeID] = true
	}

	var start []uuid.UUID
	for id := range e.Nodes {
		if !incoming[id] {
			start = append(start, id)
		}
	}
	return start
}

func (e *WorkflowEngine) findNextNodes(nodeID uuid.UUID, triggeredHandle string) []uuid.UUID {
	var next []uuid.UUID
	for _, edge := range e.Edges {
		if edge.SourceNodeID == nodeID {
			// If triggeredHandle is specified, only follow matching edges.
			// If triggeredHandle is empty or "default", follow all or default.
			// For now, strict matching if handle is provided.
			if triggeredHandle != "" && edge.SourceHandle != triggeredHandle {
				continue
			}
			next = append(next, edge.TargetNodeID)
		}
	}
	return next
}

func (e *WorkflowEngine) getIncomingEdges(nodeID uuid.UUID) []domain.WorkflowEdge {
	var incoming []domain.WorkflowEdge
	for _, edge := range e.Edges {
		if edge.TargetNodeID == nodeID {
			incoming = append(incoming, edge)
		}
	}
	return incoming
}

func (e *WorkflowEngine) updateLog(ctx context.Context, logID uuid.UUID, status domain.NodeRunLogStatus, output, errorMsg string) error {
	// We need a repository method that supports these fields.
	// The interface has Update(ctx, id, req).
	// But req has Status, LogOutput, ErrorMsg.
	// It doesn't seem to have FinishedAt in the Request struct based on previous view,
	// but the domain model has it.
	// Let's check UpdateNodeRunLogRequest again.
	// It has Status, LogOutput, ErrorMsg.
	// The repository implementation likely handles FinishedAt setting if status is terminal.

	req := &domain.UpdateNodeRunLogRequest{
		Status:    status,
		LogOutput: output,
		ErrorMsg:  errorMsg,
	}
	return e.LogRepo.Update(ctx, logID, req)
}

func (e *WorkflowEngine) failRun(ctx context.Context, msg string) error {
	now := time.Now()
	e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusFailed, &now)
	return errors.New(msg)
}

func (e *WorkflowEngine) logNodeError(ctx context.Context, nodeID uuid.UUID, msg string) {
	// Try to create a log entry for the error
	e.LogRepo.Create(ctx, &domain.CreateNodeRunLogRequest{
		RunID:  e.RunID,
		NodeID: nodeID,
		Status: domain.NodeRunLogStatusFailed,
	})
	// We can't easily update it with the message if we just created it without ID return in one line,
	// but this is a fallback.
}
