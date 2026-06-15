package service

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

var ErrNoDeviceTokens = fmt.Errorf("no device tokens registered")

type NotificationService struct {
	deviceRepo *repository.DeviceTokenRepository
	messaging  *messaging.Client
}

func NewNotificationService(deviceRepo *repository.DeviceTokenRepository, messagingClient *messaging.Client) *NotificationService {
	return &NotificationService{
		deviceRepo: deviceRepo,
		messaging:  messagingClient,
	}
}

func (s *NotificationService) SendToUser(ctx context.Context, userID uuid.UUID, notification domain.PushNotification) (*domain.PushSendResult, error) {
	tokens, err := s.deviceRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, ErrNoDeviceTokens
	}

	result := &domain.PushSendResult{}
	messages := make([]*messaging.Message, 0, len(tokens))
	tokenRefs := make([]string, 0, len(tokens))

	for _, device := range tokens {
		messages = append(messages, buildMessage(device.FCMToken, notification))
		tokenRefs = append(tokenRefs, device.FCMToken)
	}

	response, err := s.messaging.SendEach(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("send push notifications: %w", err)
	}

	for i, sendResponse := range response.Responses {
		if sendResponse.Success {
			result.SuccessCount++
			continue
		}

		result.FailureCount++
		if sendResponse.Error != nil && messaging.IsRegistrationTokenNotRegistered(sendResponse.Error) {
			invalidToken := tokenRefs[i]
			result.InvalidTokens = append(result.InvalidTokens, invalidToken)
			_ = s.deviceRepo.DeleteByFCMToken(ctx, invalidToken)
		}
	}

	return result, nil
}

func buildMessage(token string, notification domain.PushNotification) *messaging.Message {
	data := notification.Data
	if data == nil {
		data = map[string]string{}
	}

	return &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
		Data: data,
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: notification.Title,
						Body:  notification.Body,
					},
					Sound: "default",
				},
			},
		},
	}
}
