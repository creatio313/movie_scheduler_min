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

// [POST] /api/time_slots_def : 時間枠の作成
func HandleCreateTimeSlotDef(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t models.TimeSlotDef
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateTimeSlotDef(t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var startArg interface{}
		var endArg interface{}
		if t.StartTime != "" {
			startArg = t.StartTime
		}
		if t.EndTime != "" {
			endArg = t.EndTime
		}

		query := "INSERT INTO time_slots_def (project_id, slot_name, start_time, end_time) VALUES (?, ?, ?, ?) RETURNING id"
		err := db.QueryRowContext(r.Context(), query, t.ProjectID, t.SlotName, startArg, endArg).Scan(&t.ID)
		if err != nil {
			slog.Error("Failed to insert time_slot_def", "error", err, "project_id", t.ProjectID)
			http.Error(w, "Failed to create time_slot_def", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Time slot created", "time_slot_id", t.ID, "slot_name", t.SlotName, "project_id", t.ProjectID)
		response.RespondJSON(w, http.StatusCreated, t)
	}
}

// [GET] /api/time_slots_def/{id} : 時間枠1件取得
func HandleGetTimeSlotDef(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var t models.TimeSlotDef
		var startTime, endTime sql.NullString
		err := db.QueryRowContext(r.Context(), "SELECT id, project_id, slot_name, start_time, end_time FROM time_slots_def WHERE id = ?", id).
			Scan(&t.ID, &t.ProjectID, &t.SlotName, &startTime, &endTime)

		if err == sql.ErrNoRows {
			http.Error(w, "Time slot definition not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get time_slot_def", "error", err, "time_slot_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		t.StartTime = startTime.String
		t.EndTime = endTime.String

		response.RespondJSON(w, http.StatusOK, t)
	}
}

// [PUT] /api/time_slots_def/{id} : 時間枠の更新
func HandleUpdateTimeSlotDef(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var t models.TimeSlotDef
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateTimeSlotDef(t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var startArg interface{}
		var endArg interface{}
		if t.StartTime != "" {
			startArg = t.StartTime
		}
		if t.EndTime != "" {
			endArg = t.EndTime
		}

		_, err := db.ExecContext(r.Context(), "UPDATE time_slots_def SET slot_name = ?, start_time = ?, end_time = ? WHERE id = ?", t.SlotName, startArg, endArg, id)
		if err != nil {
			slog.Error("Failed to update time_slot_def", "error", err, "time_slot_id", id)
			http.Error(w, "Failed to update time_slot_def", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Time slot updated", "time_slot_id", id, "slot_name", t.SlotName)
		timeSlotID, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Invalid time_slot ID format", "error", err, "time_slot_id", id)
			http.Error(w, "Invalid time_slot ID format", http.StatusInternalServerError)
			return
		}
		t.ID = timeSlotID
		response.RespondJSON(w, http.StatusOK, t)
	}
}

// [DELETE] /api/time_slots_def/{id} : 時間枠の削除
func HandleDeleteTimeSlotDef(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		_, err := db.ExecContext(r.Context(), "DELETE FROM time_slots_def WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete time_slot_def", "error", err, "time_slot_id", id)
			http.Error(w, "Failed to delete time_slot_def", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Time slot deleted", "time_slot_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// [GET] /api/projects/{id}/time_slots_def : プロジェクトの時間枠一覧
func HandleListTimeSlotsDefByProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		rows, err := db.QueryContext(r.Context(), "SELECT id, project_id, slot_name, start_time, end_time FROM time_slots_def WHERE project_id = ? ORDER BY start_time", projectID)
		if err != nil {
			slog.Error("Failed to fetch time_slots_def", "error", err, "project_id", projectID)
			http.Error(w, "Failed to fetch time_slots_def", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]models.TimeSlotDef, 0)
		for rows.Next() {
			var t models.TimeSlotDef
			var startTime, endTime sql.NullString
			if err := rows.Scan(&t.ID, &t.ProjectID, &t.SlotName, &startTime, &endTime); err != nil {
				http.Error(w, "Failed to read time_slots_def", http.StatusInternalServerError)
				return
			}
			t.StartTime = startTime.String
			t.EndTime = endTime.String
			items = append(items, t)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read time_slots_def", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}
