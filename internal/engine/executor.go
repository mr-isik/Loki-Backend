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
"github.com/mr-isik/loki-backend/internal/engine/utils"
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

	isSubEngine bool
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
func (e *WorkflowEngine) Execute(ctx context.Context) error {
	if !e.isSubEngine {
		if err := e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusRunning, nil); err != nil {
			return fmt.Errorf("failed to start run: %w", err)
		}
	}

	// Build in-degree map: how many incoming edges each node has.
	inDegree := e.buildInDegreeMap()

	// Find start nodes (in-degree == 0).
	startNodes := e.findStartNodes(inDegree)
	if len(startNodes) == 0 {
		return e.failRun(ctx, "No start nodes found")
	}

	var wg sync.WaitGroup

	// scheduleNode launches a goroutine to process a single node.
	var scheduleNode func(nodeID uuid.UUID)
	scheduleNode = func(nodeID uuid.UUID) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if ctx.Err() != nil {
				return
			}

			triggeredHandle, err := e.processNode(ctx, nodeID)
			
			node := e.Nodes[nodeID]
			nodeType := ""
			if t, ok := node.Data["type"].(string); ok {
				nodeType = t
			}

			// Sub-Workflow execution for loops
			if err == nil && nodeType == "loop" {
				err = e.executeLoop(ctx, nodeID)
				if err != nil {
					e.errMu.Lock()
					e.nodeErrors = append(e.nodeErrors, fmt.Errorf("loop iteration failed at node %s: %w", nodeID, err))
					e.errMu.Unlock()
					return
				}
				// Force the main engine to continue ONLY along output_done branch
				triggeredHandle = "output_done"
			} else if err != nil {
				e.errMu.Lock()
				e.nodeErrors = append(e.nodeErrors, fmt.Errorf("node %s failed: %w", nodeID, err))
				e.errMu.Unlock()
				return
			}

			e.depMu.Lock()
			e.triggeredHandles[nodeID] = triggeredHandle
			e.depMu.Unlock()

			nextNodes := e.findNextNodes(nodeID, triggeredHandle)

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

	for _, nodeID := range startNodes {
		scheduleNode(nodeID)
	}

	wg.Wait()

	e.errMu.Lock()
	errs := make([]error, len(e.nodeErrors))
	copy(errs, e.nodeErrors)
	e.errMu.Unlock()

	if len(errs) > 0 {
		if !e.isSubEngine {
			now := time.Now()
			e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusFailed, &now)
		}
		return errors.Join(errs...)
	}

	if !e.isSubEngine {
		now := time.Now()
		if err := e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusCompleted, &now); err != nil {
			return fmt.Errorf("failed to complete run: %w", err)
		}
	}

	return nil
}

func (e *WorkflowEngine) buildInDegreeMap() map[uuid.UUID]int {
	inDegree := make(map[uuid.UUID]int)

	for id := range e.Nodes {
		inDegree[id] = 0
	}

	for _, edge := range e.Edges {
		if _, hasSource := e.Nodes[edge.SourceNodeID]; hasSource {
			if _, hasTarget := e.Nodes[edge.TargetNodeID]; hasTarget {
				inDegree[edge.TargetNodeID]++
			}
		}
	}

	return inDegree
}

func (e *WorkflowEngine) processNode(ctx context.Context, nodeID uuid.UUID) (string, error) {
	node, exists := e.Nodes[nodeID]
	if !exists {
		return "", fmt.Errorf("node %s not found", nodeID)
	}

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

	jsonData, _ := json.Marshal(inputData)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := executor.Execute(timeoutCtx, jsonData)

	if err != nil {
		sanitizedErr := utils.SanitizeError(err)
		
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			sanitizedErr = "Execution timed out after 10 seconds."
		}

		e.updateLog(ctx, logEntry.ID, domain.NodeRunLogStatusFailed, "", sanitizedErr)
		return "", errors.New(sanitizedErr)
	}

	e.mu.Lock()
	e.nodeOutputs[nodeID] = result.OutputData
	e.mu.Unlock()

	status := domain.NodeRunLogStatusCompleted
	if result.Status == "failed" {
		status = domain.NodeRunLogStatusFailed
	}

	if err := e.updateLog(ctx, logEntry.ID, status, result.Log, ""); err != nil {
		fmt.Printf("failed to update log: %v\n", err)
	}

	if result.Status == "failed" {
		return "", fmt.Errorf("node execution failed")
	}

	return result.TriggeredHandle, nil
}

