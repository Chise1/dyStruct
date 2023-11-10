package dyStruct

import (
	"errors"
	"fmt"
	"reflect"
)

type structImpl struct {
	fields     map[string]DyStruct
	fieldsName map[string]struct{}
	value      reflect.Value
	field      reflect.Type
}

func (s *structImpl) Set(value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	val := s.value
	for {
		if val.Kind() != reflect.Ptr {
			break
		}
		val = val.Elem()
	}
	switch valT := value.(type) {
	case reflect.Value:
		val.Set(reflect.Indirect(valT))
	default:
		val.Set(reflect.Indirect(reflect.ValueOf(value)))
	}
	return
}

func (s *structImpl) Get() (any, error) {
	return s.value.Interface(), nil
}

// can set struct.substruct field value
func (s *structImpl) ChainSet(names []string, value any) error {
	if len(names) == 0 {
		return s.Set(value)
	}
	name := names[0]

	field, err := s.getFieldImpl(name)
	if err != nil {
		return errors.New("got field error" + err.Error()) // todo 优化
	}
	return field.ChainSet(names[1:], value)
}

// can set struct.substruct field value
func (s *structImpl) ChainGet(names []string) (any, error) {
	if len(names) == 0 {
		return s.Get()
	}
	name := names[0]

	field, err := s.getFieldImpl(name)
	if err != nil {
		return nil, errors.New("got field error" + err.Error()) // todo 优化
	}
	return field.ChainGet(names[1:])
}

func (s *structImpl) Type() reflect.Type {
	return s.field
}
func (s *structImpl) ChainType(names []string) (reflect.Type, error) {
	if len(names) == 0 {
		return s.Type(), nil
	}
	name := names[0]

	field, err := s.getFieldImpl(name)
	if err != nil {
		return nil, errors.New("got field type error" + err.Error()) // todo 优化
	}
	return field.ChainType(names[1:])
}
func newStructImpl(valueOf reflect.Value) (*structImpl, error) {
	typeOf := valueOf.Type()
	fields := make(map[string]struct{}, valueOf.NumField()) // fixme 延迟生成
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)
		fields[field.Name] = struct{}{}
	}

	return &structImpl{
		fields:     make(map[string]DyStruct, valueOf.NumField()),
		fieldsName: fields,
		value:      valueOf,
		field:      valueOf.Type(),
	}, nil
}
func (s *structImpl) getFieldImpl(fieldName string) (DyStruct, error) {
	fieldImpl, found := s.fields[fieldName]
	if found {
		return fieldImpl, nil
	}
	_, found = s.fieldsName[fieldName]
	if !found {
		return nil, errors.New("do not have field " + fieldName)
	}
	typeOf := s.value.Type()
	var err error
	field, ok := typeOf.FieldByName(fieldName)
	if !ok {
		return nil, errors.New("can not found field:" + fieldName)
	}
	valueOf := s.value.FieldByName(fieldName)

	var impl DyStruct
	if field.Type.Kind() == reflect.Struct {
		impl, err = subWriter(valueOf)
		if err != nil {
			return nil, err
		}
	} else if field.Type.Kind() == reflect.Pointer { // todo暂时不要支持指针？
		fmt.Println(valueOf.Kind())
		impl, err = subWriter(valueOf.Elem())
		if err != nil {
			panic("todo")
		}
	} else if field.Type.Kind() == reflect.Map { // todo暂时不要支持指针？
		impl, err = newMapImpl(valueOf)
		if err != nil {
			panic("todo")
		}
	} else if field.Type.Kind() == reflect.Slice {
		impl, err = newSliceImpl(valueOf)
		if err != nil {
			panic("todo")
		}
	} else {
		impl, err = newScalarImpl(valueOf)
		if err != nil {
			panic("todo")
		}
	}
	s.fields[field.Name] = impl
	return impl, nil
}
