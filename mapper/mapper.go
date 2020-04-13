package mapper

import (
	"errors"
	"github.com/sdghchj/sql-rules-engine/function"
	"github.com/sdghchj/sql-rules-engine/handler"
	"github.com/sdghchj/sql-rules-engine/parser"
	"github.com/sdghchj/sql-rules-engine/utils"
)

var ErrInvalidFromKey = errors.New("invalid from-key")
var ErrInvalidToKey = errors.New("invalid to-key")
var ErrInvalidConvertingFunc = errors.New("invalid converting-function")
var ErrNilParser = errors.New("nil parser")

type Mapper interface {
	handler.Handler
	SetFieldParser(parser parser.Parser) Mapper
	AddCurrentTimestampField(toKeyPath string, convert func(interface{}) interface{}) error
	AddConstField(value interface{}, toKeyPath string) error
	AddField(fromKeyPath, toKeyPath string) error
	AddFunctionField(fromKeyPath, toKeyPath string, convert func(interface{}) interface{}) error
	AddFieldFromMultiplePath(fromKeyPaths []string, toKeyPath string, convert func([]interface{}) interface{}) error
}

type mapper struct {
	fields []FieldValueConverter
	parser parser.Parser
	funcs  function.Functions
}

func NewMapper(funcs function.Functions) Mapper {
	return &mapper{parser: parser.DefaultGoParser, funcs: funcs}
}

func (m *mapper) SetFieldParser(parser parser.Parser) Mapper {
	m.parser = parser
	return m
}

func (m *mapper) AddCurrentTimestampField(toKeyPath string, convert func(interface{}) interface{}) error {
	if toKeyPath == "" || !utils.IsValidKeyPath(toKeyPath) {
		return ErrInvalidToKey
	}
	m.fields = append(m.fields, &fromCurrentTimestampFieldValue{toPath: toKeyPath, convert: convert})
	return nil
}

func (m *mapper) AddConstField(value interface{}, toKeyPath string) error {
	if toKeyPath == "" || !utils.IsValidKeyPath(toKeyPath) {
		return ErrInvalidToKey
	}
	m.fields = append(m.fields, &constantFieldValue{toPath: toKeyPath, value: value})
	return nil
}

func (m *mapper) AddField(fromKeyPath, toKeyPath string) error {
	if !utils.IsValidKeyPath(fromKeyPath) {
		if utils.IsLiteralNumber(fromKeyPath) {
			return m.AddConstField(utils.LiteralNumber(fromKeyPath), toKeyPath)
		} else if utils.IsLiteralString(fromKeyPath) {
			return m.AddConstField(utils.LiteralString(fromKeyPath), toKeyPath)
		}
		return m.AddFunctionField(fromKeyPath, toKeyPath, nil)
	}

	if toKeyPath == "" {
		toKeyPath = fromKeyPath
	} else if !utils.IsValidKeyPath(toKeyPath) {
		return ErrInvalidToKey
	}

	m.fields = append(m.fields, &funcFieldValueConverter{fromPath: fromKeyPath, toPath: toKeyPath})
	return nil
}

func (m *mapper) AddFunctionField(fromKeyPath, toKeyPath string, convert func(interface{}) interface{}) error {
	if m.parser == nil {
		return ErrNilParser
	}
	if fromKeyPath == "" {
		return ErrInvalidFromKey
	}

	resolver, err := m.parser.Parse(fromKeyPath, m.funcs)
	if err != nil {
		return err
	}

	if toKeyPath == "" {
		toKeyPath = utils.AdjustKeyPath(fromKeyPath)
	} else if !utils.IsValidKeyPath(toKeyPath) {
		return ErrInvalidToKey
	}

	m.fields = append(m.fields, &funcFieldValueConverter{fromPath: fromKeyPath, toPath: toKeyPath, resolver: resolver, convert: convert})
	return nil
}

func (m *mapper) AddFieldFromMultiplePath(fromKeyPaths []string, toKeyPath string, convert func([]interface{}) interface{}) error {
	if m.parser == nil {
		return ErrNilParser
	}

	fromPaths := map[string]parser.Resolver{}
	for _, fromPath := range fromKeyPaths {
		if fromPath == "" {
			return ErrInvalidFromKey
		}
		resolver, err := m.parser.Parse(fromPath, m.funcs)
		if err != nil {
			return err
		}
		fromPaths[fromPath] = resolver
	}

	if toKeyPath == "" || !utils.IsValidKeyPath(toKeyPath) {
		return ErrInvalidToKey
	}
	if convert == nil {
		return ErrInvalidConvertingFunc
	}

	m.fields = append(m.fields, &fromMultipleFieldValue{fromPaths: fromPaths, toPath: toKeyPath, convert: convert})
	return nil
}

func (m *mapper) Handle(src map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	} else if len(m.fields) == 0 {
		return map[string]interface{}{}
	}
	ret := map[string]interface{}{}
	for _, v := range m.fields {
		val := v.ConvertValue(src)
		if val == nil {
			continue //skip
		}
		toPath := v.ConvertToPath()
		if toPath == "*" {
			if temp, ok := val.(map[string]interface{}); ok {
				ret = temp
			}
		} else if utils.IsValidKeyPath(toPath) {
			utils.SetByPath(ret, v.ConvertToPath(), val)
		}
	}
	return ret
}

func (f *mapper) HandleAsync(src map[string]interface{}) {

}
