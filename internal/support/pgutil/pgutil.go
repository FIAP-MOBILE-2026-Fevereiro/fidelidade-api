package pgutil

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TextOrEmpty(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func Text(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{String: value, Valid: true}
}

func Time(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func TimePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	timestamp := value.Time.UTC()
	return &timestamp
}

func Bool(value pgtype.Bool) bool {
	return value.Valid && value.Bool
}
