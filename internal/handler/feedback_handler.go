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

type FeedbackHandler struct {
	feedbackService *service.FeedbackService
}

func NewFeedbackHandler(feedbackService *service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{feedbackService: feedbackService}
}

type createFeedbackRequest struct {
	Title   string `json:"title" binding:"required"`
	Device  string `json:"device" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type updateFeedbackRequest struct {
	Title   *string `json:"title"`
	Device  *string `json:"device"`
	Content *string `json:"content"`
}

func (h *FeedbackHandler) ListForAdmin(c *gin.Context) {
	items, err := h.feedbackService.ListAllForAdmin(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToListFeedback)
		return
	}

	if items == nil {
		items = []domain.FeedbackAdminItem{}
	}

	c.JSON(http.StatusOK, items)
}

func (h *FeedbackHandler) ListByUser(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	items, err := h.feedbackService.ListByUserID(c.Request.Context(), user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToListFeedback)
		return
	}

	if items == nil {
		items = []domain.Feedback{}
	}

	c.JSON(http.StatusOK, items)
}

func (h *FeedbackHandler) Get(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	feedbackID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidFeedbackID)
		return
	}

	item, err := h.feedbackService.Get(c.Request.Context(), user.ID, feedbackID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *FeedbackHandler) Create(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	var req createFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.feedbackService.Create(c.Request.Context(), user.ID, domain.CreateFeedbackInput{
		Title:   req.Title,
		Device:  req.Device,
		Content: req.Content,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *FeedbackHandler) Update(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	feedbackID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidFeedbackID)
		return
	}

	var req updateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.feedbackService.Update(c.Request.Context(), user.ID, feedbackID, domain.UpdateFeedbackInput{
		Title:   req.Title,
		Device:  req.Device,
		Content: req.Content,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *FeedbackHandler) Delete(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	feedbackID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidFeedbackID)
		return
	}

	if err := h.feedbackService.Delete(c.Request.Context(), user.ID, feedbackID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *FeedbackHandler) handleError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		respondValidationError(c, validationErr)
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, i18n.ErrFeedbackNotFound)
	default:
		respondError(c, http.StatusInternalServerError, i18n.ErrInternalServerError)
	}
}
