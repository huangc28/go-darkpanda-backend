package contracts

type ServiceActions string

type ServiceFSMer interface {
	Current() string
	Is(state string) bool
	Event(event string, args ...interface{}) error
	SetState(state string)
}
