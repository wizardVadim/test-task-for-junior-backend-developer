package taskrecurrence

import "errors"

var (
	ErrNotFound             = errors.New("task recurrence not found")
	ErrInvalidType          = errors.New("invalid task recurrence type")
	ErrInvalidStartDate     = errors.New("start date is required")
	ErrInvalidEveryNDays    = errors.New("every_n_days must be greater than 0 for daily recurrence")
	ErrInvalidDayOfMonth    = errors.New("day_of_month must be between 1 and 30 for monthly recurrence")
	ErrUnexpectedEveryNDays = errors.New("every_n_days must be empty for this recurrence type")
	ErrUnexpectedDayOfMonth = errors.New("day_of_month must be empty for this recurrence type")
)
