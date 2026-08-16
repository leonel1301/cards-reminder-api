package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

const (
	maxContractUploadBytes = 10 << 20 // 10 MiB
	openaiFilesURL         = "https://api.openai.com/v1/files"
	openaiChatURL          = "https://api.openai.com/v1/chat/completions"
)

type ContractExtractService struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewContractExtractService(apiKey, model string) *ContractExtractService {
	if model == "" {
		model = "gpt-4o"
	}
	return &ContractExtractService{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type ContractUpload struct {
	Filename    string
	ContentType string
	Data        []byte
}

func (s *ContractExtractService) Extract(ctx context.Context, upload ContractUpload) (*domain.ContractExtraction, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("openai api key is not configured")
	}
	if len(upload.Data) == 0 {
		return nil, ValidationError{Field: "file", Message: "cannot be empty"}
	}
	if len(upload.Data) > maxContractUploadBytes {
		return nil, ValidationError{Field: "file", Message: "is too large"}
	}

	contentType := normalizeContentType(upload.ContentType, upload.Filename)
	if !isAllowedContractType(contentType) {
		return nil, ValidationError{Field: "file", Message: "must be a PDF or image"}
	}

	raw, err := s.analyze(ctx, upload.Filename, contentType, upload.Data)
	if err != nil {
		return nil, err
	}

	extraction, err := parseExtraction(raw)
	if err != nil {
		return nil, fmt.Errorf("parse extraction: %w", err)
	}
	return extraction, nil
}

func normalizeContentType(contentType, filename string) string {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if ct != "" && ct != "application/octet-stream" {
		return ct
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".heic", ".heif":
		return "image/heic"
	default:
		return ct
	}
}

func isAllowedContractType(contentType string) bool {
	switch contentType {
	case "application/pdf",
		"image/png",
		"image/jpeg",
		"image/jpg",
		"image/webp",
		"image/heic",
		"image/heif":
		return true
	default:
		return false
	}
}

func (s *ContractExtractService) analyze(ctx context.Context, filename, contentType string, data []byte) (string, error) {
	content := []any{
		map[string]any{
			"type": "text",
			"text": contractExtractionPrompt,
		},
	}

	var uploadedFileID string
	if strings.HasPrefix(contentType, "image/") {
		content = append(content, map[string]any{
			"type": "image_url",
			"image_url": map[string]any{
				"url": fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(data)),
			},
		})
	} else {
		fileID, err := s.uploadFile(ctx, filename, contentType, data)
		if err != nil {
			return "", fmt.Errorf("openai file upload: %w", err)
		}
		uploadedFileID = fileID
		content = append(content, map[string]any{
			"type": "file",
			"file": map[string]any{
				"file_id": fileID,
			},
		})
	}

	if uploadedFileID != "" {
		defer s.deleteFile(context.Background(), uploadedFileID)
	}

	payload := map[string]any{
		"model": s.model,
		"messages": []any{
			map[string]any{
				"role":    "user",
				"content": content,
			},
		},
		"response_format": map[string]any{"type": "json_object"},
		"temperature":     0.1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiChatURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai chat status %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("openai returned empty content")
	}
	return parsed.Choices[0].Message.Content, nil
}

func (s *ContractExtractService) uploadFile(ctx context.Context, filename, contentType string, data []byte) (string, error) {
	if filename == "" {
		filename = "contract.pdf"
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("purpose", "user_data")
	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {
			fmt.Sprintf(`form-data; name="file"; filename="%s"`, sanitizeFilename(filename)),
		},
		"Content-Type": {contentType},
	})
	if err != nil {
		return "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openaiFilesURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai files status %d: %s", resp.StatusCode, truncate(string(respBody), 400))
	}

	var parsed struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if parsed.ID == "" {
		return "", fmt.Errorf("openai files response missing id")
	}
	return parsed.ID, nil
}

func (s *ContractExtractService) deleteFile(ctx context.Context, fileID string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, openaiFilesURL+"/"+fileID, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}

const contractExtractionPrompt = `You are extracting structured data from a credit card contract or statement for a personal finance reminder app (Waloop).
Return ONLY valid JSON with this exact shape:
{
  "name": string|null,
  "last_four_digits": string|null,
  "issuer": string|null,
  "billing_cycle_day": number|null,
  "payment_due_day": number|null,
  "annual_fee": string|null,
  "interest_rate_summary": string|null,
  "notes": string|null,
  "summary": string,
  "confidence": "high"|"medium"|"low",
  "warnings": string[]
}

Rules:
- name: product/card name if present.
- last_four_digits: exactly 4 digits when visible; otherwise null.
- issuer: bank/issuer brand (e.g. Visa, Mastercard, BBVA, Interbank).
- billing_cycle_day / payment_due_day: day of month 1-31 when clearly stated (cut-off / due date). Prefer due date for payment_due_day and statement closing day for billing_cycle_day.
- notes: short practical notes for the user (fees, grace period, important clauses), Spanish if the document is Spanish, English if English.
- summary: 1-3 sentence overview of what you found.
- warnings: list uncertain fields or missing data.
- Do not invent numbers. Use null when unknown.
- Ignore promotional fluff; focus on dates, fees, rates, and identifiers.`

func parseExtraction(raw string) (*domain.ContractExtraction, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var extraction domain.ContractExtraction
	if err := json.Unmarshal([]byte(raw), &extraction); err != nil {
		return nil, err
	}
	if extraction.Confidence == "" {
		extraction.Confidence = "low"
	}
	if extraction.Warnings == nil {
		extraction.Warnings = []string{}
	}
	if extraction.LastFourDigits != nil {
		digits := onlyDigits(*extraction.LastFourDigits)
		if len(digits) == 4 {
			extraction.LastFourDigits = &digits
		} else {
			extraction.LastFourDigits = nil
		}
	}
	extraction.BillingCycleDay = clampDay(extraction.BillingCycleDay)
	extraction.PaymentDueDay = clampDay(extraction.PaymentDueDay)
	return &extraction, nil
}

func clampDay(day *int) *int {
	if day == nil {
		return nil
	}
	if *day < 1 || *day > 31 {
		return nil
	}
	return day
}

func onlyDigits(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, `"`, "")
	if name == "" || name == "." || name == ".." {
		return "contract.pdf"
	}
	return name
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "…"
}
