package types

import "net/http"

type Ratelimit struct {
	fn http.HandlerFunc
	w  http.ResponseWriter
	r  *http.Request
}
