package bollywood

type Message struct {
	Payload any
	Engine  *Engine
	Sender  *Actor
	Target  *Actor
}
