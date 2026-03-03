package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/creatio313/movie_scheduler/internal/models"
	"github.com/creatio313/movie_scheduler/internal/response"
)

// [POST] /api/cast_availabilities : 役者スケジュール可用性の作成
func HandleCreateCastAvailability(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c models.CastAvailability
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		query := "INSERT INTO cast_availabilities (candidate_date_id, time_slot_id, cast_id, is_available) VALUES (?, ?, ?, ?) RETURNING id"
		err := db.QueryRowContext(r.Context(), query, c.CandidateDateID, c.TimeSlotID, c.CastID, c.IsAvailable).Scan(&c.ID)
		if err != nil {
			slog.Error("Failed to insert cast_availability", "error", err, "cast_id", c.CastID)
			http.Error(w, "Failed to create cast_availability", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Cast availability created", "availability_id", c.ID, "cast_id", c.CastID, "is_available", c.IsAvailable)
		response.RespondJSON(w, http.StatusCreated, c)
	}
}

// [GET] /api/cast_availabilities/{id} : 役者スケジュール可用性1件取得
func HandleGetCastAvailability(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var c models.CastAvailability
		err := db.QueryRowContext(r.Context(), "SELECT id, candidate_date_id, time_slot_id, cast_id, is_available FROM cast_availabilities WHERE id = ?", id).
			Scan(&c.ID, &c.CandidateDateID, &c.TimeSlotID, &c.CastID, &c.IsAvailable)

		if err == sql.ErrNoRows {
			http.Error(w, "Cast availability not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get cast_availability", "error", err, "availability_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, c)
	}
}

// [PUT] /api/cast_availabilities/{id} : 役者スケジュール可用性の更新
func HandleUpdateCastAvailability(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var c models.CastAvailability
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		_, err := db.ExecContext(r.Context(), "UPDATE cast_availabilities SET is_available = ? WHERE id = ?", c.IsAvailable, id)
		if err != nil {
			slog.Error("Failed to update cast_availability", "error", err, "availability_id", id)
			http.Error(w, "Failed to update cast_availability", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Cast availability updated", "availability_id", id, "is_available", c.IsAvailable)
		availID, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Invalid availability ID format", "error", err, "availability_id", id)
			http.Error(w, "Invalid availability ID format", http.StatusInternalServerError)
			return
		}
		c.ID = availID
		response.RespondJSON(w, http.StatusOK, c)
	}
}

// [DELETE] /api/cast_availabilities/{id} : 役者スケジュール可用性の削除
func HandleDeleteCastAvailability(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		_, err := db.ExecContext(r.Context(), "DELETE FROM cast_availabilities WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete cast_availability", "error", err, "availability_id", id)
			http.Error(w, "Failed to delete cast_availability", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Cast availability deleted", "availability_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// [GET] /api/casts/{id}/cast_availabilities : 役者の可用性一覧
func HandleListCastAvailabilitiesByCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		castID := r.PathValue("id")

		rows, err := db.QueryContext(r.Context(), "SELECT id, candidate_date_id, time_slot_id, cast_id, is_available FROM cast_availabilities WHERE cast_id = ? ORDER BY candidate_date_id, time_slot_id", castID)
		if err != nil {
			slog.Error("Failed to fetch cast_availabilities", "error", err, "cast_id", castID)
			http.Error(w, "Failed to fetch cast_availabilities", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]models.CastAvailability, 0)
		for rows.Next() {
			var c models.CastAvailability
			if err := rows.Scan(&c.ID, &c.CandidateDateID, &c.TimeSlotID, &c.CastID, &c.IsAvailable); err != nil {
				http.Error(w, "Failed to read cast_availabilities", http.StatusInternalServerError)
				return
			}
			items = append(items, c)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read cast_availabilities", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}
