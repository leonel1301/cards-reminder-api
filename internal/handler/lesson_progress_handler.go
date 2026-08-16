package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type LessonProgressHandler struct {
	lessonService *service.LessonProgressService
}

func NewLessonProgressHandler(lessonService *service.LessonProgressService) *LessonProgressHandler {
	return &LessonProgressHandler{lessonService: lessonService}
}

func (h *LessonProgressHandler) List(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	response, err := h.lessonService.List(c.Request.Context(), user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToListLessons)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *LessonProgressHandler) Mark(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	item, err := h.lessonService.Mark(c.Request.Context(), user.ID, c.Param("lessonId"))
	if err != nil {
		var validation service.ValidationError
		if errors.As(err, &validation) {
			respondValidationError(c, validation)
			return
		}
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToMarkLesson)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *LessonProgressHandler) Unmark(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	if err := h.lessonService.Unmark(c.Request.Context(), user.ID, c.Param("lessonId")); err != nil {
		var validation service.ValidationError
		if errors.As(err, &validation) {
			respondValidationError(c, validation)
			return
		}
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToUnmarkLesson)
		return
	}

	c.Status(http.StatusNoContent)
}
