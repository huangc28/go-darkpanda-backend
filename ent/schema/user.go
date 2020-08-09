package schema

import (
	"time"

	"github.com/facebookincubator/ent"
	"github.com/facebookincubator/ent/schema/edge"
	"github.com/facebookincubator/ent/schema/field"
)

type Gender string

func (g Gender) toString() string {
	return string(g)
}

var (
	Male   Gender = "male"
	Female Gender = "female"
)

type PremiumKind string

func (p PremiumKind) toString() string {
	return string(p)
}

var (
	Normal PremiumKind = "normal"
	Paid   PremiumKind = "paid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.
			String("username").
			Unique(),

		field.
			Bool("phone_verified").
			Default(false),

		field.
			Int16("auth_sms_code").
			Optional(),

		field.
			Enum("gender").
			Values(Male.toString(), Female.toString()),

		field.
			Enum("premium_kind").
			Values(Normal.toString(), Paid.toString()),

		field.
			Time("premium_expiry_date"),
		field.
			Time("created_at").
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			To("refcode_invitor", UserRefCodes.Type).
			StorageKey(edge.Column("invitor_id")),

		edge.
			From("userrefcodes", UserRefCodes.Type).
			Ref("refcode_invitee"),

		edge.To("inquiry", Inquiry.Type).
			StorageKey(edge.Column("inquirer_id")).
			Unique(),

		edge.From("service_customer", Service.Type).Ref("customer"),

		edge.From("service_provider", Service.Type).Ref("service_provider"),

		edge.From("groups", Group.Type).Ref("users"),
	}
}
