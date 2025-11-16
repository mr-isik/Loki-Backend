package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNodeTemplateNotFound = errors.New("node template not found")
)

type NodeTemplate struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TypeKey     string                 `json:"type_key"`
	Category    string                 `json:"category"`
	Inputs      map[string]interface{} `json:"inputs,omitempty"`
	Outputs     map[string]interface{} `json:"outputs,omitempty"`
}

type NodeTemplateResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TypeKey     string                 `json:"type_key"`
	Category    string                 `json:"category"`
	Inputs      map[string]interface{} `json:"inputs,omitempty"`
	Outputs     map[string]interface{} `json:"outputs,omitempty"`
}

func (nt *NodeTemplate) ToResponse() *NodeTemplateResponse {
	return &NodeTemplateResponse{
		ID:          nt.ID,
		Name:        nt.Name,
		Description: nt.Description,
		TypeKey:     nt.TypeKey,
		Category:    nt.Category,
		Inputs:      nt.Inputs,
		Outputs:     nt.Outputs,
	}
}

type NodeTemplateRepository interface {
	GetAll(ctx context.Context) ([]*NodeTemplate, error)
	GetByID(ctx context.Context, id uuid.UUID) (*NodeTemplate, error)
}

type NodeTemplateService interface {
	ListNodeTemplates(ctx context.Context) ([]*NodeTemplateResponse, error)
	GetNodeTemplate(ctx context.Context, id uuid.UUID) (*NodeTemplateResponse, error)
}
