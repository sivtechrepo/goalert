package alertmetrics

import (
	"time"
)

// A Record is a daily aggregate of alert metrics for a given service.
type Record struct {
	ServiceID      string
	Date           time.Time
	AlertCount     int
	AvgTimeToAck   time.Duration
	AvgTimeToClose time.Duration
	EscalatedCount int
}
