package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // MariaDB/MySQL用標準ドライバ

	"github.com/creatio313/movie_scheduler/internal/secretmanager"
	"github.com/creatio313/movie_scheduler/internal/server"
)

func main() {
	// 1. ロガーの初期化
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. データベース接続の設定
	var db *sql.DB
	var err error

	// Try to get database credentials from Secret Manager first
	vaultID := os.Getenv("SAKURA_VAULT_ID")
	secretName := os.Getenv("SAKURA_SECRET_NAME")

	if vaultID != "" {
		// Fetch database secret from Secret Manager
		client, err := secretmanager.NewSecretClient(vaultID, secretName)
		if err != nil {
			slog.Error("Failed to create Secret Manager client", "error", err)
			os.Exit(1)
		}

		dbSecret, err := client.FetchDatabaseSecret()
		if err != nil {
			slog.Error("Failed to fetch database secret", "error", err.Error())
			os.Exit(1)
		}

		// Construct DSN from secret
		dbDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbSecret.Username, dbSecret.Password, dbSecret.Host, dbSecret.Port, dbSecret.DatabaseName)

		// Open database connection
		db, err = sql.Open("mysql", dbDsn)
		if err != nil {
			slog.Error("DB connection failed", "error", err)
			os.Exit(1)
		}
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		defer db.Close()
	} else {
		// Fallback to environment variable if Secret Manager is not configured
		dbDsn := os.Getenv("DB_DSN")
		if dbDsn != "" {
			db, err = sql.Open("mysql", dbDsn)
			if err != nil {
				slog.Error("DB connection failed", "error", err)
				os.Exit(1)
			}
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(5)
			db.SetConnMaxLifetime(5 * time.Minute)
			defer db.Close()
		} else {
			slog.Warn("Neither SAKURA_VAULT_ID nor DB_DSN is set, running without DB connection")
		}
	}

	// 3. サーバーの起動
	if err := server.Start(db); err != nil {
		os.Exit(1)
	}
}
