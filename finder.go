package jas

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

//Finder is a accessor and validator with an unified interface to get parameters from both http request form and json body.
//Form parameters take precedence over request json body.

//Finder can also be used for json data only, *http.Request is optional.
//But then you should not call any "Require" methods, those should only be used to get http request parameters.

//All the "paths" parameter can only be int or string type. int for get element in json array,
//string for form values or fields in json objects.

//All the "Find" methods return error, All the "Require" methods do panic with RequestError when error occured.
type Finder struct {
	value interface{}
	err   error
	req   *http.Request
}

var WrongTypeError = errors.New("jas.Finder: wrong type")
var IndexOutOfBoundError = errors.New("jas.Finder: index out of bound")
var EntryNotExistsError = errors.New("jas.Finder: entry not exists")
var EmptyStringError = errors.New("jas.Finder: empty string")
var EmptySliceError = errors.New("jas.Finder: empty slice")
var EmptyMapError = errors.New("jas.Finder: empty map")
var NullValueError = errors.New("jas.Finder: null value")

var TooShortError = errors.New("jas.Finder: string too short")
var TooLongError = errors.New("jas.Finder: string too long")
var NotPositiveError = errors.New("jas.Finder: not positive")
var DoNotMatchError = errors.New("jas.Finder: do not match")

var InvalidErrorFormat = "%vInvalid"
var NotPositiveErrorFormat = "%vNotPositive"
var TooShortErrorFormat = "%vTooShort"
var TooLongErrorFormat = "%vTooLong"
var MalformedJsonBody = "MalformedJsonBody"

func (finder Finder) FindString(paths ...interface{}) (string, error) {
	if s := finder.findFormString(paths...); s != "" {
		return s, nil
	}
	finder = finder.FindChild(paths...)
	if finder.err != nil {
		return "", finder.err
	}
	if s, ok := finder.value.(string); ok {
		if s == "" {
			return s, EmptyStringError
		}
		return s, nil
	}
	return "", WrongTypeError
}

// Looks up the given path and returns a default value if not present.
func (finder Finder) FindOptionalString(val string, paths ...interface{}) (string, error) {
	s, err := finder.FindString(paths...)
	if err != nil {
		switch err {
		case EmptyStringError, EntryNotExistsError, NullValueError:
			return val, nil
		}
	}
	return s, err
}

func (finder Finder) FindStringLen(min, max int, paths ...interface{}) (string, error) {
	s, err := finder.FindString(paths...)
	if err != nil {
		return s, err
	}
	if len(s) < min {
		return s, TooShortError
	}
	if len(s) >= max {
		return s, TooLongError
	}
	return s, nil
}

func (finder Finder) FindStringRuneLen(min, max int, paths ...interface{}) (string, error) {
	s, err := finder.FindString(paths...)
	if err != nil {
		return s, err
	}
	count := 0
	for _ = range s {
		count++
	}
	if count < min {
		return s, TooShortError
	}
	if count >= max {
		return s, TooLongError
	}
	return s, nil
}

func (finder Finder) FindStringMatch(reg *regexp.Regexp, paths ...interface{}) (string, error) {
	s, err := finder.FindString(paths...)
	if err != nil {
		return s, err
	}
	if !reg.MatchString(s) {
		return s, DoNotMatchError
	}
	return s, nil
}

func (finder Finder) FindSlice(paths ...interface{}) ([]interface{}, error) {
	finder = finder.FindChild(paths...)
	if finder.err != nil {
		return nil, finder.err
	}
	if s, ok := finder.value.([]interface{}); ok {
		if len(s) == 0 {
			return s, EmptySliceError
		}
		return s, nil
	}
	return nil, WrongTypeError
}

func (finder Finder) RequireSlice(paths ...interface{}) []interface{} {
	s, err := finder.FindSlice(paths...)
	if err != nil {
		doPanic(InvalidErrorFormat, paths...)
	}
	return s
}

func (finder Finder) RequireStringLen(min, max int, paths ...interface{}) string {
	s := finder.RequireString(paths...)
	if len(s) < min {
		doPanic(TooShortErrorFormat, paths...)
	}
	if len(s) >= max {
		doPanic(TooLongErrorFormat, paths...)
	}
	return s
}

func (finder Finder) RequireStringRuneLen(min, max int, paths ...interface{}) string {
	s := finder.RequireString(paths...)
	count := 0
	for _ = range s {
		count++
	}
	if count < min {
		doPanic(TooShortErrorFormat, paths...)
	}
	if count >= max {
		doPanic(TooLongErrorFormat, paths...)
	}
	return s
}

func (finder Finder) RequireStringMatch(reg *regexp.Regexp, paths ...interface{}) string {
	s := finder.RequireString(paths...)
	if !reg.MatchString(s) {
		doPanic(InvalidErrorFormat, paths...)
	}
	return s
}

func (finder Finder) FindMap(paths ...interface{}) (map[string]interface{}, error) {
	finder = finder.FindChild(paths...)
	if finder.err != nil {
		return nil, finder.err
	}
	if m, ok := finder.value.(map[string]interface{}); ok {
		if len(m) == 0 {
			return m, EmptyMapError
		}
		return m, nil
	}
	return nil, WrongTypeError
}

func (finder Finder) RequireMap(paths ...interface{}) map[string]interface{} {
	m, err := finder.FindMap(paths...)
	if err != nil {
		doPanic(InvalidErrorFormat, paths...)
	}
	return m
}

