package jas

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type Response struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error"`
}

//Context contains all the information for a single request.
//it hides the http.ResponseWriter because directly writing to http.ResponseWriter will make Context unable to work correctly.
//it embeds *http.Request and Finder, so you can call methods and access fields in Finder and *http.Request directly.
type Context struct {
	Finder
	*http.Request
	ResponseHeader http.Header
	Callback       string //jsonp callback
	Status         int
	Error          AppError
	Data           interface{} //The data to be written after the resource method has returned.
	UserId         int64
	Id             int64
	Extra          interface{} //Store extra data generated/used by hook functions, e.g. 'BeforeServe'.
	writer         io.Writer
	responseWriter http.ResponseWriter
	clientClosed   bool
	written        int
	config         *Config
	pathSegments   []string
	gaps           []string
}

var NoJsonBody = errors.New("jas.Context: no json body")

//Write and flush the data.
//It can be used for http streaming or to write a portion of large amount of data.
//If the type of the data is not []byte, it will be marshaled to json format.
func (ctx *Context) FlushData(data interface{}) (written int, err error) {
	var dataBytes []byte
	switch data.(type) {
	case []byte:
		dataBytes = data.([]byte)
	default:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			return
		}
	}
	if ctx.config.FlushDelimiter != nil {
		dataBytes = append(dataBytes, ctx.config.FlushDelimiter...)
	}
	if ctx.written == 0 && ctx.Status != 200 {
		ctx.responseWriter.WriteHeader(ctx.Status)
	}
	written, err = ctx.writer.Write(dataBytes)
	if err != nil {
		return
	}
	ctx.written += written
	if gzipWriter, ok := ctx.writer.(*gzip.Writer); ok {
		err = gzipWriter.Flush()
		if err != nil {
			return
		}
	}
	ctx.responseWriter.(http.Flusher).Flush()
	return
}

//Add response header Set-Cookie.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	ctx.ResponseHeader.Add("Set-Cookie", cookie.String())
}

//override *http.Request AddCookie method to add response header's cookie.
//Same as SetCookie.
func (ctx *Context) AddCookie(cookie *http.Cookie) {
	ctx.ResponseHeader.Add("Set-Cookie", cookie.String())
}

func (ctx *Context) deferredResponse() {
	if x := recover(); x != nil {
		var appErr AppError
		if handled, ok := x.(AppError); ok {
			appErr = handled
		} else {
			appErr = NewInternalError(x)
		}
		ctx.Error = appErr
	}
	var resp Response
	resp.Data = ctx.Data
	if ctx.Error != nil {
		ctx.Status = ctx.Error.Status()
		resp.Error = ctx.Error.Message()
	}
	var written int
	if ctx.config.HijackWrite != nil {
		ctx.responseWriter.WriteHeader(ctx.Status)
		written = ctx.config.HijackWrite(ctx.writer, ctx)
	} else {
		jsonBytes, _ := json.Marshal(resp)
		if ctx.Callback != "" { // handle JSONP
			if ctx.written == 0 {
				ctx.ResponseHeader.Set("Content-Type", "application/javascript; charset=utf-8")
				ctx.responseWriter.WriteHeader(ctx.Status)
			}
			a, _ := ctx.writer.Write([]byte(ctx.Callback + "("))
			b, _ := ctx.writer.Write(jsonBytes)
			c, _ := ctx.writer.Write([]byte(");"))
			written = a + b + c
		} else {
			if ctx.written == 0 {
				ctx.responseWriter.WriteHeader(ctx.Status)
				written, _ = ctx.writer.Write(jsonBytes)
			} else if resp.Data != nil || resp.Error != nil {
				written, _ = ctx.writer.Write(jsonBytes)
			}
		}
	}
	if ctx.Error != nil {
		ctx.written += written
		ctx.writer = nil
		ctx.Error.Log(ctx)
		if ctx.config.OnAppError != nil {
			go ctx.config.OnAppError(ctx.Error, ctx)
		}
	}
}

//Typically used in for loop condition.along with Flush.
func (ctx *Context) ClientClosed() bool {
	if ctx.clientClosed {
		return true
	}
	select {
	case <-ctx.responseWriter.(http.CloseNotifier).CloseNotify():
		ctx.clientClosed = true
	default:
	}
	return ctx.clientClosed
}

//the segment index starts at the resource segment
func (ctx *Context) PathSegment(index int) string {
	if len(ctx.pathSegments) <= index {
		return ""
	}
	return ctx.pathSegments[index]
}

//If the gap has multiple segments, the key should be
//the segment defined in resource Gap method.
//e.g. for gap ":domain/:language", use key ":domain"
//to get the first gap segment, use key ":language" to get the second gap segment.
//The first gap segment can also be gotten by empty string key "" for convenience.
func (ctx *Context) GapSegment(key string) string {
	for i := 0; i < len(ctx.gaps); i++ {
		if key == "" {
			return ctx.pathSegments[i+1]
		}
		if key == ctx.gaps[i] {
			return ctx.pathSegments[i+1]
		}
	}
	return ""
}

//It is an convenient method to validate and get the user id.
func (ctx *Context) RequireUserId() int64 {
	if ctx.UserId <= 0 {
		requerstError := NewRequestError("Unauthorized")
		requerstError.StatusCode = UnauthorizedStatusCode
		panic(requerstError)
	}
	return ctx.UserId
}

//Unmarshal the request body into the interface.
//It only works when you set Config option `DisableAutoUnmarshal` to true.
func (ctx *Context) Unmarshal(in interface{}) error {
	if !ctx.config.DisableAutoUnmarshal {
		panic("Should only call it when  'DisableAutoUnmarshal' is set to true")
	}
	if ctx.ContentLength > 0 && strings.Contains(ctx.Header.Get("Content-Type"), "application/json") {
		decoder := json.NewDecoder(ctx.Body)
		decoder.UseNumber()
		return decoder.Decode(in)
	}
	return NoJsonBody
}

//If set Config option `DisableAutoUnmarshal` to true, you should call this method first before you can get body parameters in Finder methods..
func (ctx *Context) UnmarshalInFinder() {
	if ctx.value == nil && ctx.ContentLength > 0 && strings.Contains(ctx.Header.Get("Content-Type"), "application/json") {
		var in interface{}
		decoder := json.NewDecoder(ctx.Body)
		decoder.UseNumber()
		ctx.err = decoder.Decode(&in)
		ctx.value = in
	}
}

type ContextWriter struct {
	Ctx *Context
}

func (cw ContextWriter) Write(p []byte) (n int, err error) {
	return cw.Ctx.FlushData(p)
}
