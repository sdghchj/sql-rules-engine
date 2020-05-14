package parser

import (
	"github.com/sdghchj/sql-rules-engine/function"
	"github.com/sdghchj/sql-rules-engine/utils"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

type Resolver interface {
	Evaluate(obj interface{}) interface{}
}

type goResolver struct {
	funcs function.Functions
	node  ast.Expr
}

func NewGoResolver(node ast.Expr, funcs function.Functions) Resolver {
	return &goResolver{node: node, funcs: funcs}
}

func (r *goResolver) Evaluate(obj interface{}) interface{} {
	return r.visit(r.node, obj)
}

func (r *goResolver) visitBinaryExpression(exp *ast.BinaryExpr, obj interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	var bx, by interface{}
	if exp.Y != nil {
		by = r.visit(exp.Y, obj)
	}

	switch exp.Op {
	case token.LAND:
		if by == nil {
			return false
		}
		y := reflect.ValueOf(by).Bool()
		if !y {
			return y
		}
		bx = r.visit(exp.X, obj)
		if bx == nil {
			return false
		}
		x := reflect.ValueOf(bx).Bool()
		return x
	case token.LOR:
		if by != nil {
			y := reflect.ValueOf(by).Bool()
			if y {
				return y
			}
		}

		bx = r.visit(exp.X, obj)
		if bx == nil {
			return false
		}
		x := reflect.ValueOf(bx).Bool()
		return x
	}

	if by == nil {
		goto InterfaceEqual
	}

	if exp.X != nil {
		bx = r.visit(exp.X, obj)
		if bx == nil {
			goto InterfaceEqual
		}
	}

	if x, err := utils.GetFloat64(bx); err == nil {
		if y, err := utils.GetFloat64(by); err == nil {
			switch exp.Op {
			case token.GTR:
				return x > y
			case token.LSS:
				return x < y
			case token.GEQ:
				return x >= y
			case token.LEQ:
				return x <= y
			case token.NEQ:
				return x != y
			case token.EQL:
				return x == y
			case token.ADD:
				return x + y
			case token.SUB:
				return x - y
			case token.MUL:
				return x * y
			case token.QUO: //divide
				return x / y
			case token.REM: //%
				if int64(y) == 0 {
					return nil
				}
				return int64(x) % int64(y)
			case token.AND:
				return int64(x) & int64(y)
			case token.OR:
				return int64(x) | int64(y)
			case token.XOR:
				return int64(x) ^ int64(y)
			case token.SHL:
				return int64(x) << uint(y)
			case token.SHR:
				return int64(x) >> uint(y)
			}
		}
	} else if x, ok := bx.(string); ok {
		if y, ok := by.(string); ok {
			switch exp.Op {
			case token.GTR:
				return x > y
			case token.LSS:
				return x < y
			case token.GEQ:
				return x >= y
			case token.LEQ:
				return x <= y
			case token.NEQ:
				return x != y
			case token.EQL:
				return x == y
			case token.ADD:
				return x + y
			}
		}
	}

InterfaceEqual:
	switch exp.Op {
	case token.NEQ:
		return bx != by
	case token.EQL:
		return bx == by
	}

	return nil
}

func (r *goResolver) visitArrayIndexExpression(exp *ast.IndexExpr, obj interface{}) (ret interface{}) {
	val := r.visit(exp.X, obj)
	if val == nil {
		//root means root of obj
		if idt, ok := exp.X.(*ast.Ident); ok && strings.EqualFold(idt.Name, "root") {
			val = obj
		} else {
			return nil
		}
	}
	valValue := reflect.ValueOf(val)

	//in case of array[-1], return all elements of val
	if unary, ok := exp.Index.(*ast.UnaryExpr); ok {
		if _, ok := unary.X.(*ast.BasicLit); ok && unary.Op == token.SUB {
			return val
		}
	}

	index := r.visit(exp.Index, obj)
	if index == nil {
		return nil
	}
	indexValue := reflect.ValueOf(index)

	defer func() {
		defer func() {
			if err := recover(); err != nil {
				ret = nil
			}
		}()
		if err := recover(); err != nil {
			f := indexValue.Float()
			valValue.Index(int(f)).Interface()
		}
	}()

	ret = valValue.Index(int(indexValue.Int())).Interface()
	return
}

func (r *goResolver) visitFuncExpression(exp *ast.CallExpr, obj interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	ident, ok := exp.Fun.(*ast.Ident)
	if !ok {
		return nil
	}

	var args []interface{}
	length := len(exp.Args)

	if length > 0 {
		args = make([]interface{}, length)
	}
	for i, arg := range exp.Args {
		args[i] = r.visit(arg, obj)
	}

	if r.funcs != nil && r.funcs.Exists(ident.Name) {
		ret = r.funcs.Call(ident.Name, args)
	} else {
		ret = function.DefaultFunctions.Call(ident.Name, args)
	}

	return ret
}

func (r *goResolver) visit(node ast.Expr, obj interface{}) interface{} {
	switch exp := node.(type) {
	case *ast.BinaryExpr:
		return r.visitBinaryExpression(exp, obj)
	case *ast.BasicLit:
		switch exp.Kind {
		case token.INT:
			x, errx := strconv.ParseInt(exp.Value, 10, 64)
			if errx != nil {
				return nil
			}
			return x
		case token.FLOAT:
			x, errx := strconv.ParseFloat(exp.Value, 64)
			if errx != nil {
				return nil
			}
			return x
		case token.STRING:
			return exp.Value[1 : len(exp.Value)-1] //remove "
		}
		break
	case *ast.Ident:
		switch exp.Name {
		case "nil":
			return nil
		case "true":
			return true
		case "false":
			return false
		default:
			if mp, ok := obj.(map[string]interface{}); ok {
				if val, ok := mp[exp.Name]; ok {
					return val
				}
			}
		}
		break
	case *ast.SelectorExpr:
		x := r.visit(exp.X, obj)
		if x == nil {
			return nil
		}
		if mp, ok := x.(map[string]interface{}); ok {
			if val, ok := mp[exp.Sel.Name]; ok {
				return val
			}
		} else if arrMap, ok := x.([]map[string]interface{}); ok {
			var ret []interface{}
			for _, mp := range arrMap {
				if val, ok := mp[exp.Sel.Name]; ok {
					ret = append(ret, val)
				}
			}
			return ret
		}
		break
	case *ast.CallExpr:
		return r.visitFuncExpression(exp, obj)
		break
	case *ast.IndexExpr:
		return r.visitArrayIndexExpression(exp, obj)
		break
	case *ast.UnaryExpr:
		x := r.visit(exp.X, obj)
		if x == nil {
			return x
		}
		switch exp.Op {
		case token.NOT:
			if b, ok := x.(bool); ok {
				return !b
			}
		case token.SUB:
			if n, err := utils.GetFloat64(x); err == nil {
				return -n
			}
		case token.XOR:
			if n, err := utils.GetFloat64(x); err == nil {
				return ^int64(n)
			}
		}
		return nil
	case *ast.ParenExpr:
		return r.visit(exp.X, obj)
	default:
		return nil
	}
	return nil
}
