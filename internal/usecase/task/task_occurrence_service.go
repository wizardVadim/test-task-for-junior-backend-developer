package task

import (
	"context"
	"fmt"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskoccurrencedomain "example.com/taskservice/internal/domain/task_occurrence"
)

type TaskOccurrenceService struct {
	repo     TaskOccurrenceRepository
	taskRepo TaskRepository
	now      func() time.Time
}

func NewTaskOccurrenceService(
	repo TaskOccurrenceRepository,
	taskRepo TaskRepository,
) *TaskOccurrenceService {
	return &TaskOccurrenceService{
		repo:     repo,
		taskRepo: taskRepo,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *TaskOccurrenceService) GetByDate(
	ctx context.Context,
	date time.Time,
) ([]taskoccurrencedomain.TaskOccurrence, error) {
	return s.repo.GetByDate(ctx, date)
}

func (s *TaskOccurrenceService) GetByTaskID(
	ctx context.Context,
	taskID int64,
) ([]taskoccurrencedomain.TaskOccurrence, error) {
	if taskID <= 0 {
		return nil, fmt.Errorf("%w: task id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByTaskID(ctx, taskID)
}

func (s *TaskOccurrenceService) UpdateStatus(
	ctx context.Context,
	id int64,
	status taskdomain.Status,
) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	if !status.Valid() {
		return fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	occ, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if occ == nil {
		return taskoccurrencedomain.ErrNotFound
	}

	occ.Status = status
	occ.UpdatedAt = s.now()

	if status == taskdomain.StatusDone {
		now := s.now()
		occ.CompletedAt = &now
	} else {
		occ.CompletedAt = nil
	}

	if err := s.repo.Update(ctx, occ); err != nil {
		return err
	}

	return s.refreshTaskStatus(ctx, occ.TaskID)
}

func (s *TaskOccurrenceService) refreshTaskStatus(ctx context.Context, taskID int64) error {
	taskModel, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if taskModel == nil {
		return taskdomain.ErrNotFound
	}

	occurrences, err := s.repo.GetByTaskID(ctx, taskID)
	if err != nil {
		return err
	}

	taskModel.Status = aggregateTaskStatus(occurrences)
	taskModel.UpdatedAt = s.now()

	_, err = s.taskRepo.Update(ctx, taskModel)
	return err
}

func aggregateTaskStatus(occurrences []taskoccurrencedomain.TaskOccurrence) taskdomain.Status {
	if len(occurrences) == 0 {
		return taskdomain.StatusNew
	}

	allDone := true

	for i := range occurrences {
		switch occurrences[i].Status {
		case taskdomain.StatusInProgress:
			return taskdomain.StatusInProgress
		case taskdomain.StatusDone:
			// ok
		default:
			allDone = false
		}
	}

	if allDone {
		return taskdomain.StatusDone
	}

	return taskdomain.StatusNew
}