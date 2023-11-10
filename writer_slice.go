package dynamicstruct

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type sliceImpl struct {
	value      reflect.Value
	field      reflect.StructField
	sliceToMap string //以map形式存储切片
	mapWriters map[any]Writer
}

func (s *sliceImpl) Set(value any) (err error) {
	defer func() {
		er := recover()
		if er != nil {
			err = errors.New(er.(string))
		}
	}()
	switch val := value.(type) {
	case reflect.Value:
		s.value.Set(val)
	default:
		if value == nil {
			x := reflect.MakeSlice(s.value.Type(), 0, 0)
			s.value.Set(x)
		} else {
			s.value.Set(reflect.ValueOf(value))
		}
	}
	s.mapWriters = make(map[any]Writer, s.value.Len())
	return
}

func (s *sliceImpl) Get() (any, bool) {
	return s.value.Interface(), true
}

// can set struct.substruct field value
func (s *sliceImpl) LinkSet(names []string, value any) error {
	if len(names) == 0 {
		return s.Set(value)
	}
	var val reflect.Value
	switch data := value.(type) {
	case reflect.Value:
		val = data
	default:
		val = reflect.ValueOf(value)
	}
	if s.sliceToMap != "" && value != nil { // 校验子对象id和写入id是一致
		if fmt.Sprint(val.FieldByName(s.sliceToMap)) != names[0] {
			return errors.New("id " + fmt.Sprint(val.FieldByName(s.sliceToMap)) + " not same " + names[0])
		}
	}
	atoi, err := s.computeIndex(names[0])
	if err != nil {
		return err
	}
	if len(names) == 1 {
		if value == nil {
			if atoi >= 0 && atoi < s.value.Len() {
				s.value.Set(reflect.AppendSlice(s.value.Slice(0, atoi), s.value.Slice(atoi+1, s.value.Len())))
				delete(s.mapWriters, atoi)
			}
			return nil
		}

		if atoi > s.value.Len() {
			x := reflect.MakeSlice(s.value.Type(), atoi-s.value.Len(), atoi-s.value.Len())
			s.value.Set(reflect.AppendSlice(s.value, x))
			s.value.Set(reflect.Append(s.value, val))
		} else if atoi == s.value.Len() {
			s.value.Set(reflect.Append(s.value, val))
		} else {
			s.value.Index(atoi).Set(val)
		}
		return nil
	}
	if atoi > s.value.Len() {
		x := reflect.MakeSlice(s.value.Type(), atoi-s.value.Len()+1, atoi-s.value.Len()+1)
		s.value.Set(reflect.AppendSlice(s.value, x))
	} else if atoi == s.value.Len() {
		s.value.Set(reflect.Append(s.value, reflect.Zero(s.value.Type().Elem())))
	}

	var writer Writer
	writer, err = subWriter(s.value.Index(atoi))
	if err != nil {
		return err
	}
	s.mapWriters[atoi] = writer
	return writer.LinkSet(names[1:], value)
}

// can set struct.substruct field value
func (s *sliceImpl) LinkGet(names []string) (any, bool) {
	if len(names) == 0 {
		return s.Get()
	}
	atoi, err := s.computeIndex(names[0])
	if err != nil {
		return nil, false
	}
	if atoi >= 0 && atoi < s.value.Len() {
		writer, found := s.mapWriters[atoi]
		if !found {
			writer, err = subWriter(s.value.Index(atoi))
			s.mapWriters[atoi] = writer
		}
		return writer.LinkGet(names[1:])
	}
	return nil, false

}

func (s *sliceImpl) computeIndex(atoiStr string) (int, error) {
	var atoi = -1
	if len(s.sliceToMap) == 0 {
		var err error
		if atoiStr == "*" {
			atoi = s.value.Len()
		} else {
			atoi, err = strconv.Atoi(atoiStr)
			if err != nil {
				return 0, err
			}
		}
	} else {
		var flag bool
		for i := 0; i < s.value.Len(); i++ {
			x := s.value.Index(i)
			if !x.IsZero() {
				sub := fmt.Sprint(x.FieldByName(s.sliceToMap))
				if sub == atoiStr {
					atoi = i
					flag = true
					break
				}
			}
		}
		if !flag {
			atoi = s.value.Len()
		}
	}
	return atoi, nil
}

func (s *sliceImpl) Type() reflect.Type {
	return s.field.Type
}
func (s *sliceImpl) LinkTyp(names []string) (reflect.Type, bool) {
	if len(names) == 0 {
		return s.Type(), true
	}
	ints := reflect.Zero(s.value.Type().Elem())
	writer, err := subWriter(ints)
	if err == nil {
		return writer.LinkTyp(names[1:])
	}
	return nil, false
}
func (s sliceImpl) MarshalJSON() ([]byte, error) {
	panic("not support")
}
func (s *sliceImpl) UnmarshalJSON([]byte) error {
	panic("not support")
}
