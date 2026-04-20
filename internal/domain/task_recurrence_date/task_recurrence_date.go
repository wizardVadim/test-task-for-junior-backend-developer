package taskrecurrencedate

import "time"

type TaskRecurrenceDate struct {
	ID           int64     `json:"id"`
	RecurrenceID int64     `json:"recurrence_id"`
	RunDate      time.Time `json:"run_date"`
}
