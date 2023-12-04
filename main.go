package bollywood

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type actorMap map[string]*Actor
type Registry struct {
	actors actorMap
	mu     sync.RWMutex
}

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

type Engine struct {
	registry *Registry
}

type Message struct {
	Message any
	Engine  *Engine
	Sender  *Actor
}

// ActorInterface is the interface that must be implemented by all actors.
type ActorInterface interface {
	Receive(message Message)
}

func NewEngine() *Engine {
	e := &Engine{
		registry: &Registry{
			actors: make(actorMap),
		},
	}
	err := e.Spawn("deadletter", &DeadLetter{}, nil)
	if err != nil {
		panic("could not spawn deadletter actor")
	}
	return e
}

func (r *Registry) register(name string, a *Actor) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.actors[name]; ok {
		return false
	}
	r.actors[name] = a
	return true
}

func (r *Registry) unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.actors[name]; !ok {
		return fmt.Errorf("actor %s does not exist", name)
	}
	delete(r.actors, name)
	return nil
}

func (r *Registry) get(name string) (*Actor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.actors[name]; !ok {
		return nil, false
	}
	return r.actors[name], true
}

// getAll returns a slice of all actors except the deadletter actor
func (r *Registry) getAll() []*Actor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	actors := make([]*Actor, 0, len(r.actors))
	for _, a := range r.actors {
		if a.Id != "deadletter" {
			actors = append(actors, a)
		}
	}
	return actors
}

func (e *Engine) Spawn(id string, actor ActorInterface, parent *Actor) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	newActor := &Actor{
		Id:             id,
		Implementation: actor,
		recvCh:         make(chan Message),
		Parent:         parent,
		wg:             wg,
	}
	newActor.stopped.Store(false)
	ok := e.registry.register(id, newActor)
	if !ok {
		return fmt.Errorf("an actor with that ID already exists")
	}
	// spawn the actor
	go func(a *Actor) { // == actor goroutine ==
		// send started message
		actor.Receive(Message{
			Message: ActorStarted{},
			Engine:  e,
			Sender:  nil, // system message have no sender
		})
		// receive messages, this blocks until the channel is closed
		for msg := range a.recvCh {
			actor.Receive(msg)
		}
		// actor is done
		actor.Receive(Message{
			Message: ActorStopped{},
			Engine:  e,
			Sender:  nil, // system message have no sender
		})
		// unregister the actor
		err := e.registry.unregister(id)
		if err != nil {
			fmt.Printf("ERROR: unregister %s: %s", id, err)
		}
		a.wg.Done() // signal that the actor is done
		// at this point
	}(newActor)
	return nil
}

func (e *Engine) StopByName(name string) *sync.WaitGroup {
	a, ok := e.registry.get(name)
	if !ok {
		fmt.Printf("ERROR(StopByName): actor %s not found\n", name)
		return &sync.WaitGroup{}
	}
	return e.Stop(a)
}

func (e *Engine) Stop(a *Actor) *sync.WaitGroup {
	if !a.stopped.Load() {
		close(a.recvCh)
		a.stopped.Store(true)
	} else {
		fmt.Printf("ERROR(Stop): actor (%s) already stopped\n", a.Id)
	}
	return a.wg
}

func (e *Engine) SendByName(target string, msg any, sender *Actor) {
	a, ok := e.registry.get(target)
	if !ok {
		e.SendByName("deadletter", msg, sender)
		return
	}
	a.recvCh <- Message{
		Message: msg,
		Engine:  e,
		Sender:  sender,
	}
}
func (e *Engine) Send(target *Actor, msg any, sender *Actor) {
	target.recvCh <- Message{
		Message: msg,
		Engine:  e,
		Sender:  sender,
	}
	return
}

func (e *Engine) GetActor(name string) (*Actor, bool) {
	return e.registry.get(name)
}

func (e *Engine) Shutdown() {
	// stop all actors
	for _, a := range e.registry.getAll() {
		e.Stop(a).Wait()
	}
	// stop the deadletter actor
	e.StopByName("deadletter").Wait()
}

type DeadLetter struct {
	messages []Message
}

func (d *DeadLetter) Receive(msg Message) {
	switch msg.Message.(type) {
	case ActorStarted:
	case ActorStopped:
	default:
		fmt.Printf("WARNING: dead letter: %v\n", msg.Message)
		d.messages = append(d.messages, msg)
	}
}

func (d *DeadLetter) GetMessages() []Message {
	return d.messages
}
