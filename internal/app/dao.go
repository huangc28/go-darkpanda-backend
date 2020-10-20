package app

type UserDaoer interface {
	CheckIsMaleByUuid(uuid string) (bool, error)
	CheckIsFemaleByUuid(uuid string) (bool, error)
}
