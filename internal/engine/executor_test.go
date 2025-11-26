package engine

import (
	"context"
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

	// 2. Node 1 Log
	mockLogRepo.On("Create", mock.Anything, mock.MatchedBy(func(req *domain.CreateNodeRunLogRequest) bool {
		return req.NodeID == node1ID && req.Status == domain.NodeRunLogStatusRunning
	})).Return(&domain.NodeRunLog{ID: uuid.New()}, nil)

	mockLogRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(req *domain.UpdateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusCompleted
	})).Return(nil)

	// 3. Node 2 Log
	mockLogRepo.On("Create", mock.Anything, mock.MatchedBy(func(req *domain.CreateNodeRunLogRequest) bool {
		return req.NodeID == node2ID && req.Status == domain.NodeRunLogStatusRunning
	})).Return(&domain.NodeRunLog{ID: uuid.New()}, nil)

	mockLogRepo.On("Update", mock.Anything, mock.Anything, mock.MatchedBy(func(req *domain.UpdateNodeRunLogRequest) bool {
		return req.Status == domain.NodeRunLogStatusCompleted
	})).Return(nil)

	// 4. Complete Run
	mockRunRepo.On("UpdateStatus", mock.Anything, runID, domain.WorkflowRunStatusCompleted, mock.Anything).Return(nil)

	// Execute
	engine := NewWorkflowEngine(nodes, edges, runID, workflowID, mockLogRepo, mockRunRepo)
	err := engine.Execute(context.Background())

	// Assert
	assert.NoError(t, err)
	mockRunRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}
