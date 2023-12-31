package bollywood

import (
	"sync"
	"sync/atomic"
)

// ActorInterface is the interface that must be implemented by all actors.
type ActorInterface interface {
	Receive(message Envelope)
}

type Actor struct {
	Id             string
	Implementation ActorInterface
	Parent         *Actor
	recvCh         chan Envelope
	stopped        atomic.Bool
	wg             *sync.WaitGroup // used to wait for actor to stop
}

type ActorStarted struct{}
type ActorStopped struct{}

type Receiver func(any)
