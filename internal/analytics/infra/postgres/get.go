package infra

import (
	"auto-http-fetcher/internal/analytics/domain"
	"context"
)

func (p *PGAnalyticsRepo) Get(ctx context.Context) (*domain.Analytics, error) {
	var analytics domain.Analytics
	var avgDuration float64
	var minDuration float64
	var maxDuration float64

	query := `SELECT 
		COUNT(*) as total_calls,
		COUNT(*) FILTER (WHERE status = 'success') as success_calls,
		COUNT(*) FILTER (WHERE status = 'failed') as failed_calls,
		COALESCE(AVG(duration), 0) as avg_duration,
		COALESCE(MIN(duration), 0) as min_duration,
		COALESCE(MAX(duration), 0) as max_duration,
		COALESCE(AVG(attempt), 0) as avg_attempts
	FROM responses`

	statusStatsQuery := `SELECT status_code, COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () as percentage FROM responses GROUP BY status_code`

	err := p.pool.QueryRow(ctx, query).Scan(
		&analytics.TotalCalls,
		&analytics.SuccessCalls,
		&analytics.FailedCalls,
		&avgDuration,
		&minDuration,
		&maxDuration,
		&analytics.AvgAttempts,
	)
	if err != nil {
		return nil, err
	}

	analytics.AvgDuration = int64(avgDuration)
	analytics.MinDuration = int64(minDuration)
	analytics.MaxDuration = int64(maxDuration)

	statusStatsRows, err := p.pool.Query(ctx, statusStatsQuery)
	if err != nil {
		return nil, err
	}
	defer statusStatsRows.Close()

	analytics.StatusStats = make(map[int]float64)

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
