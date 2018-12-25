# grpc-interceptor
gRPC interceptor chaining handler tool, like [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)   
与 go-grpc-middleware 的不同之处， 比如有三个拦截器程序 handler_a, handler_b, handler_c   
1、go-grpc-middleware 只支持 a -> b -> c 串行执行，也就是后一个必须等待上一个执行完以后才能执行  
2、grpc-interceptor 支持中断b执行，待后续handler执行完以后继续执行   

使用此拦截器就可以方便的实现一些middleware，比如请求处理时长，输出出入请求日志，trace，auth等公共逻辑   
> 目前只实现了grpc Unary，因为 Stream用的很少，暂不支持

# 使用
```
//编写处理程序-1
 UnaryServerMiddlewareDemo1(c *UnaryServerConext) (interface{}, error) {
	//any logic here, auth, log ,tracing, validation
	log.Printf("demo interceptor being ")
	c.Next()
	log.Printf("demo interceptor end")
	return nil, nil
}

//编写处理程序-2
func UnaryServerMiddlewareDemo2(c *UnaryServerConext) (interface{}, error) {
	//any logic here, auth, log ,tracing, validation
	log.Printf("demo interceptor being ")
	c.Next()
	log.Printf("demo interceptor end")
	return nil, nil
}

//server端使用interceptor
func NewgRPCServer() *grpc.Server {
	midware := NewInterceptorEngine().UnaryServerUse(
		UnaryServerMiddlewareDemo1,
		UnaryServerMiddlewareDemo2,
	)
	s := grpc.NewServer(grpc.UnaryInterceptor(midware.UnaryServerInterceptor()))
	return s
}

```