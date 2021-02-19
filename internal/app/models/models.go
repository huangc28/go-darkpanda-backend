// Code generated by sqlc. DO NOT EDIT.

package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

func (e *Gender) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Gender(s)
	case string:
		*e = Gender(s)
	default:
		return fmt.Errorf("unsupported scan type for Gender: %T", src)
	}
	return nil
}

type InquiryStatus string

const (
	InquiryStatusInquiring              InquiryStatus = "inquiring"
	InquiryStatusCanceled               InquiryStatus = "canceled"
	InquiryStatusExpired                InquiryStatus = "expired"
	InquiryStatusBooked                 InquiryStatus = "booked"
	InquiryStatusChatting               InquiryStatus = "chatting"
	InquiryStatusWaitForInquirerApprove InquiryStatus = "wait_for_inquirer_approve"
	InquiryStatusAsking                 InquiryStatus = "asking"
)

func (e *InquiryStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = InquiryStatus(s)
	case string:
		*e = InquiryStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for InquiryStatus: %T", src)
	}
	return nil
}

type LobbyStatus string

const (
	LobbyStatusWaiting LobbyStatus = "waiting"
	LobbyStatusPause   LobbyStatus = "pause"
	LobbyStatusExpired LobbyStatus = "expired"
	LobbyStatusLeft    LobbyStatus = "left"
	LobbyStatusAsking  LobbyStatus = "asking"
)

func (e *LobbyStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = LobbyStatus(s)
	case string:
		*e = LobbyStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for LobbyStatus: %T", src)
	}
	return nil
}

type PremiumType string

const (
	PremiumTypeNormal PremiumType = "normal"
	PremiumTypePaid   PremiumType = "paid"
)

func (e *PremiumType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = PremiumType(s)
	case string:
		*e = PremiumType(s)
	default:
		return fmt.Errorf("unsupported scan type for PremiumType: %T", src)
	}
	return nil
}

type RefCodeType string

const (
	RefCodeTypeInvitor RefCodeType = "invitor"
	RefCodeTypeManager RefCodeType = "manager"
)

func (e *RefCodeType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RefCodeType(s)
	case string:
		*e = RefCodeType(s)
	default:
		return fmt.Errorf("unsupported scan type for RefCodeType: %T", src)
	}
	return nil
}

type ServiceStatus string

const (
	ServiceStatusUnpaid          ServiceStatus = "unpaid"
	ServiceStatusToBeFulfilled   ServiceStatus = "to_be_fulfilled"
	ServiceStatusCanceled        ServiceStatus = "canceled"
	ServiceStatusFailedDueToBoth ServiceStatus = "failed_due_to_both"
	ServiceStatusGirlWaiting     ServiceStatus = "girl_waiting"
	ServiceStatusFufilling       ServiceStatus = "fufilling"
	ServiceStatusFailedDueToGirl ServiceStatus = "failed_due_to_girl"
	ServiceStatusFailedDueToMan  ServiceStatus = "failed_due_to_man"
	ServiceStatusCompleted       ServiceStatus = "completed"
	ServiceStatusNegotiating     ServiceStatus = "negotiating"
)

func (e *ServiceStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ServiceStatus(s)
	case string:
		*e = ServiceStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ServiceStatus: %T", src)
	}
	return nil
}

type ServiceType string

const (
	ServiceTypeSex      ServiceType = "sex"
	ServiceTypeDiner    ServiceType = "diner"
	ServiceTypeMovie    ServiceType = "movie"
	ServiceTypeShopping ServiceType = "shopping"
	ServiceTypeChat     ServiceType = "chat"
)

func (e *ServiceType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ServiceType(s)
	case string:
		*e = ServiceType(s)
	default:
		return fmt.Errorf("unsupported scan type for ServiceType: %T", src)
	}
	return nil
}

type Chatroom struct {
	ID           int64          `json:"id"`
	InquiryID    int32          `json:"inquiry_id"`
	ChannelUuid  sql.NullString `json:"channel_uuid"`
	MessageCount sql.NullInt32  `json:"message_count"`
	Enabled      sql.NullBool   `json:"enabled"`
	CreatedAt    time.Time      `json:"created_at"`
	ExpiredAt    time.Time      `json:"expired_at"`
	UpdatedAt    sql.NullTime   `json:"updated_at"`
	DeletedAt    sql.NullTime   `json:"deleted_at"`
}

