package interceptor

import (
	"google.golang.org/grpc"
	"log"
)

func UnaryServerMiddlewareDemo1(c *UnaryServerConext) (interface{}, error) {
	//any logic here, auth, log ,tracing, validation
	log.Printf("demo interceptor being ")
	c.Next()
	log.Printf("demo interceptor end")
	return nil, nil
}

func UnaryServerMiddlewareDemo2(c *UnaryServerConext) (interface{}, error) {
	//any logic here, auth, log ,tracing, validation
	log.Printf("demo interceptor being ")
	c.Next()
	log.Printf("demo interceptor end")
	return nil, nil
}

//Use Intercept
func NewgRPCServer() *grpc.Server {
	midware := NewInterceptorEngine().UnaryServerUse(
		UnaryServerMiddlewareDemo1,
		UnaryServerMiddlewareDemo2,
	)
	s := grpc.NewServer(grpc.UnaryInterceptor(midware.UnaryServerInterceptor()))
	return s
}
