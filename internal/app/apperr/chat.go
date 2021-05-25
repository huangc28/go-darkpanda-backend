package apperr

const (
	FailedToValidateEmitTextMessageParams                 = "7000001"
	FailedToSendTextMessage                               = "7000002"
	FailedToGetChatRoomByChannelUuid                      = "7000003"
	MessageExceedMaximumCount                             = "7000004"
	ChatRoomHasExpired                                    = "7000005"
	FailedToGetChatRoomByInquiryID                        = "7000006"
	FailedToLeaveChat                                     = "7000007"
	FailedToDeleteChat                                    = "7000008"
	FailedToLeaveAllMembers                               = "7000009"
	FailedToCreatePrivateChatRoom                         = "7000010"
	FailedToGetFemaleChatRooms                            = "7000011"
	FailedToGetMessageFromFireStore                       = "7000012"
	FailedGetHistoricalMessagesFromFireStore              = "7000013"
	FailedToValidateGetChatroomsParams                    = "7000014"
	FailedToValidateEmitServiceSettingMessageParams       = "7000015"
	FailedToGetServiceByInquiryUUID                       = "7000016"
	FailedToValidateEmitConfirmedServiceParams            = "7000017"
	FailedToChangeStatusWhenEmittingServiceConfirmMessage = "7000018"
	FailedToUpdateChatroom                                = "7000019"
	FailedToSendUpdateInquiryMessage                      = "7000020"
	UserIsNotInTheChatroom                                = "7000021"
	FailedToGetInquiryByChannelUuid                       = "7000022"
	FailedToCheckIsUserInChatroom                         = "7000023"

	FailedToUploadQRCode              = "7000024"
	FailedToGenQRCodeUuid             = "7000025"
	FailedToMarshQRCodeContent        = "7000026"
	FailedToCreateServiceQRCodeRecord = "7000027"

	ServiceStartTimeNotValid = "7000028"
)

var ChatErrorMessageMap = map[string]string{
	MessageExceedMaximumCount:       "Exceed maximum message count. chatroom is closed",
	ChatRoomHasExpired:              "Chatroom has expired, please create another inquiry to proceed chatroom",
	UserIsNotInTheChatroom:          "User is not in the chatroom",
	FailedToGetInquiryByChannelUuid: "Chatroom is not in relate to any inquiry",
}
