package dynamicstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Writer is helper interface for writing to a struct.
type Writer interface {
	// Set sets the value of the field with the given name.
	Set(value any) error
	Get() (any, bool)
	LinkSet(names []string, value any) error // slice map 删除则设置为value nil
	LinkGet(names []string) (any, bool)
	Type() reflect.Type
	LinkTyp(names []string) (reflect.Type, bool)
	json.Marshaler
	json.Unmarshaler
}

// todo add support slice map
func subWriter(value any) (writer Writer, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = errors.New(fmt.Sprint(recover()))
		}
	}()
	fields := make(map[string]Writer)
	fieldNames := make(map[string]struct{})
	valueOf, ok := value.(reflect.Value)
	if !ok {
		valueOf = reflect.ValueOf(value)
	}
	for {
		if valueOf.Kind() != reflect.Ptr && valueOf.Kind() != reflect.Interface {
			break
		}
		valueOf = valueOf.Elem()
	}
	typeOf := valueOf.Type()
	if typeOf.Kind() != reflect.Struct {
		return nil, errors.New("value must be struct ptr")
	}
	for i := 0; i < valueOf.NumField(); i++ {
		field := typeOf.Field(i)
		var impl Writer
		if field.Type.Kind() == reflect.Struct {
			impl, err = subWriter(valueOf.Field(i))
			if err != nil {
				return nil, err
			}
		} else if field.Type.Kind() == reflect.Pointer { // todo暂时不要支持指针？
			elem := field.Type.Elem()
			if elem.Kind() == reflect.Map {
				impl = &mapImpl{
					field:      field,
					mapWriters: make(map[any]Writer),
					value:      valueOf.Field(i),
				}
			} else {
				return nil, errors.New("not suport pointer")
			}

		} else if field.Type.Kind() == reflect.Map { // todo暂时不要支持指针？
			valueOf.Field(i).Set(reflect.MakeMap(valueOf.Field(i).Type()))
			impl = &mapImpl{
				field:      field,
				mapWriters: make(map[any]Writer),
				value:      valueOf.Field(i),
			}
		} else if field.Type.Kind() == reflect.Slice {
			slice := &sliceImpl{
				field:      field,
				value:      valueOf.Field(i),
				mapWriters: make(map[any]Writer),
			}
			tagStr := field.Tag.Get("dynamic")
			if len(tagStr) > 0 {
				tags := strings.Split(tagStr, ",")
				for _, tag := range tags {
					infos := strings.Split(tag, "=")
					if len(infos) == 2 {
						k, v := infos[0], infos[1]
						if k == "sliceKey" {
							if ToTitle {
								v = strings.ToTitle(v)
							}
							slice.sliceToMap = v
						}
					} else {
						log.Printf("got error tag:%s", tagStr)
					}
				}
			}
			impl = slice
		} else {
			impl = &scalarImpl{
				field: field,
				value: valueOf.Field(i),
			}
		}
		fieldNames[field.Name] = struct{}{}
		fields[field.Name] = impl
	}

	return &structImpl{
		fieldNames: fieldNames,
		fields:     fields,
		value:      valueOf,
		fieldType:  typeOf,
	}, nil
}
func NewWriter(value any) (writer Writer, err error) {
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() != reflect.Ptr {
		return nil, errors.New("value must be ptr")
	}
	var typeOf reflect.Type
	for {
		typeOf = valueOf.Type()
		if typeOf.Kind() == reflect.Struct {
			break
		}
		valueOf = valueOf.Elem()
	}
	ret, err := subWriter(value)
	return ret, err
}

func UpdateFromJson(writer Writer, linkName []string, jsonData []byte, unmarshalFunc func([]byte, any) error, validate func(s interface{}) error) error {
	typ, ok := writer.LinkTyp(linkName)
	if !ok {
		return fmt.Errorf("can not got type by %v", linkName)
	}
	instance := reflect.New(typ).Interface()
	err := unmarshalFunc(jsonData, &instance)
	if err != nil {
		return err
	}
	if validate != nil {
		if typ.Kind() == reflect.Struct {
			err = validate(instance)
			if err != nil {
				return err
			}
		} else if typ.Kind() == reflect.Slice {
			obj := reflect.ValueOf(instance).Elem()
			var errs []error
			for i := 0; i < obj.Len(); i++ {
				errs = append(errs, validate(obj.Index(i)))
			}
			if len(errs) > 0 {
				var errString []string
				for _, err := range errs {
					errString = append(errString, err.Error())
				}
				return errors.New(strings.Join(errString, ","))
			}
		}
	}
	return writer.LinkSet(linkName, reflect.Indirect(reflect.ValueOf(instance)))
}
