package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type OwnerHandler struct {
	ownerService *service.OwnerService
}

func NewOwnerHandler(ownerService *service.OwnerService) *OwnerHandler {
	return &OwnerHandler{ownerService: ownerService}
}

type createOwnerRequest struct {
	Name      string `json:"name" binding:"required"`
	SalaryDay *int   `json:"salary_day"`
}

type updateOwnerRequest struct {
	Name      *string `json:"name"`
	SalaryDay *int    `json:"salary_day"`
}

func (h *OwnerHandler) List(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	owners, err := h.ownerService.List(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list owners"})
		return
	}

	if owners == nil {
		owners = []domain.Owner{}
	}

	c.JSON(http.StatusOK, owners)
}

func (h *OwnerHandler) Get(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner id"})
		return
	}

	owner, err := h.ownerService.Get(c.Request.Context(), user.ID, ownerID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, owner)
}

func (h *OwnerHandler) Create(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var req createOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owner, err := h.ownerService.Create(c.Request.Context(), user.ID, domain.CreateOwnerInput{
		Name:      req.Name,
		SalaryDay: req.SalaryDay,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, owner)
}

func (h *OwnerHandler) Update(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner id"})
		return
	}

	var req updateOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owner, err := h.ownerService.Update(c.Request.Context(), user.ID, ownerID, domain.UpdateOwnerInput{
		Name:      req.Name,
		SalaryDay: req.SalaryDay,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, owner)
}

func (h *OwnerHandler) Delete(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner id"})
		return
	}

	if err := h.ownerService.Delete(c.Request.Context(), user.ID, ownerID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *OwnerHandler) handleError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
	case errors.Is(err, service.ErrOwnerHasCards):
		c.JSON(http.StatusConflict, gin.H{"error": "Owner has assigned cards"})
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "owner not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
