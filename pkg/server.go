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
	upgrader *websocket.Upgrader
	services map[string]*Service
}

func NewServer() *Server {
	return &Server{
		upgrader: new(websocket.Upgrader),
		services: make(map[string]*Service),
	}
}

func (s *Server) Listen(addr string) {
	http.HandleFunc("/", s.Serve)
	http.ListenAndServe(addr, nil)
}

func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer c.Close()
	socket := NewSocket(c)

	for {
		message := <-socket.Incoming
		reply, err := s.Dispatch(socket, message)
		if err != nil {
			socket.Outgoing <- Response{
				Type:    Error,
				Payload: err.Error(),
			}
		}
		if reply != nil {
			socket.Incoming <- *reply
		}
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

	method.Func.Call([]reflect.Value{
		reflect.ValueOf(service.recv),
		reflect.ValueOf(socket),
	})

	return nil, nil
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
