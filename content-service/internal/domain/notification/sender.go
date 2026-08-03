package notification

import "context"

type Sender interface {
	Send(ctx context.Context, data *NotificationData) error
}

type NotificationData struct {
	Title     string
	Body      string
	Data      map[string]string
	Receivers []string
	ImageURL  string
	Count     int
}
