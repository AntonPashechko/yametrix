package client

type HTTPClient interface {
	Send() error
}
