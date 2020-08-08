// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"github.com/facebookincubator/ent/dialect/sql"
	"github.com/huangc28/go-darkpanda-backend/ent/inquiry"
	"github.com/huangc28/go-darkpanda-backend/ent/user"
)

// Inquiry is the model entity for the Inquiry schema.
type Inquiry struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Budget holds the value of the "budget" field.
	Budget float32 `json:"budget,omitempty"`
	// ServiceType holds the value of the "service_type" field.
	ServiceType inquiry.ServiceType `json:"service_type,omitempty"`
	// InquiryStatus holds the value of the "inquiry_status" field.
	InquiryStatus inquiry.InquiryStatus `json:"inquiry_status,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the InquiryQuery when eager-loading is set.
	Edges       InquiryEdges `json:"edges"`
	inquirer_id *int
}

// InquiryEdges holds the relations/edges for other nodes in the graph.
type InquiryEdges struct {
	// Users holds the value of the users edge.
	Users *User
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// UsersOrErr returns the Users value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e InquiryEdges) UsersOrErr() (*User, error) {
	if e.loadedTypes[0] {
		if e.Users == nil {
			// The edge users was loaded in eager-loading,
			// but was not found.
			return nil, &NotFoundError{label: user.Label}
		}
		return e.Users, nil
	}
	return nil, &NotLoadedError{edge: "users"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Inquiry) scanValues() []interface{} {
	return []interface{}{
		&sql.NullInt64{},   // id
		&sql.NullFloat64{}, // budget
		&sql.NullString{},  // service_type
		&sql.NullString{},  // inquiry_status
	}
}

// fkValues returns the types for scanning foreign-keys values from sql.Rows.
func (*Inquiry) fkValues() []interface{} {
	return []interface{}{
		&sql.NullInt64{}, // inquirer_id
	}
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Inquiry fields.
func (i *Inquiry) assignValues(values ...interface{}) error {
	if m, n := len(values), len(inquiry.Columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	value, ok := values[0].(*sql.NullInt64)
	if !ok {
		return fmt.Errorf("unexpected type %T for field id", value)
	}
	i.ID = int(value.Int64)
	values = values[1:]
	if value, ok := values[0].(*sql.NullFloat64); !ok {
		return fmt.Errorf("unexpected type %T for field budget", values[0])
	} else if value.Valid {
		i.Budget = float32(value.Float64)
	}
	if value, ok := values[1].(*sql.NullString); !ok {
		return fmt.Errorf("unexpected type %T for field service_type", values[1])
	} else if value.Valid {
		i.ServiceType = inquiry.ServiceType(value.String)
	}
	if value, ok := values[2].(*sql.NullString); !ok {
		return fmt.Errorf("unexpected type %T for field inquiry_status", values[2])
	} else if value.Valid {
		i.InquiryStatus = inquiry.InquiryStatus(value.String)
	}
	values = values[3:]
	if len(values) == len(inquiry.ForeignKeys) {
		if value, ok := values[0].(*sql.NullInt64); !ok {
			return fmt.Errorf("unexpected type %T for edge-field inquirer_id", value)
		} else if value.Valid {
			i.inquirer_id = new(int)
			*i.inquirer_id = int(value.Int64)
		}
	}
	return nil
}

// QueryUsers queries the users edge of the Inquiry.
func (i *Inquiry) QueryUsers() *UserQuery {
	return (&InquiryClient{config: i.config}).QueryUsers(i)
}

// Update returns a builder for updating this Inquiry.
// Note that, you need to call Inquiry.Unwrap() before calling this method, if this Inquiry
// was returned from a transaction, and the transaction was committed or rolled back.
func (i *Inquiry) Update() *InquiryUpdateOne {
	return (&InquiryClient{config: i.config}).UpdateOne(i)
}

// Unwrap unwraps the entity that was returned from a transaction after it was closed,
// so that all next queries will be executed through the driver which created the transaction.
func (i *Inquiry) Unwrap() *Inquiry {
	tx, ok := i.config.driver.(*txDriver)
	if !ok {
		panic("ent: Inquiry is not a transactional entity")
	}
	i.config.driver = tx.drv
	return i
}

// String implements the fmt.Stringer.
func (i *Inquiry) String() string {
	var builder strings.Builder
	builder.WriteString("Inquiry(")
	builder.WriteString(fmt.Sprintf("id=%v", i.ID))
	builder.WriteString(", budget=")
	builder.WriteString(fmt.Sprintf("%v", i.Budget))
	builder.WriteString(", service_type=")
	builder.WriteString(fmt.Sprintf("%v", i.ServiceType))
	builder.WriteString(", inquiry_status=")
	builder.WriteString(fmt.Sprintf("%v", i.InquiryStatus))
	builder.WriteByte(')')
	return builder.String()
}

// Inquiries is a parsable slice of Inquiry.
type Inquiries []*Inquiry

func (i Inquiries) config(cfg config) {
	for _i := range i {
		i[_i].config = cfg
	}
}
