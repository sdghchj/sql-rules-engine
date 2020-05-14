package filter

import (
	"encoding/json"
	"github.com/sdghchj/sql-rules-engine/function"
	"github.com/sdghchj/sql-rules-engine/handler"
	"github.com/sdghchj/sql-rules-engine/parser"
	"strings"
)

type FieldFilter interface {
	handler.Handler
	Parse(match string, handlers ...handler.EventHandler) error
	Match(obj interface{}) bool
	MatchJson(json string) bool
	ParseForAsyncHandlers(match string, asyncHandlers ...handler.AsyncEventHandler) error
}

type fieldFilter struct {
	parser        parser.Parser
	funcs         function.Functions
	resolver      parser.Resolver
	handlers      []handler.EventHandler
	asyncHandlers []handler.AsyncEventHandler
}

func NewFieldFilter(funcs function.Functions) FieldFilter {
	return &fieldFilter{parser: parser.DefaultGoParser, funcs: funcs}
}

func (f *fieldFilter) Parse(match string, handlers ...handler.EventHandler) error {
	r, err := f.parser.Parse(match, f.funcs)
	if err != nil {
		return err
	}
	f.resolver = r
	f.handlers = handlers
	return nil
}

func (f *fieldFilter) ParseForAsyncHandlers(match string, asyncHandlers ...handler.AsyncEventHandler) error {
	r, err := f.parser.Parse(match, f.funcs)
	if err != nil {
		return err
	}
	f.resolver = r
	f.asyncHandlers = asyncHandlers
	return nil
}

func (f *fieldFilter) Match(obj interface{}) bool {
	if f.resolver != nil {
		ret := f.resolver.Evaluate(obj)
		if ret == nil {
			return false
		}
		if b, ok := ret.(bool); ok && b {
			return true
		}
	}
	return false
}

func (f *fieldFilter) MatchJson(jsonText string) bool {
	decoder := json.NewDecoder(strings.NewReader(jsonText))
	//decoder.UseNumber()
	var src interface{}
	err := decoder.Decode(&src)
	if err != nil {
		return false
	}
	return f.Match(src)
}

func (f *fieldFilter) Handle(obj interface{}) interface{} {
	if f.Match(obj) {
		if f.handlers != nil && len(f.handlers) > 0 {
			for _, handler := range f.handlers {
				obj = handler(obj)
			}
		}
		return obj
	}
	return nil
}

func (f *fieldFilter) HandleAsync(obj interface{}) {
	if f.Match(obj) {
		if f.asyncHandlers != nil && len(f.asyncHandlers) > 0 {
			for _, handler := range f.asyncHandlers {
				go handler(obj)
			}
		}
	}
}
