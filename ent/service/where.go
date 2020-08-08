// Code generated by entc, DO NOT EDIT.

package service

import (
	"time"

	"github.com/facebookincubator/ent/dialect/sql"
	"github.com/facebookincubator/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/huangc28/go-darkpanda-backend/ent/predicate"
)

// ID filters vertices based on their identifier.
func ID(id int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldID), id))
	})
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.In(s.C(FieldID), v...))
	})
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.NotIn(s.C(FieldID), v...))
	})
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldID), id))
	})
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldID), id))
	})
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldID), id))
	})
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldID), id))
	})
}

// UUID applies equality check predicate on the "uuid" field. It's identical to UUIDEQ.
func UUID(v uuid.UUID) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldUUID), v))
	})
}

// Price applies equality check predicate on the "price" field. It's identical to PriceEQ.
func Price(v float32) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldPrice), v))
	})
}

// Duration applies equality check predicate on the "duration" field. It's identical to DurationEQ.
func Duration(v int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDuration), v))
	})
}

// AppointmentTime applies equality check predicate on the "appointment_time" field. It's identical to AppointmentTimeEQ.
func AppointmentTime(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldAppointmentTime), v))
	})
}

// Lng applies equality check predicate on the "lng" field. It's identical to LngEQ.
func Lng(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldLng), v))
	})
}

// Lat applies equality check predicate on the "lat" field. It's identical to LatEQ.
func Lat(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldLat), v))
	})
}

// GirlReady applies equality check predicate on the "girl_ready" field. It's identical to GirlReadyEQ.
func GirlReady(v bool) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldGirlReady), v))
	})
}

// ManReady applies equality check predicate on the "man_ready" field. It's identical to ManReadyEQ.
func ManReady(v bool) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldManReady), v))
	})
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldCreatedAt), v))
	})
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldUpdatedAt), v))
	})
}

// UUIDEQ applies the EQ predicate on the "uuid" field.
func UUIDEQ(v uuid.UUID) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldUUID), v))
	})
}

// UUIDNEQ applies the NEQ predicate on the "uuid" field.
func UUIDNEQ(v uuid.UUID) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldUUID), v))
	})
}

// UUIDIn applies the In predicate on the "uuid" field.
func UUIDIn(vs ...uuid.UUID) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldUUID), v...))
	})
}

// UUIDNotIn applies the NotIn predicate on the "uuid" field.
func UUIDNotIn(vs ...uuid.UUID) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldUUID), v...))
	})
}

// UUIDGT applies the GT predicate on the "uuid" field.
func UUIDGT(v uuid.UUID) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldUUID), v))
	})
}

// UUIDGTE applies the GTE predicate on the "uuid" field.
func UUIDGTE(v uuid.UUID) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldUUID), v))
	})
}

// UUIDLT applies the LT predicate on the "uuid" field.
func UUIDLT(v uuid.UUID) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldUUID), v))
	})
}

// UUIDLTE applies the LTE predicate on the "uuid" field.
func UUIDLTE(v uuid.UUID) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldUUID), v))
	})
}

// PriceEQ applies the EQ predicate on the "price" field.
func PriceEQ(v float32) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldPrice), v))
	})
}

// PriceNEQ applies the NEQ predicate on the "price" field.
func PriceNEQ(v float32) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldPrice), v))
	})
}

// PriceIn applies the In predicate on the "price" field.
func PriceIn(vs ...float32) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldPrice), v...))
	})
}

// PriceNotIn applies the NotIn predicate on the "price" field.
func PriceNotIn(vs ...float32) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldPrice), v...))
	})
}

// PriceGT applies the GT predicate on the "price" field.
func PriceGT(v float32) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldPrice), v))
	})
}

// PriceGTE applies the GTE predicate on the "price" field.
func PriceGTE(v float32) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldPrice), v))
	})
}

// PriceLT applies the LT predicate on the "price" field.
func PriceLT(v float32) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldPrice), v))
	})
}

// PriceLTE applies the LTE predicate on the "price" field.
func PriceLTE(v float32) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldPrice), v))
	})
}

