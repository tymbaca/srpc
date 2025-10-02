package srpc

import (
	"context"
	"reflect"
)

func NewServer(l ServerListener, codec Codec) *Server {
	return &Server{
		services: make(map[string]service),
	}
}

type Server struct {
	listener ServerListener
	codec    Codec
	services map[string]service
}

func Register[T any](s *Server, impl T) {
	t := reflect.TypeFor[T]()
	v := reflect.ValueOf(impl)

	service := service{
		typ: t,
		val: v,
	}
	service.methods = getMethods(v)

	s.services[t.Name()] = service
}

func (s *Server) Start(ctx context.Context) error {
	for {
		s.listener.Accept()
	}
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

type service struct {
	typ reflect.Type
	val reflect.Value

	methods map[string]method
}

type method struct {
	val reflect.Value
}

func (m *method) call() {
}
