package alertmetrics

import (
	"time"

	"github.com/target/goalert/util/timeutil"
)

// An AlertMetricDataPoint for calculating alert metrics
type AlertDataPoint struct {
	ID        int       `json:"id"`
	ServiceID string    `json:"service_id"`
	TimeToAck timeutil.ISODuration `json:"time_to_ack"`
	TimeToClose timeutil.ISODuration `json:"time_to_close"`
	Escalated bool `json:"escalated"`
	Timestamp time.Time `json:"timestamp"`
}