// DurationEQ applies the EQ predicate on the "duration" field.
func DurationEQ(v int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDuration), v))
	})
}

// DurationNEQ applies the NEQ predicate on the "duration" field.
func DurationNEQ(v int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldDuration), v))
	})
}

// DurationIn applies the In predicate on the "duration" field.
func DurationIn(vs ...int) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldDuration), v...))
	})
}

// DurationNotIn applies the NotIn predicate on the "duration" field.
func DurationNotIn(vs ...int) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldDuration), v...))
	})
}

// DurationGT applies the GT predicate on the "duration" field.
func DurationGT(v int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldDuration), v))
	})
}

// DurationGTE applies the GTE predicate on the "duration" field.
func DurationGTE(v int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldDuration), v))
	})
}

// DurationLT applies the LT predicate on the "duration" field.
func DurationLT(v int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldDuration), v))
	})
}

// DurationLTE applies the LTE predicate on the "duration" field.
func DurationLTE(v int) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldDuration), v))
	})
}

// AppointmentTimeEQ applies the EQ predicate on the "appointment_time" field.
func AppointmentTimeEQ(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldAppointmentTime), v))
	})
}

// AppointmentTimeNEQ applies the NEQ predicate on the "appointment_time" field.
func AppointmentTimeNEQ(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldAppointmentTime), v))
	})
}

// AppointmentTimeIn applies the In predicate on the "appointment_time" field.
func AppointmentTimeIn(vs ...time.Time) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldAppointmentTime), v...))
	})
}

// AppointmentTimeNotIn applies the NotIn predicate on the "appointment_time" field.
func AppointmentTimeNotIn(vs ...time.Time) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldAppointmentTime), v...))
	})
}

// AppointmentTimeGT applies the GT predicate on the "appointment_time" field.
func AppointmentTimeGT(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldAppointmentTime), v))
	})
}

// AppointmentTimeGTE applies the GTE predicate on the "appointment_time" field.
func AppointmentTimeGTE(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldAppointmentTime), v))
	})
}

// AppointmentTimeLT applies the LT predicate on the "appointment_time" field.
func AppointmentTimeLT(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldAppointmentTime), v))
	})
}

// AppointmentTimeLTE applies the LTE predicate on the "appointment_time" field.
func AppointmentTimeLTE(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldAppointmentTime), v))
	})
}

// LngEQ applies the EQ predicate on the "lng" field.
func LngEQ(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldLng), v))
	})
}

// LngNEQ applies the NEQ predicate on the "lng" field.
func LngNEQ(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldLng), v))
	})
}

// LngIn applies the In predicate on the "lng" field.
func LngIn(vs ...float64) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldLng), v...))
	})
}

// LngNotIn applies the NotIn predicate on the "lng" field.
func LngNotIn(vs ...float64) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldLng), v...))
	})
}

// LngGT applies the GT predicate on the "lng" field.
func LngGT(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldLng), v))
	})
}

// LngGTE applies the GTE predicate on the "lng" field.
func LngGTE(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldLng), v))
	})
}

// LngLT applies the LT predicate on the "lng" field.
func LngLT(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldLng), v))
	})
}

// LngLTE applies the LTE predicate on the "lng" field.
func LngLTE(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldLng), v))
	})
}

// LatEQ applies the EQ predicate on the "lat" field.
func LatEQ(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldLat), v))
	})
}

// LatNEQ applies the NEQ predicate on the "lat" field.
func LatNEQ(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldLat), v))
	})
}

// LatIn applies the In predicate on the "lat" field.
func LatIn(vs ...float64) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldLat), v...))
	})
}

// LatNotIn applies the NotIn predicate on the "lat" field.
func LatNotIn(vs ...float64) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldLat), v...))
	})
}

// LatGT applies the GT predicate on the "lat" field.
func LatGT(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldLat), v))
	})
}

// LatGTE applies the GTE predicate on the "lat" field.
func LatGTE(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldLat), v))
	})
}

// LatLT applies the LT predicate on the "lat" field.
func LatLT(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldLat), v))
	})
}

// LatLTE applies the LTE predicate on the "lat" field.
func LatLTE(v float64) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldLat), v))
	})
}