type ChatroomUser struct {
	ID         int64        `json:"id"`
	ChatroomID int32        `json:"chatroom_id"`
	UserID     int32        `json:"user_id"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  sql.NullTime `json:"updated_at"`
	DeletedAt  sql.NullTime `json:"deleted_at"`
}

type Image struct {
	ID        int64        `json:"id"`
	UserID    int32        `json:"user_id"`
	Url       string       `json:"url"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

type LobbyUser struct {
	ID          int64        `json:"id"`
	ChannelUuid string       `json:"channel_uuid"`
	InquiryID   int32        `json:"inquiry_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   sql.NullTime `json:"updated_at"`
	DeletedAt   sql.NullTime `json:"deleted_at"`
	ExpiredAt   time.Time    `json:"expired_at"`
	LobbyStatus LobbyStatus  `json:"lobby_status"`
}

type Payment struct {
	ID         int64          `json:"id"`
	PayerID    int32          `json:"payer_id"`
	PayeeID    int32          `json:"payee_id"`
	ServiceID  int32          `json:"service_id"`
	Price      string         `json:"price"`
	RecTradeID sql.NullString `json:"rec_trade_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  sql.NullTime   `json:"updated_at"`
	DeletedAt  sql.NullTime   `json:"deleted_at"`
}

type Service struct {
	ID                int64          `json:"id"`
	Uuid              uuid.UUID      `json:"uuid"`
	CustomerID        sql.NullInt32  `json:"customer_id"`
	ServiceProviderID sql.NullInt32  `json:"service_provider_id"`
	Price             sql.NullString `json:"price"`
	Duration          sql.NullInt32  `json:"duration"`
	AppointmentTime   sql.NullTime   `json:"appointment_time"`
	Lng               sql.NullString `json:"lng"`
	Lat               sql.NullString `json:"lat"`
	ServiceType       ServiceType    `json:"service_type"`
	GirlReady         sql.NullBool   `json:"girl_ready"`
	ManReady          sql.NullBool   `json:"man_ready"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         sql.NullTime   `json:"updated_at"`
	DeletedAt         sql.NullTime   `json:"deleted_at"`
	Budget            sql.NullString `json:"budget"`
	InquiryID         int32          `json:"inquiry_id"`
	ServiceStatus     ServiceStatus  `json:"service_status"`
}

type ServiceInquiry struct {
	ID              int64          `json:"id"`
	InquirerID      sql.NullInt32  `json:"inquirer_id"`
	Budget          string         `json:"budget"`
	ServiceType     ServiceType    `json:"service_type"`
	InquiryStatus   InquiryStatus  `json:"inquiry_status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       sql.NullTime   `json:"updated_at"`
	DeletedAt       sql.NullTime   `json:"deleted_at"`
	Uuid            string         `json:"uuid"`
	Price           sql.NullString `json:"price"`
	Duration        sql.NullInt32  `json:"duration"`
	AppointmentTime sql.NullTime   `json:"appointment_time"`
	Lng             sql.NullString `json:"lng"`
	Lat             sql.NullString `json:"lat"`
	// Time that this inquiry will be invalid.
	ExpiredAt sql.NullTime  `json:"expired_at"`
	PickerID  sql.NullInt32 `json:"picker_id"`
}

type User struct {
	ID                int64          `json:"id"`
	Username          string         `json:"username"`
	PhoneVerified     bool           `json:"phone_verified"`
	AuthSmsCode       sql.NullInt32  `json:"auth_sms_code"`
	Gender            Gender         `json:"gender"`
	PremiumType       PremiumType    `json:"premium_type"`
	PremiumExpiryDate sql.NullTime   `json:"premium_expiry_date"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         sql.NullTime   `json:"updated_at"`
	DeletedAt         sql.NullTime   `json:"deleted_at"`
	Uuid              string         `json:"uuid"`
	PhoneVerifyCode   sql.NullString `json:"phone_verify_code"`
	AvatarUrl         sql.NullString `json:"avatar_url"`
	Nationality       sql.NullString `json:"nationality"`
	Region            sql.NullString `json:"region"`
	Age               sql.NullInt32  `json:"age"`
	Height            sql.NullString `json:"height"`
	Weight            sql.NullString `json:"weight"`
	Habbits           sql.NullString `json:"habbits"`
	Description       sql.NullString `json:"description"`
	BreastSize        sql.NullString `json:"breast_size"`
	Mobile            sql.NullString `json:"mobile"`
}

type UserRating struct {
	ID         int64          `json:"id"`
	FromUserID sql.NullInt32  `json:"from_user_id"`
	ToUserID   sql.NullInt32  `json:"to_user_id"`
	Rating     sql.NullInt32  `json:"rating"`
	Comments   sql.NullString `json:"comments"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  sql.NullTime   `json:"updated_at"`
	DeletedAt  sql.NullTime   `json:"deleted_at"`
}

type UserRefcode struct {
	ID          int64         `json:"id"`
	InvitorID   int32         `json:"invitor_id"`
	InviteeID   sql.NullInt32 `json:"invitee_id"`
	RefCode     string        `json:"ref_code"`
	RefCodeType RefCodeType   `json:"ref_code_type"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   sql.NullTime  `json:"updated_at"`
	DeletedAt   sql.NullTime  `json:"deleted_at"`
	// Time that this referral code will be invalid.
	ExpiredAt sql.NullTime `json:"expired_at"`
}
