package server

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/time/rate"

	"github.com/creatio313/movie_scheduler/internal/handlers"
	"github.com/creatio313/movie_scheduler/internal/middleware"
)

// DefaultPort はデフォルトのサーバーポート
const DefaultPort = "8080"

// SetupRouter は HTTPルーターを設定してハンドラーを登録します
func SetupRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// ヘルスチェック用エンドポイント (DBの死活監視は含めない)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API用エンドポイント
	// Projects CRUD
	mux.HandleFunc("POST /api/projects", handlers.HandleCreateProject(db))
	mux.HandleFunc("GET /api/projects/{id}", handlers.HandleGetProject(db))
	mux.HandleFunc("PUT /api/projects/{id}", handlers.HandleUpdateProject(db))
	mux.HandleFunc("DELETE /api/projects/{id}", handlers.HandleDeleteProject(db))

	// Casts CRUD
	mux.HandleFunc("POST /api/casts", handlers.HandleCreateCast(db))
	mux.HandleFunc("GET /api/casts/{id}", handlers.HandleGetCast(db))
	mux.HandleFunc("PUT /api/casts/{id}", handlers.HandleUpdateCast(db))
	mux.HandleFunc("DELETE /api/casts/{id}", handlers.HandleDeleteCast(db))
	mux.HandleFunc("GET /api/projects/{id}/casts", handlers.HandleListCastsByProject(db))

	// Scenes CRUD
	mux.HandleFunc("POST /api/scenes", handlers.HandleCreateScene(db))
	mux.HandleFunc("GET /api/scenes/{id}", handlers.HandleGetScene(db))
	mux.HandleFunc("PUT /api/scenes/{id}", handlers.HandleUpdateScene(db))
	mux.HandleFunc("DELETE /api/scenes/{id}", handlers.HandleDeleteScene(db))
	mux.HandleFunc("GET /api/projects/{id}/scenes", handlers.HandleListScenesByProject(db))
	mux.HandleFunc("GET /api/projects/{id}/scene_availabilities", handlers.HandleListSceneAvailabilitiesByProject(db))

	// CandidateDates CRUD
	mux.HandleFunc("POST /api/candidate_dates", handlers.HandleCreateCandidateDate(db))
	mux.HandleFunc("GET /api/candidate_dates/{id}", handlers.HandleGetCandidateDate(db))
	mux.HandleFunc("PUT /api/candidate_dates/{id}", handlers.HandleUpdateCandidateDate(db))
	mux.HandleFunc("DELETE /api/candidate_dates/{id}", handlers.HandleDeleteCandidateDate(db))
	mux.HandleFunc("GET /api/projects/{id}/candidate_dates", handlers.HandleListCandidateDatesByProject(db))

	// TimeSlotDefs CRUD
	mux.HandleFunc("POST /api/time_slots_def", handlers.HandleCreateTimeSlotDef(db))
	mux.HandleFunc("GET /api/time_slots_def/{id}", handlers.HandleGetTimeSlotDef(db))
	mux.HandleFunc("PUT /api/time_slots_def/{id}", handlers.HandleUpdateTimeSlotDef(db))
	mux.HandleFunc("DELETE /api/time_slots_def/{id}", handlers.HandleDeleteTimeSlotDef(db))
	mux.HandleFunc("GET /api/projects/{id}/time_slots_def", handlers.HandleListTimeSlotsDefByProject(db))

	// SceneAllowedTimeSlots CRUD
	mux.HandleFunc("POST /api/scene_allowed_time_slots", handlers.HandleCreateSceneAllowedTimeSlot(db))
	mux.HandleFunc("GET /api/scene_allowed_time_slots/{id}", handlers.HandleGetSceneAllowedTimeSlot(db))
	mux.HandleFunc("DELETE /api/scene_allowed_time_slots/{id}", handlers.HandleDeleteSceneAllowedTimeSlot(db))
	mux.HandleFunc("GET /api/scenes/{id}/scene_allowed_time_slots", handlers.HandleListSceneAllowedTimeSlotsByScene(db))

	// SceneRequiredCasts CRUD
	mux.HandleFunc("POST /api/scene_required_casts", handlers.HandleCreateSceneRequiredCast(db))
	mux.HandleFunc("GET /api/scene_required_casts/{id}", handlers.HandleGetSceneRequiredCast(db))
	mux.HandleFunc("DELETE /api/scene_required_casts/{id}", handlers.HandleDeleteSceneRequiredCast(db))
	mux.HandleFunc("GET /api/scenes/{id}/scene_required_casts", handlers.HandleListSceneRequiredCastsByScene(db))

	// CastAvailabilities CRUD
	mux.HandleFunc("POST /api/cast_availabilities", handlers.HandleCreateCastAvailability(db))
	mux.HandleFunc("GET /api/cast_availabilities/{id}", handlers.HandleGetCastAvailability(db))
	mux.HandleFunc("PUT /api/cast_availabilities/{id}", handlers.HandleUpdateCastAvailability(db))
	mux.HandleFunc("DELETE /api/cast_availabilities/{id}", handlers.HandleDeleteCastAvailability(db))
	mux.HandleFunc("GET /api/casts/{id}/cast_availabilities", handlers.HandleListCastAvailabilitiesByCast(db))

	// レート制限ミドルウェアを適用（1秒あたり10リクエスト、バースト20）
	rateLimiter := middleware.NewRateLimiter(rate.Limit(10), 20)

	// CORSとレート制限ミドルウェアを適用
	return middleware.CORSMiddleware(rateLimiter.RateLimitMiddleware(mux))
}

// Start はHTTPサーバーを起動します
func Start(db *sql.DB) error {
	handler := SetupRouter(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	slog.Info("Server is starting", "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed", "error", err)
		return err
	}
	return nil
}
