package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskrecurrencedomain "example.com/taskservice/internal/domain/task_recurrence"
	taskrecurrencedatedomain "example.com/taskservice/internal/domain/task_recurrence_date"
)

type taskMutationDTO struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      taskdomain.Status  `json:"status"`
	Recurrence  *taskRecurrenceDTO `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64              `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      taskdomain.Status  `json:"status"`
	Recurrence  *taskRecurrenceDTO `json:"recurrence,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func newTaskDTO(
	task *taskdomain.Task,
	recurrence *taskrecurrencedomain.TaskRecurrence,
	recurrenceDate []taskrecurrencedatedomain.TaskRecurrenceDate,
) taskDTO {
	var recurrenceDTO *taskRecurrenceDTO

	if recurrence != nil {
		r := newTaskRecurrenceDTO(recurrence, recurrenceDate)
		recurrenceDTO = &r
	}

	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Recurrence:  recurrenceDTO,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}
