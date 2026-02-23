package notifier

import "fmt"

// Factory creates a new Notifier instance with default values.
type Factory func() Notifier

// Registry holds all registered notifier plugins.
type Registry struct {
	factories map[string]Factory
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

// Register adds a notifier factory to the registry. It returns an error if a
// plugin with the same name is already registered.
func (r *Registry) Register(name string, f Factory) error {
	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("notifier %q already registered", name)
	}
	r.factories[name] = f
	return nil
}

// All returns a copy of all registered factories.
func (r *Registry) All() map[string]Factory {
	cp := make(map[string]Factory, len(r.factories))
	for k, v := range r.factories {
		cp[k] = v
	}
	return cp
}
