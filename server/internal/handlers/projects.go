package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/creatio313/movie_scheduler/internal/models"
	"github.com/creatio313/movie_scheduler/internal/response"
	"github.com/creatio313/movie_scheduler/internal/validators"
)

// [POST] /api/projects : プロジェクトの作成
func HandleCreateProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p models.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateProject(p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// MariaDBの RETURNING 句を使って、生成されたUUIDとタイムスタンプを取得する
		query := "INSERT INTO projects (title, description) VALUES (?, ?) RETURNING id"
		err := db.QueryRowContext(r.Context(), query, p.Title, p.Description).Scan(&p.ID)
		if err != nil {
			slog.Error("Failed to insert project", "error", err, "title", p.Title)
			http.Error(w, "Failed to create project", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Project created", "project_id", p.ID, "title", p.Title)
		response.RespondJSON(w, http.StatusCreated, p)
	}
}

// [GET] /api/projects/{id} : プロジェクト1件取得
func HandleGetProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id") // Go 1.22の新機能

		var p models.Project
		var desc sql.NullString
		err := db.QueryRowContext(r.Context(), "SELECT id, title, description FROM projects WHERE id = ?", id).
			Scan(&p.ID, &p.Title, &desc)

		if err == sql.ErrNoRows {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get project", "error", err, "project_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		p.Description = desc.String
		response.RespondJSON(w, http.StatusOK, p)
	}
}

// [PUT] /api/projects/{id} : プロジェクトの更新
func HandleUpdateProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var p models.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateProject(p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.ExecContext(r.Context(), "UPDATE projects SET title = ?, description = ? WHERE id = ?", p.Title, p.Description, id)
		if err != nil {
			slog.Error("Failed to update project", "error", err, "project_id", id)
			http.Error(w, "Failed to update project", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Project updated", "project_id", id, "title", p.Title)
		p.ID = id
		response.RespondJSON(w, http.StatusOK, p)
	}
}

// [DELETE] /api/projects/{id} : プロジェクトの削除
func HandleDeleteProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		// CASCADE制約があるため、関連するキャストやシーン情報も自動的に削除されます
		_, err := db.ExecContext(r.Context(), "DELETE FROM projects WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete project", "error", err, "project_id", id)
			http.Error(w, "Failed to delete project", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Project deleted", "project_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}
