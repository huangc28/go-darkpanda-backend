package schema

import (
	"github.com/facebookincubator/ent"
	"github.com/facebookincubator/ent/schema/edge"
	"github.com/facebookincubator/ent/schema/field"
)

type RefCodeType string

func (t RefCodeType) toString() string {
	return string(t)
}

var (
	Manager RefCodeType = "manager"
	Invitor RefCodeType = "invitor"
)

// UserRefCodes holds the schema definition for the UserRefCodes entity.
type UserRefCodes struct {
	ent.Schema
}

func (UserRefCodes) Config() ent.Config {
	return ent.Config{
		Table: "user_ref_codes",
	}
}

// Fields of the UserRefCodes.
func (UserRefCodes) Fields() []ent.Field {
	return []ent.Field{
		field.String("ref_code"),

		field.
			Enum("ref_code_type").
			Values(
				Manager.toString(),
				Invitor.toString(),
			),
	}
}

// Edges of the UserRefCodes.
func (UserRefCodes) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("users", User.Type).
			Ref("refcode_invitor").
			Unique(),

		edge.
			To("refcode_invitee", User.Type).
			StorageKey(edge.Column("invitee_id")).
			Unique(),
	}
}
