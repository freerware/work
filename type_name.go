package work

import "fmt"

// TypeName represents an entity's type.
type TypeName string

// TypeNameOf provides the type name for the provided entity.
func TypeNameOf(entity interface{}) TypeName {
	return TypeName(fmt.Sprintf("%T", entity))
}

// String provides the string representation of the type name.
func (t TypeName) String() string {
	return string(t)
}
