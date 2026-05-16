package server

import (
	"sync"
	"time"
)

type fixedWindowLimiter struct {
	mu      sync.Mutex
	now     func() time.Time
	window  time.Duration
	limit   int
	buckets map[string]rateBucket
}

type rateBucket struct {
	start time.Time
	count int
}

func newFixedWindowLimiter(window time.Duration, limit int) *fixedWindowLimiter {
	return &fixedWindowLimiter{
		now:     time.Now,
		window:  window,
		limit:   limit,
		buckets: map[string]rateBucket{},
	}
}

func (l *fixedWindowLimiter) Allow(key string) bool {
	if key == "" || l.limit <= 0 {
		return false
	}
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()
	bucket := l.buckets[key]
	if bucket.start.IsZero() || now.Sub(bucket.start) >= l.window {
		l.buckets[key] = rateBucket{start: now, count: 1}
		l.pruneLocked(now)
		return true
	}
	if bucket.count >= l.limit {
		return false
	}
	bucket.count++
	l.buckets[key] = bucket
	return true
}

func (l *fixedWindowLimiter) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = map[string]rateBucket{}
}

func (l *fixedWindowLimiter) pruneLocked(now time.Time) {
	for key, bucket := range l.buckets {
		if now.Sub(bucket.start) >= l.window {
			delete(l.buckets, key)
		}
	}
}
