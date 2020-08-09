package schema

import (
	"github.com/facebookincubator/ent"
	"github.com/facebookincubator/ent/schema/field"
)

// GroupUsers holds the schema definition for the GroupUsers entity.
type GroupUsers struct {
	ent.Schema
}

func (GroupUsers) Config() ent.Config {
	return ent.Config{
		Table: "group_users",
	}
}

// Fields of the GroupUsers.
func (GroupUsers) Fields() []ent.Field {
	//return nil
	return []ent.Field{
		field.Enum("auth_type").Values("user", "admin"),
	}
}

// Edges of the GroupUsers.
func (GroupUsers) Edges() []ent.Edge {
	return nil
}
