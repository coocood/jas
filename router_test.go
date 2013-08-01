package jas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var empty = fmt.Sprint("")

type Hello struct {
}

func (*Hello) Get(ctx *Context) {
	ctx.Data = "hello world"
}

func TestHelloWorld(t *testing.T) {
	assert := NewAssert(t)
	req, _ := http.NewRequest("GET", "http://localhost/hello?name=world", nil)
	router := NewRouter(new(Hello))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(`{"data":"hello world","error":null}`, string(recorder.Body.Bytes()))
}

type Users struct {
}

func (u *Users) ImageUrl(ctx *Context) {}

func (u *Users) PhotosId(ctx *Context) {
	ctx.Data = ctx.Id
}

func (u *Users) Gap() string {
	return ":username"
}

type UsersId struct{}

func (ui *UsersId) ImageUrl(ctx *Context) {}

func (ui *UsersId) PostPost(ctx *Context) {}

func (ui *UsersId) GetPost(ctx *Context) {}

type Stringusers struct{}

func (*Stringusers) Get(ctx *Context) {}

type stringusers struct{}

func (u *stringusers) Gap() string {
	return ":id"
}

func (*stringusers) Get(ctx *Context) {}

func TestRouter(t *testing.T) {
	assert := NewAssert(t)
	router := NewRouter(new(Users))
	router.BasePath = "/base/"
	paths := strings.Split(router.HandledPaths(false), "\n")
	assert.Equal("GET /users/:username/image_url", paths[0])
	assert.Equal("GET /users/:username/photos/:id", paths[1])

	paths = strings.Split(router.HandledPaths(true), "\n")
	assert.Equal("GET /base/users/:username/image_url", paths[0])
	assert.Equal("GET /base/users/:username/photos/:id", paths[1])

	req, _ := http.NewRequest("GET", "http://localhost/base/users/adam/image_url", nil)
	path, id, segments, gaps := router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	ctx := new(Context)
	ctx.req = req
	ctx.pathSegments = segments
	ctx.gaps = gaps
	assert.Equal("adam", ctx.GapSegment(":username"))
	assert.Equal("adam", ctx.GapSegment(""))

	assert.Equal("GET /users/:username/image_url", path)
	assert.Equal(0, id)
	assert.Equal("users/adam/image_url", strings.Join(segments, "/"))
	assert.Equal(":username", strings.Join(gaps, "/"))
	_, ok := router.methodMap[path]
	assert.True(ok)

	req, _ = http.NewRequest("GET", "http://localhost/base/users/jack/photos/56", nil)
	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("GET /users/:username/photos/:id", path)
	assert.Equal(56, id)
	assert.Equal("users/jack/photos/56", strings.Join(segments, "/"))
	assert.Equal(":username", strings.Join(gaps, "/"))
	_, ok = router.methodMap[path]
	assert.True(ok)

	router = NewRouter(new(UsersId))
	router.BasePath = "/base/1/"
	paths = strings.Split(router.HandledPaths(false), "\n")
	assert.Equal("GET /users/:id/image_url", paths[0])
	assert.Equal("GET /users/:id/post", paths[1])
	assert.Equal("POST /users/:id/post", paths[2])

	req, _ = http.NewRequest("GET", "http://localhost/base/1/users/5/image_url", nil)

	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("GET /users/:id/image_url", path)
	assert.Equal(5, id)
	assert.Equal("users/5/image_url", strings.Join(segments, "/"))
	_, ok = router.methodMap[path]
	assert.True(ok)

	router.AllowIntegerGap = true
	req, _ = http.NewRequest("POST", "http://localhost/base/1/users/6/post", nil)
	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("POST /users/:id/post", path)
	_, ok = router.methodMap[path]
	assert.True(ok)

	req, _ = http.NewRequest("GET", "http://localhost/base/1/users/3/post", nil)
	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("GET /users/:id/post", path)
	_, ok = router.methodMap[path]
	assert.True(ok)

	router = NewRouter(new(UsersId), new(Users))
	router.BasePath = "/base/1/"
	req, _ = http.NewRequest("GET", "http://localhost/base/1/users/5/post", nil)
	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("GET /users/:id/post", path)

	router.AllowIntegerGap = true
	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("GET /users/:username/post", path)

	router = NewRouter(new(Stringusers), new(stringusers))
	router.BasePath = "/base/2/"
	req, _ = http.NewRequest("GET", "http://localhost/base/2/stringusers", nil)
	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("GET /stringusers", path)
	req, _ = http.NewRequest("GET", "http://localhost/base/2/stringusers/51f959801a2a7c3300000000", nil)
	path, id, segments, gaps = router.resolvePath(req.Method, req.URL.Path[len(router.BasePath):])
	assert.Equal("GET /stringusers/:id", path)
}

