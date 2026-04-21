package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskoccurrencedomain "example.com/taskservice/internal/domain/task_occurrence"
	taskrecurrencedomain "example.com/taskservice/internal/domain/task_recurrence"
	taskrecurrencedatedomain "example.com/taskservice/internal/domain/task_recurrence_date"
)

type GeneratorService struct {
	taskRepo                TaskRepository
	taskRecurrenceRepo      TaskRecurrenceRepository
	taskRecurrenceDatesRepo TaskRecurrenceDateRepository
	taskOccurrenceRepo      TaskOccurrenceRepository
	now                     func() time.Time
}

func NewGeneratorService(
	taskRepo TaskRepository,
	taskRecurrenceRepo TaskRecurrenceRepository,
	taskRecurrenceDatesRepo TaskRecurrenceDateRepository,
	taskOccurrenceRepo TaskOccurrenceRepository,
) *GeneratorService {
	return &GeneratorService{
		taskRepo:                taskRepo,
		taskRecurrenceRepo:      taskRecurrenceRepo,
		taskRecurrenceDatesRepo: taskRecurrenceDatesRepo,
		taskOccurrenceRepo:      taskOccurrenceRepo,
		now:                     func() time.Time { return time.Now().UTC() },
	}
}

func (s *GeneratorService) GenerateRecent(ctx context.Context) error {
	now := s.now()
	since := now.Add(-10 * time.Minute)

	recurrences, err := s.taskRecurrenceRepo.ListUpdatedSince(ctx, since)
	if err != nil {
		return err
	}

	from := startOfDayUTC(now)
	to := from.AddDate(0, 0, 2)

	if err := s.generateRecurringForRange(ctx, recurrences, from, to); err != nil {
		return err
	}

	tasksWithoutRecurrence, err := s.taskRepo.ListWithoutRecurrenceUpdatedSince(ctx, since)
	if err != nil {
		return err
	}

	if err := s.generateSingleOccurrences(ctx, tasksWithoutRecurrence, from); err != nil {
		return err
	}

	return nil
}

func (s *GeneratorService) GenerateDaily(ctx context.Context) error {
	now := s.now()

	recurrences, err := s.taskRecurrenceRepo.ListActive(ctx)
	if err != nil {
		return err
	}

	from := startOfDayUTC(now)
	to := from.AddDate(0, 0, 7)

	if err := s.generateRecurringForRange(ctx, recurrences, from, to); err != nil {
		return err
	}

	tasksWithoutRecurrence, err := s.taskRepo.ListWithoutRecurrence(ctx)
	if err != nil {
		return err
	}

	if err := s.generateSingleOccurrences(ctx, tasksWithoutRecurrence, from); err != nil {
		return err
	}

	return nil
}

func (s *GeneratorService) generateRecurringForRange(
	ctx context.Context,
	recurrences []taskrecurrencedomain.TaskRecurrence,
	from, to time.Time,
) error {
	from = startOfDayUTC(from)
	to = startOfDayUTC(to)

	for i := range recurrences {
		recurrence := recurrences[i]

		taskModel, err := s.taskRepo.GetByID(ctx, recurrence.TaskID)
		if err != nil {
			return err
		}
		if taskModel == nil {
			continue
		}

		var dates []taskrecurrencedatedomain.TaskRecurrenceDate
		if recurrence.Type == taskrecurrencedomain.TypeSpecificDates {
			dates, err = s.taskRecurrenceDatesRepo.GetByRecurrenceID(ctx, recurrence.ID)
			if err != nil {
				return err
			}
		}

		for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
			if !matchesRecurrence(day, &recurrence, dates) {
				continue
			}

			recurrenceID := recurrence.ID
			now := s.now()

			occ := &taskoccurrencedomain.TaskOccurrence{
				TaskID:        taskModel.ID,
				RecurrenceID:  &recurrenceID,
				ScheduledDate: startOfDayUTC(day),
				Status:        taskdomain.StatusNew,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := s.taskOccurrenceRepo.Create(ctx, occ); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *GeneratorService) generateSingleOccurrences(
	ctx context.Context,
	tasks []taskdomain.Task,
	day time.Time,
) error {
	day = startOfDayUTC(day)

	for i := range tasks {
		now := s.now()

		occ := &taskoccurrencedomain.TaskOccurrence{
			TaskID:        tasks[i].ID,
			RecurrenceID:  nil,
			ScheduledDate: day,
			Status:        taskdomain.StatusNew,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.taskOccurrenceRepo.Create(ctx, occ); err != nil {
			return err
		}
	}

	return nil
}

func matchesRecurrence(
	date time.Time,
	recurrence *taskrecurrencedomain.TaskRecurrence,
	dates []taskrecurrencedatedomain.TaskRecurrenceDate,
) bool {
	if recurrence == nil || !recurrence.IsActive {
		return false
	}

	date = startOfDayUTC(date)
	start := startOfDayUTC(recurrence.StartDate)

	if date.Before(start) {
		return false
	}

	switch recurrence.Type {
	case taskrecurrencedomain.TypeDaily:
		if recurrence.EveryNDays == nil || *recurrence.EveryNDays <= 1 {
			return true
		}
		diffDays := daysBetween(start, date)
		return diffDays%*recurrence.EveryNDays == 0

	case taskrecurrencedomain.TypeMonthly:
		if recurrence.DayOfMonth == nil {
			return false
		}
		return date.Day() == *recurrence.DayOfMonth

	case taskrecurrencedomain.TypeSpecificDates:
		for _, d := range dates {
			if sameDayUTC(d.RunDate, date) {
				return true
			}
		}
		return false

	case taskrecurrencedomain.TypeEvenDays:
		return date.Day()%2 == 0

	case taskrecurrencedomain.TypeOddDays:
		return date.Day()%2 != 0
	}

	return false
}

func startOfDayUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func sameDayUTC(a, b time.Time) bool {
	return startOfDayUTC(a).Equal(startOfDayUTC(b))
}

func daysBetween(from, to time.Time) int {
	from = startOfDayUTC(from)
	to = startOfDayUTC(to)
	return int(to.Sub(from).Hours() / 24)
}