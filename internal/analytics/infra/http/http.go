package http

import "auto-http-fetcher/internal/analytics/service"

type Handler struct {
	analytics *service.AnalyticsService
}

func NewHandler(analytics *service.AnalyticsService) *Handler {
	return &Handler{
		analytics: analytics,
	}
}
