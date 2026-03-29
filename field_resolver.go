package raymond

// FieldResolver is an optional interface that context objects can implement
// to provide custom field lookup. When the template context implements this
// interface, raymond calls GetField instead of using reflect to access map
// keys or struct fields. This enables layered/virtual views that avoid
// copying data into flat maps.
//
// Return (value, true) if the field exists, or (nil, false) if not found.
type FieldResolver interface {
	GetField(name string) (interface{}, bool)
}
