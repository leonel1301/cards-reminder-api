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

type CardHandler struct {
	cardService *service.CardService
}

func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

type createCardRequest struct {
	Name            string     `json:"name" binding:"required"`
	LastFourDigits  string     `json:"last_four_digits" binding:"required"`
	Issuer          *string    `json:"issuer"`
	BillingCycleDay int        `json:"billing_cycle_day" binding:"required"`
	PaymentDueDay   int        `json:"payment_due_day" binding:"required"`
	ColorHex        *string    `json:"color_hex"`
	Notes           *string    `json:"notes"`
	OwnerID         *uuid.UUID `json:"owner_id"`
}

type updateCardRequest struct {
	Name            *string    `json:"name"`
	LastFourDigits  *string    `json:"last_four_digits"`
	Issuer          *string    `json:"issuer"`
	BillingCycleDay *int       `json:"billing_cycle_day"`
	PaymentDueDay   *int       `json:"payment_due_day"`
	ColorHex        *string    `json:"color_hex"`
	Notes           *string    `json:"notes"`
	IsActive        *bool      `json:"is_active"`
	OwnerID         *uuid.UUID `json:"owner_id"`
}

func (h *CardHandler) List(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cards, err := h.cardService.List(c.Request.Context(), user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToListCards)
		return
	}

	if cards == nil {
		cards = []domain.Card{}
	}

	c.JSON(http.StatusOK, cards)
}

func (h *CardHandler) Get(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	card, err := h.cardService.Get(c.Request.Context(), user.ID, cardID)
	if err != nil {
		h.handleCardError(c, err)
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) Create(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	var req createCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.cardService.Create(c.Request.Context(), user.ID, domain.CreateCardInput{
		Name:            req.Name,
		LastFourDigits:  req.LastFourDigits,
		Issuer:          req.Issuer,
		BillingCycleDay: req.BillingCycleDay,
		PaymentDueDay:   req.PaymentDueDay,
		ColorHex:        req.ColorHex,
		Notes:           req.Notes,
		OwnerID:         req.OwnerID,
	})
	if err != nil {
		h.handleCardError(c, err)
		return
	}

	c.JSON(http.StatusCreated, card)
}

func (h *CardHandler) Update(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	var req updateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.cardService.Update(c.Request.Context(), user.ID, cardID, domain.UpdateCardInput{
		Name:            req.Name,
		LastFourDigits:  req.LastFourDigits,
		Issuer:          req.Issuer,
		BillingCycleDay: req.BillingCycleDay,
		PaymentDueDay:   req.PaymentDueDay,
		ColorHex:        req.ColorHex,
		Notes:           req.Notes,
		IsActive:        req.IsActive,
		OwnerID:         req.OwnerID,
	})
	if err != nil {
		h.handleCardError(c, err)
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) Delete(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	if err := h.cardService.Delete(c.Request.Context(), user.ID, cardID); err != nil {
		h.handleCardError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CardHandler) handleCardError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		respondValidationError(c, validationErr)
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, i18n.ErrCardNotFound)
	default:
		respondError(c, http.StatusInternalServerError, i18n.ErrInternalServerError)
	}
}
