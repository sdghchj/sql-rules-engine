package handler

type EventHandler func(obj interface{}) interface{}
type AsyncEventHandler func(obj interface{})

type Handler interface {
	Handle(obj interface{}) interface{}
	HandleAsync(obj interface{})
}
