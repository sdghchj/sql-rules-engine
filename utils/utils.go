package utils

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var ErrTypeError = errors.New("type error")

func GetInt64(val interface{}) (int64, error) {
	switch n := val.(type) {
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case int8:
		return int64(n), nil
	case uint8:
		return int64(n), nil
	case int16:
		return int64(n), nil
	case uint16:
		return int64(n), nil
	case json.Number:
		v, err := n.Int64()
		if err != nil {
			panic(err)
		}
		return v, nil
	}
	return 0, ErrTypeError
}

func MustGetInt64(val interface{}) int64 {
	n, err := GetInt64(val)
	if err != nil {
		panic(err)
	}
	return n
}

func GetFloat64(val interface{}) (float64, error) {
	switch n := val.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case uint32:
		return float64(n), nil
	case int8:
		return float64(n), nil
	case uint8:
		return float64(n), nil
	case int16:
		return float64(n), nil
	case uint16:
		return float64(n), nil
	case json.Number:
		v, err := n.Float64()
		if err != nil {
			panic(err)
		}
		return v, nil
	}
	return 0, ErrTypeError
}

func MustGetFloat64(val interface{}) float64 {
	n, err := GetFloat64(val)
	if err != nil {
		panic(err)
	}
	return n
}

func IsLiteralNumber(keyPath string) bool {
	ok, err := regexp.Match(`^[0-9]+.{0,1}[0-9]*$`, []byte(keyPath))
	if err != nil || !ok {
		return false
	}
	return true
}

func LiteralNumber(keyPath string) interface{} {
	if strings.ContainsRune(keyPath, '.') {
		f, err := strconv.ParseFloat(keyPath, 64)
		if err != nil {
			return nil
		}
		return f
	}
	i, err := strconv.ParseInt(keyPath, 10, 64)
	if err != nil {
		return nil
	}
	return i
}

func IsLiteralString(keyPath string) bool {
	return keyPath[0] == keyPath[len(keyPath)-1] && (keyPath[0] == '"' || keyPath[0] == '\'')
}

func LiteralString(keyPath string) string {
	return keyPath[1 : len(keyPath)-1]
}

func IsValidKeyPath(keyPath string) bool {
	if keyPath == "" {
		return false
	}

	if keyPath == "*" {
		return true
	}

	keys := strings.Split(keyPath, ".")
	for _, key := range keys {
		//variable name
		ok, err := regexp.Match(`^[a-zA-Z_][a-zA-Z0-9_]*$`, []byte(key))
		if err != nil || !ok {
			return false
		}
	}
	return true
}

func IsValidKeyPaths(keyPaths []string) bool {
	if len(keyPaths) == 0 {
		return false
	}

	for _, path := range keyPaths {
		if IsValidKeyPath(path) {
			return false
		}
	}
	return true
}

func AdjustKeyPath(keyPath string) string {
	key := []byte(keyPath)
	for i := 0; i < len(key); i++ {
		if key[i] >= 'A' && key[i] <= 'Z' || key[i] >= 'a' && key[i] <= 'z' || key[i] == '_' || key[i] >= '0' && key[i] <= '9' && i > 0 && key[i-1] != '.' {

		} else {
			key[i] = '_'
		}
	}
	return string(key)
}

func GetByPath(obj interface{}, keyPath string) interface{} {
	keys := strings.Split(keyPath, ".")
	var val interface{}
	for _, k := range keys {
		if obj != nil {
			if mp, ok := obj.(map[string]interface{}); ok {
				if val, ok = mp[k]; ok {
					obj, _ = val.(map[string]interface{})
					continue
				}
			}
		}
		return nil
	}
	return val
}

func SetByPath(obj map[string]interface{}, keyPath string, value interface{}) {
	keys := strings.Split(keyPath, ".")
	depth := len(keys)
	var val interface{}
	if depth > 1 {
		var ok bool
		for i := 0; i < depth-1; i++ {
			if val, ok = obj[keys[i]]; ok {
				if obj, ok = val.(map[string]interface{}); ok {
					continue
				}
			}
			temp := map[string]interface{}{}
			obj[keys[i]] = temp
			obj = temp
		}
	}
	obj[keys[depth-1]] = value
}
