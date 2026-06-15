package backend

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 2*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func rateLimiter(next http.HandlerFunc, maxRequests int, window time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			next(w, r)
			return
		}

		if time.Since(v.lastSeen) > window {
			v.count = 1
			v.lastSeen = time.Now()
			mu.Unlock()
			next(w, r)
			return
		}

		v.count++
		v.lastSeen = time.Now()

		if v.count > maxRequests {
			mu.Unlock()
			http.Error(w, "Trop de requêtes, réessayez plus tard", http.StatusTooManyRequests)
			return
		}
		mu.Unlock()
		next(w, r)
	}
}
