// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/contracts/chat.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	db "github.com/huangc28/go-darkpanda-backend/db"
	contracts "github.com/huangc28/go-darkpanda-backend/internal/app/contracts"
	models "github.com/huangc28/go-darkpanda-backend/internal/app/models"
	sqlx "github.com/jmoiron/sqlx"
	reflect "reflect"
)

// MockChatServicer is a mock of ChatServicer interface
type MockChatServicer struct {
	ctrl     *gomock.Controller
	recorder *MockChatServicerMockRecorder
}

// MockChatServicerMockRecorder is the mock recorder for MockChatServicer
type MockChatServicerMockRecorder struct {
	mock *MockChatServicer
}

// NewMockChatServicer creates a new mock instance
func NewMockChatServicer(ctrl *gomock.Controller) *MockChatServicer {
	mock := &MockChatServicer{ctrl: ctrl}
	mock.recorder = &MockChatServicerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChatServicer) EXPECT() *MockChatServicerMockRecorder {
	return m.recorder
}

// CreateAndJoinChatroom mocks base method
func (m *MockChatServicer) CreateAndJoinChatroom(inquiryID int64, userIDs ...int64) (*models.Chatroom, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{inquiryID}
	for _, a := range userIDs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateAndJoinChatroom", varargs...)
	ret0, _ := ret[0].(*models.Chatroom)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAndJoinChatroom indicates an expected call of CreateAndJoinChatroom
func (mr *MockChatServicerMockRecorder) CreateAndJoinChatroom(inquiryID interface{}, userIDs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{inquiryID}, userIDs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAndJoinChatroom", reflect.TypeOf((*MockChatServicer)(nil).CreateAndJoinChatroom), varargs...)
}

// WithTx mocks base method
func (m *MockChatServicer) WithTx(tx *sqlx.Tx) contracts.ChatServicer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTx", tx)
	ret0, _ := ret[0].(contracts.ChatServicer)
	return ret0
}

// WithTx indicates an expected call of WithTx
func (mr *MockChatServicerMockRecorder) WithTx(tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTx", reflect.TypeOf((*MockChatServicer)(nil).WithTx), tx)
}

// MockChatDaoer is a mock of ChatDaoer interface
type MockChatDaoer struct {
	ctrl     *gomock.Controller
	recorder *MockChatDaoerMockRecorder
}

// MockChatDaoerMockRecorder is the mock recorder for MockChatDaoer
type MockChatDaoerMockRecorder struct {
	mock *MockChatDaoer
}

// NewMockChatDaoer creates a new mock instance
func NewMockChatDaoer(ctrl *gomock.Controller) *MockChatDaoer {
	mock := &MockChatDaoer{ctrl: ctrl}
	mock.recorder = &MockChatDaoerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChatDaoer) EXPECT() *MockChatDaoerMockRecorder {
	return m.recorder
}

// WithTx mocks base method
func (m *MockChatDaoer) WithTx(tx *sqlx.Tx) contracts.ChatDaoer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTx", tx)
	ret0, _ := ret[0].(contracts.ChatDaoer)
	return ret0
}

// WithTx indicates an expected call of WithTx
func (mr *MockChatDaoerMockRecorder) WithTx(tx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTx", reflect.TypeOf((*MockChatDaoer)(nil).WithTx), tx)
}

// WithConn mocks base method
func (m *MockChatDaoer) WithConn(conn db.Conn) contracts.ChatDaoer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithConn", conn)
	ret0, _ := ret[0].(contracts.ChatDaoer)
	return ret0
}

// WithConn indicates an expected call of WithConn
func (mr *MockChatDaoerMockRecorder) WithConn(conn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithConn", reflect.TypeOf((*MockChatDaoer)(nil).WithConn), conn)
}

