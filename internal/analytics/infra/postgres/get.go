package infra

import (
	"auto-http-fetcher/internal/analytics/domain"
	"context"
)

func (p *PGAnalyticsRepo) Get(ctx context.Context) (*domain.Analytics, error) {
	var analytics domain.Analytics
	query := `SELECT 
		COUNT(*) as total_calls,
		COUNT(*) FILTER (WHERE status = 'success') as success_calls,
		COUNT(*) FILTER (WHERE status = 'failed') as failed_calls,
		AVG(duration) as avg_duration,
		MIN(duration) as min_duration,
		MAX(duration) as max_duration,
		AVG(attempt) as avg_attempts
	FROM responses`
	statusStatsQuery := `SELECT status_code, COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () as percentage FROM responses GROUP BY status_code`

	err := p.pool.QueryRow(ctx, query).Scan(&analytics.TotalCalls, &analytics.SuccessCalls, &analytics.FailedCalls, &analytics.AvgDuration, &analytics.MinDuration, &analytics.MaxDuration, &analytics.AvgAttempts)
	if err != nil {
		return nil, err
	}

	statusStatsRows, err := p.pool.Query(ctx, statusStatsQuery)
	if err != nil {
		return nil, err
	}

	analytics.StatusStats = make(map[int]float64)

	defer statusStatsRows.Close()

	for statusStatsRows.Next() {
		var statusCode int
		var percentage float64

		if err = statusStatsRows.Scan(&statusCode, &percentage); err != nil {
			return nil, err
		}

		analytics.StatusStats[statusCode] = percentage
	}

	return &analytics, statusStatsRows.Err()
}
