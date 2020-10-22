package chat

import "time"

const MaxMassageCount = 200

func IsExceedMaxMessageCount(count int) bool {
	return count+1 > MaxMassageCount
}

func IsChatroomExpired(expT time.Time) bool {
	return expT.Before(time.Now())
}
