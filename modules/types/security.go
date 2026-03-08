package types

import "net/http"

type Ratelimit struct {
	Fn    http.HandlerFunc
	W     http.ResponseWriter
	R     *http.Request
	Limit chan struct{}
}
