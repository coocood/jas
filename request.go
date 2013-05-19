package jas

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	baseLocal = "http://localhost/"
)

func NewGetRequest(baseUrlOrPath, path string, nameValues ...interface{}) *http.Request {
	url := concateUrl(baseUrlOrPath, path)
	if len(nameValues) > 0 {
		query := NameValuesToUrlValues(nameValues...).Encode()
		url += "?" + query
	}
	req, _ := http.NewRequest(
		"GET", url, nil,
	)
	return req
}

func NewPostFormRequest(baseUrlOrPath, path string, nameValues ...interface{}) *http.Request {
	var reader io.Reader
	if len(nameValues) > 0 {
		pairs := NameValuesToUrlValues(nameValues...)
		reader = strings.NewReader(pairs.Encode())
	}
	req, _ := http.NewRequest("POST", concateUrl(baseUrlOrPath, path), reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func NewPostJsonRequest(baseUrlOrPath, path string, jsonData []byte, nameValues ...interface{}) *http.Request {
	url := concateUrl(baseUrlOrPath, path)
	if len(nameValues) > 0 {
		query := NameValuesToUrlValues(nameValues...).Encode()
		url += "?" + query
	}
	req, _ := http.NewRequest(
		"POST", url, bytes.NewReader(jsonData),
	)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NameValuesToUrlValues(nameValues ...interface{}) url.Values {
	if len(nameValues)%2 != 0 {
		panic(fmt.Sprint("name value pair not even:", nameValues, len(nameValues)))
	}
	nameValuesStrings := make([]string, len(nameValues))
	for i := 0; i < len(nameValues); i++ {
		value := nameValues[i]
		var str string
		if byt, ok := value.([]byte); ok {
			str = string(byt)
		} else {
			str = fmt.Sprint(value)
		}
		nameValuesStrings[i] = str
	}
	values := url.Values{}
	for i := 0; i < len(nameValuesStrings); i += 2 {
		values.Set(nameValuesStrings[i], nameValuesStrings[i+1])
	}
	return values
}

func concateUrl(base, path string) string {
	if base == "" {
		base = baseLocal
	} else if !strings.HasPrefix(base, "http") {
		if base[0] == '/' {
			base = base[1:]
		}
		base = baseLocal + base
	}
	if base[len(base)-1] != '/' {
		base += "/"
	}
	if path != "" && path[0] == '/' {
		path = path[1:]
	}
	return base + path
}
