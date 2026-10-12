package pilot

// Registry maps registered pilot IDs to pilot implementations.
type Registry struct {
	byID map[ID]Pilot
}

// NewRegistry returns a registry with the Manual pilot registered as
// the default.
func NewRegistry() *Registry {
	r := &Registry{byID: map[ID]Pilot{}}
	r.Register(ManualID, Manual())
	return r
}

// Register makes p available under id.
func (r *Registry) Register(id ID, p Pilot) { r.byID[id] = p }

// Lookup returns the pilot registered under id.
func (r *Registry) Lookup(id ID) (Pilot, bool) {
	p, ok := r.byID[id]
	return p, ok
}
