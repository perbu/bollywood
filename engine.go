package bollywood

import (
	"fmt"
	"sync"
)

type Engine struct {
	registry *Registry
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

func (e *Engine) Spawn(id string, actor ActorInterface, parent *Actor) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	newActor := &Actor{
		Id:             id,
		Implementation: actor,
		recvCh:         make(chan Envelope),
		Parent:         parent,
		wg:             wg,
	}
	newActor.stopped.Store(false)
	ok := e.registry.register(id, newActor)
	if !ok {
		return &DuplicateActorError{Id: id}
	}
	// spawn the actor
	go func(a *Actor) { // == actor goroutine ==
		// send started message
		actor.Receive(Envelope{
			Message: ActorStarted{},
			Engine:  e,
			Sender:  nil, // system message have no sender
		})
		// receive messages, this blocks until the channel is closed
		for msg := range a.recvCh {
			actor.Receive(msg)
		}
		// actor is done
		actor.Receive(Envelope{
			Message: ActorStopped{},
			Engine:  e,
			Sender:  nil, // system message have no sender
		})
		// unregister the actor
		err := e.registry.unregister(id)
		if err != nil {
			panic(err)
		}
		a.wg.Done() // signal that the actor is done
		// at this point
	}(newActor)
	return nil
}

func (e *Engine) Stop(name string) *sync.WaitGroup {
	a, ok := e.registry.get(name)
	if !ok {
		return &sync.WaitGroup{}
	}
	if !a.stopped.Load() {
		close(a.recvCh)
		a.stopped.Store(true)
	} else {
		fmt.Printf("ERROR(Stop): actor (%s) already stopped\n", a.Id)
	}
	return a.wg
}

func (e *Engine) Send(target string, msg any, sender *Actor) {
	a, ok := e.registry.get(target)
	if !ok {
		e.Send("deadletter", msg, sender)
		return
	}
	a.recvCh <- Envelope{
		Target:  a,
		Message: msg,
		Engine:  e,
		Sender:  sender,
	}
}

func (e *Engine) GetActor(name string) (*Actor, bool) {
	return e.registry.get(name)
}
