package function

import (
	"reflect"
	"strings"
)

type Functions interface {
	Init(i interface{}) Functions
	RegisterFunc(name string, f func(value []interface{}) interface{}) Functions
	Exists(name string) bool
	Call(name string, args []interface{}) interface{}
}

type functions struct {
	funcs map[string]func(value []interface{}) interface{}
}

var DefaultFunctions Functions

func init() {
	if DefaultFunctions != nil {
		return
	}
	DefaultFunctions = NewFunctions()
	DefaultFunctions.Init(&defaultFunctor)
}

func NewFunctions() Functions {
	return &functions{funcs: make(map[string]func(value []interface{}) interface{})}
}

func (fs *functions) Init(i interface{}) Functions {
	val := reflect.ValueOf(i)
	for i := 0; i < val.NumMethod(); i++ {
		if m, ok := val.Method(i).Interface().(func([]interface{}) interface{}); ok && m != nil {
			DefaultFunctions.RegisterFunc(val.Type().Method(i).Name, m)
		}
	}
	return fs
}

func (fs *functions) RegisterFunc(name string, f func([]interface{}) interface{}) Functions {
	fs.funcs[strings.ToLower(name)] = f
	return fs
}

func (fs *functions) Exists(name string) bool {
	_, ok := fs.funcs[strings.ToLower(name)]
	return ok
}

func (fs *functions) Call(name string, args []interface{}) interface{} {
	if f, ok := fs.funcs[strings.ToLower(name)]; ok {
		return f(args)
	}
	return nil
}
