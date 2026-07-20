package notification

import "context"

type Dispatcher interface {
	SendSMS(ctx context.Context, phone string, message string)
}
