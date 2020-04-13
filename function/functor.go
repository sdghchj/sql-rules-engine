package function

import (
	"github.com/sdghchj/sql-rules-engine/utils"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

type functor int

var defaultFunctor functor

func (*functor) Len(args []interface{}) (length interface{}) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	length = len(args)
	if length == 1 {
		if args[0] == nil {
			return nil
		}
		return reflect.ValueOf(args[0]).Len()
	}
	return length
}

func (*functor) Sum(args []interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			ret = 0
		}
	}()

	length := len(args)
	if length == 0 {
		return 0
	}

	if length == 1 {
		if args[0] == nil {
			return nil
		}
		valType := reflect.ValueOf(args[0])
		switch valType.Kind() {
		case reflect.Array, reflect.Slice:
			var sum float64 = 0
			length = valType.Len()
			for i := 0; i < length; i++ {
				sum += utils.MustGetFloat64(valType.Index(i).Interface())
			}
			ret = sum
			return
		default:
			return valType.Float()
		}
	}

	var sum float64 = 0
	for i := 0; i < length; i++ {
		if args[i] == nil {
			continue
		}
		sum += utils.MustGetFloat64(args[i])
	}
	ret = sum
	return
}

func (*functor) Average(args []interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			ret = 0
		}
	}()

	length := len(args)
	if length == 0 {
		return nil
	}

	if length == 1 {
		if args[0] == nil {
			return nil
		}
		valType := reflect.ValueOf(args[0])
		var sum float64 = 0
		switch valType.Kind() {
		case reflect.Array, reflect.Slice:
			length = valType.Len()
			for i := 0; i < length; i++ {
				sum += utils.MustGetFloat64(valType.Index(i).Interface())
			}
			return sum / float64(length)
		default:
			return valType.Float()
		}
	}

	var sum float64 = 0
	for i := 0; i < length; i++ {
		if args[i] == nil {
			continue
		}
		sum += utils.MustGetFloat64(args[i])
	}
	return sum / float64(length)
}

func findExtremum(args []interface{}, cmp func(val, base float64) bool) (ret interface{}) {
	length := len(args)
	if length == 0 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	if length == 1 {
		valType := reflect.ValueOf(args[0])
		switch valType.Kind() {
		case reflect.Array, reflect.Slice:
			length = valType.Len()
			if length == 0 {
				return nil
			} else if length == 1 {
				return utils.MustGetFloat64(valType.Index(0).Interface())
			}
			var max = utils.MustGetFloat64(valType.Index(0).Interface())
			for i := 1; i < length; i++ {
				val := utils.MustGetFloat64(valType.Index(i).Interface())
				if cmp(val, max) {
					max = val
				}
			}
			ret = max
			return
		default:
			return nil
		}
	}

	var max = utils.MustGetFloat64(args[0])
	for i := 1; i < length; i++ {
		val := utils.MustGetFloat64(args[i])
		if cmp(val, max) {
			max = val
		}
	}
	ret = max
	return
}

func (*functor) Max(args []interface{}) (ret interface{}) {
	return findExtremum(args, func(val, base float64) bool {
		return val > base
	})
}

func (*functor) Min(args []interface{}) (ret interface{}) {
	return findExtremum(args, func(val, base float64) bool {
		return val < base
	})
}

func (*functor) Array(args []interface{}) (ret interface{}) {
	return args
}

func (*functor) Substr(args []interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	n := len(args)
	if n == 0 || args[0] == nil {
		return nil
	}

	text := reflect.ValueOf(args[0]).String()
	if n >= 3 {
		pos := reflect.ValueOf(args[1]).Int()
		length := reflect.ValueOf(args[2]).Int()
		return text[pos : pos+length]
	} else if n == 2 {
		pos := reflect.ValueOf(args[1]).Int()
		return text[pos:]
	}
	return text
}

func (*functor) InRange(args []interface{}) (ret interface{}) {
	if len(args) < 3 {
		return nil
	}

	{
		target, err := utils.GetFloat64(args[0])
		if err != nil {
			goto STRING
		}
		left, err := utils.GetFloat64(args[1])
		if err != nil {
			goto STRING
		}
		right, err := utils.GetFloat64(args[2])
		if err != nil {
			goto STRING
		}

		return left <= right && target >= left && target < right || (left > right && (target >= left || target < right))
	}

STRING:
	{
		target, ok := args[0].(string)
		if !ok {
			return nil
		}
		left, ok := args[1].(string)
		if !ok {
			return nil
		}
		right, ok := args[2].(string)
		if !ok {
			return nil
		}
		return left <= right && target >= left && target < right || (left > right && (target >= left || target < right))
	}
}

func (*functor) Timestamp(args []interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	length := len(args)
	if length == 0 {
		return nil
	} else if length > 6 {
		length = 6
	}

	var dt = [6]int{0, 0, 0, 0, 0, 0}
	for i := 0; i < length; i++ {
		dt[i] = int(utils.MustGetFloat64(args[i]))
	}
	return time.Date(dt[0], time.Month(dt[1]), dt[2], dt[3], dt[4], dt[5], 0, time.Local).Unix()
}

