package notification_sender

import (
	"context"

	domainNotification "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
)

type noOpSender struct{}

func NewNoOp() domainNotification.Sender {
	return &noOpSender{}
}

func (n *noOpSender) Send(_ context.Context, _ *domainNotification.NotificationData) error {
	return nil
}
