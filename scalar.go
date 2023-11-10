package dyStruct

import (
	"errors"
	"reflect"
)

// int float string
type scalarImpl struct {
	field reflect.Type
	value reflect.Value
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

func (s *scalarImpl) Get() (any, error) {
	return s.value.Interface(), nil
}

// can set struct.substruct field value
func (s *scalarImpl) ChainSet(_ []string, value any) error {
	return s.Set(value)
}

// can set struct.substruct field value
func (s *scalarImpl) ChainGet(_ []string) (any, error) {
	return s.Get()
}

func (s *scalarImpl) Type() reflect.Type {
	return s.field
}
func (s *scalarImpl) ChainType(names []string) (reflect.Type, error) {
	return s.Type(), nil
}
func newScalarImpl(value reflect.Value) (*scalarImpl, error) {
	return &scalarImpl{field: value.Type(), value: value}, nil
}
