package http

import (
	"auto-http-fetcher/internal/analytics/domain"
	"auto-http-fetcher/internal/analytics/service"
)

type Handler struct {
	analytics *service.AnalyticsService
}

type AnalyticsResponse = domain.Analytics

func NewHandler(analytics *service.AnalyticsService) *Handler {
	return &Handler{
		analytics: analytics,
	}
}
