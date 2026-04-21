package postgres

import (
	"context"
	"errors"

	taskrecurrencedomain "example.com/taskservice/internal/domain/task_recurrence"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRecurrenceRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRecurrenceRepository(pool *pgxpool.Pool) *TaskRecurrenceRepository {
	return &TaskRecurrenceRepository{pool: pool}
}

func (r *TaskRecurrenceRepository) Create(
	ctx context.Context,
	recurrence *taskrecurrencedomain.TaskRecurrence,
) (*taskrecurrencedomain.TaskRecurrence, error) {
	const query = `
		INSERT INTO task_recurrences (
			task_id,
			type,
			start_date,
			every_n_days,
			day_of_month,
			is_active,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, task_id, type, start_date, every_n_days, day_of_month, is_active, created_at, updated_at
	`

	var created taskrecurrencedomain.TaskRecurrence

	err := r.pool.QueryRow(
		ctx,
		query,
		recurrence.TaskID,
		recurrence.Type,
		recurrence.StartDate,
		recurrence.EveryNDays,
		recurrence.DayOfMonth,
		recurrence.IsActive,
		recurrence.CreatedAt,
		recurrence.UpdatedAt,
	).Scan(
		&created.ID,
		&created.TaskID,
		&created.Type,
		&created.StartDate,
		&created.EveryNDays,
		&created.DayOfMonth,
		&created.IsActive,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *TaskRecurrenceRepository) GetByTaskID(
	ctx context.Context,
	taskID int64,
) (*taskrecurrencedomain.TaskRecurrence, error) {
	const query = `
		SELECT id, task_id, type, start_date, every_n_days, day_of_month, is_active, created_at, updated_at
		FROM task_recurrences
		WHERE task_id = $1
		LIMIT 1
	`

	var recurrence taskrecurrencedomain.TaskRecurrence

	err := r.pool.QueryRow(ctx, query, taskID).Scan(
		&recurrence.ID,
		&recurrence.TaskID,
		&recurrence.Type,
		&recurrence.StartDate,
		&recurrence.EveryNDays,
		&recurrence.DayOfMonth,
		&recurrence.IsActive,
		&recurrence.CreatedAt,
		&recurrence.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &recurrence, nil
}

func (r *TaskRecurrenceRepository) DeleteByTaskID(ctx context.Context, taskID int64) error {
	const query = `DELETE FROM task_recurrences WHERE task_id = $1`
	_, err := r.pool.Exec(ctx, query, taskID)
	return err
}
