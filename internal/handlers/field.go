package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rule-engine/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type FieldServiceInterface interface {
	CreateField(ctx context.Context, namespace string, field *domain.Field) error
	ListFields(ctx context.Context, namespace string) ([]*domain.Field, error)
	GetField(ctx context.Context, namespace, fieldID string) (*domain.Field, error)
}

type FieldHandler struct {
	fieldService FieldServiceInterface
}

func NewFieldHandler(fieldService FieldServiceInterface) *FieldHandler {
	return &FieldHandler{fieldService: fieldService}
}

// CreateFieldRequest represents the request body for creating a field
type CreateFieldRequest struct {
	FieldID     string `json:"fieldId" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
}

// CreateFieldResponse represents the response for creating a field
type CreateFieldResponse struct {
	Success bool `json:"success"`
	Field   struct {
		FieldID     string    `json:"fieldId"`
		Type        string    `json:"type"`
		Description string    `json:"description"`
		CreatedAt   time.Time `json:"createdAt"`
		CreatedBy   string    `json:"createdBy"`
	} `json:"field"`
}

// ListFields GET /v1/namespaces/{id}/fields
func (h *FieldHandler) ListFields(c *gin.Context) {
	namespace := c.Param("id")

	fields, err := h.fieldService.ListFields(c.Request.Context(), namespace)
	if err != nil {
		apiErr, ok := err.(*domain.APIError)
		if ok {
			c.JSON(apiErr.HTTPStatus(), apiErr)
		} else {
			c.JSON(http.StatusInternalServerError, domain.ErrListError)
		}
		return
	}

	// Convert to response format as per API spec
	var response []gin.H
	for _, field := range fields {
		response = append(response, gin.H{
			"fieldId":     field.FieldID,
			"type":        field.Type,
			"description": field.Description,
			"createdAt":   field.CreatedAt,
			"createdBy":   field.CreatedBy,
		})
	}

	c.JSON(http.StatusOK, response)
}

// CreateField POST /v1/namespaces/{id}/fields
func (h *FieldHandler) CreateField(c *gin.Context) {
	namespace := c.Param("id")

	var req CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrValidationError)
		return
	}

	// Get createdBy from JWT context
	createdBy, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, domain.ErrInternalError)
		return
	}

	field := &domain.Field{
		FieldID:     req.FieldID,
		Type:        req.Type,
		Description: req.Description,
		CreatedBy:   createdBy.(string),
	}

	err := h.fieldService.CreateField(c.Request.Context(), namespace, field)
	if err != nil {
		apiErr, ok := err.(*domain.APIError)
		if ok {
			c.JSON(apiErr.HTTPStatus(), apiErr)
		} else {
			c.JSON(http.StatusInternalServerError, domain.ErrInternalError)
		}
		return
	}

	// Get the created field to return in response
	createdField, err := h.fieldService.GetField(c.Request.Context(), namespace, req.FieldID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrInternalError)
		return
	}

	response := CreateFieldResponse{
		Success: true,
	}
	response.Field.FieldID = createdField.FieldID
	response.Field.Type = createdField.Type
	response.Field.Description = createdField.Description
	response.Field.CreatedAt = createdField.CreatedAt
	response.Field.CreatedBy = createdField.CreatedBy

	c.JSON(http.StatusCreated, response)
}

type mockFieldService struct{ mock.Mock }

func (m *mockFieldService) CreateField(ctx context.Context, namespace string, field *domain.Field) error {
	args := m.Called(ctx, namespace, field)
	return args.Error(0)
}

func (m *mockFieldService) ListFields(ctx context.Context, namespace string) ([]*domain.Field, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).([]*domain.Field), args.Error(1)
}

func (m *mockFieldService) GetField(ctx context.Context, namespace, fieldID string) (*domain.Field, error) {
	args := m.Called(ctx, namespace, fieldID)
	return args.Get(0).(*domain.Field), args.Error(1)
}

func TestFieldHandler_ListFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(mockFieldService)
	h := NewFieldHandler(mockService)

	r := gin.Default()
	r.GET("/v1/namespaces/:namespace/fields", func(c *gin.Context) { h.ListFields(c) })

	t.Run("success", func(t *testing.T) {
		fields := []*domain.Field{{FieldID: "salary", Type: "number", Description: "desc", CreatedAt: time.Now(), CreatedBy: "admin"}}
		mockService.On("ListFields", mock.Anything, "demo").Return(fields, nil)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/namespaces/demo/fields", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "salary", resp[0]["fieldId"])
	})

	t.Run("error", func(t *testing.T) {
		mockService.On("ListFields", mock.Anything, "fail").Return(nil, domain.ErrListError)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/namespaces/fail/fields", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "LIST_ERROR", resp["code"])
	})
}

func TestFieldHandler_CreateField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(mockFieldService)
	h := NewFieldHandler(mockService)

	r := gin.Default()
	r.POST("/v1/namespaces/:namespace/fields", func(c *gin.Context) {
		c.Set("client_id", "admin")
		h.CreateField(c)
	})

	field := &domain.Field{FieldID: "salary", Type: "number", Description: "desc", CreatedBy: "admin"}

	t.Run("success", func(t *testing.T) {
		mockService.On("CreateField", mock.Anything, "demo", mock.AnythingOfType("*domain.Field")).Return(nil)
		mockService.On("ListFields", mock.Anything, "demo").Return([]*domain.Field{field}, nil)
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"fieldId":"salary","type":"number","description":"desc"}`)
		req, _ := http.NewRequest("POST", "/v1/namespaces/demo/fields", body)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("validation error", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"type":"number"}`) // missing fieldId
		req, _ := http.NewRequest("POST", "/v1/namespaces/demo/fields", body)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "VALIDATION_ERROR", resp["code"])
	})

	t.Run("service error", func(t *testing.T) {
		mockService.On("CreateField", mock.Anything, "fail", mock.AnythingOfType("*domain.Field")).Return(domain.ErrFieldAlreadyExists)
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"fieldId":"salary","type":"number"}`)
		req, _ := http.NewRequest("POST", "/v1/namespaces/fail/fields", body)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "FIELD_ALREADY_EXISTS", resp["code"])
	})
}
