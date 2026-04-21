package postgres

import (
	"context"
	"errors"
	"time"

	taskoccurrencedomain "example.com/taskservice/internal/domain/task_occurrence"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskOccurrenceRepository struct {
	pool *pgxpool.Pool
}

func NewTaskOccurrenceRepository(pool *pgxpool.Pool) *TaskOccurrenceRepository {
	return &TaskOccurrenceRepository{pool: pool}
}

func (r *TaskOccurrenceRepository) Update(ctx context.Context, occ *taskoccurrencedomain.TaskOccurrence) error {
	const query = `
		UPDATE task_occurrences
		SET
			status = $1,
			completed_at = $2,
			updated_at = $3
		WHERE id = $4
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		occ.Status,
		occ.CompletedAt,
		occ.UpdatedAt,
		occ.ID,
	)

	return err
}

func (r *TaskOccurrenceRepository) Create(
	ctx context.Context,
	occ *taskoccurrencedomain.TaskOccurrence,
) error {
	const query = `
		INSERT INTO task_occurrences (
			task_id,
			recurrence_id,
			scheduled_date,
			status,
			completed_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (task_id, scheduled_date) DO NOTHING
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		occ.TaskID,
		occ.RecurrenceID,
		occ.ScheduledDate,
		occ.Status,
		occ.CompletedAt,
		occ.CreatedAt,
		occ.UpdatedAt,
	)

	return err
}

func (r *TaskOccurrenceRepository) GetByDate(
	ctx context.Context,
	date time.Time,
) ([]taskoccurrencedomain.TaskOccurrence, error) {

	const query = `
		SELECT id, task_id, recurrence_id, scheduled_date, status, completed_at, created_at, updated_at
		FROM task_occurrences
		WHERE scheduled_date = $1
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query, toDateOnly(date))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []taskoccurrencedomain.TaskOccurrence

	for rows.Next() {
		var occ taskoccurrencedomain.TaskOccurrence

		err := rows.Scan(
			&occ.ID,
			&occ.TaskID,
			&occ.RecurrenceID,
			&occ.ScheduledDate,
			&occ.Status,
			&occ.CompletedAt,
			&occ.CreatedAt,
			&occ.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, occ)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func toDateOnly(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (r *TaskOccurrenceRepository) GetByTaskID(
	ctx context.Context,
	taskID int64,
) ([]taskoccurrencedomain.TaskOccurrence, error) {
	const query = `
		SELECT id, task_id, recurrence_id, scheduled_date, status, completed_at, created_at, updated_at
		FROM task_occurrences
		WHERE task_id = $1
		ORDER BY scheduled_date, id
	`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []taskoccurrencedomain.TaskOccurrence

	for rows.Next() {
		var occ taskoccurrencedomain.TaskOccurrence

		err := rows.Scan(
			&occ.ID,
			&occ.TaskID,
			&occ.RecurrenceID,
			&occ.ScheduledDate,
			&occ.Status,
			&occ.CompletedAt,
			&occ.CreatedAt,
			&occ.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, occ)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *TaskOccurrenceRepository) GetByID(
	ctx context.Context,
	id int64,
) (*taskoccurrencedomain.TaskOccurrence, error) {
	const query = `
		SELECT id, task_id, recurrence_id, scheduled_date, status, completed_at, created_at, updated_at
		FROM task_occurrences
		WHERE id = $1
	`

	var occ taskoccurrencedomain.TaskOccurrence

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&occ.ID,
		&occ.TaskID,
		&occ.RecurrenceID,
		&occ.ScheduledDate,
		&occ.Status,
		&occ.CompletedAt,
		&occ.CreatedAt,
		&occ.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &occ, nil
}
