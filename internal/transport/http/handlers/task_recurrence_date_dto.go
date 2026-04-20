package handlers

import (
	"time"

	taskrecurrencedatedomain "example.com/taskservice/internal/domain/task_recurrence_date"
)

type taskRecurrenceDateDTO struct {
	ID           int64     `json:"id"`
	RecurrenceID int64     `json:"recurrence_id"`
	RunDate      time.Time `json:"run_date"`
}

func newTaskRecurrenceDateDTO(taskRecurrenceDate *taskrecurrencedatedomain.TaskRecurrenceDate) taskRecurrenceDateDTO {
	return taskRecurrenceDateDTO{
		ID:           taskRecurrenceDate.ID,
		RecurrenceID: taskRecurrenceDate.RecurrenceID,
		RunDate:      taskRecurrenceDate.RunDate,
	}
}
