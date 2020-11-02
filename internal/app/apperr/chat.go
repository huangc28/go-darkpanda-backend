package apperr

const (
	FailedToValidateEmitTextMessageParams = "7000001"
	FailedToSendTextMessage               = "7000002"
	FailedToGetChatRoomByChannelUuid      = "7000003"
	MessageExceedMaximumCount             = "7000004"
	ChatRoomHasExpired                    = "7000005"
	FailedToGetChatRoomByInquiryID        = "7000006"
	FailedToLeaveChat                     = "7000007"
	FailedToDeleteChat                    = "7000008"
	FailedToLeaveAllMembers               = "7000009"
	FailedToCreatePrivateChatRoom         = "7000010"
	FailedToGetFemaleChatRooms            = "7000011"
)

var ChatErrorMessageMap = map[string]string{
	MessageExceedMaximumCount: "Exceed maximum message count. chatroom is closed",
	ChatRoomHasExpired:        "Chatroom has expired, please create another inquiry to proceed chatroom",
}
