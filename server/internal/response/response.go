package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// RespondJSON は指定されたステータスコードとペイロードでJSON形式でレスポンスを返します
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			slog.Error("Failed to encode response", "error", err)
		}
	}
}
