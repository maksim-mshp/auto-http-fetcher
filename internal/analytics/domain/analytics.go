package domain

type Analytics struct {
	TotalCalls   int
	SuccessCalls int
	FailedCalls  int
	StatusStats  map[int]float64
	AvgDuration  int64
	MinDuration  int64
	MaxDuration  int64
	AvgAttempts  float64
}
