package db

import "strings"

func ComposeFieldsSQLString(fields ...string) string {
	if len(fields) == 0 {
		fields = append(fields, "*")
	}

	fieldsStr := strings.TrimSuffix(strings.Join(fields, ","), ",")

	return fieldsStr
}
