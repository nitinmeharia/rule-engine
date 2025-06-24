package domain

import (
	"encoding/json"
	"time"
)

// Request DTOs
type CreateNamespaceRequest struct {
	ID          string `json:"id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type CreateFieldRequest struct {
	FieldID     string `json:"fieldId" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type CreateFunctionRequest struct {
	ID     string   `json:"id" binding:"required"`
	Type   string   `json:"type" binding:"required"`
	Args   []string `json:"args,omitempty"`
	Values []string `json:"values,omitempty"`
}

type UpdateFunctionRequest struct {
	Type   string   `json:"type" binding:"required"`
	Args   []string `json:"args,omitempty"`
	Values []string `json:"values,omitempty"`
}

type CreateRuleRequest struct {
	ID         string          `json:"id" binding:"required"`
	Logic      string          `json:"logic" binding:"required"`
	Conditions json.RawMessage `json:"conditions" binding:"required"`
}

type UpdateRuleRequest struct {
	Logic      string          `json:"logic" binding:"required"`
	Conditions json.RawMessage `json:"conditions" binding:"required"`
}

type CreateTerminalRequest struct {
	TerminalID string `json:"terminalId" binding:"required"`
}

type CreateWorkflowRequest struct {
	ID      string          `json:"id" binding:"required"`
	StartAt string          `json:"startAt" binding:"required"`
	Steps   json.RawMessage `json:"steps" binding:"required"`
}

type UpdateWorkflowRequest struct {
	StartAt string          `json:"startAt" binding:"required"`
	Steps   json.RawMessage `json:"steps" binding:"required"`
}

// Response DTOs
type NamespaceResponse struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

type FieldResponse struct {
	FieldID     string    `json:"fieldId"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

type FunctionResponse struct {
	ID          string     `json:"id"`
	Version     int32      `json:"version"`
	Status      string     `json:"status"`
	Type        string     `json:"type"`
	Args        []string   `json:"args"`
	Values      []string   `json:"values"`
	ReturnType  string     `json:"returnType"`
	CreatedAt   time.Time  `json:"createdAt"`
	CreatedBy   string     `json:"createdBy"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	PublishedBy *string    `json:"publishedBy,omitempty"`
}

type RuleResponse struct {
	ID          string          `json:"id"`
	Version     int32           `json:"version"`
	Status      string          `json:"status"`
	Logic       string          `json:"logic"`
	Conditions  json.RawMessage `json:"conditions"`
	CreatedAt   time.Time       `json:"createdAt"`
	CreatedBy   string          `json:"createdBy"`
	PublishedAt *time.Time      `json:"publishedAt,omitempty"`
	PublishedBy *string         `json:"publishedBy,omitempty"`
}

type CreateRuleResponse struct {
	Status string       `json:"status"`
	Rule   RuleResponse `json:"rule"`
}

type PublishRuleResponse struct {
	Status string `json:"status"`
}

type CreateFunctionResponse struct {
	Status   string           `json:"status"`
	Function FunctionResponse `json:"function"`
}

type UpdateFunctionResponse struct {
	Function FunctionResponse `json:"function"`
}

type TerminalResponse struct {
	TerminalID string    `json:"terminalId"`
	CreatedAt  time.Time `json:"createdAt"`
	CreatedBy  string    `json:"createdBy"`
}

type WorkflowResponse struct {
	ID          string          `json:"id"`
	Version     int32           `json:"version"`
	Status      string          `json:"status"`
	StartAt     string          `json:"startAt"`
	Steps       json.RawMessage `json:"steps"`
	CreatedAt   time.Time       `json:"createdAt"`
	CreatedBy   string          `json:"createdBy"`
	PublishedAt *time.Time      `json:"publishedAt,omitempty"`
	PublishedBy *string         `json:"publishedBy,omitempty"`
}

type PublishWorkflowResponse struct {
	Status string `json:"status"`
}

// Error Response
type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Success Response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}
