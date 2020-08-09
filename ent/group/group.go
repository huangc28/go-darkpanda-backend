// Code generated by entc, DO NOT EDIT.

package group

import (
	"fmt"
)

const (
	// Label holds the string label denoting the group type in the database.
	Label = "group"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldGroupType holds the string denoting the group_type field in the database.
	FieldGroupType = "group_type"

	// EdgeUsers holds the string denoting the users edge name in mutations.
	EdgeUsers = "users"

	// Table holds the table name of the group in the database.
	Table = "groups"
	// UsersTable is the table the holds the users relation/edge. The primary key declared below.
	UsersTable = "group_users"
	// UsersInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UsersInverseTable = "users"
)

// Columns holds all SQL columns for group fields.
var Columns = []string{
	FieldID,
	FieldGroupType,
}

var (
	// UsersPrimaryKey and UsersColumn2 are the table columns denoting the
	// primary key for the users relation (M2M).
	UsersPrimaryKey = []string{"group_id", "user_id"}
)

// GroupType defines the type for the group_type enum field.
type GroupType string

// GroupType values.
const (
	GroupTypeAgroup GroupType = "Agroup"
	GroupTypeBgroup GroupType = "Bgroup"
)

func (gt GroupType) String() string {
	return string(gt)
}

// GroupTypeValidator is a validator for the "group_type" field enum values. It is called by the builders before save.
func GroupTypeValidator(gt GroupType) error {
	switch gt {
	case GroupTypeAgroup, GroupTypeBgroup:
		return nil
	default:
		return fmt.Errorf("group: invalid enum value for group_type field: %q", gt)
	}
}
