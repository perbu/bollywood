package bollywood

type DuplicateActorError struct {
	Id string
}

func (e *DuplicateActorError) Error() string {
	return "duplicate actor: " + e.Id
}

type NoSuchActorError struct {
	Id string
}

func (e *NoSuchActorError) Error() string {
	return "no such actor: " + e.Id
}
