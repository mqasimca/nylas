package utils

import (
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name     string
		minDelay time.Duration
	}{
		{
			name:     "creates rate limiter with 500ms delay",
			minDelay: 500 * time.Millisecond,
		},
		{
			name:     "creates rate limiter with 1s delay",
			minDelay: 1 * time.Second,
		},
		{
			name:     "creates rate limiter with 100ms delay",
			minDelay: 100 * time.Millisecond,
		},
		{
			name:     "creates rate limiter with zero delay",
			minDelay: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.minDelay)
			if rl == nil {
				t.Fatal("expected non-nil RateLimiter")
			}
			if rl.minDelay != tt.minDelay {
				t.Errorf("expected minDelay %v, got %v", tt.minDelay, rl.minDelay)
			}
			if !rl.lastCall.IsZero() {
				t.Error("expected lastCall to be zero initially")
			}
		})
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	t.Run("first wait is immediate", func(t *testing.T) {
		rl := NewRateLimiter(100 * time.Millisecond)

		start := time.Now()
		rl.Wait()
		elapsed := time.Since(start)

		// First call should not wait
		if elapsed > 50*time.Millisecond {
			t.Errorf("first Wait() took too long: %v", elapsed)
		}
	})

	t.Run("second wait is delayed", func(t *testing.T) {
		rl := NewRateLimiter(100 * time.Millisecond)

		// First call
		rl.Wait()

		// Second call should wait
		start := time.Now()
		rl.Wait()
		elapsed := time.Since(start)

		// Should have waited approximately 100ms
		if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
			t.Errorf("second Wait() should wait ~100ms, got %v", elapsed)
		}
	})

	t.Run("waits correct duration after partial elapsed time", func(t *testing.T) {
		rl := NewRateLimiter(200 * time.Millisecond)

		// First call
		rl.Wait()

		// Wait 100ms (half of minDelay)
		time.Sleep(100 * time.Millisecond)

		// Second call should wait remaining ~100ms
		start := time.Now()
		rl.Wait()
		elapsed := time.Since(start)

		// Should have waited approximately 100ms (200ms - 100ms already elapsed)
		if elapsed < 90*time.Millisecond || elapsed > 150*time.Millisecond {
			t.Errorf("Wait() should wait ~100ms, got %v", elapsed)
		}
	})

	t.Run("no wait if enough time has passed", func(t *testing.T) {
		rl := NewRateLimiter(100 * time.Millisecond)

		// First call
		rl.Wait()

		// Wait longer than minDelay
		time.Sleep(150 * time.Millisecond)

		// Second call should not wait
		start := time.Now()
		rl.Wait()
		elapsed := time.Since(start)

		if elapsed > 50*time.Millisecond {
			t.Errorf("Wait() should not wait, got %v", elapsed)
		}
	})
}

func TestRateLimiter_TryWait(t *testing.T) {
	t.Run("first try is allowed", func(t *testing.T) {
		rl := NewRateLimiter(100 * time.Millisecond)

		allowed := rl.TryWait()
		if !allowed {
			t.Error("first TryWait() should be allowed")
		}
	})

	t.Run("second immediate try is not allowed", func(t *testing.T) {
		rl := NewRateLimiter(100 * time.Millisecond)

		// First call
		allowed1 := rl.TryWait()
		if !allowed1 {
			t.Fatal("first TryWait() should be allowed")
		}

		// Immediate second call
		allowed2 := rl.TryWait()
		if allowed2 {
			t.Error("immediate second TryWait() should not be allowed")
		}
	})

	t.Run("try is allowed after minDelay", func(t *testing.T) {
		rl := NewRateLimiter(100 * time.Millisecond)

		// First call
		rl.TryWait()

		// Wait for minDelay
		time.Sleep(110 * time.Millisecond)

		// Second call should be allowed
		allowed := rl.TryWait()
		if !allowed {
			t.Error("TryWait() should be allowed after minDelay")
		}
	})

	t.Run("try is not allowed before minDelay", func(t *testing.T) {
		rl := NewRateLimiter(200 * time.Millisecond)

		// First call
		rl.TryWait()

		// Wait less than minDelay
		time.Sleep(100 * time.Millisecond)

		// Second call should not be allowed
		allowed := rl.TryWait()
		if allowed {
			t.Error("TryWait() should not be allowed before minDelay")
		}
	})

	t.Run("non-blocking behavior", func(t *testing.T) {
		rl := NewRateLimiter(1 * time.Second)

		// First call
		rl.TryWait()

		// Immediate second call should return quickly (not block)
		start := time.Now()
		rl.TryWait()
		elapsed := time.Since(start)

		if elapsed > 50*time.Millisecond {
			t.Errorf("TryWait() should be non-blocking, took %v", elapsed)
		}
	})
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(50 * time.Millisecond)
	iterations := 100

	var wg sync.WaitGroup
	wg.Add(2)

	// Run two goroutines concurrently calling Wait
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			rl.Wait()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			rl.Wait()
		}
	}()

	// Should complete without deadlock or panic
	wg.Wait()
}

func TestRateLimiter_ConcurrentTryWait(t *testing.T) {
	rl := NewRateLimiter(50 * time.Millisecond)
	iterations := 100

	var wg sync.WaitGroup
	wg.Add(2)

	allowed1 := 0
	allowed2 := 0

	// Run two goroutines concurrently calling TryWait
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			if rl.TryWait() {
				allowed1++
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			if rl.TryWait() {
				allowed2++
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Should complete without deadlock or panic
	wg.Wait()

	// At least some calls should have been allowed
	total := allowed1 + allowed2
	if total == 0 {
		t.Error("expected some TryWait calls to be allowed")
	}
}

func TestRateLimiter_ZeroDelay(t *testing.T) {
	rl := NewRateLimiter(0)

	// With zero delay, all calls should be immediate
	for i := 0; i < 10; i++ {
		start := time.Now()
		rl.Wait()
		elapsed := time.Since(start)

		// Should not wait
		if elapsed > 50*time.Millisecond {
			t.Errorf("Wait() with zero delay should be immediate, took %v", elapsed)
		}
	}
}

func TestRateLimiter_TryWaitZeroDelay(t *testing.T) {
	rl := NewRateLimiter(0)

	// With zero delay, all calls should be allowed
	for i := 0; i < 10; i++ {
		allowed := rl.TryWait()
		if !allowed {
			t.Errorf("TryWait() with zero delay should always be allowed")
		}
	}
}
