package bollywood

import (
	"log/slog"
	"reflect"
)

type DeadLetter struct {
	messages []Envelope
}

func (d *DeadLetter) Receive(msg Envelope) {
	switch msg.Message.(type) {
	case ActorStarted:
	case ActorStopped:
	default:
		var sender string
		switch msg.Sender {
		case nil:
			sender = "<nil>"
		default:
			sender = msg.Sender.Id
		}
		slog.Warn("dead letter", "sender", sender, "target", msg.Target, "type",
			reflect.TypeOf(msg.Message).String())
		d.messages = append(d.messages, msg)
	}
}

func (d *DeadLetter) GetMessages() []Envelope {
	return d.messages
}
