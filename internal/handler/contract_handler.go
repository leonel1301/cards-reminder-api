package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type ContractHandler struct {
	contractService *service.ContractExtractService
	usageRepo       *repository.ContractUsageRepository
}

func NewContractHandler(
	contractService *service.ContractExtractService,
	usageRepo *repository.ContractUsageRepository,
) *ContractHandler {
	return &ContractHandler{
		contractService: contractService,
		usageRepo:       usageRepo,
	}
}

func (h *ContractHandler) GetUsage(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	used, err := h.usageRepo.GetAnalyzeCount(c.Request.Context(), user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToGetContractUsage)
		return
	}

	c.JSON(http.StatusOK, domain.NewContractUsage(used))
}

// Analyze accepts multipart form field "file" (PDF or image) and returns
// structured credit-card contract fields extracted by GPT.
func (h *ContractHandler) Analyze(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrContractFileRequired)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrContractFileRequired)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, (10<<20)+1))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrContractFileRequired)
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	if _, err := h.usageRepo.TryConsume(c.Request.Context(), user.ID, domain.ContractAnalyzeLimit); err != nil {
		if errors.Is(err, repository.ErrContractAnalyzeLimitReached) {
			respondError(c, http.StatusTooManyRequests, i18n.ErrContractAnalyzeLimitReached)
			return
		}
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToGetContractUsage)
		return
	}

	result, err := h.contractService.Extract(c.Request.Context(), service.ContractUpload{
		Filename:    fileHeader.Filename,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		_ = h.usageRepo.Release(c.Request.Context(), user.ID)

		var validation service.ValidationError
		if errors.As(err, &validation) {
			respondValidationError(c, validation)
			return
		}
		if strings.Contains(strings.ToLower(err.Error()), "api key") {
			respondError(c, http.StatusServiceUnavailable, i18n.ErrContractAIUnavailable)
			return
		}
		respondError(c, http.StatusBadGateway, i18n.ErrContractAnalyzeFailed)
		return
	}

	c.JSON(http.StatusOK, result)
}
