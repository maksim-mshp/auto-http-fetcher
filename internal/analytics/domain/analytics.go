package domain

import "time"

type Analytics struct {
	TotalCalls   int
	SuccessCalls int
	FailedCalls  int
	StatusStats  map[int]float64
	AvgDuration  time.Duration
	MinDuration  time.Duration
	MaxDuration  time.Duration
	AvgAttempts  float64
}
