package client

import "net/http"

type Client interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}
