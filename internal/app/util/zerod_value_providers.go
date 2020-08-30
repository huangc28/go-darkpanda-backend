package util

import "database/sql"

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
