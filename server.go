package srpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/tymbaca/srpc/codec"
	"github.com/tymbaca/srpc/enc"
	"github.com/tymbaca/srpc/internal/fx"
	"github.com/tymbaca/srpc/internal/pipe"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/metadata"
	"github.com/tymbaca/srpc/status"
	"github.com/tymbaca/srpc/transport"
)

// NewServer creates new [Server] with provided codec and options.
func NewServer(codec codec.Codec, opts ...ServerOption) *Server {
	s := &Server{
		enc:            encContext,
		codec:          codec,
		logger:         logger.NoopLogger{},
		streamResponse: false,
		services:       make(map[string]service),
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

// Server is the RPC server. It holds registered services with methods
// and calls them when it gets the request with matching ServiceName.
type Server struct {
	enc              enc.Context
	codec            codec.Codec
	logger           logger.Logger
	streamResponse   bool
	connErrorHandler func(error) error

	services   map[string]service
	l          transport.Listener
	middlwares []Middleware
}

type Handler func(ctx context.Context, reqmeta enc.Request, req any) (resp any, err error)

type Middleware func(next Handler) Handler

type service struct {
	name string
	typ  reflect.Type
	val  reflect.Value

	methods map[string]method
}

type method struct {
	val reflect.Value
}

// Register registers the provided service impl into the server.
// It registers it with the name of provided T, e.g.:
// - `Register(s, MyServiceImpl{})` will register it as "MyServiceImpl",
// - `Register[MyService](s, MyServiceImpl{})` will register it as "MyService".
func Register[T any](s *Server, impl T) {
	registerWithName(s, impl, "")
}

// RegisterWithName registers the provided service impl into the server with provided name.
// If name == "", nothing will happen.
func RegisterWithName[T any](s *Server, impl T, name string) {
	if name == "" {
		return
	}
	registerWithName(s, impl, name)
}

func registerWithName[T any](s *Server, impl T, name string) {
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

// Start starts server with provided listener. It blocks until server (listener) get closed.
// In normal scenario, if [Server.Close] was called, Start will exit with nil.
// See [transport.Listener.Accept] for details.
func (s *Server) Start(ctx context.Context, l transport.Listener) error {
	s.l = l
	ctx, cancel := fx.CloseOnCancel(ctx, s)
	defer cancel()

	for {
		conn, err := s.l.Accept()
		if errors.Is(err, transport.ErrListenerClosed) {
			return nil
		}
		if errors.Is(err, transport.ErrListenerBadClose) {
			return err
		}
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}

		go func() {
			ctx, cancel := fx.CloseOnCancel(ctx, conn)
			defer cancel()

			err := s.handleConn(ctx, conn)
			if err != nil {
				if s.connErrorHandler != nil {
					if err = s.connErrorHandler(err); err != nil {
						s.logger.Error(err.Error())
					}
					return
				} else {
					s.logger.Error(err.Error())
				}
			}
		}()
	}
}

// Close closes currently running listener causing [Server.Start] to exit with nil.
func (s *Server) Close() error {
	if s.l != nil {
		return s.l.Close()
	}

	return nil
}

func (s *Server) handleConn(ctx context.Context, conn transport.Conn) (err error) {
	req, err := enc.ReadRequest(s.enc, conn)
	if err != nil {
		return err
	}

	resp := s.handleReq(ctx, req)

	if err := drain(req.Body); err != nil {
		return fmt.Errorf("drain remaining request body: %w", err)
	}

	return enc.WriteResponse(s.enc, conn, resp)
}

func (s *Server) handleReq(ctx context.Context, req enc.Request) (resp enc.Response) {
	serviceName, methodName, ok := req.ServiceMethod.Split()
	if !ok {
		return respErrorf(nil, status.InvalidServiceMethod, "")
	}

	service, ok := s.services[serviceName]
	if !ok {
		return respErrorf(nil, status.ServiceNotFound, "")
	}

	method, ok := service.methods[methodName]
	if !ok {
		return respErrorf(nil, status.MethodNotFound, "")
	}

	return s.call(method, ctx, req)
}

func (s *Server) call(method method, ctx context.Context, reqmeta enc.Request) enc.Response {
	ctx = metadata.ToContext(ctx, reqmeta.Metadata.Map())

	var respMD metadata.Metadata
	ctx = metadata.ResponseToContext(ctx, &respMD)

	req := reflect.New(method.val.Type().In(1))
	err := s.codec.Decode(reqmeta.Body, req.Interface())
	if err != nil {
		return respErrorf(respMD, status.InvalidArgument, "can't decode input values: %w", err)
	}

	var h Handler = func(ctx context.Context, reqmeta enc.Request, req any) (resp any, err error) {
		retVals := method.val.Call(toValues(ctx, req))

		ret := retVals[0].Interface()
		if !retVals[1].IsNil() {
			err := retVals[1].Interface().(error)
			return nil, err
		}

		return ret, nil
	}

	for _, m := range s.middlwares {
		h = m(h)
	}

	ret, err := h(ctx, reqmeta, req.Elem().Interface())
	if err != nil {
		if code, ok := status.FromError(err); ok {
			// to prevent duplicate code description on client, e.g. "InvalidArgument: InvalidArgument: <errorText>"
			errMsg := strings.TrimSuffix(err.Error(), code.String()+": ") // (yes, i know)
			return respError(respMD, code, errMsg)
		}

		return respErrorf(respMD, status.ErrorFromService, "error from service: %w", err)
	}

	if s.streamResponse {
		return resp(respMD, status.OK,
			pipe.ToReader(func(w io.Writer) error { return s.codec.Encode(w, ret) }),
		)
	}

	bodyBuf := bytes.NewBuffer(nil) // TODO: use sync.Pool
	if err := s.codec.Encode(bodyBuf, ret); err != nil {
		return respErrorf(respMD, status.InternalError, "can't encode return values: %w", err)
	}

	return resp(respMD, status.OK, bodyBuf)
}

func resp(md metadata.Metadata, statusCode status.Code, body io.Reader) enc.Response {
	resp := enc.Response{
		Metadata:   enc.NewMetadata(md),
		StatusCode: statusCode,
		Error:      nil,
		Body:       body,
	}

	return resp
}

func respError(md metadata.Metadata, statusCode status.Code, errorMsg string) enc.Response {
	resp := enc.Response{
		Metadata:   enc.NewMetadata(md),
		StatusCode: statusCode,
		Error:      errors.New(errorMsg),
		Body:       nil,
	}

	return resp
}

func respErrorf(md metadata.Metadata, statusCode status.Code, errorMsg string, errorMsgArgs ...any) enc.Response {
	resp := enc.Response{
		Metadata:   enc.NewMetadata(md),
		StatusCode: statusCode,
		Error:      fmt.Errorf(errorMsg, errorMsgArgs...),
		Body:       nil,
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

func drain(r io.Reader) error {
	if r == nil {
		return nil
	}
	_, err := io.Copy(io.Discard, r)
	return err
}
