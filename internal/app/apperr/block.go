package apperr

var (
	FailedToUnblockNotBlockedUser = "2300001"
	FailedToCheckHasBlockedUser   = "2300002"
	FailedToUnblockUser           = "2300003"
	FailedToBlockUser             = "2300004"
	UnableToFindBlockee           = "2300005"
)

var blockErrorCodeMsgMap = map[string]string{
	FailedToUnblockNotBlockedUser: "failed to unblock user that has not been blocked by you",
	UnableToFindBlockee:           "the person to block is not found",
}
