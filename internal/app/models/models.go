// Code generated by sqlc. DO NOT EDIT.

package models

import (
	"database/sql"
	"fmt"
	"time"
)

type CancelCause string

const (
	CancelCauseNone                            CancelCause = "none"
	CancelCauseGirlCancelBeforeAppointmentTime CancelCause = "girl_cancel_before_appointment_time"
	CancelCauseGirlCancelAfterAppointmentTime  CancelCause = "girl_cancel_after_appointment_time"
	CancelCauseGuyCancelBeforeAppointmentTime  CancelCause = "guy_cancel_before_appointment_time"
	CancelCauseGuyCancelAfterAppointmentTime   CancelCause = "guy_cancel_after_appointment_time"
	CancelCausePaymentFailed                   CancelCause = "payment_failed"
)

func (e *CancelCause) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = CancelCause(s)
	case string:
		*e = CancelCause(s)
	default:
		return fmt.Errorf("unsupported scan type for CancelCause: %T", src)
	}
	return nil
}

type ChatroomType string

const (
	ChatroomTypeInquiryChat ChatroomType = "inquiry_chat"
	ChatroomTypeServiceChat ChatroomType = "service_chat"
)

func (e *ChatroomType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ChatroomType(s)
	case string:
		*e = ChatroomType(s)
	default:
		return fmt.Errorf("unsupported scan type for ChatroomType: %T", src)
	}
	return nil
}

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

type InquiryType string

const (
	InquiryTypeDirect InquiryType = "direct"
	InquiryTypeRandom InquiryType = "random"
)