func (finder Finder) RequireString(paths ...interface{}) string {
	s, err := finder.FindString(paths...)
	if err != nil {
		doPanic(InvalidErrorFormat, paths...)
	}
	return s
}

func (finder Finder) FindInt(paths ...interface{}) (int64, error) {
	if s := finder.findFormString(paths...); s != "" {
		return strconv.ParseInt(s, 10, 64)
	}
	num, err := finder.findNumber(paths...)
	if err != nil {
		return 0, err
	}
	return num.Int64()
}

func (finder Finder) FindPositiveInt(paths ...interface{}) (int64, error) {
	integer, err := finder.FindInt(paths...)
	if err != nil {
		return integer, err
	}
	if integer <= 0 {
		return integer, NotPositiveError
	}
	return integer, nil
}

func (finder Finder) RequireInt(paths ...interface{}) int64 {
	i, err := finder.FindInt(paths...)
	if err != nil {
		doPanic(InvalidErrorFormat, paths...)
	}
	return i
}

func (finder Finder) RequirePositiveInt(paths ...interface{}) int64 {
	i := finder.RequireInt(paths...)
	if i <= 0 {
		doPanic(NotPositiveErrorFormat, paths...)
	}
	return i
}

func (finder Finder) FindFloat(paths ...interface{}) (float64, error) {
	if s := finder.findFormString(paths...); s != "" {
		return strconv.ParseFloat(s, 64)
	}
	num, err := finder.findNumber(paths...)
	if err != nil {
		return 0, err
	}
	return num.Float64()
}

func (finder Finder) RequireFloat(paths ...interface{}) float64 {
	f, err := finder.FindFloat(paths...)
	if err != nil {
		doPanic(InvalidErrorFormat, paths...)
	}
	return f
}

func (finder Finder) RequirePositiveFloat(paths ...interface{}) float64 {
	f, err := finder.FindFloat(paths...)
	if err != nil {
		doPanic(InvalidErrorFormat, paths...)
	} else if f < 0 {
		doPanic(NotPositiveErrorFormat, paths...)
	}
	return f
}

func (finder Finder) FindBool(paths ...interface{}) (bool, error) {
	if s := finder.findFormString(paths...); s != "" {
		return strconv.ParseBool(s)
	}
	finder = finder.FindChild(paths...)
	if finder.err != nil {
		return false, finder.err
	}
	if b, ok := finder.value.(bool); ok {
		return b, nil
	}
	return false, WrongTypeError
}

// return the length of []interface or map[string]interface{}
// return -1 if the value not found or has wrong type.
func (finder Finder) Len(paths ...interface{}) int {
	finder = finder.FindChild(paths...)
	if finder.err != nil {
		return -1
	}
	switch finder.value.(type) {
	case []interface{}:
		return len(finder.value.([]interface{}))
	case map[string]interface{}:
		return len(finder.value.(map[string]interface{}))
	}
	return -1
}

func (finder Finder) FindChild(paths ...interface{}) Finder {
	finder.req = nil
	if finder.value == nil {
		finder.err = NullValueError
		return finder
	}
	for _, path := range paths {
		switch path.(type) {
		case string:
			finder = finder.findChildInMap(path.(string))
			if finder.err != nil {
				return finder
			}
		case int:
			finder = finder.findChildInSlice(path.(int))
			if finder.err != nil {
				return finder
			}
		default:
			panic("path type can only be string or int")
		}
	}
	return finder
}

//Construct a Finder with *http.Request.
func FinderWithRequest(req *http.Request) Finder {
	finder := Finder{}
	finder.req = req
	return finder
}

//Construct a Finder with json formatted data.
func FinderWithBytes(data []byte) Finder {
	finder := Finder{}
	var in interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	finder.err = decoder.Decode(&in)
	finder.value = in
	return finder
}

func (finder Finder) findChildInSlice(index int) Finder {
	slice, ok := finder.value.([]interface{})
	if !ok {
		finder.err = WrongTypeError
		return finder
	}
	if len(slice) <= index || index < 0 {
		finder.err = IndexOutOfBoundError
		return finder
	}
	finder.value = slice[index]
	if finder.value == nil {
		finder.err = NullValueError
	}
	return finder
}

func (finder Finder) findChildInMap(key string) Finder {
	m, ok := finder.value.(map[string]interface{})
	if !ok {
		finder.err = WrongTypeError
		return finder
	}
	finder.value, ok = m[key]
	if !ok {
		finder.err = EntryNotExistsError
	} else if finder.value == nil {
		finder.err = NullValueError
	}
	return finder
}

func (finder Finder) findNumber(paths ...interface{}) (json.Number, error) {
	finder = finder.FindChild(paths...)
	if finder.err != nil {
		return "", finder.err
	}
	if num, ok := finder.value.(json.Number); ok {
		return num, nil
	}
	return "", WrongTypeError
}

func doPanic(format string, paths ...interface{}) {
	keyPath := "value"
	if len(paths) > 0 {
		lastPath := paths[len(paths)-1]
		if s, ok := lastPath.(string); ok {
			keyPath = s
		}
	}
	requestErrorString := fmt.Sprintf(format, keyPath)
	requerstError := NewRequestError(requestErrorString)
	panic(requerstError)
}

func (finder Finder) findFormString(paths ...interface{}) string {
	if finder.req != nil && len(paths) == 1 {
		key, ok := paths[0].(string)
		if ok {
			return finder.req.FormValue(key)
		}
	}
	return ""
}
