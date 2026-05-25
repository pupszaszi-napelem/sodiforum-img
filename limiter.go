package main

import (
	"net"
	"sync"
	"time"
)

const (
	rateWindow = time.Minute
	rateLimit  = 20
)

type rateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{hits: make(map[string][]time.Time)}
}

func (l *rateLimiter) allow(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	now := time.Now()
	cutoff := now.Add(-rateWindow)

	l.mu.Lock()
	defer l.mu.Unlock()

	hits := l.hits[host]
	kept := hits[:0]
	for _, hit := range hits {
		if hit.After(cutoff) {
			kept = append(kept, hit)
		}
	}
	if len(kept) >= rateLimit {
		l.hits[host] = kept
		return false
	}
	l.hits[host] = append(kept, now)
	return true
}
