package bollywood

import (
	"fmt"
	"sync"
)

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
			Payload: ActorStarted{},
			Engine:  e,
			Sender:  nil, // system message have no sender
		})
		// receive messages, this blocks until the channel is closed
		for msg := range a.recvCh {
			actor.Receive(msg)
		}
		// actor is done
		actor.Receive(Message{
			Payload: ActorStopped{},
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
		Payload: msg,
		Engine:  e,
		Sender:  sender,
	}
}
func (e *Engine) Send(target *Actor, msg any, sender *Actor) {
	target.recvCh <- Message{
		Payload: msg,
		Engine:  e,
		Sender:  sender,
		Target:  target,
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
