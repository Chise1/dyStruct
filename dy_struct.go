package dyStruct

import "reflect"

var SplitSep = "." // split fields

// DyStruct is helper interface for writing to a struct.
type DyStruct interface {
	// Set sets the value of the field with the given name.
	Set(value any) error
	Get() (any, error)
	ChainSet(chainName string, value any) error // if field is slice or map, value is nil,it would be delete.
	ChainGet(chainName string) (any, error)
	chainSet(names []string, value any) error
	chainGet(names []string) (any, error)
	// Type use to create new instance
	Type() reflect.Type
	chainType(chainName []string) (reflect.Type, error)
	ChainType(name string) (reflect.Type, error)
}
