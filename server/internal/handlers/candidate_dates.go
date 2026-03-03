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

// [POST] /api/candidate_dates : 候補日の作成
func HandleCreateCandidateDate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cd models.CandidateDate
		if err := json.NewDecoder(r.Body).Decode(&cd); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateCandidateDate(cd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query := "INSERT INTO candidate_dates (project_id, target_date) VALUES (?, ?) RETURNING id"
		err := db.QueryRowContext(r.Context(), query, cd.ProjectID, cd.TargetDate).Scan(&cd.ID)
		if err != nil {
			slog.Error("Failed to insert candidate_date", "error", err, "project_id", cd.ProjectID)
			http.Error(w, "Failed to create candidate_date", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Candidate date created", "candidate_date_id", cd.ID, "target_date", cd.TargetDate, "project_id", cd.ProjectID)
		response.RespondJSON(w, http.StatusCreated, cd)
	}
}

// [GET] /api/candidate_dates/{id} : 候補日1件取得
func HandleGetCandidateDate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var cd models.CandidateDate
		err := db.QueryRowContext(r.Context(), "SELECT id, project_id, target_date FROM candidate_dates WHERE id = ?", id).
			Scan(&cd.ID, &cd.ProjectID, &cd.TargetDate)

		if err == sql.ErrNoRows {
			http.Error(w, "Candidate date not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get candidate_date", "error", err, "candidate_date_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, cd)
	}
}

// [PUT] /api/candidate_dates/{id} : 候補日の更新
func HandleUpdateCandidateDate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var cd models.CandidateDate
		if err := json.NewDecoder(r.Body).Decode(&cd); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateCandidateDate(cd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.ExecContext(r.Context(), "UPDATE candidate_dates SET target_date = ? WHERE id = ?", cd.TargetDate, id)
		if err != nil {
			slog.Error("Failed to update candidate_date", "error", err, "candidate_date_id", id)
			http.Error(w, "Failed to update candidate_date", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Candidate date updated", "candidate_date_id", id, "target_date", cd.TargetDate)
		cdID, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Invalid candidate_date ID format", "error", err, "candidate_date_id", id)
			http.Error(w, "Invalid candidate_date ID format", http.StatusInternalServerError)
			return
		}
		cd.ID = cdID
		response.RespondJSON(w, http.StatusOK, cd)
	}
}

// [DELETE] /api/candidate_dates/{id} : 候補日の削除
func HandleDeleteCandidateDate(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		_, err := db.ExecContext(r.Context(), "DELETE FROM candidate_dates WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete candidate_date", "error", err, "candidate_date_id", id)
			http.Error(w, "Failed to delete candidate_date", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Candidate date deleted", "candidate_date_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// [GET] /api/projects/{id}/candidate_dates : プロジェクトの候補日一覧
func HandleListCandidateDatesByProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		rows, err := db.QueryContext(r.Context(), "SELECT id, project_id, target_date FROM candidate_dates WHERE project_id = ? ORDER BY target_date", projectID)
		if err != nil {
			slog.Error("Failed to fetch candidate_dates", "error", err, "project_id", projectID)
			http.Error(w, "Failed to fetch candidate_dates", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]models.CandidateDate, 0)
		for rows.Next() {
			var cd models.CandidateDate
			if err := rows.Scan(&cd.ID, &cd.ProjectID, &cd.TargetDate); err != nil {
				http.Error(w, "Failed to read candidate_dates", http.StatusInternalServerError)
				return
			}
			items = append(items, cd)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read candidate_dates", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}
