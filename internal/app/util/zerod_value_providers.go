package util

import (
	"database/sql"

	"github.com/shopspring/decimal"
)

// SQLNilString provide sql compatible nullable string value. If the given string pointer is a nil pointer,
// returns storable empty string value in sql compatible fashion.
func SQLNilString(str *string) sql.NullString {
	s := sql.NullString{
		String: "",
		Valid:  false,
	}

	if str != nil {
		s = sql.NullString{
			String: *str,
			Valid:  true,
		}
	}

	return s
}

func SQLNilInt32(num *int) sql.NullInt32 {
	n := sql.NullInt32{
		Int32: 0,
		Valid: false,
	}

	if num != nil {
		n = sql.NullInt32{
			Int32: int32(*num),
			Valid: true,
		}
	}

	return n
}

func SQLNilFloat64(f *float64) sql.NullString {
	n := sql.NullString{
		String: "",
		Valid:  false,
	}

	if f != nil {
		n = sql.NullString{
			String: decimal.NewFromFloat(*f).String(),
			Valid:  true,
		}
	}

	return n
}
