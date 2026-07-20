package notification

import "context"

type Service struct {
	provider Provider
}

func NewService(provider Provider) *Service {
	return &Service{provider: provider}
}

func (s *Service) Send(ctx context.Context, phone, message string) error {

	return s.provider.Send(ctx, phone, message)

}
