package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskoccurrencedomain "example.com/taskservice/internal/domain/task_occurrence"
	taskrecurrencedomain "example.com/taskservice/internal/domain/task_recurrence"
	taskrecurrencedatedomain "example.com/taskservice/internal/domain/task_recurrence_date"
)

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type TaskRecurrenceRepository interface {
	Create(ctx context.Context, recurrence *taskrecurrencedomain.TaskRecurrence) (*taskrecurrencedomain.TaskRecurrence, error)
	GetByTaskID(ctx context.Context, taskID int64) (*taskrecurrencedomain.TaskRecurrence, error)
	Update(ctx context.Context, recurrence *taskrecurrencedomain.TaskRecurrence) (*taskrecurrencedomain.TaskRecurrence, error)
	DeleteByTaskID(ctx context.Context, taskID int64) error
}

type TaskRecurrenceDateRepository interface {
	CreateMany(ctx context.Context, dates []taskrecurrencedatedomain.TaskRecurrenceDate) error
	GetByRecurrenceID(ctx context.Context, recurrenceID int64) ([]taskrecurrencedatedomain.TaskRecurrenceDate, error)
	DeleteByRecurrenceID(ctx context.Context, recurrenceID int64) error
}

type TaskOccurrenceRepository interface {
	Create(ctx context.Context, occ *taskoccurrencedomain.TaskOccurrence) error
	GetByDate(ctx context.Context, date time.Time) ([]taskoccurrencedomain.TaskOccurrence, error)
	Update(ctx context.Context, occ *taskoccurrencedomain.TaskOccurrence) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*TaskDetails, error)
	GetByID(ctx context.Context, id int64) (*TaskDetails, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*TaskDetails, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type TaskDetails struct {
	Task            *taskdomain.Task
	Recurrence      *taskrecurrencedomain.TaskRecurrence
	RecurrenceDates []taskrecurrencedatedomain.TaskRecurrenceDate
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Recurrence  *RecurrenceInput
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	Recurrence  *RecurrenceInput
}

type RecurrenceInput struct {
	Type       taskrecurrencedomain.Type
	StartDate  time.Time
	EveryNDays *int
	DayOfMonth *int
	IsActive   bool
	Dates      []RecurrenceDateInput
}

type RecurrenceDateInput struct {
	RunDate time.Time
}
