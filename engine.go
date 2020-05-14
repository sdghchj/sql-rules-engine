package engine

import (
	"encoding/json"
	"errors"
	"github.com/sdghchj/sql-rules-engine/function"
	"github.com/sdghchj/sql-rules-engine/handler"
	"github.com/sdghchj/sql-rules-engine/rule"
	"strings"
	"sync"
)

type Engine interface {
	ParseRuleEvent(name string, match string, handlers ...handler.EventHandler) (rule.Rule, error)
	ParseRuleAsyncEvent(name string, match string, asyncHandlers ...handler.AsyncEventHandler) (rule.Rule, error)
	ParseSql(sql string) (rule.Rule, error)
	RegisterRuleFunction(name string, fun func(rule.Rule) func(values []interface{}) interface{}) Engine
	PutRule(name string, rule rule.Rule) Engine
	//Handle(map[string]interface{}) map[string]interface{}
	HandleAsync(obj interface{})
	HandleJsonAsync(jsonText string) error
	ConvertJson(name string, jsonText string) (string, error)
}

type jsonEngine struct {
	defaultPretty bool
	rules         sync.Map //map[string]rule.Rule
	//	rulesLock     sync.Mutex
	funcs map[string]func(rule.Rule) func(values []interface{}) interface{}
}

var ErrNoRuleFound = errors.New("no rule found")

func NewJsonEngine(defaultPretty bool) Engine {
	return &jsonEngine{defaultPretty: defaultPretty}
}

func (e *jsonEngine) RegisterRuleFunction(name string, fun func(rule.Rule) func(values []interface{}) interface{}) Engine {
	if e.funcs == nil {
		e.funcs = make(map[string]func(rule.Rule) func(values []interface{}) interface{})
	}
	e.funcs[name] = fun
	return e
}

func (e *jsonEngine) ParseRuleEvent(name string, match string, handlers ...handler.EventHandler) (rule.Rule, error) {
	jsonRule := rule.NewJsonRule(e.defaultPretty)
	err := jsonRule.AddEventHandler(match, handlers...)
	if err != nil {
		return nil, err
	}
	e.PutRule(name, jsonRule)
	return jsonRule, nil
}

func (e *jsonEngine) ParseRuleAsyncEvent(name string, match string, asyncHandlers ...handler.AsyncEventHandler) (rule.Rule, error) {
	jsonRule := rule.NewJsonRule(e.defaultPretty)
	err := jsonRule.AddEventAsyncHandler(match, asyncHandlers...)
	if err != nil {
		return nil, err
	}
	e.PutRule(name, jsonRule)
	return jsonRule, nil
}

func (e *jsonEngine) ParseSql(sql string) (rule.Rule, error) {
	jsonRule := rule.NewJsonRule(e.defaultPretty)

	var ruleFunctions function.Functions
	if len(e.funcs) > 0 {
		ruleFunctions = function.NewFunctions()
		for name, fun := range e.funcs {
			ruleFunctions.RegisterFunc(name, fun(jsonRule))
		}
	}

	err := jsonRule.AddConvertHandlerBySql(sql, ruleFunctions)
	if err != nil {
		return nil, err
	}
	e.PutRule(jsonRule.Name(), jsonRule)
	return jsonRule, nil
}

func (e *jsonEngine) getRule(name string) (r rule.Rule) {
	if i, ok := e.rules.Load(name); ok {
		if r, ok = i.(rule.Rule); ok {
			return r
		}
	}
	return nil
}

func (e *jsonEngine) PutRule(name string, rule rule.Rule) Engine {
	if rule == nil {
		e.rules.Delete(name)
		return e
	}
	e.rules.Store(name, rule)
	return e
}

/*func (e *jsonEngine) Handle(src map[string]interface{}) map[string]interface{}{
	return src
}*/

func (e *jsonEngine) HandleAsync(obj interface{}) {
	e.rules.Range(func(key, value interface{}) bool {
		if r, ok := value.(rule.Rule); ok {
			r.HandleAsync(obj)
		}
		return true
	})
	return
}

func (e *jsonEngine) HandleJsonAsync(jsonText string) error {
	decoder := json.NewDecoder(strings.NewReader(jsonText))
	//decoder.UseNumber()
	src := map[string]interface{}{}
	err := decoder.Decode(&src)
	if err != nil {
		return err
	}
	e.HandleAsync(src)
	return nil
}

func (e *jsonEngine) ConvertJson(name string, jsonText string) (string, error) {
	if rule := e.getRule(name); rule != nil {
		return rule.ConvertJson(jsonText)
	}
	return "", ErrNoRuleFound
}
