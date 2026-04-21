package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskoccurrencedomain "example.com/taskservice/internal/domain/task_occurrence"
)

type taskOccurrenceDTO struct {
	ID            int64             `json:"id"`
	TaskID        int64             `json:"task_id"`
	RecurrenceID  *int64            `json:"recurrence_id"`
	ScheduledDate time.Time         `json:"scheduled_date"`
	Status        taskdomain.Status `json:"status"`
	CompletedAt   *time.Time        `json:"completed_at"`
	CreatedAt     time.Time         `json:"created_at,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at,omitempty"`
}

func newTaskOccurrenceDTO(taskOccurrence *taskoccurrencedomain.TaskOccurrence) taskOccurrenceDTO {
	return taskOccurrenceDTO{
		ID:            taskOccurrence.ID,
		TaskID:        taskOccurrence.TaskID,
		RecurrenceID:  taskOccurrence.RecurrenceID,
		ScheduledDate: taskOccurrence.ScheduledDate,
		Status:        taskOccurrence.Status,
		CompletedAt:   taskOccurrence.CompletedAt,
		CreatedAt:     taskOccurrence.CreatedAt,
		UpdatedAt:     taskOccurrence.UpdatedAt,
	}
}
