package bollywood

import (
	"log/slog"
	"reflect"
)

type DeadLetter struct {
	messages []Message
}

func (d *DeadLetter) Receive(msg Message) {
	switch msg.Payload.(type) {
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
			reflect.TypeOf(msg.Payload).String())
		d.messages = append(d.messages, msg)
	}
}

func (d *DeadLetter) GetMessages() []Message {
	return d.messages
}
