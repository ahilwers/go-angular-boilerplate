package http

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, ErrorResponse{Error: message}, status)
}

// parsePaginationParams extracts page and limit from query parameters
// Returns (page, limit) or (0, 0) if not provided or invalid
func parsePaginationParams(r *http.Request) (int, int) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	if pageStr == "" || limitStr == "" {
		return 0, 0
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, 0
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 0, 0
	}

	return page, limit
}
