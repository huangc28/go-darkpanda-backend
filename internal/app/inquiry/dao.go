package inquiry

type UserDaoer interface {
	CheckIsMaleByUuid(uuid string) (bool, error)
	CheckIsFeMaleByUuid(uuid string) (bool, error)
}
