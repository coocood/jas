JAS
===

JAS (JSON API Server) is a simple and powerful REST API framework for Go.

##Requirement

Require Go 1.1, Go 1.03 is not supported.

##Key Features

* No need to manually define any url routing rules, the rules are defined by your resource struct names and method names.
No more inconsistencies between your url path and your method name.

* Generate all the handled url paths seperated by "\n", so it can be used for reference or detect api changes.

* Unobtrusive, JAS router is just a http.Handler, you can make it work with other http.Handlers as well as have multiple JAS routers on the same server.

* Support HTTP Streaming, you can keep an connection open and send real-time data to the client, and get notified when the connection is closed on the client side.

* Support extract parameters from json request body in any level of depth, so it can be used like JSON RPC.

* Get and validate request parameter at the same time, support validate a integer, a string's min and max length or rune length, and match a regular expression.

* Generate default response with the parameter name when validation failed, optionally log the error in Common Log Format.

* Wrap all unhandled error into InternalError type, write response to the client with default message, and log the stacktraces and request infomation in Common Log Format.
and provide a optional function parameter to do the extra error handlling work.

* Provide an interface to handle errors, so you can define your own error type if the two default implementation can not meet your requirement.

* Support gzip.

* Highly configuarable.

##Install

    go get github.com/coocood/jas

Only depends on a small assert package `github.com/coocood/assrt` for testing.

##Get Started

Define a struct type and its methods, methods should have one argement of type *jas.Context, no return value.

    type Hello struct {}

    func (*Hello) Get (ctx *jas.Context) { // `GET /v1/hello`
    	ctx.Data = "hello world"
    	//response: `{"data":"hello world","error":null}`
    }

    func main () {
        router := jas.NewRouter(new(Hello))
        router.BasePath = "/v1/"
        fmt.Println(router.HandledPaths(true))
        //output: `GET /v1/hello`
        http.Handle(router.BasePath, router)
        http.ListenAndServe(":8080", nil)
    }


HTTP GET POST PUT DELETE should be the prefix of the method name. Methods with no prefix defaults to handle GET request.

    func (*Hello) Foo (ctx *jas.Context) {} // `GET /v1/hello/foo`

    func (*Hello) PostFoo (ctx *jas.Context) {} // `POST /v1/hello/foo`

    func (*Hello) PostPost (ctx *jas.Context) {} // `POST /v1/hello/post`

    func (*Hello) GetPost (ctx *jas.Context) {} // `GET /v1/hello/post`

    func (*Hello) PutFoo (ctx *jas.Context) {} // `PUT /v1/hello/foo`

    func (*Hello) DeleteFoo (ctx *jas.Context) {} // `DELETE /v1/hello/foo`


And you can make it more RESTful by put an Id path between the resource name and method name, the id value can be get from *jas.Context:
Resource name with `Id` suffix do the trick.

    type HelloId struct {}

    func (*HelloId) Foo (ctx *jas.Context) {// `GET /v1/hello/:id/foo`
        id := ctx.Id
        _ = id
    }

Find methods will return err if the parameter value is invalid.
Require methods will stop the execution in the method and respond an error message if the parameter value is invalid.

    func (*Hello) Foo (ctx *jas.Context) {
        name := ctx.RequireString("name") // will stop execution and response `{"data":null,"error":"nameInvalid"} if "name" parameter is not provided.
        age := ctx.RequirePositiveInt("age")
        grade, err := ctx.FindPositiveInt("grade")
        password := ctx.RequireStringLen(6, 60, "password") // 6, 60 is the min and max length, error message can be "passwordTooShort" or "passwordTooLong"
        email := ctx.RequireStringMatch(emailRegexp, "email") // emailRegexp is a *regexp.Regexp instance.error message would be "emailInvalid"
        _, _, _, _, _, _ = name, age, grade, err,password, email
    }

Get json body parameter:
Assume we have a request with json body

    `{
        "foo":[
            {"name":"abc"},
            {"id":200}
        ]
    }`

Then we can get the value with Find or Require methods

    func (*Hello) Bar (ctx *jas.Context) {
        name, _ := ctx.Find("foo", 0, "name") // "abc"
        id:= ctx.RequirePositiveInt("foo", 1, "id") //200
    }


HTTP streaming :

    func (*Hello) Feed (ctx *jas.Context) {
        for !ctx.ClientClosed() {
            ctx.FlushData([]byte("some data"))
            time.Sleep(time.Second)
        }
    }


##Documentation

See [Gowalker](http://gowalker.org/github.com/coocood/jas) for complete documentation.

##LICENSE

JAS is distributed under the terms of the MIT License. See [LICENSE](https://github.com/coocood/jas/blob/master/LICENSE) for details.