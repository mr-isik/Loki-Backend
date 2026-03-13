package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repositories
type MockRunRepo struct {
	mock.Mock
}

func (m *MockRunRepo) Create(ctx context.Context, workflowID uuid.UUID) (*domain.WorkflowRun, error) {
	args := m.Called(ctx, workflowID)
	return args.Get(0).(*domain.WorkflowRun), args.Error(1)
}
func (m *MockRunRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowRun, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.WorkflowRun), args.Error(1)
}
func (m *MockRunRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.WorkflowRunStatus, finishedAt *time.Time) error {
	args := m.Called(ctx, id, status, finishedAt)
	return args.Error(0)
}
func (m *MockRunRepo) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID, limit, offset int) ([]*domain.WorkflowRun, int, error) {
	args := m.Called(ctx, workflowID, limit, offset)
	return args.Get(0).([]*domain.WorkflowRun), args.Int(1), args.Error(2)
}

type MockLogRepo struct {
	mock.Mock
}

func (m *MockLogRepo) Create(ctx context.Context, req *domain.CreateNodeRunLogRequest) (*domain.NodeRunLog, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.NodeRunLog), args.Error(1)
}
func (m *MockLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.NodeRunLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.NodeRunLog), args.Error(1)
}
func (m *MockLogRepo) GetByRunID(ctx context.Context, runID uuid.UUID) ([]*domain.NodeRunLog, error) {
	args := m.Called(ctx, runID)
	return args.Get(0).([]*domain.NodeRunLog), args.Error(1)
}
func (m *MockLogRepo) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateNodeRunLogRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func TestWorkflowEngine_Execute_SimpleFlow(t *testing.T) {
	// Setup
	runID := uuid.New()
	workflowID := uuid.New()
	node1ID := uuid.New()
	node2ID := uuid.New()

	// Nodes
	nodes := []domain.WorkflowNode{
		{
			ID:         node1ID,
			WorkflowID: workflowID,
			Data: map[string]interface{}{
				"type": "set_data",
				"data": map[string]interface{}{"foo": "bar"},
			},
		},
		{
			ID:         node2ID,
			WorkflowID: workflowID,
			Data: map[string]interface{}{
				"type": "set_data", // Just using set_data again as a dummy
				"data": map[string]interface{}{"baz": "qux"},
			},
		},
	}

	// Edges: Node1 -> Node2
	edges := []domain.WorkflowEdge{
		{
			ID:           uuid.New(),
			WorkflowID:   workflowID,
			SourceNodeID: node1ID,
			TargetNodeID: node2ID,
			SourceHandle: "output",
			TargetHandle: "input",
		},
	}

	// Mocks
	mockRunRepo := new(MockRunRepo)
	mockLogRepo := new(MockLogRepo)

	// Expectations
	// 1. Start Run
	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusRunning, mock.Anything).Return(nil)

	// 2. Node logs (order may vary in parallel, use generic matchers)
	mockLogRepo.On("Create", mock.Anything, mock.MatchedBy(func(req *domain.CreateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusRunning
	})).Return(&domain.NodeRunLog{ID: uuid.New()}, nil)

	mockLogRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(req *domain.UpdateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusCompleted
	})).Return(nil)

	// 3. Complete Run
	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusCompleted, mock.Anything).Return(nil)

	// Execute
	engine := NewWorkflowEngine(nodes, edges, runID, workflowID, mockLogRepo, mockRunRepo)
	err := engine.Execute(context.Background())

	// Assert
	assert.NoError(t, err)
	mockRunRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestWorkflowEngine_Execute_ParallelBranches(t *testing.T) {
	// Topology:
	//   StartNode
	//    /     \
	//  BranchA  BranchB    (parallel — no dependency between them)
	//    \     /
	//   MergeNode

	runID := uuid.New()
	workflowID := uuid.New()
	startID := uuid.New()
	branchAID := uuid.New()
	branchBID := uuid.New()
	mergeID := uuid.New()

	nodes := []domain.WorkflowNode{
		{ID: startID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "set_data", "data": map[string]interface{}{"start": true}}},
		{ID: branchAID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "set_data", "data": map[string]interface{}{"branch": "A"}}},
		{ID: branchBID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "set_data", "data": map[string]interface{}{"branch": "B"}}},
		{ID: mergeID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "merge"}},
	}

	edges := []domain.WorkflowEdge{
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: startID, TargetNodeID: branchAID, SourceHandle: "output", TargetHandle: "input"},
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: startID, TargetNodeID: branchBID, SourceHandle: "output", TargetHandle: "input"},
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: branchAID, TargetNodeID: mergeID, SourceHandle: "output", TargetHandle: "input"},
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: branchBID, TargetNodeID: mergeID, SourceHandle: "output", TargetHandle: "input"},
	}

	mockRunRepo := new(MockRunRepo)
	mockLogRepo := new(MockLogRepo)

	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusRunning, mock.Anything).Return(nil)
	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusCompleted, mock.Anything).Return(nil)

	mockLogRepo.On("Create", mock.Anything, mock.MatchedBy(func(req *domain.CreateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusRunning
	})).Return(&domain.NodeRunLog{ID: uuid.New()}, nil)

	mockLogRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(req *domain.UpdateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusCompleted
	})).Return(nil)

	engine := NewWorkflowEngine(nodes, edges, runID, workflowID, mockLogRepo, mockRunRepo)

	err := engine.Execute(context.Background())

	assert.NoError(t, err)

	// Both branches + start + merge = 4 nodes processed.
	engine.mu.RLock()
	assert.Len(t, engine.nodeOutputs, 4, "All 4 nodes should have produced output")
	engine.mu.RUnlock()

	// We expect log Create to be called 4 times (once per node).
	mockLogRepo.AssertNumberOfCalls(t, "Create", 4)
}

