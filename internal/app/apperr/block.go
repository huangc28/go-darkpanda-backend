package apperr

var (
	FailedToUnblockNotBlockedUser = "2300001"
	FailedToCheckHasBlockedUser   = "2300002"
	FailedToUnblockUser           = "2300003"
)

var blockErrorCodeMsgMap = map[string]string{
	FailedToUnblockNotBlockedUser: "failed to unblock user that has not been blocked by you.",
}
