package interceptor

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
)

type UnaryClientHandler func(c *UnaryClientConext) error
type UnaryServerHandler func(c *UnaryServerConext) (interface{}, error)

type UnaryClientConext struct {
	Ctx      context.Context
	Method   string
	Req      interface{}
	Reply    interface{}
	Cc       *grpc.ClientConn
	Invoker  grpc.UnaryInvoker
	Opts     []grpc.CallOption
	handlers []UnaryClientHandler
	index    int8
	Err      error
	engine   *InterceptorEngine
	keys     map[string]interface{}
}

func (c *UnaryClientConext) reset(ctx context.Context, method string, req,
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) {
	c.Ctx = ctx
	c.Method = method
	c.Req = req
	c.Reply = reply
	c.Cc = cc
	c.Invoker = invoker
	c.index = -1
	c.handlers = []UnaryClientHandler{}
}
func (c *UnaryClientConext) Next() {
	c.index++
	for s := int8(len(c.handlers)); c.index < s; c.index++ {
		err := c.handlers[c.index](c)
		if s-c.index == 1 {
			c.Err = err
		}
	}
}

// Set: store a new key/value pair for this context.
func (c *UnaryClientConext) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = value
}

// Get: get a value for key
func (c *UnaryClientConext) Get(key string) (value interface{}, exists bool) {
	value, exists = c.keys[key]
	return
}

type InterceptorEngine struct {
	ucHandles []UnaryClientHandler
	usHandles []UnaryServerHandler

	ucPool sync.Pool
	usPool sync.Pool
}

func NewInterceptorEngine() *InterceptorEngine {
	engine := &InterceptorEngine{}
	engine.ucPool.New = func() interface{} {
		return engine.allocateUCContext()
	}
	engine.usPool.New = func() interface{} {
		return engine.allocateUSContext()
	}
	return engine
}
func (engine *InterceptorEngine) allocateUCContext() *UnaryClientConext {
	return &UnaryClientConext{engine: engine}
}

func (engine *InterceptorEngine) allocateUSContext() *UnaryServerConext {
	return &UnaryServerConext{engine: engine}
}
func (engine *InterceptorEngine) UnaryClientUse(handler ...UnaryClientHandler) *InterceptorEngine {
	engine.ucHandles = append(engine.ucHandles, handler...)
	return engine
}

func (engine *InterceptorEngine) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		c := engine.ucPool.Get().(*UnaryClientConext)
		c.reset(ctx, method, req, reply, cc, invoker, opts...)
		c.handlers = append(c.handlers, engine.ucHandles...)
		c.handlers = append(c.handlers, func(tctx *UnaryClientConext) error {
			return invoker(tctx.Ctx, tctx.Method, tctx.Req, tctx.Reply, tctx.Cc, tctx.Opts...)
		})
		c.Next()
		engine.ucPool.Put(c)
		return c.Err
	}
}

type UnaryServerConext struct {
	Ctx      context.Context
	Req      interface{}
	Info     *grpc.UnaryServerInfo
	Handler  grpc.UnaryHandler
	handlers []UnaryServerHandler
	index    int8
	Reply    interface{}
	Err      error
	engine   *InterceptorEngine
	keys     map[string]interface{}
}

func (c *UnaryServerConext) reset(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) {
	c.Ctx = ctx
	c.Req = req
	c.Info = info
	c.Handler = handler
	c.index = -1
	c.handlers = []UnaryServerHandler{}
}

func (c *UnaryServerConext) Next() {
	c.index++
	for s := int8(len(c.handlers)); c.index < s; c.index++ {
		reply, err := c.handlers[c.index](c)
		if s-c.index == 1 {
			c.Err = err
			c.Reply = reply
		}
	}
}

// Set: store a new key/value pair for this context.
func (c *UnaryServerConext) Set(key string, value interface{}) {
	if c.keys == nil {
		c.keys = make(map[string]interface{})
	}
	c.keys[key] = value
}

// Get: get a value for key
func (c *UnaryServerConext) Get(key string) (value interface{}, exists bool) {
	value, exists = c.keys[key]
	return
}

func (engine *InterceptorEngine) UnaryServerUse(handler ...UnaryServerHandler) *InterceptorEngine {
	engine.usHandles = append(engine.usHandles, handler...)
	return engine
}

func (engine *InterceptorEngine) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		c := engine.usPool.Get().(*UnaryServerConext)
		c.reset(ctx, req, info, handler)
		c.handlers = append(c.handlers, engine.usHandles...)
		c.handlers = append(c.handlers, func(c *UnaryServerConext) (interface{}, error) {
			return handler(c.Ctx, c.Req)
		})
		c.Next()
		engine.usPool.Put(c)
		return c.Reply, c.Err
	}
}
