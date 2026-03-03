package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter はIPアドレスごとにレート制限を管理します
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter は新しいレート制限マネージャーを作成します
// r: 1秒あたりの許可リクエスト数
// b: バーストサイズ（短時間に許可する最大リクエスト数）
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter は指定されたIPアドレスのリミッターを取得または作成します
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware はレート制限を適用するミドルウェアを提供します
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// クライアントIPを取得（X-Forwarded-Forヘッダーを優先）
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.Header.Get("X-Real-IP")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}

		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