func (*functor) CurrentTimestamp(args []interface{}) interface{} {
	return time.Now().Unix()
}

func (*functor) Year(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return time.Unix(int64(utils.MustGetFloat64(args[0])), 0).Year()
}

func (*functor) Month(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return time.Unix(int64(utils.MustGetFloat64(args[0])), 0).Month()
}

func (*functor) Day(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return time.Unix(int64(utils.MustGetFloat64(args[0])), 0).Day()
}

func (*functor) Hour(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return time.Unix(int64(utils.MustGetFloat64(args[0])), 0).Hour()
}

func (*functor) Minute(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return time.Unix(int64(utils.MustGetFloat64(args[0])), 0).Minute()
}

func (*functor) Second(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return time.Unix(int64(utils.MustGetFloat64(args[0])), 0).Second()
}

func (*functor) Regex(args []interface{}) interface{} {
	if len(args) < 2 || args[0] == nil || args[1] == nil {
		return false
	}

	text, ok := args[0].(string)
	if !ok {
		return false
	}

	pattern, ok := args[1].(string)
	if !ok {
		return false
	}

	match, err := regexp.MatchString(pattern, text)
	return err == nil && match
}

const DIFF = 0.000001

func (*functor) In(args []interface{}) (ret interface{}) {
	length := len(args)
	if length < 2 || args[0] == nil || args[1] == nil {
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			if length == 2 {
				ret = args[0] == args[1]
			} else {
				ret = false
			}
		}
	}()

	if length == 2 {
		arrayType := reflect.ValueOf(args[1])
		n := arrayType.Len()
		for i := 0; i < n; i++ {
			if arrayType.Index(i).Interface() == args[0] {
				return true
			}
		}
	} else {
		//really fuck golang's type and json take all number as float64 by default
		if num, err := utils.GetFloat64(args[0]); err == nil {
			var rnum float64
			for i := 1; i < length; i++ {
				if rnum, err = utils.GetFloat64(args[i]); err == nil && math.Abs(num-rnum) < DIFF {
					return true
				}
			}
		} else {
			for i := 1; i < length; i++ {
				if args[0] == args[i] {
					return true
				}
			}
		}
	}

	return false
}

func (*functor) Int(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	if text, ok := args[0].(string); ok {
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return n
		}
		return nil
	} else if n, err := utils.GetInt64(args[0]); err == nil {
		return n
	} else {
		return int64(reflect.ValueOf(args[0]).Float())
	}

	return nil
}

func (*functor) Float(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	if text, ok := args[0].(string); ok {
		if n, err := strconv.ParseFloat(text, 64); err == nil {
			return n
		}
		return nil
	} else if n, err := utils.GetFloat64(args[0]); err == nil {
		return n
	}

	return nil
}

func (*functor) String(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	if n, err := utils.GetInt64(args[0]); err == nil {
		return strconv.FormatInt(n, 10)
	} else if n, err := utils.GetFloat64(args[0]); err == nil {
		return strconv.FormatFloat(n, 'f', -1, 64)
	}

	if text, ok := args[0].(string); ok {
		return text
	}

	return nil
}

func (*functor) Abs(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		defer func() {
			if err := recover(); err != nil {
				ret = nil
			}
		}()
		if err := recover(); err != nil {
			ret = math.Abs(utils.MustGetFloat64(args[0]))
		}
	}()

	n := utils.MustGetInt64(args[0])
	if n < 0 {
		return -n
	}
	return n
}

func (*functor) NullIf(args []interface{}) interface{} {
	if len(args) < 2 || args[0] == nil || args[1] == nil {
		return nil
	}

	if num, err := utils.GetFloat64(args[0]); err == nil {
		if rnum, err := utils.GetFloat64(args[1]); err == nil && math.Abs(num-rnum) < DIFF {
			return nil
		}
	} else if args[0] == args[1] {
		return nil
	}

	return args[0]
}

func (*functor) IfNull(args []interface{}) interface{} {
	if len(args) < 1 || args[0] == nil {
		return true
	}

	return false
}

func (*functor) Power(args []interface{}) (ret interface{}) {
	if len(args) < 2 || args[0] == nil || args[1] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return math.Pow(utils.MustGetFloat64(args[0]), utils.MustGetFloat64(args[1]))
}

func (*functor) Sqrt(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return math.Sqrt(utils.MustGetFloat64(args[0]))
}

func (*functor) Exp(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return math.Exp(utils.MustGetFloat64(args[0]))
}

func (*functor) Ceil(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return math.Ceil(utils.MustGetFloat64(args[0]))
}

func (*functor) Floor(args []interface{}) (ret interface{}) {
	if len(args) < 1 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	return math.Floor(utils.MustGetFloat64(args[0]))
}

func (*functor) Iif(args []interface{}) (ret interface{}) {
	if len(args) < 3 || args[0] == nil {
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	if reflect.ValueOf(args[0]).Bool() {
		return args[1]
	}
	return args[2]
}
