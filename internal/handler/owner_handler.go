package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
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
		respondUnauthenticated(c)
		return
	}

	owners, err := h.ownerService.List(c.Request.Context(), user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToListOwners)
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
		respondUnauthenticated(c)
		return
	}

	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidOwnerID)
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
		respondUnauthenticated(c)
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
		respondUnauthenticated(c)
		return
	}

	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidOwnerID)
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
		respondUnauthenticated(c)
		return
	}

	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidOwnerID)
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
		respondValidationError(c, validationErr)
	case errors.Is(err, service.ErrOwnerHasCards):
		respondError(c, http.StatusConflict, i18n.ErrOwnerHasAssignedCards)
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, i18n.ErrOwnerNotFound)
	default:
		respondError(c, http.StatusInternalServerError, i18n.ErrInternalServerError)
	}
}
