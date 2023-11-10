package dynamicstruct

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
)

type structImpl struct {
	fieldNames map[string]struct{}
	fields     map[string]Writer
	value      reflect.Value
	fieldType  reflect.Type
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
	s.fields = make(map[string]Writer, val.NumField())
	return
}

func (s *structImpl) Get() (any, bool) {
	return s.value.Interface(), true
}

// can set struct.substruct field value
func (s *structImpl) LinkSet(names []string, value any) error {
	if len(names) == 0 {
		return s.Set(value)
	}
	name := names[0]
	if ToTitle {
		name = UpperFirstAlphabet(name)
	}
	field, ok := s.getField(name)
	if !ok {
		return errors.New("not found field " + name)
	}
	return field.LinkSet(names[1:], value)
}

// can set struct.substruct field value
func (s *structImpl) LinkGet(names []string) (any, bool) {
	if len(names) == 0 {
		return s.Get()
	}
	name := names[0]
	if ToTitle {
		name = UpperFirstAlphabet(name)
	}
	field, ok := s.getField(name)
	if !ok {
		return nil, false
	}
	return field.LinkGet(names[1:])
}

func (s *structImpl) Type() reflect.Type {
	return s.fieldType
}
func (s *structImpl) LinkTyp(names []string) (reflect.Type, bool) {
	if len(names) == 0 {
		return s.Type(), true
	}
	name := names[0]
	if ToTitle {
		name = UpperFirstAlphabet(name)
	}
	field, ok := s.getField(name)
	if !ok {
		return nil, false
	}
	return field.LinkTyp(names[1:])
}

func (s *structImpl) getField(fieldName string) (Writer, bool) {
	writer, found := s.fields[fieldName]
	if found {
		return writer, true
	}
	_, found = s.fieldNames[fieldName]
	if found {
		var err error
		field, _ := s.fieldType.FieldByName(fieldName)
		val := s.value.FieldByName(fieldName)
		var impl Writer
		if field.Type.Kind() == reflect.Struct {
			impl, err = subWriter(val)
			if err != nil {
				log.Printf("get Field Writer error: " + err.Error())
				return nil, false
			}
		} else if field.Type.Kind() == reflect.Pointer { // todo暂时不要支持指针？
			elem := field.Type.Elem()
			if elem.Kind() == reflect.Map {
				impl = &mapImpl{
					field:      field,
					mapWriters: make(map[any]Writer),
					value:      val,
				}
			} else {
				log.Printf("get fieldType writer error: not suport pointer ")
				return nil, false
			}
		} else if field.Type.Kind() == reflect.Map { // todo暂时不要支持指针？
			val.Set(reflect.MakeMap(val.Type()))
			impl = &mapImpl{
				field:      field,
				mapWriters: make(map[any]Writer),
				value:      val,
			}
		} else if field.Type.Kind() == reflect.Slice {
			slice := &sliceImpl{
				field:      field,
				value:      val,
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
				value: val,
			}
		}
		s.fields[field.Name] = impl
		return impl, true
	}
	return nil, false
}
func (s structImpl) MarshalJSON() ([]byte, error) {
	obj, _ := s.Get()
	return json.Marshal(obj)
}
func (s *structImpl) UnmarshalJSON(data []byte) error {
	typ := s.Type()
	instance := reflect.New(typ).Interface()
	err := json.Unmarshal(data, &instance)
	if err != nil {
		return err
	}
	return s.Set(instance)
}
