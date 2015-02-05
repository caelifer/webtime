package service

// Service - generic service interface. If service relies on by-directional communication,
// returned value from the channel should include a "reply" channel
type Service interface {
	Service(quit <-chan bool) <-chan interface{}
}
