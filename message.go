package bollywood

type Envelope struct {
	Message any
	Engine  *Engine // for the convenience of the actor
	Sender  *Actor  // can be nil
	Target  *Actor  // should never be nil
}
