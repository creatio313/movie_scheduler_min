package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/creatio313/movie_scheduler/internal/models"
	"github.com/creatio313/movie_scheduler/internal/response"
	"github.com/creatio313/movie_scheduler/internal/validators"
)

// [POST] /api/scenes : シーンの作成
func HandleCreateScene(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var s models.Scene
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateScene(s); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query := "INSERT INTO scenes (project_id, scene_name, description) VALUES (?, ?, ?) RETURNING id"
		err := db.QueryRowContext(r.Context(), query, s.ProjectID, s.SceneName, s.Description).Scan(&s.ID)
		if err != nil {
			slog.Error("Failed to insert scene", "error", err, "project_id", s.ProjectID)
			http.Error(w, "Failed to create scene", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Scene created", "scene_id", s.ID, "scene_name", s.SceneName, "project_id", s.ProjectID)
		response.RespondJSON(w, http.StatusCreated, s)
	}
}

// [GET] /api/scenes/{id} : シーン1件取得
func HandleGetScene(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var s models.Scene
		var desc sql.NullString
		err := db.QueryRowContext(r.Context(), "SELECT id, project_id, scene_name, description FROM scenes WHERE id = ?", id).
			Scan(&s.ID, &s.ProjectID, &s.SceneName, &desc)

		if err == sql.ErrNoRows {
			http.Error(w, "Scene not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get scene", "error", err, "scene_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		s.Description = desc.String

		response.RespondJSON(w, http.StatusOK, s)
	}
}

// [PUT] /api/scenes/{id} : シーンの更新
func HandleUpdateScene(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var s models.Scene
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateScene(s); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.ExecContext(r.Context(), "UPDATE scenes SET scene_name = ?, description = ? WHERE id = ?", s.SceneName, s.Description, id)
		if err != nil {
			slog.Error("Failed to update scene", "error", err, "scene_id", id)
			http.Error(w, "Failed to update scene", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Scene updated", "scene_id", id, "scene_name", s.SceneName)
		sceneID, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Invalid scene ID format", "error", err, "scene_id", id)
			http.Error(w, "Invalid scene ID format", http.StatusInternalServerError)
			return
		}
		s.ID = sceneID
		response.RespondJSON(w, http.StatusOK, s)
	}
}

// [DELETE] /api/scenes/{id} : シーンの削除
func HandleDeleteScene(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		_, err := db.ExecContext(r.Context(), "DELETE FROM scenes WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete scene", "error", err, "scene_id", id)
			http.Error(w, "Failed to delete scene", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Scene deleted", "scene_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// [GET] /api/projects/{id}/scenes : プロジェクトのシーン一覧
func HandleListScenesByProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		rows, err := db.QueryContext(r.Context(), "SELECT id, project_id, scene_name, description FROM scenes WHERE project_id = ? ORDER BY id", projectID)
		if err != nil {
			slog.Error("Failed to fetch scenes", "error", err, "project_id", projectID)
			http.Error(w, "Failed to fetch scenes", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]models.Scene, 0)
		for rows.Next() {
			var s models.Scene
			var desc sql.NullString
			if err := rows.Scan(&s.ID, &s.ProjectID, &s.SceneName, &desc); err != nil {
				http.Error(w, "Failed to read scenes", http.StatusInternalServerError)
				return
			}
			s.Description = desc.String
			items = append(items, s)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read scenes", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}

// SceneAvailabilityRow はシーンごとの撮影可能日時を表す構造体
type SceneAvailabilityRow struct {
	SceneID    int    `json:"scene_id"`
	SceneName  string `json:"scene_name"`
	TargetDate string `json:"target_date"`
	TimeSlotID int    `json:"time_slot_id"`
	SlotName   string `json:"slot_name"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
}

// [GET] /api/projects/{id}/scene_availabilities : シーンの撮影可能日時一覧
func HandleListSceneAvailabilitiesByProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		query := `
SELECT
	sc.id AS scene_id,
	sc.scene_name AS scene_name,
	cd.target_date AS target_date,
	tsd.id AS time_slot_id,
	tsd.slot_name AS slot_name,
	tsd.start_time AS start_time,
	tsd.end_time AS end_time
FROM scenes sc
JOIN candidate_dates cd ON cd.project_id = sc.project_id
JOIN time_slots_def tsd ON tsd.project_id = sc.project_id
JOIN scene_allowed_time_slots sats ON sats.scene_id = sc.id AND sats.time_slot_id = tsd.id
JOIN scene_required_casts src ON src.scene_id = sc.id
JOIN cast_availabilities ca
	ON ca.candidate_date_id = cd.id
	AND ca.time_slot_id = tsd.id
	AND ca.cast_id = src.cast_id
	AND ca.is_available = 1
WHERE sc.project_id = ?
GROUP BY sc.id, cd.id, tsd.id
HAVING COUNT(DISTINCT ca.cast_id) = (
	SELECT COUNT(*) FROM scene_required_casts WHERE scene_id = sc.id
)
ORDER BY sc.id, cd.target_date, tsd.start_time
`

		rows, err := db.QueryContext(r.Context(), query, projectID)
		if err != nil {
			slog.Error("Failed to fetch scene availabilities", "error", err, "project_id", projectID)
			http.Error(w, "Failed to fetch scene availabilities", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]SceneAvailabilityRow, 0)
		for rows.Next() {
			var row SceneAvailabilityRow
			var startTime, endTime sql.NullString
			if err := rows.Scan(&row.SceneID, &row.SceneName, &row.TargetDate, &row.TimeSlotID, &row.SlotName, &startTime, &endTime); err != nil {
				http.Error(w, "Failed to read scene availabilities", http.StatusInternalServerError)
				return
			}
			row.StartTime = startTime.String
			row.EndTime = endTime.String
			items = append(items, row)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read scene availabilities", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}