// ServiceTypeEQ applies the EQ predicate on the "service_type" field.
func ServiceTypeEQ(v ServiceType) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldServiceType), v))
	})
}

// ServiceTypeNEQ applies the NEQ predicate on the "service_type" field.
func ServiceTypeNEQ(v ServiceType) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldServiceType), v))
	})
}

// ServiceTypeIn applies the In predicate on the "service_type" field.
func ServiceTypeIn(vs ...ServiceType) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldServiceType), v...))
	})
}

// ServiceTypeNotIn applies the NotIn predicate on the "service_type" field.
func ServiceTypeNotIn(vs ...ServiceType) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldServiceType), v...))
	})
}

// ServiceStatusEQ applies the EQ predicate on the "service_status" field.
func ServiceStatusEQ(v ServiceStatus) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldServiceStatus), v))
	})
}

// ServiceStatusNEQ applies the NEQ predicate on the "service_status" field.
func ServiceStatusNEQ(v ServiceStatus) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldServiceStatus), v))
	})
}

// ServiceStatusIn applies the In predicate on the "service_status" field.
func ServiceStatusIn(vs ...ServiceStatus) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldServiceStatus), v...))
	})
}

// ServiceStatusNotIn applies the NotIn predicate on the "service_status" field.
func ServiceStatusNotIn(vs ...ServiceStatus) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldServiceStatus), v...))
	})
}

// ServiceStatusIsNil applies the IsNil predicate on the "service_status" field.
func ServiceStatusIsNil() predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.IsNull(s.C(FieldServiceStatus)))
	})
}

// ServiceStatusNotNil applies the NotNil predicate on the "service_status" field.
func ServiceStatusNotNil() predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NotNull(s.C(FieldServiceStatus)))
	})
}

// GirlReadyEQ applies the EQ predicate on the "girl_ready" field.
func GirlReadyEQ(v bool) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldGirlReady), v))
	})
}

// GirlReadyNEQ applies the NEQ predicate on the "girl_ready" field.
func GirlReadyNEQ(v bool) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldGirlReady), v))
	})
}

// ManReadyEQ applies the EQ predicate on the "man_ready" field.
func ManReadyEQ(v bool) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldManReady), v))
	})
}

// ManReadyNEQ applies the NEQ predicate on the "man_ready" field.
func ManReadyNEQ(v bool) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldManReady), v))
	})
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldCreatedAt), v))
	})
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldCreatedAt), v))
	})
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldCreatedAt), v...))
	})
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldCreatedAt), v...))
	})
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldCreatedAt), v))
	})
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldCreatedAt), v))
	})
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldCreatedAt), v))
	})
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldCreatedAt), v))
	})
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldUpdatedAt), v))
	})
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldUpdatedAt), v))
	})
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldUpdatedAt), v...))
	})
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.Service {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Service(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldUpdatedAt), v...))
	})
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldUpdatedAt), v))
	})
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldUpdatedAt), v))
	})
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldUpdatedAt), v))
	})
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldUpdatedAt), v))
	})
}

// HasCustomer applies the HasEdge predicate on the "customer" edge.
func HasCustomer() predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(CustomerTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, CustomerTable, CustomerColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasCustomerWith applies the HasEdge predicate on the "customer" edge with a given conditions (other predicates).
func HasCustomerWith(preds ...predicate.User) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(CustomerInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, CustomerTable, CustomerColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasServiceProvider applies the HasEdge predicate on the "service_provider" edge.
func HasServiceProvider() predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(ServiceProviderTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, ServiceProviderTable, ServiceProviderColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasServiceProviderWith applies the HasEdge predicate on the "service_provider" edge with a given conditions (other predicates).
func HasServiceProviderWith(preds ...predicate.User) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(ServiceProviderInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, ServiceProviderTable, ServiceProviderColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups list of predicates with the AND operator between them.
func And(predicates ...predicate.Service) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups list of predicates with the OR operator between them.
func Or(predicates ...predicate.Service) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Service) predicate.Service {
	return predicate.Service(func(s *sql.Selector) {
		p(s.Not())
	})
}
