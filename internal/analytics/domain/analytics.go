package domain

type Analytics struct {
	TotalCalls   int `example:"42"`
	SuccessCalls int `example:"37"`
	FailedCalls  int `example:"5"`
	StatusStats  map[int]float64
	AvgDuration  int64   `example:"250000000"`
	MinDuration  int64   `example:"100000000"`
	MaxDuration  int64   `example:"900000000"`
	AvgAttempts  float64 `example:"1.2"`
}
