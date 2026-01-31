package mcp

// Registry holds available context providers.
// Allows dynamic registration for future extensibility.
type Registry struct {
	providers map[string]ContextProvider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ContextProvider),
	}
}

func (r *Registry) Register(p ContextProvider) {
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (ContextProvider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) All() []ContextProvider {
	result := make([]ContextProvider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}
