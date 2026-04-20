package taskoccurrence

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type TaskOccurrence struct {
	ID            int64
	TaskID        int64
	RecurrenceID  *int64
	ScheduledDate time.Time
	Status        taskdomain.Status
	CompletedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}