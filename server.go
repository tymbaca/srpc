package srpc

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"

	"github.com/tymbaca/srpc/pkg/atomic"
	"github.com/tymbaca/srpc/pkg/pipe"
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

	l      ServerListener
	closed atomic.Value[bool]
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
	s.closed.Store(false)
	s.l = l
	defer s.l.Close()

	for !s.closed.Load() {
		if err := ctx.Err(); err != nil {
			return err
		}

		conn, err := s.l.Accept()
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

	return nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.l.Close()
}

func (s *Server) handleConn(ctx context.Context, conn ServerConn) (err error) {
	defer conn.Close()
	req := conn.Request()

	serviceName, methodName, ok := req.ServiceMethod.Split()
	if !ok {
		return conn.Send(ctx, respError(req, StatusInvalidServiceMethod, ""))
	}

	service, ok := s.services[serviceName]
	if !ok {
		return conn.Send(ctx, respError(req, StatusServiceNotFound, ""))
	}

	method, ok := service.methods[methodName]
	if !ok {
		return conn.Send(ctx, respError(req, StatusMethodNotFound, ""))
	}

	resp := s.call(method, ctx, req)

	return conn.Send(ctx, resp)
}

func (s *Server) call(m method, ctx context.Context, req Request) Response {
	// TODO: put metadata in context

	assert(m.val.Type().NumIn() == 2)
	assert(m.val.Type().In(0) == reflect.TypeFor[context.Context]())

	arg := reflect.New(m.val.Type().In(1)).Interface()
	err := s.codec.Decode(req.Body, arg)
	if err != nil {
		return respError(req, StatusBadRequest, "can't decode: %w", err)
	}

	retVals := m.val.Call(toValues(ctx, arg))
	assert(len(retVals) == 2)
	assert(reflect.TypeOf(retVals[1]) == reflect.TypeFor[error]())

	ret := retVals[0].Interface()
	err = retVals[1].Interface().(error)
	if err != nil {
		return respError(req, StatusErrorFromService, "error from service: %w", err)
	}

	return resp(req, StatusOK, pipe.ToReader(func(w io.Writer) error {
		return s.codec.Encode(w, ret)
	}))
}

func resp(req Request, statusCode StatusCode, body io.Reader) Response {
	resp := Response{
		ServiceMethod: req.ServiceMethod,
		Metadata:      Metadata{},
		StatusCode:    statusCode,
		Error:         nil,
		Body:          body,
	}

	return resp
}

func respError(req Request, statusCode StatusCode, errorMsg string, errorMsgArgs ...any) Response {
	resp := Response{
		ServiceMethod: req.ServiceMethod,
		Metadata:      Metadata{},
		StatusCode:    statusCode,
		Error:         fmt.Errorf(errorMsg, errorMsgArgs...),
		Body:          nil,
	}

	return resp
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

func toValues(ins ...any) []reflect.Value {
	outs := make([]reflect.Value, 0, len(ins))
	for _, in := range ins {
		v := reflect.ValueOf(in)
		outs = append(outs, v)
	}

	return outs
}
