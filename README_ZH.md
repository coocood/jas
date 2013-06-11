JAS
===

JAS (JSON API Server) 是一个简单的功能强大REST API框架

[![Build Status](https://drone.io/github.com/coocood/jas/status.png)](https://drone.io/github.com/coocood/jas/latest)

## 版本支持

仅支持Go 1.1

## 特性

* 无需手动定义URL路由规则， 路由规则由资源struct名和方法名决定，保证了方法名和URL路径的一致性。

* 生成所有已处理的URL路径，用"\n"分割，可以作为API参考资料或检测API的改变。

* 非侵入式，JAS router只是一个http.Handler，可以和其他http.Handler一起使用，或在一个服务端口使用用多个JAS router。

* 支持HTTP Streaming, 可以保持长连接，向客户端发送实时数据，并在客户端关闭时得到通知。

* 支持从JSON request body的任意深度的路径提取参数，可以像JSON RPC一样用。

* 提取参数的同时进行基本的验证，支持验证整数，string的长度，string的rune长度，string正则匹配。

* 如果验证失败，响应默认的包含参数名的消息，（可选）将错误用Common Log Format写入日志。

* 把未处理的错误包装成InternalError类型，响应默认消息，将stacktrace和请求信息用Common Log Format写入日志，支持自定义的错误回调函数。

* 错误使用interface类型，这样当自带的错误类型不能满足需求，可以自定义错误类型。

* 支持gzip。

* 丰富的配置选项。

## 性能

JAS是一个net/http包的轻量包装，每次请求增加了大约1000ns的处理时间，意味着在10000 QPS的应用场景下，达到99%的性能。

但是JAS的性能要好于比任何用正则定义的路由解决方案，一个单独的match操作通常需要1000ns。

JAS router不需要正则操作，路由的性能不会随着添加更多的资源和方法而降低。

## 安装

    go get github.com/coocood/jas

## 快速上手

定义一个struct及其方法，方法需要包含一个参数*jas.Context，没有返回值。

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


## 文档

完整的文档在 [Gowalker](http://gowalker.org/github.com/coocood/jas) 或 [godoc](http://godoc.org/github.com/coocood/jas).

## License

MIT License.[LICENSE](https://github.com/coocood/jas/blob/master/LICENSE).