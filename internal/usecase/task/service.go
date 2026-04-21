package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskrecurrencedomain "example.com/taskservice/internal/domain/task_recurrence"
	taskrecurrencedatedomain "example.com/taskservice/internal/domain/task_recurrence_date"
)

type Service struct {
	taskRepo                TaskRepository
	taskRecurrenceRepo      TaskRecurrenceRepository
	taskRecurrenceDatesRepo TaskRecurrenceDateRepository
	now                     func() time.Time
}

func NewService(
	taskRepo TaskRepository,
	taskRecurrenceRepo TaskRecurrenceRepository,
	taskRecurrenceDatesRepo TaskRecurrenceDateRepository,
) *Service {
	return &Service{
		taskRepo:                taskRepo,
		taskRecurrenceRepo:      taskRecurrenceRepo,
		taskRecurrenceDatesRepo: taskRecurrenceDatesRepo,
		now:                     func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*TaskDetails, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()

	taskModel := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	createdTask, err := s.taskRepo.Create(ctx, taskModel)
	if err != nil {
		return nil, err
	}

	var createdRecurrence *taskrecurrencedomain.TaskRecurrence
	var createdDates []taskrecurrencedatedomain.TaskRecurrenceDate

	if normalized.Recurrence != nil {
		recurrenceModel := &taskrecurrencedomain.TaskRecurrence{
			TaskID:     createdTask.ID,
			Type:       normalized.Recurrence.Type,
			StartDate:  normalized.Recurrence.StartDate,
			EveryNDays: normalized.Recurrence.EveryNDays,
			DayOfMonth: normalized.Recurrence.DayOfMonth,
			IsActive:   normalized.Recurrence.IsActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		createdRecurrence, err = s.taskRecurrenceRepo.Create(ctx, recurrenceModel)
		if err != nil {
			return nil, err
		}

		if len(normalized.Recurrence.Dates) > 0 {
			dates := make([]taskrecurrencedatedomain.TaskRecurrenceDate, 0, len(normalized.Recurrence.Dates))
			for _, d := range normalized.Recurrence.Dates {
				dates = append(dates, taskrecurrencedatedomain.TaskRecurrenceDate{
					RecurrenceID: createdRecurrence.ID,
					RunDate:      d.RunDate,
				})
			}

			err = s.taskRecurrenceDatesRepo.CreateMany(ctx, dates)
			if err != nil {
				return nil, err
			}
			createdDates = dates
		}
	}

	return &TaskDetails{
		Task:            createdTask,
		Recurrence:      createdRecurrence,
		RecurrenceDates: createdDates,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*TaskDetails, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	taskModel, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	recurrence, err := s.taskRecurrenceRepo.GetByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	var dates []taskrecurrencedatedomain.TaskRecurrenceDate
	if recurrence != nil {
		dates, err = s.taskRecurrenceDatesRepo.GetByRecurrenceID(ctx, recurrence.ID)
		if err != nil {
			return nil, err
		}
	}

	return &TaskDetails{
		Task:            taskModel,
		Recurrence:      recurrence,
		RecurrenceDates: dates,
	}, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*TaskDetails, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	taskModel := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updatedTask, err := s.taskRepo.Update(ctx, taskModel)
	if err != nil {
		return nil, err
	}

	var updatedRecurrence *taskrecurrencedomain.TaskRecurrence
	var updatedDates []taskrecurrencedatedomain.TaskRecurrenceDate

	// простой вариант для тестового:
	// удаляем старое правило и создаём новое заново
	if err := s.taskRecurrenceRepo.DeleteByTaskID(ctx, id); err != nil {
		return nil, err
	}

	if normalized.Recurrence != nil {
		now := s.now()

		recurrenceModel := &taskrecurrencedomain.TaskRecurrence{
			TaskID:     updatedTask.ID,
			Type:       normalized.Recurrence.Type,
			StartDate:  normalized.Recurrence.StartDate,
			EveryNDays: normalized.Recurrence.EveryNDays,
			DayOfMonth: normalized.Recurrence.DayOfMonth,
			IsActive:   normalized.Recurrence.IsActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		updatedRecurrence, err = s.taskRecurrenceRepo.Create(ctx, recurrenceModel)
		if err != nil {
			return nil, err
		}

		if len(normalized.Recurrence.Dates) > 0 {
			dates := make([]taskrecurrencedatedomain.TaskRecurrenceDate, 0, len(normalized.Recurrence.Dates))
			for _, d := range normalized.Recurrence.Dates {
				dates = append(dates, taskrecurrencedatedomain.TaskRecurrenceDate{
					RecurrenceID: updatedRecurrence.ID,
					RunDate:      d.RunDate,
				})
			}

			err = s.taskRecurrenceDatesRepo.CreateMany(ctx, dates)
			if err != nil {
				return nil, err
			}
			updatedDates = dates
		}
	}

	return &TaskDetails{
		Task:            updatedTask,
		Recurrence:      updatedRecurrence,
		RecurrenceDates: updatedDates,
	}, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.taskRepo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.taskRepo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateRecurrenceInput(input.Recurrence); err != nil {
		return CreateInput{}, err
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validateRecurrenceInput(input.Recurrence); err != nil {
		return UpdateInput{}, err
	}

	return input, nil
}

func validateRecurrenceInput(input *RecurrenceInput) error {
	if input == nil {
		return nil
	}

	if !input.Type.Valid() {
		return fmt.Errorf("%w: invalid recurrence type", ErrInvalidInput)
	}

	if input.StartDate.IsZero() {
		return fmt.Errorf("%w: recurrence start_date is required", ErrInvalidInput)
	}

	switch input.Type {
	case taskrecurrencedomain.TypeDaily:
		if input.EveryNDays != nil && *input.EveryNDays <= 0 {
			return fmt.Errorf("%w: every_n_days must be positive", ErrInvalidInput)
		}
		if input.DayOfMonth != nil {
			return fmt.Errorf("%w: day_of_month is not allowed for daily recurrence", ErrInvalidInput)
		}
		if len(input.Dates) > 0 {
			return fmt.Errorf("%w: dates are not allowed for daily recurrence", ErrInvalidInput)
		}

	case taskrecurrencedomain.TypeMonthly:
		if input.DayOfMonth == nil || *input.DayOfMonth < 1 || *input.DayOfMonth > 31 {
			return fmt.Errorf("%w: invalid day_of_month", ErrInvalidInput)
		}
		if input.EveryNDays != nil {
			return fmt.Errorf("%w: every_n_days is not allowed for monthly recurrence", ErrInvalidInput)
		}
		if len(input.Dates) > 0 {
			return fmt.Errorf("%w: dates are not allowed for monthly recurrence", ErrInvalidInput)
		}

	case taskrecurrencedomain.TypeSpecificDates:
		if len(input.Dates) == 0 {
			return fmt.Errorf("%w: dates are required for specific_dates", ErrInvalidInput)
		}
		if input.EveryNDays != nil {
			return fmt.Errorf("%w: every_n_days is not allowed for specific_dates", ErrInvalidInput)
		}
		if input.DayOfMonth != nil {
			return fmt.Errorf("%w: day_of_month is not allowed for specific_dates", ErrInvalidInput)
		}

	case taskrecurrencedomain.TypeEvenDays, taskrecurrencedomain.TypeOddDays:
		if input.EveryNDays != nil {
			return fmt.Errorf("%w: every_n_days is not allowed for this recurrence type", ErrInvalidInput)
		}
		if input.DayOfMonth != nil {
			return fmt.Errorf("%w: day_of_month is not allowed for this recurrence type", ErrInvalidInput)
		}
		if len(input.Dates) > 0 {
			return fmt.Errorf("%w: dates are not allowed for this recurrence type", ErrInvalidInput)
		}
	}

	return nil
}
