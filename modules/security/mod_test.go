package security

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHeaders(t *testing.T) {
	// Check for "Auth Header" field in custom headers section
	// then verify if true or not.
}

func TestRateLimit(t *testing.T) {
	slowFn := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}

	t.Run("Logic_Block", func(t *testing.T) {
		limit := make(chan struct{}, 1) // Only 1 slot
		w1, r1 := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
		w2, r2 := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
		rl1 := &Ratelimit{Fn: slowFn, W: w1, R: r1, Limit: limit}
		rl1.NewThread()
		rl2 := &Ratelimit{Fn: slowFn, W: w2, R: r2, Limit: limit}
		rl2.NewThread()
		if len(limit) != 1 {
			t.Errorf("Expected 1 slot occupied, got %d", len(limit))
		}
	})

	t.Run("Key_Return", func(t *testing.T) {
		limit := make(chan struct{}, 1)
		w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
		rl := &Ratelimit{Fn: slowFn, W: w, R: r, Limit: limit}
		rl.NewThread()
		time.Sleep(100 * time.Millisecond)
		if len(limit) != 0 {
			t.Errorf("Key was not returned to the channel")
		}
	})

	t.Run("Stress_Flood", func(t *testing.T) {
		const maxLimit = 15
		const totalRequests = 1000
		limit := make(chan struct{}, maxLimit)
		var wg sync.WaitGroup
		for i := 0; i < totalRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
				rl := &Ratelimit{Fn: slowFn, W: w, R: r, Limit: limit}
				rl.NewThread()
			}()
		}
		wg.Wait()
		t.Log("Stress test completed without panics")
	})

	t.Run("Stress_Metrics", func(t *testing.T) {
		const maxLimit = 15
		const totalRequests = 1000
		limit := make(chan struct{}, maxLimit)
		var accepted int64
		var rejected int64
		var wg sync.WaitGroup
		totalFunc := func() {
			defer wg.Done()
			w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
			rl := &Ratelimit{
				Fn:    slowFn,
				W:     w,
				R:     r,
				Limit: limit,
			}
			rl.NewThread()
			if w.Code == http.StatusTooManyRequests {
				atomic.AddInt64(&rejected, 1)
			} else {
				atomic.AddInt64(&accepted, 1)
			}
		}
		for i := 0; i < totalRequests; i++ {
			wg.Add(1)
			go totalFunc()
			time.Sleep(2 * time.Millisecond)
		}
		wg.Wait()
		t.Logf("--- Final Stats ---")
		t.Logf("Total:    %d", totalRequests)
		t.Logf("Accepted: %d (Managed to get a key)", accepted)
		t.Logf("Rejected: %d (Saw 'Server Busy')", rejected)
	})
}
