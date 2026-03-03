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

// [POST] /api/casts : キャストの作成
func HandleCreateCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c models.Cast
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateCast(c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query := "INSERT INTO casts (project_id, name, role_name) VALUES (?, ?, ?) RETURNING id, created_at"
		err := db.QueryRowContext(r.Context(), query, c.ProjectID, c.Name, c.RoleName).Scan(&c.ID, &c.CreatedAt)
		if err != nil {
			slog.Error("Failed to insert cast", "error", err, "project_id", c.ProjectID)
			http.Error(w, "Failed to create cast", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Cast created", "cast_id", c.ID, "name", c.Name, "role_name", c.RoleName, "project_id", c.ProjectID)
		response.RespondJSON(w, http.StatusCreated, c)
	}
}

// [GET] /api/casts/{id} : キャスト1件取得
func HandleGetCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var c models.Cast
		err := db.QueryRowContext(r.Context(), "SELECT id, project_id, name, role_name, created_at FROM casts WHERE id = ?", id).
			Scan(&c.ID, &c.ProjectID, &c.Name, &c.RoleName, &c.CreatedAt)

		if err == sql.ErrNoRows {
			http.Error(w, "Cast not found", http.StatusNotFound)
			return
		} else if err != nil {
			slog.Error("Failed to get cast", "error", err, "cast_id", id)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, c)
	}
}

// [PUT] /api/casts/{id} : キャストの更新
func HandleUpdateCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var c models.Cast
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// 入力バリデーション
		if err := validators.ValidateCast(c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.ExecContext(r.Context(), "UPDATE casts SET name = ?, role_name = ? WHERE id = ?", c.Name, c.RoleName, id)
		if err != nil {
			slog.Error("Failed to update cast", "error", err, "cast_id", id)
			http.Error(w, "Failed to update cast", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Cast updated", "cast_id", id, "name", c.Name, "role_name", c.RoleName)
		castID, err := strconv.Atoi(id)
		if err != nil {
			slog.Error("Invalid cast ID format", "error", err, "cast_id", id)
			http.Error(w, "Invalid cast ID format", http.StatusInternalServerError)
			return
		}
		c.ID = castID
		response.RespondJSON(w, http.StatusOK, c)
	}
}

// [DELETE] /api/casts/{id} : キャストの削除
func HandleDeleteCast(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		_, err := db.ExecContext(r.Context(), "DELETE FROM casts WHERE id = ?", id)
		if err != nil {
			slog.Error("Failed to delete cast", "error", err, "cast_id", id)
			http.Error(w, "Failed to delete cast", http.StatusInternalServerError)
			return
		}

		// 監査ログ
		slog.Info("Cast deleted", "cast_id", id)
		w.WriteHeader(http.StatusNoContent)
	}
}

// [GET] /api/projects/{id}/casts : プロジェクトのキャスト一覧
func HandleListCastsByProject(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")

		rows, err := db.QueryContext(r.Context(), "SELECT id, project_id, name, role_name, created_at FROM casts WHERE project_id = ? ORDER BY created_at, id", projectID)
		if err != nil {
			slog.Error("Failed to fetch casts", "error", err, "project_id", projectID)
			http.Error(w, "Failed to fetch casts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := make([]models.Cast, 0)
		for rows.Next() {
			var c models.Cast
			if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.RoleName, &c.CreatedAt); err != nil {
				http.Error(w, "Failed to read casts", http.StatusInternalServerError)
				return
			}
			items = append(items, c)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Failed to read casts", http.StatusInternalServerError)
			return
		}

		response.RespondJSON(w, http.StatusOK, items)
	}
}
