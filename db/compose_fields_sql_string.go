package db

import (
	"bytes"
	"fmt"
	"strings"
)

func ComposeFieldsSQLString(fields ...string) string {
	if len(fields) == 0 {
		fields = append(fields, "*")
	}

	fieldsStr := strings.TrimSuffix(strings.Join(fields, ","), ",")

	return fieldsStr
}

// Compose a list that the sequal can compare against.
// WHERE name IN ('apple', 'orange', 'banana')
func ComposeStringList(strVals ...string) string {
	if len(strVals) == 0 {
		strVals = append(strVals, "'*'")
	}

	var buffer bytes.Buffer

	for _, strVal := range strVals {
		buffer.WriteString(fmt.Sprintf("'%s',", strVal))
	}

	return strings.TrimSuffix(buffer.String(), ",")
}
