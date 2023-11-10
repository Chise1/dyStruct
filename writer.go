package dyStruct

import (
	"errors"
	"fmt"
	"reflect"
)

// DyStruct is helper interface for writing to a struct.
type DyStruct interface {
	// Set sets the value of the field with the given name.
	Set(value any) error
	Get() (any, error)
	Type() reflect.Type
	ChainSet(chainName []string, value any) error // if field is slice or map, value is nil,it would be delete.
	ChainGet(chainName []string) (any, error)
	ChainType(name []string) (reflect.Type, error)
}

func subWriter(valueOf reflect.Value) (writer DyStruct, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprint(recover()))
		}
	}()
	switch valueOf.Kind() {
	case reflect.Ptr:
		return subWriter(valueOf.Elem())
	case reflect.Int, reflect.Float64, reflect.String, reflect.Bool:
		return newScalarImpl(valueOf)
	case reflect.Struct:
		return newStructImpl(valueOf)
	default:
		return nil, errors.New("not support: " + valueOf.Kind().String())
	}
}

func NewWriter(value any) (writer DyStruct, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprint(recover()))
		}
	}()
	valueOf, ok := value.(reflect.Value)
	if !ok {
		valueOf = reflect.ValueOf(value)
	}
	switch valueOf.Kind() {
	case reflect.Ptr:
		return subWriter(valueOf.Elem())
	default:
		return nil, errors.New("must be use pointer")
	}
}

func UpdateFromJson(writer DyStruct, linkName []string, jsonData []byte) error {
	return UpdateFromJsonWithValidate(writer, linkName, jsonData, nil, nil)
}
func UpdateFromJsonWithValidate(writer DyStruct, linkName []string, jsonData []byte, unmarshalFunc func([]byte, any) error, validate func(s interface{}) error) error {
	typ, err := writer.ChainType(linkName)
	if err != nil { // todo 输出报错
		return fmt.Errorf("can not got type by %v", linkName)
	}
	instance := reflect.New(typ).Interface()
	err = unmarshalFunc(jsonData, &instance)
	if err != nil {
		return err
	}
	if validate != nil {
		err = validate(instance)
		if err != nil {
			return err
		}
	}
	fmt.Println(instance)
	return writer.ChainSet(linkName, reflect.Indirect(reflect.ValueOf(instance)))
}
