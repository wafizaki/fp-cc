package handlers

import (
	"context"
	"net/http"
	"strconv"
	"tenant-manager/models"
	"tenant-manager/services"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	service *services.TenantService
}

func NewTenantHandler(service *services.TenantService) *TenantHandler {
	return &TenantHandler{
		service: service,
	}
}

func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req models.CreateTenantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Invalid request body", err))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("Validation failed", err))
		return
	}

	ctx := context.Background()
	tenant, err := h.service.CreateTenant(ctx, req.Name)
	if err != nil {
		if contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, models.NewErrorResponse(
				"Tenant already exists",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to create tenant", err))
		return
	}

	c.JSON(http.StatusCreated, models.NewSuccessResponse("Tenant created successfully", tenant))
}

func (h *TenantHandler) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	ctx := context.Background()
	tenants, meta, err := h.service.ListTenants(ctx, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to retrieve tenants", err))
		return
	}

	c.JSON(http.StatusOK, models.NewPaginatedResponse("Tenants retrieved successfully", tenants, meta))
}

func (h *TenantHandler) GetTenant(c *gin.Context) {
	name := c.Param("name")

	ctx := context.Background()
	tenant, err := h.service.GetTenant(ctx, name)
	if err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Tenant not found",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to retrieve tenant", err))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse("Tenant retrieved successfully", tenant))
}

func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	name := c.Param("name")

	ctx := context.Background()
	if err := h.service.DeleteTenant(ctx, name); err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Tenant not found",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to delete tenant", err))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse("Tenant deleted successfully", nil))
}

func (h *TenantHandler) StopContainer(c *gin.Context) {
	name := c.Param("name")

	ctx := context.Background()
	if err := h.service.StopTenantContainer(ctx, name); err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Tenant not found",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to stop container", err))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse("Container stopped successfully", nil))
}

func (h *TenantHandler) StartContainer(c *gin.Context) {
	name := c.Param("name")

	ctx := context.Background()
	if err := h.service.StartTenantContainer(ctx, name); err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.NewErrorResponse(
				"Tenant not found",
				err,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("Failed to start container", err))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse("Container started successfully", gin.H{
		"name":   name,
		"status": "running",
	}))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr)+1 && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

