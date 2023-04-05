package client

type HttpClient interface {
	Send() error
}
