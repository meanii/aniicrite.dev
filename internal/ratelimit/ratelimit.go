// Package ratelimit provides a small in-memory failed-attempt limiter,
// used to throttle admin login brute-force attempts per client IP.
package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	count   int
	resetAt time.Time
}

// Limiter blocks a key once it accumulates max failures within window.
// A successful action should call Reset to clear the key immediately.
type Limiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	max       int
	window    time.Duration
	nowFn     func() time.Time
	lastSwept time.Time
}

// New builds a Limiter allowing up to max failures per window.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		buckets: make(map[string]*bucket),
		max:     max,
		window:  window,
		nowFn:   time.Now,
	}
}

// Allow reports whether key may attempt again (i.e. is not currently blocked).
func (l *Limiter) Allow(key string) bool {
	now := l.nowFn()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sweep(now)
	b, ok := l.buckets[key]
	if !ok {
		return true
	}
	if now.After(b.resetAt) {
		delete(l.buckets, key)
		return true
	}
	return b.count < l.max
}

// Fail records a failed attempt for key, starting its window on the first miss.
func (l *Limiter) Fail(key string) {
	now := l.nowFn()
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{count: 1, resetAt: now.Add(l.window)}
		return
	}
	b.count++
}

// Reset clears a key after a successful action.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// sweep drops expired buckets, at most once per window, under the held lock.
func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.lastSwept) < l.window {
		return
	}
	l.lastSwept = now
	for k, b := range l.buckets {
		if now.After(b.resetAt) {
			delete(l.buckets, k)
		}
	}
}
