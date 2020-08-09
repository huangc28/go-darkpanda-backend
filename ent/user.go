// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"github.com/facebookincubator/ent/dialect/sql"
	"github.com/huangc28/go-darkpanda-backend/ent/inquiry"
	"github.com/huangc28/go-darkpanda-backend/ent/user"
)

// User is the model entity for the User schema.
type User struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Username holds the value of the "username" field.
	Username string `json:"username,omitempty"`
	// PhoneVerified holds the value of the "phone_verified" field.
	PhoneVerified bool `json:"phone_verified,omitempty"`
	// AuthSmsCode holds the value of the "auth_sms_code" field.
	AuthSmsCode int16 `json:"auth_sms_code,omitempty"`
	// Gender holds the value of the "gender" field.
	Gender user.Gender `json:"gender,omitempty"`
	// PremiumKind holds the value of the "premium_kind" field.
	PremiumKind user.PremiumKind `json:"premium_kind,omitempty"`
	// PremiumExpiryDate holds the value of the "premium_expiry_date" field.
	PremiumExpiryDate time.Time `json:"premium_expiry_date,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the UserQuery when eager-loading is set.
	Edges UserEdges `json:"edges"`
}

// UserEdges holds the relations/edges for other nodes in the graph.
type UserEdges struct {
	// RefcodeInvitor holds the value of the refcode_invitor edge.
	RefcodeInvitor []*UserRefCodes
	// Userrefcodes holds the value of the userrefcodes edge.
	Userrefcodes []*UserRefCodes
	// Inquiry holds the value of the inquiry edge.
	Inquiry *Inquiry
	// ServiceCustomer holds the value of the service_customer edge.
	ServiceCustomer []*Service
	// ServiceProvider holds the value of the service_provider edge.
	ServiceProvider []*Service
	// Groups holds the value of the groups edge.
	Groups []*Group
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [6]bool
}

// RefcodeInvitorOrErr returns the RefcodeInvitor value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) RefcodeInvitorOrErr() ([]*UserRefCodes, error) {
	if e.loadedTypes[0] {
		return e.RefcodeInvitor, nil
	}
	return nil, &NotLoadedError{edge: "refcode_invitor"}
}

// UserrefcodesOrErr returns the Userrefcodes value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) UserrefcodesOrErr() ([]*UserRefCodes, error) {
	if e.loadedTypes[1] {
		return e.Userrefcodes, nil
	}
	return nil, &NotLoadedError{edge: "userrefcodes"}
}

// InquiryOrErr returns the Inquiry value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e UserEdges) InquiryOrErr() (*Inquiry, error) {
	if e.loadedTypes[2] {
		if e.Inquiry == nil {
			// The edge inquiry was loaded in eager-loading,
			// but was not found.
			return nil, &NotFoundError{label: inquiry.Label}
		}
		return e.Inquiry, nil
	}
	return nil, &NotLoadedError{edge: "inquiry"}
}

// ServiceCustomerOrErr returns the ServiceCustomer value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) ServiceCustomerOrErr() ([]*Service, error) {
	if e.loadedTypes[3] {
		return e.ServiceCustomer, nil
	}
	return nil, &NotLoadedError{edge: "service_customer"}
}

// ServiceProviderOrErr returns the ServiceProvider value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) ServiceProviderOrErr() ([]*Service, error) {
	if e.loadedTypes[4] {
		return e.ServiceProvider, nil
	}
	return nil, &NotLoadedError{edge: "service_provider"}
}

// GroupsOrErr returns the Groups value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) GroupsOrErr() ([]*Group, error) {
	if e.loadedTypes[5] {
		return e.Groups, nil
	}
	return nil, &NotLoadedError{edge: "groups"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*User) scanValues() []interface{} {
	return []interface{}{
		&sql.NullInt64{},  // id
		&sql.NullString{}, // username
		&sql.NullBool{},   // phone_verified
		&sql.NullInt64{},  // auth_sms_code
		&sql.NullString{}, // gender
		&sql.NullString{}, // premium_kind
		&sql.NullTime{},   // premium_expiry_date
		&sql.NullTime{},   // created_at
		&sql.NullTime{},   // updated_at
	}
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the User fields.
func (u *User) assignValues(values ...interface{}) error {
	if m, n := len(values), len(user.Columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	value, ok := values[0].(*sql.NullInt64)
	if !ok {
		return fmt.Errorf("unexpected type %T for field id", value)
	}
	u.ID = int(value.Int64)
	values = values[1:]
	if value, ok := values[0].(*sql.NullString); !ok {
		return fmt.Errorf("unexpected type %T for field username", values[0])
	} else if value.Valid {
		u.Username = value.String
	}
	if value, ok := values[1].(*sql.NullBool); !ok {
		return fmt.Errorf("unexpected type %T for field phone_verified", values[1])
	} else if value.Valid {
		u.PhoneVerified = value.Bool
	}
	if value, ok := values[2].(*sql.NullInt64); !ok {
		return fmt.Errorf("unexpected type %T for field auth_sms_code", values[2])
	} else if value.Valid {
		u.AuthSmsCode = int16(value.Int64)
	}
	if value, ok := values[3].(*sql.NullString); !ok {
		return fmt.Errorf("unexpected type %T for field gender", values[3])
	} else if value.Valid {
		u.Gender = user.Gender(value.String)
	}
	if value, ok := values[4].(*sql.NullString); !ok {
		return fmt.Errorf("unexpected type %T for field premium_kind", values[4])
	} else if value.Valid {
		u.PremiumKind = user.PremiumKind(value.String)
	}
	if value, ok := values[5].(*sql.NullTime); !ok {
		return fmt.Errorf("unexpected type %T for field premium_expiry_date", values[5])
	} else if value.Valid {
		u.PremiumExpiryDate = value.Time
	}
	if value, ok := values[6].(*sql.NullTime); !ok {
		return fmt.Errorf("unexpected type %T for field created_at", values[6])
	} else if value.Valid {
		u.CreatedAt = value.Time
	}
	if value, ok := values[7].(*sql.NullTime); !ok {
		return fmt.Errorf("unexpected type %T for field updated_at", values[7])
	} else if value.Valid {
		u.UpdatedAt = value.Time
	}
	return nil
}

// QueryRefcodeInvitor queries the refcode_invitor edge of the User.
func (u *User) QueryRefcodeInvitor() *UserRefCodesQuery {
	return (&UserClient{config: u.config}).QueryRefcodeInvitor(u)
}

// QueryUserrefcodes queries the userrefcodes edge of the User.
func (u *User) QueryUserrefcodes() *UserRefCodesQuery {
	return (&UserClient{config: u.config}).QueryUserrefcodes(u)
}

// QueryInquiry queries the inquiry edge of the User.
func (u *User) QueryInquiry() *InquiryQuery {
	return (&UserClient{config: u.config}).QueryInquiry(u)
}

// QueryServiceCustomer queries the service_customer edge of the User.
func (u *User) QueryServiceCustomer() *ServiceQuery {
	return (&UserClient{config: u.config}).QueryServiceCustomer(u)
}

// QueryServiceProvider queries the service_provider edge of the User.
func (u *User) QueryServiceProvider() *ServiceQuery {
	return (&UserClient{config: u.config}).QueryServiceProvider(u)
}

// QueryGroups queries the groups edge of the User.
func (u *User) QueryGroups() *GroupQuery {
	return (&UserClient{config: u.config}).QueryGroups(u)
}

// Update returns a builder for updating this User.
// Note that, you need to call User.Unwrap() before calling this method, if this User
// was returned from a transaction, and the transaction was committed or rolled back.
func (u *User) Update() *UserUpdateOne {
	return (&UserClient{config: u.config}).UpdateOne(u)
}

// Unwrap unwraps the entity that was returned from a transaction after it was closed,
// so that all next queries will be executed through the driver which created the transaction.
func (u *User) Unwrap() *User {
	tx, ok := u.config.driver.(*txDriver)
	if !ok {
		panic("ent: User is not a transactional entity")
	}
	u.config.driver = tx.drv
	return u
}

// String implements the fmt.Stringer.
func (u *User) String() string {
	var builder strings.Builder
	builder.WriteString("User(")
	builder.WriteString(fmt.Sprintf("id=%v", u.ID))
	builder.WriteString(", username=")
	builder.WriteString(u.Username)
	builder.WriteString(", phone_verified=")
	builder.WriteString(fmt.Sprintf("%v", u.PhoneVerified))
	builder.WriteString(", auth_sms_code=")
	builder.WriteString(fmt.Sprintf("%v", u.AuthSmsCode))
	builder.WriteString(", gender=")
	builder.WriteString(fmt.Sprintf("%v", u.Gender))
	builder.WriteString(", premium_kind=")
	builder.WriteString(fmt.Sprintf("%v", u.PremiumKind))
	builder.WriteString(", premium_expiry_date=")
	builder.WriteString(u.PremiumExpiryDate.Format(time.ANSIC))
	builder.WriteString(", created_at=")
	builder.WriteString(u.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", updated_at=")
	builder.WriteString(u.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Users is a parsable slice of User.
type Users []*User

func (u Users) config(cfg config) {
	for _i := range u {
		u[_i].config = cfg
	}
}
