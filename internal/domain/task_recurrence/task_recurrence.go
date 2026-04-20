package taskrecurrence

import (
	"fmt"
	"time"
)

type Type string

const (
	TypeDaily         Type = "daily"
	TypeMonthly       Type = "monthly"
	TypeSpecificDates Type = "specific_dates"
	TypeEvenDays      Type = "even_days"
	TypeOddDays       Type = "odd_days"
)

type TaskRecurrence struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	Type       Type      `json:"type"`
	StartDate  time.Time `json:"start_date"`
	EveryNDays *int      `json:"every_n_days"`
	DayOfMonth *int      `json:"day_of_month"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (t Type) Valid() bool {
	switch t {
	case TypeDaily, TypeMonthly, TypeSpecificDates, TypeEvenDays, TypeOddDays:
		return true
	default:
		return false
	}
}

func (r TaskRecurrence) Validate() error {
	if !r.Type.Valid() {
		return ErrInvalidType
	}

	if r.StartDate.IsZero() {
		return ErrInvalidStartDate
	}

	switch r.Type {
	case TypeDaily:
		if r.EveryNDays == nil || *r.EveryNDays <= 0 {
			return ErrInvalidEveryNDays
		}
		if r.DayOfMonth != nil {
			return ErrUnexpectedDayOfMonth
		}

	case TypeMonthly:
		if r.DayOfMonth == nil || *r.DayOfMonth < 1 || *r.DayOfMonth > 30 {
			return ErrInvalidDayOfMonth
		}
		if r.EveryNDays != nil {
			return ErrUnexpectedEveryNDays
		}

	case TypeSpecificDates, TypeEvenDays, TypeOddDays:
		if r.EveryNDays != nil {
			return ErrUnexpectedEveryNDays
		}
		if r.DayOfMonth != nil {
			return ErrUnexpectedDayOfMonth
		}

	default:
		return fmt.Errorf("%w: %s", ErrInvalidType, r.Type)
	}

	return nil
}
