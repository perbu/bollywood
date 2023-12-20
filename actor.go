package bollywood

import (
	"sync"
	"sync/atomic"
)

type Actor struct {
	Id             string
	Implementation ActorInterface
	Parent         *Actor
	recvCh         chan Message
	stopped        atomic.Bool
	wg             *sync.WaitGroup // used to wait for actor to stop
}

type ActorStarted struct{}
type ActorStopped struct{}

type Receiver func(any)
