package notification

type SMS interface {
	Phone() string
	Message() string
}
