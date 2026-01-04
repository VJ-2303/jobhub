package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	rate    int
	burst   int
	clients map[string]*visitor
	mu      sync.Mutex
}

type visitor struct {
	limiter  *rate.Limiter
	lastseen time.Time
}

func NewLimiter(rate, burst int) *Limiter {
	return &Limiter{
		rate:    rate,
		burst:   burst,
		clients: make(map[string]*visitor),
	}
}

func (l *Limiter) CleanupClients() {
	for {
		time.Sleep(1 * time.Minute)

		l.mu.Lock()
		for ip, v := range l.clients {
			if time.Since(v.lastseen) > 3*time.Minute {
				delete(l.clients, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *Limiter) GetClient(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, found := l.clients[ip]

	if !found {
		limiter := rate.NewLimiter(rate.Limit(l.rate), l.burst)
		v := &visitor{limiter: limiter, lastseen: time.Now()}
		l.clients[ip] = v
		return limiter
	}
	l.clients[ip].lastseen = time.Now()
	return v.limiter
}

func (app *application) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := readUserIP(r)

		limiter := app.limiter.GetClient(ip)
		if limiter.Allow() == false {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Helper to handle Proxy Headers correctly
func readUserIP(r *http.Request) string {
	// Check X-Real-IP (Standard for Nginx)
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Check X-Forwarded-For (Standard for Load Balancers)
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
