package schema

import (
	"time"

	"github.com/facebookincubator/ent"
	"github.com/facebookincubator/ent/schema/edge"
	"github.com/facebookincubator/ent/schema/field"
	"github.com/google/uuid"
)

type ServiceStatus string

func (ss ServiceStatus) toString() string {
	return string(ss)
}

const (
	ToBeFulFilled   ServiceStatus = "to_be_fulfilled"
	ServiceCanceled ServiceStatus = "canceled"
	FailedDueToBoth ServiceStatus = "failed_due_to_both"
	GirlWaiting     ServiceStatus = "girl_waiting"
	Fulfilling      ServiceStatus = "fulfilling"
	Completed       ServiceStatus = "completed"
	FailedDueToGirl ServiceStatus = "failed_due_to_girl"
	FailedDueToMan  ServiceStatus = "failed_due_to_man"
)

// Service holds the schema definition for the Service entity.
type Service struct {
	ent.Schema
}

// Fields of the Service.
func (Service) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}),
		field.Float32("price"),
		field.Int("duration"),
		field.Time("appointment_time"),
		field.Float("lng").Comment("longtitude of service location."),
		field.Float("lat").Comment("latitude of service location"),
		field.Enum("service_type").Values(
			Sex.toString(),
			Diner.toString(),
			Movie.toString(),
			Shopping.toString(),
			Chat.toString(),
		),
		field.
			Enum("service_status").
			Values(
				ToBeFulFilled.toString(),
				ServiceCanceled.toString(),
				FailedDueToBoth.toString(),
				GirlWaiting.toString(),
				Fulfilling.toString(),
				Completed.toString(),
				FailedDueToGirl.toString(),
				FailedDueToMan.toString(),
			).
			Optional(),

		field.
			Bool("girl_ready").
			Default(false),

		field.
			Bool("man_ready").
			Default(false),

		field.
			Time("created_at").
			Default(time.Now),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Service.
func (Service) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			To("customer", User.Type).
			StorageKey(edge.Column("customer_id")).
			Unique(),

		edge.
			To("service_provider", User.Type).
			StorageKey(edge.Column("service_provider_id")).
			Unique(),
	}
}
