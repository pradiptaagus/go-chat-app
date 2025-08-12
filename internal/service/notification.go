package service

import (
	"context"
	"log"
)

type NotificationPayload struct {
	message string
	client  Client
}

type Notification interface {
	SendNotification(ctx context.Context, payload *NotificationPayload)
	ReceivedNotification(ctx context.Context) <-chan NotificationPayload
}

type NotificationImpl struct {
	notification chan NotificationPayload
}

func (notification *NotificationImpl) SendNotification(ctx context.Context, payload *NotificationPayload) {
	select {
	case <-ctx.Done():
		log.Printf("Sending notification was cancelled: %s to client %v", payload.message, payload.client)
		return
	default:
		notification.notification <- *payload
		log.Printf("Sending notification to client %v: %s", payload.client, payload.message)
	}
}

func (notification *NotificationImpl) ReceivedNotification(ctx context.Context) <-chan NotificationPayload {
	select {
	case <-ctx.Done():
		log.Println("Receiving notification was cancelled")
		return nil
	default:
		return notification.notification
	}
}

func NewNotification() Notification {
	return &NotificationImpl{
		notification: make(chan NotificationPayload, 2),
	}
}