type Error struct {
}

func (h *Error) Request(ctx *Context) {
	panic(NewRequestError("request error"))
}

func (h *Error) Internal(ctx *Context) {
	regexp.MustCompile(`\1`)
}

func TestError(t *testing.T) {
	assert := NewAssert(t)
	buffer := bytes.NewBuffer(nil)
	router := NewRouter(new(Error))
	router.RequestErrorLogger = log.New(buffer, "", 0)
	req := NewGetRequest("", "error/request")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(`{"data":null,"error":"request error"}`, recorder.Body.String())
	loggedLine := buffer.String()
	assert.True(strings.Index(loggedLine, "request error") != -1)

	buffer = bytes.NewBuffer(nil)
	router.InternalErrorLogger = log.New(buffer, "", 0)
	req = NewGetRequest("", "error/internal")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(`{"data":null,"error":"InternalError"}`, recorder.Body.String())
	loggedLine = buffer.String()
	assert.True(strings.Index(loggedLine, "regexp/regexp.go") != -1)
	assert.True(strings.Index(loggedLine, "router_test.go") != -1)
}

type Jsonp struct {
}

func (jp *Jsonp) Get(ctx *Context) {
	ctx.Callback = ctx.FormValue("callback")
	ctx.Data = "jsonp"
}

func TestJsonp(t *testing.T) {
	assert := NewAssert(t)
	router := NewRouter(new(Jsonp))
	req := NewGetRequest("", "jsonp", "callback", "dosomething")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	assert.Equal(`dosomething({"data":"jsonp","error":null});`, recorder.Body.String())
}

var routerUsers = NewRouter(new(Users))
var reqUsers, _ = http.NewRequest("GET", "http://locoalhost/users/john/photos/5", nil)

func BenchmarkUsersPhotosIdJas(b *testing.B) {
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		routerUsers.ServeHTTP(recorder, reqUsers)
	}
}

func BenchmarkUsersPhotosIdBasic(b *testing.B) {
	var reg = regexp.MustCompile("/users/\\w+/photos/\\d+/?.*")
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		url := reqUsers.URL
		if reg.MatchString(url.Path) {
			handleUsersPhotosId(recorder, reqUsers)
		}
	}
}

func handleUsersPhotosId(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	id, _ := strconv.ParseInt(segments[len(segments)-1], 10, 64)
	resp := Response{}
	resp.Data = id
	jsonBytes, _ := json.Marshal(resp)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonBytes)
}

var reqHello, _ = http.NewRequest("GET", "http://locoalhost/hello?name=gopher", nil)

func BenchmarkHelloJas(b *testing.B) {
	var routerHello = NewRouter(new(Hello))
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		routerHello.ServeHTTP(recorder, reqHello)
	}
}

func BenchmarkHelloBasic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		recorder := httptest.NewRecorder()
		handleHello(recorder, reqHello)
	}
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	resp.Data = "hello world"
	jsonBytes, _ := json.Marshal(resp)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(jsonBytes)
}
