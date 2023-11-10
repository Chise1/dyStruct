package dyStruct

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScalarImpl(t *testing.T) {
	var s = 0
	writer, err := NewWriter(&s)
	assert.Equal(t, nil, err)
	err = writer.Set(10)
	assert.Equal(t, nil, err)
	val, err := writer.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, val, 10)
	type selfInt int
	var s2 selfInt = 0
	writer, err = NewWriter(&s2)
	assert.Equal(t, nil, err)
	var subIntx selfInt = 10
	err = writer.Set(subIntx)
	assert.Equal(t, nil, err)
	val, err = writer.Get()
	assert.Equal(t, nil, err)
	assert.Equal(t, true, val.(selfInt) == 10)
}
func TestSampleStruct(t *testing.T) {
	var intP = 1
	var obj = NewStruct().AddField("IntP", &intP, ``).
		AddField("Int", 0, ``).Build()
	instance := obj.New()
	writer, err := NewWriter(instance)
	assert.Equal(t, nil, err)
	err = writer.ChainSet([]string{"IntP"}, 10)
	assert.Equal(t, nil, err)
	data := `{"IntP":20,"Int":20}`
	err = UpdateFromJson(writer, nil, []byte(data))
	assert.Equal(t, nil, err)
	writerPrintString(writer)
}
func writerPrintString(w DyStruct) {
	get, err := w.Get()
	if err != nil {
		panic(err)
	}
	marshal, err := json.Marshal(get)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshal))
}