// CreateChat mocks base method
func (m *MockChatDaoer) CreateChat(inquiryID int64) (*models.Chatroom, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateChat", inquiryID)
	ret0, _ := ret[0].(*models.Chatroom)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateChat indicates an expected call of CreateChat
func (mr *MockChatDaoerMockRecorder) CreateChat(inquiryID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateChat", reflect.TypeOf((*MockChatDaoer)(nil).CreateChat), inquiryID)
}

// JoinChat mocks base method
func (m *MockChatDaoer) JoinChat(chatID int64, userIDs ...int64) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{chatID}
	for _, a := range userIDs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "JoinChat", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// JoinChat indicates an expected call of JoinChat
func (mr *MockChatDaoerMockRecorder) JoinChat(chatID interface{}, userIDs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{chatID}, userIDs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "JoinChat", reflect.TypeOf((*MockChatDaoer)(nil).JoinChat), varargs...)
}

// LeaveChat mocks base method
func (m *MockChatDaoer) LeaveChat(chatID int64, userIDs ...int64) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{chatID}
	for _, a := range userIDs {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LeaveChat", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// LeaveChat indicates an expected call of LeaveChat
func (mr *MockChatDaoerMockRecorder) LeaveChat(chatID interface{}, userIDs ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{chatID}, userIDs...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeaveChat", reflect.TypeOf((*MockChatDaoer)(nil).LeaveChat), varargs...)
}

// LeaveAllMemebers mocks base method
func (m *MockChatDaoer) LeaveAllMemebers(chatroomID int64) ([]models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LeaveAllMemebers", chatroomID)
	ret0, _ := ret[0].([]models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LeaveAllMemebers indicates an expected call of LeaveAllMemebers
func (mr *MockChatDaoerMockRecorder) LeaveAllMemebers(chatroomID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LeaveAllMemebers", reflect.TypeOf((*MockChatDaoer)(nil).LeaveAllMemebers), chatroomID)
}

// GetChatRoomByChannelUUID mocks base method
func (m *MockChatDaoer) GetChatRoomByChannelUUID(chanelUUID string, fields ...string) (*models.Chatroom, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{chanelUUID}
	for _, a := range fields {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetChatRoomByChannelUUID", varargs...)
	ret0, _ := ret[0].(*models.Chatroom)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChatRoomByChannelUUID indicates an expected call of GetChatRoomByChannelUUID
func (mr *MockChatDaoerMockRecorder) GetChatRoomByChannelUUID(chanelUUID interface{}, fields ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{chanelUUID}, fields...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChatRoomByChannelUUID", reflect.TypeOf((*MockChatDaoer)(nil).GetChatRoomByChannelUUID), varargs...)
}

// GetChatRoomByInquiryID mocks base method
func (m *MockChatDaoer) GetChatRoomByInquiryID(inquiryID int64, fields ...string) (*models.Chatroom, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{inquiryID}
	for _, a := range fields {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetChatRoomByInquiryID", varargs...)
	ret0, _ := ret[0].(*models.Chatroom)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChatRoomByInquiryID indicates an expected call of GetChatRoomByInquiryID
func (mr *MockChatDaoerMockRecorder) GetChatRoomByInquiryID(inquiryID interface{}, fields ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{inquiryID}, fields...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChatRoomByInquiryID", reflect.TypeOf((*MockChatDaoer)(nil).GetChatRoomByInquiryID), varargs...)
}

// DeleteChatRoom mocks base method
func (m *MockChatDaoer) DeleteChatRoom(ID int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteChatRoom", ID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteChatRoom indicates an expected call of DeleteChatRoom
func (mr *MockChatDaoerMockRecorder) DeleteChatRoom(ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteChatRoom", reflect.TypeOf((*MockChatDaoer)(nil).DeleteChatRoom), ID)
}

// GetFemaleInquiryChatRooms mocks base method
func (m *MockChatDaoer) GetFemaleInquiryChatRooms(userID, offset, perPage int64) ([]models.InquiryChatRoom, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFemaleInquiryChatRooms", userID, offset, perPage)
	ret0, _ := ret[0].([]models.InquiryChatRoom)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFemaleInquiryChatRooms indicates an expected call of GetFemaleInquiryChatRooms
func (mr *MockChatDaoerMockRecorder) GetFemaleInquiryChatRooms(userID, offset, perPage interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFemaleInquiryChatRooms", reflect.TypeOf((*MockChatDaoer)(nil).GetFemaleInquiryChatRooms), userID, offset, perPage)
}

// UpdateChatByUuid mocks base method
func (m *MockChatDaoer) UpdateChatByUuid(params contracts.UpdateChatByUuidParams) (*models.Chatroom, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateChatByUuid", params)
	ret0, _ := ret[0].(*models.Chatroom)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateChatByUuid indicates an expected call of UpdateChatByUuid
func (mr *MockChatDaoerMockRecorder) UpdateChatByUuid(params interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateChatByUuid", reflect.TypeOf((*MockChatDaoer)(nil).UpdateChatByUuid), params)
}

// IsUserInChatroom mocks base method
func (m *MockChatDaoer) IsUserInChatroom(userUuid, chatroomUuid string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsUserInChatroom", userUuid, chatroomUuid)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsUserInChatroom indicates an expected call of IsUserInChatroom
func (mr *MockChatDaoerMockRecorder) IsUserInChatroom(userUuid, chatroomUuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsUserInChatroom", reflect.TypeOf((*MockChatDaoer)(nil).IsUserInChatroom), userUuid, chatroomUuid)
}

// GetInquiryByChannelUuid mocks base method
func (m *MockChatDaoer) GetInquiryByChannelUuid(channelUuid string) (*models.ServiceInquiry, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInquiryByChannelUuid", channelUuid)
	ret0, _ := ret[0].(*models.ServiceInquiry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInquiryByChannelUuid indicates an expected call of GetInquiryByChannelUuid
func (mr *MockChatDaoerMockRecorder) GetInquiryByChannelUuid(channelUuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInquiryByChannelUuid", reflect.TypeOf((*MockChatDaoer)(nil).GetInquiryByChannelUuid), channelUuid)
}

// GetCompleteChatroomInfoById mocks base method
func (m *MockChatDaoer) GetCompleteChatroomInfoById(id int) (*models.CompleteChatroomInfoModel, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCompleteChatroomInfoById", id)
	ret0, _ := ret[0].(*models.CompleteChatroomInfoModel)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCompleteChatroomInfoById indicates an expected call of GetCompleteChatroomInfoById
func (mr *MockChatDaoerMockRecorder) GetCompleteChatroomInfoById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCompleteChatroomInfoById", reflect.TypeOf((*MockChatDaoer)(nil).GetCompleteChatroomInfoById), id)
}

// GetChatroomByServiceId mocks base method
func (m *MockChatDaoer) GetChatroomByServiceId(srvId int) (*models.Chatroom, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChatroomByServiceId", srvId)
	ret0, _ := ret[0].(*models.Chatroom)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChatroomByServiceId indicates an expected call of GetChatroomByServiceId
func (mr *MockChatDaoerMockRecorder) GetChatroomByServiceId(srvId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChatroomByServiceId", reflect.TypeOf((*MockChatDaoer)(nil).GetChatroomByServiceId), srvId)
}

// DeleteChatroomByServiceId mocks base method
func (m *MockChatDaoer) DeleteChatroomByServiceId(srvId int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteChatroomByServiceId", srvId)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteChatroomByServiceId indicates an expected call of DeleteChatroomByServiceId
func (mr *MockChatDaoerMockRecorder) DeleteChatroomByServiceId(srvId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteChatroomByServiceId", reflect.TypeOf((*MockChatDaoer)(nil).DeleteChatroomByServiceId), srvId)
}

// DeleteChatroomByInquiryId mocks base method
func (m *MockChatDaoer) DeleteChatroomByInquiryId(iqId int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteChatroomByInquiryId", iqId)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteChatroomByInquiryId indicates an expected call of DeleteChatroomByInquiryId
func (mr *MockChatDaoerMockRecorder) DeleteChatroomByInquiryId(iqId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteChatroomByInquiryId", reflect.TypeOf((*MockChatDaoer)(nil).DeleteChatroomByInquiryId), iqId)
}
