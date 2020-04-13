package handler

type EventHandler func(map[string]interface{}) map[string]interface{}
type AsyncEventHandler func(map[string]interface{})

type Handler interface {
	Handle(src map[string]interface{}) map[string]interface{}
	HandleAsync(src map[string]interface{})
}
