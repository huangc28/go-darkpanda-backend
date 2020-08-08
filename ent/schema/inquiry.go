package schema

import (
	"github.com/facebookincubator/ent"
	"github.com/facebookincubator/ent/schema/edge"
	"github.com/facebookincubator/ent/schema/field"
)

type ServiceType string

func (st ServiceType) toString() string {
	return string(st)
}

var (
	Sex      ServiceType = "sex"
	Diner    ServiceType = "diner"
	Movie    ServiceType = "movie"
	Shopping ServiceType = "shopping"
	Chat     ServiceType = "chat"
)

type InquiryStatus string

func (st InquiryStatus) toString() string {
	return string(st)
}

var (
	Inquiring InquiryStatus = "inquiring"
	Booked    InquiryStatus = "booked"
	Canceled  InquiryStatus = "canceled"
	Expired   InquiryStatus = "expired"
)

// Inquiry holds the schema definition for the Inquiry entity.
type Inquiry struct {
	ent.Schema
}

// Fields of the Inquiry.
func (Inquiry) Fields() []ent.Field {
	return []ent.Field{
		field.Float32("budget"),

		field.Enum("service_type").Values(
			Sex.toString(),
			Diner.toString(),
			Movie.toString(),
			Shopping.toString(),
			Chat.toString(),
		),

		field.Enum("inquiry_status").Values(
			Inquiring.toString(),
			Booked.toString(),
			Canceled.toString(),
			Expired.toString(),
		),
	}
}

// Edges of the Inquiry.
func (Inquiry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("users", User.Type).
			Ref("inquiry").
			Unique(),
	}
}
