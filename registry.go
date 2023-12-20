package bollywood

import "fmt"

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
