package postgres

import (
	"context"

	taskrecurrencedatedomain "example.com/taskservice/internal/domain/task_recurrence_date"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRecurrenceDateRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRecurrenceDateRepository(pool *pgxpool.Pool) *TaskRecurrenceDateRepository {
	return &TaskRecurrenceDateRepository{pool: pool}
}

func (r *TaskRecurrenceDateRepository) CreateMany(
	ctx context.Context,
	dates []taskrecurrencedatedomain.TaskRecurrenceDate,
) error {
	if len(dates) == 0 {
		return nil
	}

	const query = `
		INSERT INTO task_recurrence_dates (
			recurrence_id,
			run_date
		)
		VALUES ($1, $2)
	`

	batch := &pgx.Batch{}

	for _, d := range dates {
		batch.Queue(query, d.RecurrenceID, d.RunDate)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range dates {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TaskRecurrenceDateRepository) GetByRecurrenceID(
	ctx context.Context,
	recurrenceID int64,
) ([]taskrecurrencedatedomain.TaskRecurrenceDate, error) {
	const query = `
		SELECT id, recurrence_id, run_date
		FROM task_recurrence_dates
		WHERE recurrence_id = $1
		ORDER BY run_date
	`

	rows, err := r.pool.Query(ctx, query, recurrenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []taskrecurrencedatedomain.TaskRecurrenceDate

	for rows.Next() {
		var d taskrecurrencedatedomain.TaskRecurrenceDate

		err := rows.Scan(
			&d.ID,
			&d.RecurrenceID,
			&d.RunDate,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, d)
	}

	return result, nil
}

func (r *TaskRecurrenceDateRepository) DeleteByRecurrenceID(
	ctx context.Context,
	recurrenceID int64,
) error {
	const query = `
		DELETE FROM task_recurrence_dates
		WHERE recurrence_id = $1
	`

	_, err := r.pool.Exec(ctx, query, recurrenceID)
	return err
}