func (e *WorkflowEngine) executeLoop(ctx context.Context, loopNodeID uuid.UUID) error {
	e.mu.RLock()
	loopOutputs := e.nodeOutputs[loopNodeID]
	e.mu.RUnlock()

	itemsInterface, ok := loopOutputs["items"]
	if !ok {
		return nil
	}

	items, err := toSliceInterface(itemsInterface)
	if err != nil {
		return fmt.Errorf("invalid items type: %v", err)
	}

	subNodes, subEdges := e.getSubgraph(loopNodeID, "output_item")
	if len(subNodes) == 0 {
		return nil 
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(items))

	for i, item := range items {
		wg.Add(1)
		go func(index int, loopItem interface{}) {
			defer wg.Done()
			
			var nodesList []domain.WorkflowNode
			for _, n := range subNodes {
				nodesList = append(nodesList, n)
			}

			subEngine := NewWorkflowEngine(nodesList, subEdges, e.RunID, e.WorkflowID, e.LogRepo, e.RunRepo)
			subEngine.isSubEngine = true
			
			subEngine.mu.Lock()
			subEngine.nodeOutputs[loopNodeID] = map[string]interface{}{
				"output_item": loopItem,
				"index":       index,
			}
			subEngine.mu.Unlock()

			if err := subEngine.Execute(ctx); err != nil {
				errCh <- fmt.Errorf("iteration %d failed: %w", index, err)
			}
		}(i, item)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (e *WorkflowEngine) getSubgraph(startNodeID uuid.UUID, startHandle string) (map[uuid.UUID]domain.WorkflowNode, []domain.WorkflowEdge) {
	subNodes := make(map[uuid.UUID]domain.WorkflowNode)
	var subEdges []domain.WorkflowEdge

	visitedNodes := make(map[uuid.UUID]bool)
	queue := []uuid.UUID{}

	for _, edge := range e.Edges {
		if edge.SourceNodeID == startNodeID && edge.SourceHandle == startHandle {
			subEdges = append(subEdges, edge)
			if !visitedNodes[edge.TargetNodeID] {
				visitedNodes[edge.TargetNodeID] = true
				queue = append(queue, edge.TargetNodeID)
				subNodes[edge.TargetNodeID] = e.Nodes[edge.TargetNodeID]
			}
		}
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, edge := range e.Edges {
			if edge.SourceNodeID == curr {
				subEdges = append(subEdges, edge)
				if !visitedNodes[edge.TargetNodeID] {
					visitedNodes[edge.TargetNodeID] = true
					queue = append(queue, edge.TargetNodeID)
					subNodes[edge.TargetNodeID] = e.Nodes[edge.TargetNodeID]
				}
			}
		}
	}

	return subNodes, subEdges
}


func (e *WorkflowEngine) findStartNodes(inDegree map[uuid.UUID]int) []uuid.UUID {
	var start []uuid.UUID
	for id, count := range inDegree {
		if count == 0 {
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
			// Only consider targets that exist in the engine (important for subengines)
			if _, exists := e.Nodes[edge.TargetNodeID]; exists {
				next = append(next, edge.TargetNodeID)
			}
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
	if !e.isSubEngine {
		now := time.Now()
		e.RunRepo.UpdateStatus(ctx, e.RunID, domain.WorkflowRunStatusFailed, &now)
	}
	return errors.New(msg)
}

func (e *WorkflowEngine) logNodeError(ctx context.Context, nodeID uuid.UUID, msg string) {
	e.LogRepo.Create(ctx, &domain.CreateNodeRunLogRequest{
		RunID:  e.RunID,
		NodeID: nodeID,
		Status: domain.NodeRunLogStatusFailed,
	})
}

func toSliceInterface(v interface{}) ([]interface{}, error) {
	if v == nil {
		return []interface{}{}, nil
	}
	switch val := v.(type) {
	case []interface{}:
		return val, nil
	case []string:
		res := make([]interface{}, len(val))
		for i, v := range val {
			res[i] = v
		}
		return res, nil
	case []int:
		res := make([]interface{}, len(val))
		for i, v := range val {
			res[i] = v
		}
		return res, nil
	default:
		if s, ok := v.(string); ok {
			var res []interface{}
			if err := json.Unmarshal([]byte(s), &res); err == nil {
				return res, nil
			}
		}
		return nil, fmt.Errorf("input is not a slice or valid JSON array")
	}
}
