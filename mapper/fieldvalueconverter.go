package mapper

import (
	"github.com/sdghchj/sql-rules-engine/parser"
	"github.com/sdghchj/sql-rules-engine/utils"
	"time"
)

type FieldValueConverter interface {
	ConvertValue(map[string]interface{}) interface{}
	ConvertToPath() string
}

type fromCurrentTimestampFieldValue struct {
	toPath  string
	convert func(interface{}) interface{}
}

func (v *fromCurrentTimestampFieldValue) ConvertValue(src map[string]interface{}) interface{} {
	t := time.Now().Unix()
	if v.convert != nil {
		return v.convert(t)
	}
	return t
}

func (v *fromCurrentTimestampFieldValue) ConvertToPath() string {
	return v.toPath
}

type constantFieldValue struct {
	value  interface{}
	toPath string
}

func (v *constantFieldValue) ConvertValue(src map[string]interface{}) interface{} {
	return v.value
}

func (v *constantFieldValue) ConvertToPath() string {
	return v.toPath
}

type funcFieldValueConverter struct {
	fromPath string
	toPath   string
	resolver parser.Resolver
	convert  func(interface{}) interface{}
}

func (v *funcFieldValueConverter) ConvertValue(src map[string]interface{}) interface{} {
	if v.fromPath == "*" {
		return src
	} else if utils.IsLiteralString(v.fromPath) {
		return utils.LiteralString(v.fromPath)
	} else if utils.IsLiteralNumber(v.fromPath) {
		utils.LiteralNumber(v.fromPath)
	} else if v.resolver != nil {
		return v.resolver.Evaluate(src)
	}
	val := utils.GetByPath(src, v.fromPath)
	if v.convert != nil {
		val = v.convert(val)
	}
	return val
}

func (v *funcFieldValueConverter) ConvertToPath() string {
	if v.toPath != "" {
		return v.toPath
	}
	return v.fromPath
}

type fromMultipleFieldValue struct {
	fromPaths map[string]parser.Resolver
	toPath    string
	convert   func([]interface{}) interface{}
}

func (v *fromMultipleFieldValue) ConvertValue(src map[string]interface{}) interface{} {
	n := len(v.fromPaths)
	if n == 0 {
		return nil
	}
	if v.convert == nil {
		return nil
	}
	values := make([]interface{}, n)
	i := 0
	for path, resolver := range v.fromPaths {
		if path == "*" {
			values[i] = src
		} else if resolver != nil {
			values[i] = resolver.Evaluate(src)
		} else {
			values[i] = utils.GetByPath(src, path)
		}
		i++
	}
	return v.convert(values)
}

func (v *fromMultipleFieldValue) ConvertToPath() string {
	return v.toPath
}
