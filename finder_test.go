
package jas

import (
	"testing"
	"github.com/coocood/assrt"
	"regexp"
)


func TestFinderString(t *testing.T){
	assert := assrt.NewAssert(t)
	s := `{"a":"","b":null,"c":true,"d":12, "e":"str"}`
	req := NewPostJsonRequest("", "", []byte(s), "e", "E", "o", "O")
	f := FinderWithRequest(req)
	_, err := f.FindString("a")
	assert.Equal(EmptyStringError, err)
	_, err = f.FindString("b")
	assert.Equal(NullValueError, err)
	_, err = f.FindString("c")
	assert.Equal(WrongTypeError, err)
	_, err = f.FindString("d")
	assert.Equal(WrongTypeError, err)
	e, err := f.FindString("e")
	assert.Nil(err)
	assert.Equal("E", e)
	_, err = f.FindString("f")
	assert.Equal(EntryNotExistsError, err)
	o, err := f.FindString("o")
	assert.Equal("O", o)
}


func TestFinderInt(t *testing.T){
	assert := assrt.NewAssert(t)
	req := NewGetRequest("", "", "a", 1, "b", 2)
	f := FinderWithRequest(req)
	a := f.RequireInt("a")
	assert.Equal(1,a)
	b := f.RequireInt("b")
	assert.Equal(2,b)
	jsonData := []byte(`{"chars":["a","b", "c"], "obj":{"x": 100}}`)
	f = FinderWithBytes(jsonData)
	assert.Equal("b", f.RequireString("chars", 1))
	assert.Equal(100, f.RequireInt("obj","x"))
}

func TestFinderStringLen(t *testing.T){
	assert := assrt.NewAssert(t)
	req := NewGetRequest("", "", "a", "1234567", "b", "语言文字")
	f := FinderWithRequest(req)
	_, err := f.FindStringLen(8,10, "a")
	assert.NotNil(err)
	_, err = f.FindStringLen(3, 7, "a")
	assert.NotNil(err)
	_, err = f.FindStringLen(7, 10, "a")
	assert.Nil(err)
	_, err = f.FindStringRuneLen(5, 8, "b")
	assert.NotNil(err)
	_, err = f.FindStringRuneLen(1, 4, "b")
	assert.NotNil(err)
	_, err = f.FindStringRuneLen(2, 6, "b")
	assert.Nil(err)
}

func TestFinderRegexp(t *testing.T) {
	assert := assrt.NewAssert(t)
	req := NewGetRequest("","", "a", "abcderg", "b", "语言文字")
	f := FinderWithRequest(req)
	_, err := f.FindStringMatch(regexp.MustCompile("\\w+"), "a")
	assert.Nil(err)
	_, err = f.FindStringMatch(regexp.MustCompile("\\d+"), "a")
	assert.NotNil(err)
}
