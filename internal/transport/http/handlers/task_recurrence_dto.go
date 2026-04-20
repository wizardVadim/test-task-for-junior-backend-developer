package handlers

import (
	"time"

	taskrecurrencedomain "example.com/taskservice/internal/domain/task_recurrence"
	taskrecurrencedatedomain "example.com/taskservice/internal/domain/task_recurrence_date"
)

type taskRecurrenceDTO struct {
	ID         int64                     `json:"id"`
	TaskID     int64                     `json:"task_id"`
	Type       taskrecurrencedomain.Type `json:"type"`
	StartDate  time.Time                 `json:"start_date"`
	EveryNDays *int                      `json:"every_n_days"`
	DayOfMonth *int                      `json:"day_of_month"`
	IsActive   bool                      `json:"is_active"`
	CreatedAt  time.Time                 `json:"created_at"`
	UpdatedAt  time.Time                 `json:"updated_at"`
	Dates      []taskRecurrenceDateDTO   `json:"dates,omitempty"`
}

func newTaskRecurrenceDTO(
	recurrence *taskrecurrencedomain.TaskRecurrence,
	dates []taskrecurrencedatedomain.TaskRecurrenceDate,
) taskRecurrenceDTO {
	dto := taskRecurrenceDTO{
		ID:         recurrence.ID,
		TaskID:     recurrence.TaskID,
		Type:       recurrence.Type,
		StartDate:  recurrence.StartDate,
		EveryNDays: recurrence.EveryNDays,
		DayOfMonth: recurrence.DayOfMonth,
		IsActive:   recurrence.IsActive,
		CreatedAt:  recurrence.CreatedAt,
		UpdatedAt:  recurrence.UpdatedAt,
	}

	if len(dates) > 0 {
		dto.Dates = make([]taskRecurrenceDateDTO, 0, len(dates))
		for _, d := range dates {
			dto.Dates = append(dto.Dates, taskRecurrenceDateDTO{
				ID:           d.ID,
				RecurrenceID: d.RecurrenceID,
				RunDate:      d.RunDate,
			})
		}
	}

	return dto
}
