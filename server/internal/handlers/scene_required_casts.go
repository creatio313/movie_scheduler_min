package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/creatio313/movie_scheduler/internal/models"
	"github.com/creatio313/movie_scheduler/internal/response"
)

// [POST] /api/scene_required_casts : シーン必要役者の作成
func HandleCreateSceneRequiredCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s models.SceneRequiredCast
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		query := "INSERT INTO scene_required_casts (scene_id, cast_id) VALUES (?, ?) RETURNING id"
		err := db.QueryRowContext(r.Context(), query, s.SceneID, s.CastID).Scan(&s.ID)
		if err != nil {
			slog.Error("Failed to insert scene_required_cast", "error", err, "scene_id", s.SceneID)
			http.Error(w, "Failed to create scene_required_cast", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Scene required cast created", "id", s.ID, "scene_id", s.SceneID, "cast_id", s.CastID)
		response.RespondJSON(w, http.StatusCreated, s)
	}
}

// [GET] /api/scene_required_casts/{id} : シーン必要役者1件取得
func HandleGetSceneRequiredCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var s models.SceneRequiredCast
		err := db.QueryRowContext(r.Context(), "SELECT id, scene_id, cast_id FROM scene_required_casts WHERE id = ?", id).
			Scan(&s.ID, &s.SceneID, &s.CastID)

		if err == sql.ErrNoRows {
			http.Error(w, "Scene required cast not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get scene_required_cast", "error", err, "id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, s)
	}
}

// [DELETE] /api/scene_required_casts/{id} : シーン必要役者の削除
func HandleDeleteSceneRequiredCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		_, err := db.ExecContext(r.Context(), "DELETE FROM scene_required_casts WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete scene_required_cast", "error", err, "id", id)
			http.Error(w, "Failed to delete scene_required_cast", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Scene required cast deleted", "id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// [GET] /api/scenes/{id}/scene_required_casts : シーンに必要な役者一覧
func HandleListSceneRequiredCastsByScene(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sceneID := r.PathValue("id")

		rows, err := db.QueryContext(r.Context(), "SELECT id, scene_id, cast_id FROM scene_required_casts WHERE scene_id = ? ORDER BY id", sceneID)
		if err != nil {
			slog.Error("Failed to fetch scene_required_casts", "error", err, "scene_id", sceneID)
			http.Error(w, "Failed to fetch scene_required_casts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]models.SceneRequiredCast, 0)
		for rows.Next() {
			var s models.SceneRequiredCast
			if err := rows.Scan(&s.ID, &s.SceneID, &s.CastID); err != nil {
				http.Error(w, "Failed to read scene_required_casts", http.StatusInternalServerError)
				return
			}
			items = append(items, s)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read scene_required_casts", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}
