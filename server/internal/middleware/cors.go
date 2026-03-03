package middleware

import (
	"net/http"
	"os"
)

// CORSMiddleware は CORS対応のミドルウェアを提供します
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 環境変数からオリジンを取得（開発環境ではワイルドカード、本番環境では具体的なドメイン）
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "*" // デフォルトは開発環境用
		}

		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Credentialsは本番環境（ワイルドカード以外）のみ
		if allowedOrigin != "*" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
