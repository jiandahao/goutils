package convhttp

import "encoding/json"

// Binder binder
type Binder interface {
	// Name returns the binding engine name
	Name() string
	// Bind binds the passed struct pointer using the specified binding engine.
	Bind(data []byte, obj interface{}) error
}

// JSONBinder json binder
type JSONBinder struct{}

// Name returns the binding engine name
func (b *JSONBinder) Name() string {
	return "json_binder"
}

// Bind binds the passed struct pointer using json binding engine.
func (b *JSONBinder) Bind(data []byte, obj interface{}) error {
	return json.Unmarshal(data, obj)
}
