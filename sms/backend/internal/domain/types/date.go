package types

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const dateFormat = "2006-01-02"

// DateOnly 纯日期类型，JSON 序列化为 "2006-01-02" 格式。
type DateOnly struct {
	time.Time
}

func ParseDateOnly(s string) (DateOnly, error) {
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return DateOnly{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return DateOnly{Time: t}, nil
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(dateFormat) + `"`), nil
}

func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" || s == `""` {
		return nil
	}
	s = s[1 : len(s)-1]
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	d.Time = t
	return nil
}

// Scan implements sql.Scanner (pgx compatible).
func (d *DateOnly) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		d.Time = v
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan %T into DateOnly", src)
	}
}

// Value implements driver.Valuer.
func (d DateOnly) Value() (driver.Value, error) {
	return d.Time, nil
}

// String returns the date in YYYY-MM-DD format.
func (d DateOnly) String() string {
	return d.Format(dateFormat)
}