func (e *InquiryType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = InquiryType(s)
	case string:
		*e = InquiryType(s)
	default:
		return fmt.Errorf("unsupported scan type for InquiryType: %T", src)
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

type OrderStatus string

const (
	OrderStatusInit     OrderStatus = "init"
	OrderStatusOrdering OrderStatus = "ordering"
	OrderStatusSuccess  OrderStatus = "success"
	OrderStatusFailed   OrderStatus = "failed"
)

func (e *OrderStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = OrderStatus(s)
	case string:
		*e = OrderStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for OrderStatus: %T", src)
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

type ServiceOptionsType string

const (
	ServiceOptionsTypeDefault ServiceOptionsType = "default"
	ServiceOptionsTypeCustom  ServiceOptionsType = "custom"
)

func (e *ServiceOptionsType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ServiceOptionsType(s)
	case string:
		*e = ServiceOptionsType(s)
	default:
		return fmt.Errorf("unsupported scan type for ServiceOptionsType: %T", src)
	}
	return nil
}

type ServiceStatus string

const (
	ServiceStatusUnpaid        ServiceStatus = "unpaid"
	ServiceStatusPaymentFailed ServiceStatus = "payment_failed"
	ServiceStatusToBeFulfilled ServiceStatus = "to_be_fulfilled"
	ServiceStatusCanceled      ServiceStatus = "canceled"
	ServiceStatusExpired       ServiceStatus = "expired"
	ServiceStatusFulfilling    ServiceStatus = "fulfilling"
	ServiceStatusCompleted     ServiceStatus = "completed"
	ServiceStatusNegotiating   ServiceStatus = "negotiating"
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

type VerifyStatus string

const (
	VerifyStatusPending      VerifyStatus = "pending"
	VerifyStatusVerifying    VerifyStatus = "verifying"
	VerifyStatusVerified     VerifyStatus = "verified"
	VerifyStatusVerifyFailed VerifyStatus = "verify_failed"
)

func (e *VerifyStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = VerifyStatus(s)
	case string:
		*e = VerifyStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for VerifyStatus: %T", src)
	}
	return nil
}

type BankAccount struct {
	ID            int32        `json:"id"`
	UserID        int32        `json:"user_id"`
	BankName      string       `json:"bank_name"`
	Branch        string       `json:"branch"`
	AccountNumber string       `json:"account_number"`
	VerifyStatus  VerifyStatus `json:"verify_status"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     sql.NullTime `json:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`
}

type BlockList struct {
	ID            int32        `json:"id"`
	UserID        int32        `json:"user_id"`
	BlockedUserID int32        `json:"blocked_user_id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     sql.NullTime `json:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`
}

type Chatroom struct {
	ID           int64          `json:"id"`
	InquiryID    int32          `json:"inquiry_id"`
	ChannelUuid  sql.NullString `json:"channel_uuid"`
	MessageCount sql.NullInt32  `json:"message_count"`
	CreatedAt    time.Time      `json:"created_at"`
	ExpiredAt    time.Time      `json:"expired_at"`
	UpdatedAt    sql.NullTime   `json:"updated_at"`
	DeletedAt    sql.NullTime   `json:"deleted_at"`
	ChatroomType ChatroomType   `json:"chatroom_type"`
}

type ChatroomUser struct {
	ID         int64        `json:"id"`
	ChatroomID int32        `json:"chatroom_id"`
	UserID     int32        `json:"user_id"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  sql.NullTime `json:"updated_at"`
	DeletedAt  sql.NullTime `json:"deleted_at"`
}

type CoinOrder struct {
	ID      int32 `json:"id"`
	BuyerID int32 `json:"buyer_id"`
	// cost to buy, currency in TWD
	Cost        string         `json:"cost"`
	OrderStatus OrderStatus    `json:"order_status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
	DeletedAt   sql.NullTime   `json:"deleted_at"`
	PackageID   sql.NullInt32  `json:"package_id"`
	Quantity    int32          `json:"quantity"`
	RecTradeID  sql.NullString `json:"rec_trade_id"`
	Raw         sql.NullString `json:"raw"`
}

type CoinPackage struct {
	ID       int64          `json:"id"`
	DbCoins  sql.NullInt32  `json:"db_coins"`
	Cost     sql.NullString `json:"cost"`
	Currency sql.NullString `json:"currency"`
	Name     sql.NullString `json:"name"`
}

type Image struct {
	ID        int64        `json:"id"`
	UserID    int32        `json:"user_id"`
	Url       string       `json:"url"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
}

type Payment struct {
	ID        int64        `json:"id"`
	PayerID   int32        `json:"payer_id"`
	ServiceID int32        `json:"service_id"`
	Price     string       `json:"price"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
	Refunded  sql.NullBool `json:"refunded"`
}

type Service struct {
	ID                int64          `json:"id"`
	Uuid              sql.NullString `json:"uuid"`
	CustomerID        sql.NullInt32  `json:"customer_id"`
	ServiceProviderID sql.NullInt32  `json:"service_provider_id"`
	Price             sql.NullString `json:"price"`
	Duration          sql.NullInt32  `json:"duration"`
	AppointmentTime   sql.NullTime   `json:"appointment_time"`
	ServiceType       string         `json:"service_type"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         sql.NullTime   `json:"updated_at"`
	DeletedAt         sql.NullTime   `json:"deleted_at"`
	Budget            sql.NullString `json:"budget"`
	InquiryID         int32          `json:"inquiry_id"`
	ServiceStatus     ServiceStatus  `json:"service_status"`
	Address           sql.NullString `json:"address"`
	StartTime         sql.NullTime   `json:"start_time"`
	EndTime           sql.NullTime   `json:"end_time"`
	CancellerID       sql.NullInt32  `json:"canceller_id"`
	// cause states the intention of cancelling a service.
	CancelCause CancelCause    `json:"cancel_cause"`
	MatchingFee sql.NullString `json:"matching_fee"`
	Currency    sql.NullString `json:"currency"`
}

type ServiceInquiry struct {
	ID              int64          `json:"id"`
	InquirerID      sql.NullInt32  `json:"inquirer_id"`
	Budget          string         `json:"budget"`
	InquiryStatus   InquiryStatus  `json:"inquiry_status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       sql.NullTime   `json:"updated_at"`
	DeletedAt       sql.NullTime   `json:"deleted_at"`
	Uuid            string         `json:"uuid"`
	Duration        sql.NullInt32  `json:"duration"`
	AppointmentTime sql.NullTime   `json:"appointment_time"`
	Lng             sql.NullString `json:"lng"`
	Lat             sql.NullString `json:"lat"`
	// Time that this inquiry will be invalid.
	ExpiredAt         sql.NullTime   `json:"expired_at"`
	PickerID          sql.NullInt32  `json:"picker_id"`
	Address           sql.NullString `json:"address"`
	InquiryType       InquiryType    `json:"inquiry_type"`
	ExpectServiceType sql.NullString `json:"expect_service_type"`
	Currency          sql.NullString `json:"currency"`
}

type ServiceOption struct {
	ID                 int64              `json:"id"`
	Name               string             `json:"name"`
	Description        sql.NullString     `json:"description"`
	Price              sql.NullString     `json:"price"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          sql.NullTime       `json:"updated_at"`
	DeletedAt          sql.NullTime       `json:"deleted_at"`
	ServiceOptionsType ServiceOptionsType `json:"service_options_type"`
	Duration           sql.NullInt32      `json:"duration"`
}

type ServiceQrcode struct {
	ID        int64          `json:"id"`
	ServiceID int32          `json:"service_id"`
	Uuid      sql.NullString `json:"uuid"`
	Url       sql.NullString `json:"url"`
}

type ServiceRating struct {
	ID        int64          `json:"id"`
	RaterID   sql.NullInt32  `json:"rater_id"`
	ServiceID sql.NullInt32  `json:"service_id"`
	Rating    sql.NullInt32  `json:"rating"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
	DeletedAt sql.NullTime   `json:"deleted_at"`
	Comments  sql.NullString `json:"comments"`
	RateeID   sql.NullInt32  `json:"ratee_id"`
}

type User struct {
	ID                int64          `json:"id"`
	Username          string         `json:"username"`
	PhoneVerified     bool           `json:"phone_verified"`
	Gender            Gender         `json:"gender"`
	PremiumType       PremiumType    `json:"premium_type"`
	PremiumExpiryDate sql.NullTime   `json:"premium_expiry_date"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         sql.NullTime   `json:"updated_at"`
	DeletedAt         sql.NullTime   `json:"deleted_at"`
	Uuid              string         `json:"uuid"`
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
	FcmTopic          sql.NullString `json:"fcm_topic"`
}

type UserBalance struct {
	ID     int64 `json:"id"`
	UserID int32 `json:"user_id"`
	// use for update when reading the column
	Balance   string       `json:"balance"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"deleted_at"`
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

type UserServiceOption struct {
	ID              int64         `json:"id"`
	UsersID         sql.NullInt32 `json:"users_id"`
	ServiceOptionID sql.NullInt32 `json:"service_option_id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       sql.NullTime  `json:"updated_at"`
	DeletedAt       sql.NullTime  `json:"deleted_at"`
}
