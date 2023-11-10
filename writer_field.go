package dynamicstruct

import (
	"errors"
	"reflect"
)

type scalarImpl struct {
	field reflect.StructField
	value reflect.Value
}

func (s *scalarImpl) Val() reflect.Value {
	return reflect.Value{}
}
func (s *scalarImpl) Set(value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	s.value.Set(reflect.ValueOf(value))
	return
}

func (s *scalarImpl) Get() (any, bool) {
	return s.value.Interface(), true
}

// can set struct.substruct field value
func (s *scalarImpl) LinkSet(_ []string, value any) error {
	return s.Set(value)
}

// can set struct.substruct field value
func (s *scalarImpl) LinkGet(_ []string) (any, bool) {
	return s.Get()
}

func (s *scalarImpl) Type() reflect.Type {
	return s.field.Type
}
func (s *scalarImpl) LinkTyp(names []string) (reflect.Type, bool) {
	return s.Type(), true
}
func (s scalarImpl) MarshalJSON() ([]byte, error) {
	panic("not support")
}
func (s *scalarImpl) UnmarshalJSON([]byte) error {
	panic("not support")
}
