package http

import (
	"context"
	"encoding/json"
	"net/http"
)

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	analytics, err := h.analytics.Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