func TestWorkflowEngine_Execute_NodeFailure_NoBlockOthers(t *testing.T) {
	// Topology:
	//   Start
	//    / \
	//  Bad  Good   (parallel branches)
	//
	// "Bad" uses an unknown node type → fails.
	// "Good" uses set_data → succeeds.
	// Workflow should report failure but Good should still complete.

	runID := uuid.New()
	workflowID := uuid.New()
	startID := uuid.New()
	badID := uuid.New()
	goodID := uuid.New()

	nodes := []domain.WorkflowNode{
		{ID: startID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "set_data", "data": map[string]interface{}{"init": true}}},
		{ID: badID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "nonexistent_node_type"}},
		{ID: goodID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "set_data", "data": map[string]interface{}{"good": true}}},
	}

	edges := []domain.WorkflowEdge{
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: startID, TargetNodeID: badID, SourceHandle: "output", TargetHandle: "input"},
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: startID, TargetNodeID: goodID, SourceHandle: "output", TargetHandle: "input"},
	}

	mockRunRepo := new(MockRunRepo)
	mockLogRepo := new(MockLogRepo)

	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusRunning, mock.Anything).Return(nil)
	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusFailed, mock.Anything).Return(nil)

	mockLogRepo.On("Create", mock.Anything, mock.MatchedBy(func(req *domain.CreateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusRunning
	})).Return(&domain.NodeRunLog{ID: uuid.New()}, nil)

	// Good node will complete successfully
	mockLogRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(req *domain.UpdateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusCompleted
	})).Return(nil)

	// Bad node will fail
	mockLogRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(req *domain.UpdateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusFailed
	})).Return(nil)

	engine := NewWorkflowEngine(nodes, edges, runID, workflowID, mockLogRepo, mockRunRepo)
	err := engine.Execute(context.Background())

	// Workflow should fail because one node failed.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("node %s failed", badID))

	// But the good node should have been executed and produced output.
	engine.mu.RLock()
	_, goodRan := engine.nodeOutputs[goodID]
	engine.mu.RUnlock()
	assert.True(t, goodRan, "Good node should have run despite bad node's failure")

	// Run status should be failed.
	mockRunRepo.AssertCalled(t, "UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusFailed, mock.Anything)
}

func TestWorkflowEngine_Execute_NoStartNodes(t *testing.T) {
	// All nodes have incoming edges → no start node.
	runID := uuid.New()
	workflowID := uuid.New()
	nodeAID := uuid.New()
	nodeBID := uuid.New()

	nodes := []domain.WorkflowNode{
		{ID: nodeAID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "set_data"}},
		{ID: nodeBID, WorkflowID: workflowID, Data: map[string]interface{}{"type": "set_data"}},
	}

	// Circular: A → B, B → A
	edges := []domain.WorkflowEdge{
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: nodeAID, TargetNodeID: nodeBID, SourceHandle: "output", TargetHandle: "input"},
		{ID: uuid.New(), WorkflowID: workflowID, SourceNodeID: nodeBID, TargetNodeID: nodeAID, SourceHandle: "output", TargetHandle: "input"},
	}

	mockRunRepo := new(MockRunRepo)
	mockLogRepo := new(MockLogRepo)

	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusRunning, mock.Anything).Return(nil)
	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusFailed, mock.Anything).Return(nil)

	engine := NewWorkflowEngine(nodes, edges, runID, workflowID, mockLogRepo, mockRunRepo)
	err := engine.Execute(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No start nodes found")
}
