package alertmetrics

import (
	"time"
)

// An AlertMetricDataPoint for calculating alert metrics
type AlertDataPoint struct {
	ID          int           `json:"id"`
	ServiceID   string        `json:"service_id"`
	TimeToAck   time.Duration `json:"time_to_ack"`
	TimeToClose time.Duration `json:"time_to_close"`
	Escalated   bool          `json:"escalated"`
	Timestamp   time.Time     `json:"timestamp"`
}
