package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	tooManyRequestsMessage = "too many requests"
	xForwardedForHeader    = "X-Forwarded-For"
	clientsTTL             = 5 * time.Minute
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientLimiter
	rps      rate.Limit
	burst    int
	stopChan chan struct{}
}

func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients:  make(map[string]*clientLimiter),
		rps:      rate.Limit(rps),
		burst:    burst,
		stopChan: make(chan struct{}),
	}
	go limiter.cleanupLoop()
	return limiter
}

func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !l.allow(ip) {
			http.Error(w, tooManyRequestsMessage, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *IPRateLimiter) Close() {
	close(l.stopChan)
}

func (l *IPRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	client, ok := l.clients[ip]
	if !ok {
		client = &clientLimiter{
			limiter:  rate.NewLimiter(l.rps, l.burst),
			lastSeen: time.Now(),
		}
		l.clients[ip] = client
	}
	client.lastSeen = time.Now()
	return client.limiter.Allow()
}

func (l *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(clientsTTL)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.stopChan:
			return
		}
	}
}

func (l *IPRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for ip, client := range l.clients {
		if now.Sub(client.lastSeen) > clientsTTL {
			delete(l.clients, ip)
		}
	}
}

func clientIP(r *http.Request) string {
	xff := r.Header.Get(xForwardedForHeader)
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
