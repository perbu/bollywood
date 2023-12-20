package bollywood_test

import (
	"fmt"
	"github.com/perbu/bollywood"
	"sync"
	"testing"
	"time"
)

func TestEngine_Send(t *testing.T) {
	e := bollywood.NewEngine()
	err := e.Spawn("baker", &baker{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// grab the newly spawned actor
	a, ok := e.GetActor("baker")
	if !ok {
		t.Fatal("actor 'baker' not found")
	}
	// cast the implementation to the baker struct so, we can access the done waitgroup and the data
	b := a.Implementation.(*baker)
	e.Send("baker", bakeBread{}, nil)
	e.Send("baker", bakeBread{}, nil)
	e.Send("baker", bakeCake{}, nil)
	e.Stop("baker").Wait()
	if err != nil {
		t.Fatal(err)
	}
	b.done.Wait()
	fmt.Println("send ok")

	// check the results. We should have 2 breads and 1 cake
	if b.noOfBread != 2 {
		t.Fatal("wrong number of breads, expected 2 got", b.noOfBread)
	}
	if b.noOfCakes != 1 {
		t.Fatal("wrong number of cakes, expected 1 got", b.noOfCakes)
	}
}

func TestEngine_Deadletter(t *testing.T) {
	e := bollywood.NewEngine()
	e.Send("baker", bakeBread{}, nil)
	time.Sleep(time.Millisecond)
	dl, ok := e.GetActor("deadletter")
	if !ok {
		t.Fatal("deadletter actor not found")
	}
	// cast the implementation to the deadletter struct so, we can access the data
	d := dl.Implementation.(*bollywood.DeadLetter)
	if len(d.GetMessages()) != 1 {
		t.Fatal("wrong number of messages in deadletter, expected 1 got", len(d.GetMessages()))
	}

}

func BenchmarkEngine_Send(b *testing.B) {
	e := bollywood.NewEngine()
	err := e.Spawn("assistant", &assistant{}, nil)
	if err != nil {
		b.Fatal(err)
	}
	a, ok := e.GetActor("assistant")
	if !ok {
		b.Fatal("actor assistant not found")
	}
	b.ResetTimer()
	// lets send it a lot of messages
	for i := 0; i < b.N; i++ {
		e.Send("assistant", helpBake{"cake"}, a)
	}

}

type baker struct {
	noOfCakes int
	noOfBread int
	done      sync.WaitGroup
}

type bakeBread struct{}
type bakeCake struct{}

func (b *baker) Receive(msg bollywood.Envelope) {
	switch msg.Message.(type) {
	case bollywood.ActorStarted:
		fmt.Println("ActorStarted baker, spawning assistant")
		err := msg.Engine.Spawn("assistant", &assistant{}, msg.Sender)
		if err != nil {
			panic(err)
		}
		b.done.Add(1)
	case bollywood.ActorStopped:
		fmt.Println("ActorStopped baker, stopping assistant")
		msg.Engine.Stop("assistant").Wait()
		defer b.done.Done()
		break
	case bakeCake:
		msg.Engine.Send("assistant", helpBake{"cake"}, msg.Sender)
		b.noOfCakes++
		fmt.Println("baked a cake, we now have ", b.noOfCakes, " cakes")
	case bakeBread:
		msg.Engine.Send("assistant", helpBake{"bread"}, msg.Sender)
		b.noOfBread++
		fmt.Println("baked a bread, we now have ", b.noOfBread, " breads")
	}
}

type helpBake struct {
	what string
}

type assistant struct {
	helps int
}

func (a *assistant) Receive(msg bollywood.Envelope) {
	switch msg.Message.(type) {
	case bollywood.ActorStarted:
	case bollywood.ActorStopped:
		break
	case helpBake:
		a.helps++
	}
}
