package http

import (
	"encoding/json"
	"net/http"
)

// HandleGet godoc
// @Summary		Получить аналитику
// @Description	Возвращает агрегированные метрики по выполненным HTTP-запросам: количество вызовов, успешность, статусы, длительность и попытки.
// @Tags		Аналитика
// @Produce		json
// @Success		200 {object} AnalyticsResponse
// @Failure		500 {string} string
// @Router		/analytics [get]
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
