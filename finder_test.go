package jas

import (
	"net/http/httptest"
	"regexp"
	"testing"
)

const (
	jsonData = `{"a":"","b":null,"c":true,"d":12,"e":"str"}`
)

type JsonModel struct {
	A string      `json:"a"`
	B interface{} `json:"b"`
	C bool        `json:"c"`
	D int         `json:"d"`
	E string      `json:"e"`
}

type UnmarshalRes struct{}

func (*UnmarshalRes) Post(ctx *Context) {
	ctx.Data, _ = ctx.FindMap()
}

func (*UnmarshalRes) PostUnmarshal(ctx *Context) {
	jm := JsonModel{}
	err := ctx.Unmarshal(&jm)
	if err == nil {
		ctx.Data = jm
	} else {
		ctx.Error = NewRequestError("invalid json body")
	}
}

func TestUnmarshal(t *testing.T) {
	req := NewPostJsonRequest("", "/unmarshal_res", []byte(jsonData))
	router := NewRouter(new(UnmarshalRes))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert := NewAssert(t)
	assert.Equal(`{"data":{"a":"","b":null,"c":true,"d":12,"e":"str"},"error":null}`, recorder.Body.String())

	req = NewPostJsonRequest("", "/unmarshal_res/unmarshal", []byte(jsonData))
	router.DisableAutoUnmarshal = true
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(`{"data":{"a":"","b":null,"c":true,"d":12,"e":"str"},"error":null}`, recorder.Body.String())
}

func TestFinderString(t *testing.T) {
	assert := NewAssert(t)
	req := NewPostJsonRequest("", "/test_finder", []byte(jsonData), "e", "E", "o", "O")
	ctx := new(Context)
	ctx.Request = req
	ctx.Finder = FinderWithRequest(req)
	ctx.UnmarshalInFinder()
	f := ctx.Finder
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

	req = NewPostJsonRequest("", "/test_finder", []byte(jsonData))
	ctx = new(Context)
	ctx.config = new(Config)
	ctx.config.DisableAutoUnmarshal = true
	ctx.Request = req
	ctx.Finder = FinderWithRequest(req)
	tb := JsonModel{}
	ctx.Unmarshal(&tb)
	assert.Equal("", tb.A)
	assert.Nil(tb.B)
	assert.True(tb.C)
	assert.Equal(12, tb.D)
	assert.Equal("str", tb.E)
}

func TestFinderInt(t *testing.T) {
	assert := NewAssert(t)
	req := NewGetRequest("", "", "a", 1, "b", 2)
	f := FinderWithRequest(req)
	a := f.RequireInt("a")
	assert.Equal(1, a)
	b := f.RequireInt("b")
	assert.Equal(2, b)
	jsonData := []byte(`{"chars":["a","b", "c"], "obj":{"x": 100}}`)
	f = FinderWithBytes(jsonData)
	assert.Equal("b", f.RequireString("chars", 1))
	assert.Equal(100, f.RequireInt("obj", "x"))
}

func TestFinderStringLen(t *testing.T) {
	assert := NewAssert(t)
	req := NewGetRequest("", "", "a", "1234567", "b", "语言文字")
	f := FinderWithRequest(req)
	_, err := f.FindStringLen(8, 10, "a")
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
	assert := NewAssert(t)
	req := NewGetRequest("", "", "a", "abcderg", "b", "语言文字")
	f := FinderWithRequest(req)
	_, err := f.FindStringMatch(regexp.MustCompile("\\w+"), "a")
	assert.Nil(err)
	_, err = f.FindStringMatch(regexp.MustCompile("\\d+"), "a")
	assert.NotNil(err)
}

func TestFinderOptionalString(t *testing.T) {
	assert := NewAssert(t)
	req := NewPostJsonRequest("", "/test_finder", []byte(jsonData), "year", "2013", "month", "May")
	ctx := new(Context)
	ctx.Request = req
	ctx.Finder = FinderWithRequest(req)
	ctx.UnmarshalInFinder()
	f := ctx.Finder
	s, err := f.FindOptionalString("default", "xyz")
	assert.Nil(err)
	assert.Equal("default", s, "Should return default value.")

	// Use the found value
	s, err = f.FindOptionalString("not_used", "month")
	assert.Nil(err)
	assert.Equal("May", s, "Should ignore default if string found.")
}
