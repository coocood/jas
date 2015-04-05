JAS
===

JAS (JSON API Server) is a simple and powerful REST API framework for Go. [中文版 README](https://github.com/coocood/jas/blob/master/README_ZH.md)

[![Build Status](https://drone.io/github.com/coocood/jas/status.png)](https://drone.io/github.com/coocood/jas/latest)
[![Build Status](https://travis-ci.org/coocood/jas.png?branch=master)](https://travis-ci.org/coocood/jas)

## Requirement

Require Go 1.1+.

## Features

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

## Performance

JAS is a thin layer on top of net/http package, it adds about 1000ns operation time on every request, which means 99% of the performance when the qps number is around 10000.

But JAS will be faster than any regular expression routing solution. a single regular experssion match operation usually takes about 1000ns.

JAS router do not use regular expression, the routing performance would be constant as you define more resource and methods.

## Install

    go get github.com/coocood/jas

Only depends on a small assert package `github.com/coocood/assrt` for testing.

## Get Started

Define a struct type and its methods, methods should have one argument of type *jas.Context, no return value.

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


## Documentation

See [Gowalker](http://gowalker.org/github.com/coocood/jas) or [godoc](http://godoc.org/github.com/coocood/jas) for complete documentation.

## LICENSE

JAS is distributed under the terms of the MIT License. See [LICENSE](https://github.com/coocood/jas/blob/master/LICENSE) for details.

## Contributiors

[Jacob Olsen](https://github.com/jakeo), [doun](https://github.com/doun), [Olav](https://github.com/oal)
