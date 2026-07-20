package notification

import (
	"context"
	"log"
)

type AsyncDispatcher struct {
	service *Service
}

func NewAsyncDispatcher(service *Service) *AsyncDispatcher {
	return &AsyncDispatcher{service: service}
}

func (d *AsyncDispatcher) SendSMS(ctx context.Context, phone, message string) {

	go func() {

		if err := d.service.Send(context.Background(), phone, message); err != nil {
			log.Println(err)
		}

	}()

}
