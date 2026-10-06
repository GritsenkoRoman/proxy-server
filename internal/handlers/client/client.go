package client

import "net/http"

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
