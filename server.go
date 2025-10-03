package srpc

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sync/atomic"

	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/pkg/pipe"
)

func NewServer(codec Codec, opts ...ServerOption) *Server {
	s := &Server{
		services: make(map[string]service),
		codec:    codec,
		logger:   logger.NoopLogger{},
	}
	s.closed.Store(false)

	for _, o := range opts {
		o(s)
	}

	return s
}

type Server struct {
	codec    Codec
	services map[string]service

	l      Listener
	closed atomic.Bool

	logger logger.Logger
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
	RegisterWithName(s, impl, "")
}

func RegisterWithName[T any](s *Server, impl T, name string) {
	t := reflect.TypeFor[T]()
	v := reflect.ValueOf(impl)

	if name == "" {
		name = t.Name()
		if name == "" {
			name = t.Elem().Name()
		}
	}

	service := service{
		name: name,
		typ:  t,
		val:  v,
	}
	service.methods = getMethods(v)

	s.services[name] = service
}

func (s *Server) Start(ctx context.Context, l Listener) error {
	s.l = l
	defer s.Close()

	for !s.closed.Load() {
		if err := ctx.Err(); err != nil {
			return err
		}

		conn, err := s.l.Accept()
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}

		go func() {
			err := s.handleConn(ctx, conn)
			if err != nil {
				s.logger.Error(err.Error())
			}
		}()
	}

	return nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.l != nil {
		return s.l.Close()
	}

	return nil
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

	argVal := reflect.New(m.val.Type().In(1))
	err := s.codec.Decode(req.Body, argVal.Interface())
	if err != nil {
		return respError(req, StatusBadRequest, "can't decode: %w", err)
	}

	retVals := m.val.Call(toValues(ctx, argVal.Elem().Interface()))
	// assert(len(retVals) == 2)
	// assert(reflect.TypeOf(retVals[1]) == reflect.TypeFor[error]())

	ret := retVals[0].Interface()
	if !retVals[1].IsNil() {
		return respError(req, StatusErrorFromService, "error from service: %w", retVals[1].Interface().(error))
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
		Error:         tern(errorMsg != "", fmt.Errorf(errorMsg, errorMsgArgs...), fmt.Errorf("code: %s", statusCode)),
		Body:          nil,
	}

	return resp
}

func getMethods(v reflect.Value) map[string]method {
	methods := make(map[string]method)
	for i := range v.NumMethod() {
		m := v.Method(i)
		name := v.Type().Method(i).Name

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
