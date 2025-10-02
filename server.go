package srpc

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
)

func NewServer(codec Codec) *Server {
	return &Server{
		services: make(map[string]service),
		codec:    codec,
	}
}

type Server struct {
	codec    Codec
	services map[string]service
}

type service struct {
	name string
	typ  reflect.Type
	val  reflect.Value

	methods map[string]method
}

type method struct {
	val reflect.Value
}

func Register[T any](s *Server, impl T) {
	t := reflect.TypeFor[T]()
	v := reflect.ValueOf(impl)

	name := t.Name()
	if name == "" {
		slog.Debug(`t.Name() was "", getting t.Elem().Name()`)
		name = t.Elem().Name()
	}

	service := service{
		name: name,
		typ:  t,
		val:  v,
	}
	service.methods = getMethods(v)

	s.services[name] = service
}

func (s *Server) Start(ctx context.Context, l ServerListener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error(err.Error())
		}

		go func() {
			err := s.handleConn(ctx, conn)
			if err != nil {
				slog.Error(err.Error())
			}
		}()
	}
}

func (s *Server) handleConn(ctx context.Context, conn ServerConn) (err error) {
	defer conn.Close()
	req := conn.Request()

	serviceName, methodName, ok := req.ServiceMethod.Split()
	if !ok {
		return replyError(ctx, conn, req, StatusInvalidServiceMethod, "")
	}

	service, ok := s.services[serviceName]
	if !ok {
		return replyError(ctx, conn, req, StatusServiceNotFound, "")
	}

	method, ok := service.methods[methodName]
	if !ok {
		return replyError(ctx, conn, req, StatusMethodNotFound, "")
	}

	resp, err := method.call(ctx, req)
	if err != nil {
		return reply(ctx, conn, resp)
	}

	return nil
}

func (s *Server) call(m method, ctx context.Context, req Request) (Response, error) {
	// TODO: put metadata in context
	arg := reflect.New(m.val.Type().In(1)).Interface()
	err := s.codec.Decode(req.Body, arg)

	if err != nil {
		return Response{}
	}

}

func reply(ctx context.Context, conn ServerConn, resp Response) error {
	return conn.Send(ctx, resp)
}

func replyError(ctx context.Context, conn ServerConn, req Request, statusCode StatusCode, errorMsg string, errorMsgArgs ...any) error {
	resp := Response{
		ServiceMethod: req.ServiceMethod,
		Metadata:      Metadata{},
		StatusCode:    statusCode,
		Error:         fmt.Errorf(errorMsg, errorMsgArgs...),
		Body:          nil,
	}
	conn.Send(ctx)
}

func getMethods(v reflect.Value) map[string]method {
	methods := make(map[string]method)
	for i := range v.NumMethod() {
		m := v.Method(i)
		name := m.Type().Name()

		if isSuitableMethod(m) {
			methods[name] = method{val: m}
		}
	}

	return methods
}

func isSuitableMethod(method reflect.Value) bool {
	typ := method.Type()
	if typ.NumIn() != 2 {
		return false
	}

	if typ.In(0) != reflect.TypeFor[context.Context]() {
		return false
	}

	if typ.NumOut() != 2 {
		return false
	}

	if typ.Out(1) != reflect.TypeFor[error]() {
		return false
	}

	return true
}
