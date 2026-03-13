package engine

import (
	"fmt"
	"sync"

	"github.com/mr-isik/loki-backend/internal/domain"
)

// NodeFactory is a constructor function that returns a new instance of an INodeExecutor.
type NodeFactory func() domain.INodeExecutor

// NodeRegistry holds registered node factories keyed by their type string.
// It is safe for concurrent use.
type NodeRegistry struct {
	mu        sync.RWMutex
	factories map[string]NodeFactory
}

// NewNodeRegistry creates a new empty NodeRegistry.
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		factories: make(map[string]NodeFactory),
	}
}

// Register adds a node factory for the given typeKey.
// If a factory is already registered for that key, it will be overwritten.
func (r *NodeRegistry) Register(typeKey string, factory NodeFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[typeKey] = factory
}

// Get looks up the factory for typeKey, creates a new executor instance, and returns it.
func (r *NodeRegistry) Get(typeKey string) (domain.INodeExecutor, error) {
	r.mu.RLock()
	factory, ok := r.factories[typeKey]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown node type: %s", typeKey)
	}
	return factory(), nil
}

// Has returns true if a factory has been registered for the given typeKey.
func (r *NodeRegistry) Has(typeKey string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[typeKey]
	return ok
}

// RegisteredTypes returns a slice of all registered type keys.
func (r *NodeRegistry) RegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.factories))
	for k := range r.factories {
		types = append(types, k)
	}
	return types
}
