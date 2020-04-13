package parser

import (
	"github.com/sdghchj/sql-rules-engine/function"
	"github.com/sdghchj/sql-rules-engine/utils"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
)

type Resolver interface {
	Evaluate(src map[string]interface{}) interface{}
}

type goResolver struct {
	funcs function.Functions
	node  ast.Expr
}

func NewGoResolver(node ast.Expr, funcs function.Functions) Resolver {
	return &goResolver{node: node, funcs: funcs}
}

func (r *goResolver) Evaluate(src map[string]interface{}) interface{} {
	return r.visit(r.node, &src)
}

func (r *goResolver) visitBinaryExpression(exp *ast.BinaryExpr, src *map[string]interface{}) (ret interface{}) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
		}
	}()

	var bx, by interface{}
	if exp.Y != nil {
		by = r.visit(exp.Y, src)
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
		bx = r.visit(exp.X, src)
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

		bx = r.visit(exp.X, src)
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
		bx = r.visit(exp.X, src)
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

func (r *goResolver) visitArrayIndexExpression(exp *ast.IndexExpr, src *map[string]interface{}) (ret interface{}) {
	val := r.visit(exp.X, src)
	if val == nil {
		return nil
	}
	valValue := reflect.ValueOf(val)

	index := r.visit(exp.Index, src)
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

func (r *goResolver) visitFuncExpression(exp *ast.CallExpr, src *map[string]interface{}) (ret interface{}) {
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
		args[i] = r.visit(arg, src)
	}

	if r.funcs != nil && r.funcs.Exists(ident.Name) {
		ret = r.funcs.Call(ident.Name, args)
	} else {
		ret = function.DefaultFunctions.Call(ident.Name, args)
	}

	return ret
}

func (r *goResolver) visit(node ast.Expr, src *map[string]interface{}) interface{} {
	switch exp := node.(type) {
	case *ast.BinaryExpr:
		return r.visitBinaryExpression(exp, src)
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
			if val, ok := (*src)[exp.Name]; ok {
				return val
			}
		}
		break
	case *ast.SelectorExpr:
		sel := r.visit(exp.X, src)
		if sel != nil {
			if mp, ok := sel.(map[string]interface{}); ok {
				if val, ok := mp[exp.Sel.Name]; ok {
					return val
				}
			}
		}
		break
	case *ast.CallExpr:
		return r.visitFuncExpression(exp, src)
		break
	case *ast.IndexExpr:
		return r.visitArrayIndexExpression(exp, src)
		break
	case *ast.UnaryExpr:
		if exp.Op == token.NOT {
			val := r.visit(exp.X, src)
			if val != nil {
				if b, ok := val.(bool); ok {
					return !b
				}
			}
		} else if exp.Op == token.XOR {
			val := r.visit(exp.X, src)
			if val != nil {
				if n, err := utils.GetFloat64(val); err == nil {
					return ^int64(n)
				}
			}
		}
		return nil
	case *ast.ParenExpr:
		return r.visit(exp.X, src)
	default:
		return nil
	}
	return nil
}
