package bollywood

import (
	"sync"
)

type actorMap map[string]*Actor
type Registry struct {
	actors actorMap
	mu     sync.RWMutex
}

type Engine struct {
	registry *Registry
}

type Message struct {
	Payload any
	Engine  *Engine
	Sender  *Actor
	Target  *Actor
}

// ActorInterface is the interface that must be implemented by all actors.
type ActorInterface interface {
	Receive(message Message)
}
