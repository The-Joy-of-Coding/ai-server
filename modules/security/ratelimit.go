package security

import (
	"net/http"

	"ai-server/modules/types"
)

type Ratelimit types.Ratelimit

func (l *Ratelimit) NewThread() {
	select {
	case l.Limit <- struct{}{}:
		go func() {
			defer func() { <-l.Limit }()
			l.Fn(l.W, l.R)
		}()
	default:
		http.Error(
			l.W, "Server Busy: Too many concurrent requests",
			http.StatusTooManyRequests,
		)
	}
}
