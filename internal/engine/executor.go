package engine

import (
	"context"
	"errors"
	"fmt"
	"encoding/json"
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

	// Parallel execution state
	depMu            sync.Mutex
	errMu            sync.Mutex
	nodeErrors       []error
	triggeredHandles map[uuid.UUID]string // sourceNodeID → triggeredHandle
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
		Nodes:            nodeMap,
		Edges:            edges,
		RunID:            runID,
		WorkflowID:       workflowID,
		LogRepo:          logRepo,
		RunRepo:          runRepo,
		nodeOutputs:      make(map[uuid.UUID]map[string]interface{}),
		triggeredHandles: make(map[uuid.UUID]string),
	}
}

// Execute runs the workflow DAG with parallel execution of independent nodes.
// Nodes at the same topological level (independent branches) run concurrently.
// Dependencies are respected via in-degree tracking (Kahn's algorithm).
func (e *WorkflowEngine) Execute(ctx context.Context) error {
	if err := e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusRunning, nil); err != nil {
		return fmt.Errorf("failed to start run: %w", err)
	}

	// Build in-degree map: how many incoming edges each node has.
	inDegree := e.buildInDegreeMap()

	// Find start nodes (in-degree == 0).
	startNodes := e.findStartNodes()
	if len(startNodes) == 0 {
		return e.failRun(ctx, "No start nodes found")
	}

	var wg sync.WaitGroup

	// scheduleNode launches a goroutine to process a single node.
	// After completion, it decrements in-degrees of downstream nodes
	// and schedules any that become ready.
	var scheduleNode func(nodeID uuid.UUID)
	scheduleNode = func(nodeID uuid.UUID) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Check context cancellation.
			if ctx.Err() != nil {
				return
			}

			// Process the node.
			triggeredHandle, err := e.processNode(ctx, nodeID)
			if err != nil {
				e.errMu.Lock()
				e.nodeErrors = append(e.nodeErrors, fmt.Errorf("node %s failed: %w", nodeID, err))
				e.errMu.Unlock()
				// Do NOT activate downstream nodes — isolate the failure.
				return
			}

			// Store the triggered handle for this node so edge filtering works.
			e.depMu.Lock()
			e.triggeredHandles[nodeID] = triggeredHandle
			e.depMu.Unlock()

			// Find downstream nodes respecting the triggered handle.
			nextNodes := e.findNextNodes(nodeID, triggeredHandle)

			// Decrement in-degree for each downstream and schedule if ready.
			for _, nextID := range nextNodes {
				ready := false

				e.depMu.Lock()
				inDegree[nextID]--
				if inDegree[nextID] == 0 {
					ready = true
				}
				e.depMu.Unlock()

				if ready {
					scheduleNode(nextID)
				}
			}
		}()
	}

	// Launch all start nodes in parallel.
	for _, nodeID := range startNodes {
		scheduleNode(nodeID)
	}

	// Wait for all goroutines to finish.
	wg.Wait()

	// Check for errors.
	e.errMu.Lock()
	errs := make([]error, len(e.nodeErrors))
	copy(errs, e.nodeErrors)
	e.errMu.Unlock()

	if len(errs) > 0 {
		now := time.Now()
		e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusFailed, &now)
		return errors.Join(errs...)
	}

	now := time.Now()
	if err := e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusCompleted, &now); err != nil {
		return fmt.Errorf("failed to complete run: %w", err)
	}

	return nil
}

// buildInDegreeMap calculates the number of incoming edges for each node.
// Only edges that are "relevant" are counted — for nodes whose source has
// a triggered handle filter, we count all edges initially and decrement
// dynamically during execution.
func (e *WorkflowEngine) buildInDegreeMap() map[uuid.UUID]int {
	inDegree := make(map[uuid.UUID]int)

	// Initialize all nodes with 0.
	for id := range e.Nodes {
		inDegree[id] = 0
	}

	// Count incoming edges.
	for _, edge := range e.Edges {
		inDegree[edge.TargetNodeID]++
	}

	return inDegree
}

// processNode executes a single node.
func (e *WorkflowEngine) processNode(ctx context.Context, nodeID uuid.UUID) (string, error) {
	node, exists := e.Nodes[nodeID]
	if !exists {
		return "", fmt.Errorf("node %s not found", nodeID)
	}

	// 1. Create Log Entry (Running)
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
			if val, valOk := sourceOutput[edge.SourceHandle]; valOk {
				inputsFromUpstream[edge.TargetHandle] = val
			} else {
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
}
