/*
To build a REST API you need to define resources.

A resource is a struct with one or more exported pointer methods that accept only one argument of type `*jas.Context`

A `*jas.Context` has everything you need to handle the http request, it embeds a anonymous *http.Request field,
so you can call *http.Requst methods directly with *jas.Context.

The resource name and method name will be converted to snake case in the url path by default.(can be changed in config).

HTTP GET POST PUT DELETE should be the prefix of the method name.

Methods with no prefix will handle GET request.

Other HTTP request with methods like "HEAD", "OPTIONS" will be routed to resource "Get" method.

Examples:

	type Users struct {}

    func (*Users) Photo (ctx *jas.Context) {} // `GET /users/photo`

    func (*Users) PostPhoto (ctx *jas.Context) {} // `POST /users/photo`

    func (*Users) PostPost (ctx *jas.Context) {} // `POST /users/post`

    func (*Users) GetPost (ctx *jas.Context) {} // `GET /users/post`

    func (*Users) PutPhoto (ctx *jas.Context) {} // `PUT /users/photo`

    func (*Users) DeletePhoto (ctx *jas.Context) {} // `DELETE /users/photo`

On your application start, make a new *jas.Router with jas.NewRouter method, pass all your resources to it.

	router := jas.NewRouter(new(Users), new(Posts), new(Photos)

Then you can config the router, see Config type for detail.

    router.BasePath = "/v1/"
	router.EnableGzip = true

You can get all the handled path printed. they are separated by '\n'

	fmt.Println(router.HandledPaths(true)) // true for with base path. false for without base path.

Finally, set the router as http handler and Listen.

	http.Handle(router.BasePath, router)
    http.ListenAndServe(":8080", nil)

You can make it more RESTful by put an Id path between the resource name and method name.

The id value can be obtained from *jas.Context, resource name with `Id` suffix do the trick.

    type UsersId struct {}

    func (*UsersId) Photo (ctx *jas.Context) {// `GET /users/:id/photo`
        id := ctx.Id
        _ = id
    }

If resource implements ResourceWithGap interface, the handled path will has gap segments between resource name and method name.

If method has a suffix "Id", the handled path will append an `:id` segment after the method segment.

You can obtain the Id value from *jas.Context, but it only works with none Id resource, because there is only one Id field in *jas.Context.

	type Users struct {}

	func (*Users) Gap() string {
		return ":name"
	}

    func (*Users) Photo (ctx *jas.Context) {// `GET /users/:name/photo`
    	// suppose the request path is `/users/john/photo`
        name := ctx.GapSegment("") // "john"
        _ = name
    }

    func (*Users) PhotoId (ctx *jas.Context) { `GET /users/:name/photo/:id`
    	// suppose the request path is `/users/john/photo/7`
    	id := ctx.Id // 7
        _ = id
    }

Find methods will return error if the parameter value is invalid.
Require methods will stop the execution in the method and respond an error message if the parameter value is invalid.

    func (*Users) Photo (ctx *jas.Context) {
        // will stop execution and response `{"data":null,"error":"nameInvalid"} if "name" parameter is not given..
        name := ctx.RequireString("name")
        age := ctx.RequirePositiveInt("age")
        grade, err := ctx.FindPositiveInt("grade")

        // 6, 60 is the min and max length, error message can be "passwordTooShort" or "passwordTooLong"
        password := ctx.RequireStringLen(6, 60, "password")

        // emailRegexp is a *regexp.Regexp instance.error message would be "emailInvalid"
        email := ctx.RequireStringMatch(emailRegexp, "email")
        _, _, _, _, _, _ = name, age, grade, err,password, email
    }

Get json body parameter:
Assume we have a request with json body

    {
        "photo":[
            {"name":"abc"},
            {"id":200}
        ]
    }

Then we can get the value with Find or Require methods.
Find and Require methods accept varargs, the type can be either string or int.
string argument to get value from json object, int argumnet to get value form json array.

    func (*Users) Bar (ctx *jas.Context) {
        name, _ := ctx.Find("photo", 0, "name") // "abc"
        id:= ctx.RequirePositiveInt("photo", 1, "id") //200
    }

If you want unmarshal json body to struct, the `DisableAutoUnmarshal` option must be set to true.

	router.DisableAutoUnmarshal = true

Then you can call `Unmarshal` method to unmarshal json body:

	ctx.Unmarshal(&myStruct)

HTTP streaming :

FlushData will write []byte data without any modification, other data types will be marshaled to json format.

    func (*Users) Feed (ctx *jas.Context) {
        for !ctx.ClientClosed() {
            ctx.FlushData([]byte("some data"))
            time.Sleep(time.Second)
        }
    }

*/
package jas
