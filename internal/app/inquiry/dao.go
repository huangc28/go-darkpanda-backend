package inquiry

type UserDaoer interface {
	CheckIsMaleByUuid(uuid string) (bool, error)
}
