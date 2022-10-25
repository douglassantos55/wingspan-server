package pkg

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	ErrServiceNotFound    = errors.New("Service not found")
	ErrMethodNotFound     = errors.New("Method not found")
	ErrNilServiceProvided = errors.New("Cannot register nil service")
	ErrNoMethodsAvailable = errors.New("There are no methods exported for this service")
)

type Service struct {
	recv    any
	methods map[string]reflect.Method
}

type Server struct {
	server   *http.Server
	upgrader *websocket.Upgrader
	services map[string]*Service
}

func NewServer() *Server {
	return &Server{
		server:   new(http.Server),
		upgrader: new(websocket.Upgrader),
		services: make(map[string]*Service),
	}
}

func (s *Server) Close() {
	s.server.Close()
}

func (s *Server) Listen(addr string) {
	s.server.Addr = addr
	s.server.Handler = http.HandlerFunc(s.Serve)
	s.server.ListenAndServe()
}

func (s *Server) handleMessage(socket *Socket, message Message) {
	reply, err := s.Dispatch(socket, message)
	if err != nil {
		socket.Outgoing <- Response{
			Type:    Error,
			Payload: err.Error(),
		}
		return
	}
	if reply != nil {
		s.handleMessage(socket, *reply)
	}
}

func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer c.Close()
	socket := NewSocket(c)

	for message := range socket.Incoming {
		s.handleMessage(socket, message)
	}
}

func (s *Server) Dispatch(socket *Socket, message Message) (*Message, error) {
	parts := strings.Split(message.Method, ".")
	service, ok := s.services[parts[0]]
	if !ok {
		return nil, ErrServiceNotFound
	}

	methods := service.methods
	method, ok := methods[parts[1]]
	if !ok {
		return nil, ErrMethodNotFound
	}

	params := []reflect.Value{
		reflect.ValueOf(service.recv),
		reflect.ValueOf(socket),
	}

	if message.Params != nil {
		params = append(params, reflect.ValueOf(message.Params))
	}

	var err error

	returnValues := method.Func.Call(params)
	reply := returnValues[0].Interface().(*Message)
	if returnValues[1].Interface() != nil {
		err = returnValues[1].Interface().(error)
	}

	return reply, err
}

func (s *Server) Register(name string, service any) error {
	t := reflect.TypeOf(service)
	if t == nil {
		return ErrNilServiceProvided
	}

	methods := make(map[string]reflect.Method)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if m.IsExported() {
			methods[m.Name] = m
		}
	}

	if len(methods) == 0 {
		return ErrNoMethodsAvailable
	}

	s.services[name] = &Service{
		recv:    service,
		methods: methods,
	}

	return nil
}
