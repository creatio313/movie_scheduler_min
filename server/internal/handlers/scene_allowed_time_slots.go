package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/creatio313/movie_scheduler/internal/models"
	"github.com/creatio313/movie_scheduler/internal/response"
)

// [POST] /api/scene_allowed_time_slots : シーン許可時間枠の作成
func HandleCreateSceneAllowedTimeSlot(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s models.SceneAllowedTimeSlot
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		query := "INSERT INTO scene_allowed_time_slots (scene_id, time_slot_id) VALUES (?, ?) RETURNING id"
		err := db.QueryRowContext(r.Context(), query, s.SceneID, s.TimeSlotID).Scan(&s.ID)
		if err != nil {
			slog.Error("Failed to insert scene_allowed_time_slot", "error", err, "scene_id", s.SceneID)
			http.Error(w, "Failed to create scene_allowed_time_slot", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Scene allowed time slot created", "id", s.ID, "scene_id", s.SceneID, "time_slot_id", s.TimeSlotID)
		response.RespondJSON(w, http.StatusCreated, s)
	}
}

// [GET] /api/scene_allowed_time_slots/{id} : シーン許可時間枠1件取得
func HandleGetSceneAllowedTimeSlot(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var s models.SceneAllowedTimeSlot
		err := db.QueryRowContext(r.Context(), "SELECT id, scene_id, time_slot_id FROM scene_allowed_time_slots WHERE id = ?", id).
			Scan(&s.ID, &s.SceneID, &s.TimeSlotID)

		if err == sql.ErrNoRows {
			http.Error(w, "Scene allowed time slot not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get scene_allowed_time_slot", "error", err, "id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, s)
	}
}

// [DELETE] /api/scene_allowed_time_slots/{id} : シーン許可時間枠の削除
func HandleDeleteSceneAllowedTimeSlot(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		_, err := db.ExecContext(r.Context(), "DELETE FROM scene_allowed_time_slots WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete scene_allowed_time_slot", "error", err, "id", id)
			http.Error(w, "Failed to delete scene_allowed_time_slot", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Scene allowed time slot deleted", "id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// [GET] /api/scenes/{id}/scene_allowed_time_slots : シーンで撮影可能な時間枠一覧
func HandleListSceneAllowedTimeSlotsByScene(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sceneID := r.PathValue("id")

		rows, err := db.QueryContext(r.Context(), "SELECT id, scene_id, time_slot_id FROM scene_allowed_time_slots WHERE scene_id = ? ORDER BY id", sceneID)
		if err != nil {
			slog.Error("Failed to fetch scene_allowed_time_slots", "error", err, "scene_id", sceneID)
			http.Error(w, "Failed to fetch scene_allowed_time_slots", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]models.SceneAllowedTimeSlot, 0)
		for rows.Next() {
			var s models.SceneAllowedTimeSlot
			if err := rows.Scan(&s.ID, &s.SceneID, &s.TimeSlotID); err != nil {
				http.Error(w, "Failed to read scene_allowed_time_slots", http.StatusInternalServerError)
				return
			}
			items = append(items, s)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read scene_allowed_time_slots", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}
