package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sdghchj/sql-rules-engine/filter"
	"github.com/sdghchj/sql-rules-engine/function"
	"github.com/sdghchj/sql-rules-engine/handler"
	"github.com/sdghchj/sql-rules-engine/mapper"
	"go/scanner"
	"go/token"
	"strings"
)

type Rule interface {
	Name() string
	AddEventHandler(match string, handlers ...handler.EventHandler) error
	AddEventAsyncHandler(match string, asyncHandlers ...handler.AsyncEventHandler) error
	AddConvertHandlerBySql(sql string, funcs function.Functions) error
	AddHandler(cvt handler.Handler) Rule
	InsertHandler(index int, cvt handler.Handler) Rule
	Handle(obj interface{}) interface{}
	HandleAsync(obj interface{})
	ConvertJson(jsonText string) (string, error)
}

type jsonRule struct {
	name     string
	pretty   bool
	handlers []handler.Handler
}

var ErrorSqlError = errors.New("sql error")

func NewJsonRule(pretty bool) Rule {
	return &jsonRule{pretty: pretty}
}

func (r *jsonRule) Name() string {
	return r.name
}

func (r *jsonRule) AddHandler(cvt handler.Handler) Rule {
	r.handlers = append(r.handlers, cvt)
	return r
}

func (r *jsonRule) InsertHandler(index int, cvt handler.Handler) Rule {
	if index < 0 || index >= len(r.handlers) {
		r.handlers = append(r.handlers, cvt)
	} else {
		r.handlers = append(r.handlers, cvt)
		for i := len(r.handlers) - 1; i > index; i-- {
			r.handlers[i] = r.handlers[i-1]
		}
		r.handlers[index] = cvt
	}
	return r
}

func (r *jsonRule) Handle(obj interface{}) interface{} {
	for _, cvt := range r.handlers {
		if obj == nil {
			break
		}
		obj = cvt.Handle(obj)
	}
	return obj
}

func (r *jsonRule) HandleAsync(obj interface{}) {
	for _, cvt := range r.handlers {
		cvt.HandleAsync(obj)
	}
}

func (r *jsonRule) ConvertJson(jsonText string) (string, error) {
	decoder := json.NewDecoder(strings.NewReader(jsonText))
	//decoder.UseNumber()
	var obj interface{}
	err := decoder.Decode(&obj)
	if err != nil {
		return "", err
	}

	for _, cvt := range r.handlers {
		if obj == nil {
			break
		}
		obj = cvt.Handle(obj)
	}
	var bin []byte
	if r.pretty {
		bin, err = json.MarshalIndent(obj, "", "    ")
	} else {
		bin, err = json.Marshal(obj)
	}

	if err != nil {
		return "", err
	}
	return string(bin), err
}

func (r *jsonRule) AddEventHandler(match string, handlers ...handler.EventHandler) error {
	filter := filter.NewFieldFilter(nil)
	err := filter.Parse(match, handlers...)
	if err != nil {
		return err
	}
	r.AddHandler(filter)
	return nil
}

func (r *jsonRule) AddEventAsyncHandler(match string, asyncHandlers ...handler.AsyncEventHandler) error {
	filter := filter.NewFieldFilter(nil)
	err := filter.ParseForAsyncHandlers(match, asyncHandlers...)
	if err != nil {
		return err
	}
	r.AddHandler(filter)
	return nil
}

func (r *jsonRule) AddConvertHandlerBySql(sql string, funcs function.Functions) error {

	type sqlToken struct {
		pos token.Pos
		tok token.Token
		lit string
	}

	type sqlField struct {
		name  string
		alias string
	}

	var tokens []*sqlToken
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(sql))
	s.Init(file, []byte(sql), nil, 0)
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		} else if tok != token.SEMICOLON { //SEMICOLON means \n
			tokens = append(tokens, &sqlToken{pos, tok, lit})
		}
	}

	getError := func(t *sqlToken) error {
		return errors.New(fmt.Sprintf("%s\t%s\t%q\n", fset.Position(t.pos), t.tok, t.lit))
	}

	//remove auto-added ;
	length := len(tokens)

	for i := 1; i < length; i++ {
		switch tokens[i].tok {
		case token.IDENT, token.INT, token.FLOAT, token.STRING, token.CHAR:
			break
		case token.LPAREN, token.LBRACK, token.RPAREN, token.RBRACK,
			token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
			token.ASSIGN, token.EQL, token.NEQ, token.GTR, token.GEQ, token.LSS, token.LEQ,
			token.AND, token.OR, token.NOT, token.XOR, token.LOR, token.LAND,
			token.COMMA, token.PERIOD,
			token.SHL, token.SHR: //tokens[i].tok.IsOperator()
			tokens[i].lit = tokens[i].tok.String()
			break
		default:
			return getError(tokens[i])
		}
	}

	if !strings.EqualFold(tokens[0].lit, "select") {
		return getError(tokens[0])
	}

	var fields []*sqlField
	var table, where string

	lparen := 0
	lbrack := 0
	pos := 1
	step := 1

	for i := 1; i < length; i++ {
		switch tokens[i].tok {
		case token.COMMA: // ,
			if lparen == 0 && lbrack == 0 && step == 1 {
				if pos == i || pos+1 == i && tokens[pos].tok == tokens[i].tok {
					return getError(tokens[i])
				}
				var field string
				for j := pos; j < i; j++ {
					field += tokens[j].lit
				}
				fields = append(fields, &sqlField{name: field})
				pos = i + 1
			}
			break
		case token.IDENT:
			if lparen == 0 && lbrack == 0 {
				if step == 1 {
					if strings.EqualFold(tokens[i].lit, "as") {
						if pos == i || pos+1 == i && tokens[pos].tok == tokens[i].tok {
							return ErrorSqlError
						}
						var field string
						for j := pos; j < i; j++ {
							field += tokens[j].lit
						}

						i++
						//A.B.C
						pos = i
						for ; i < length; i++ {
							if ((i-pos)&1) == 0 && tokens[i].tok == token.IDENT ||
								((i-pos)&1) == 1 && tokens[i].tok == token.PERIOD {
								continue
							} else {
								break
							}
						}
						if ((i - pos) & 1) == 0 {
							return getError(tokens[i-1])
						}

						var alias string
						for j := pos; j < i; j++ {
							alias += tokens[j].lit
						}

						fields = append(fields, &sqlField{name: field, alias: alias})

						if tokens[i].tok == token.COMMA {
							pos = i + 1
						} else if tokens[i].tok == token.IDENT && strings.EqualFold(tokens[i].lit, "from") {
							i--
						} else {
							return getError(tokens[i])
						}
					} else if strings.EqualFold(tokens[i].lit, "from") {
						step++

						if i > pos {
							var field string
							for j := pos; j < i; j++ {
								field += tokens[j].lit
							}
							fields = append(fields, &sqlField{name: field})
							pos = i + 1
						} else if tokens[pos-1].tok == token.COMMA {
							return getError(tokens[i])
						}

						i++
						if i >= length {
							return getError(tokens[i-1])
						}

						if tokens[i].tok == token.IDENT {
							table = tokens[i].lit
						} else if tokens[i].tok == token.CHAR || tokens[i].tok == token.STRING {
							table = tokens[i].lit[1 : len(tokens[i].lit)-1]
						} else {
							return getError(tokens[i])
						}

						i++
						if i < length {
							if tokens[i].tok != token.IDENT || !strings.EqualFold(tokens[i].lit, "where") || i+1 >= length {
								return getError(tokens[i])
							}

							where = "where"

							step++
							pos = i + 1
						}
					}
				}
			}

			if step == 3 {
				if strings.EqualFold(tokens[i].lit, "and") {
					tokens[i].lit = "&&"
				} else if strings.EqualFold(tokens[i].lit, "or") {
					tokens[i].lit = "||"
				} else if strings.EqualFold(tokens[i].lit, "not") {
					tokens[i].lit = "!"
				} else if strings.EqualFold(tokens[i].lit, "null") {
					tokens[i].lit = "nil"
				}
			}
			break
		case token.LPAREN: // (
			lparen++
			break
		case token.LBRACK: // [
			lbrack++
			break
		case token.RPAREN: // )
			if lparen < 1 {
				return getError(tokens[i])
			}
			lparen--
			break
		case token.RBRACK: // ]
			if lbrack < 1 {
				return getError(tokens[i])
			}
			lbrack--
		case token.ASSIGN:
			if step == 3 {
				tokens[i].lit = "=="
			} else {
				return getError(tokens[i])
			}
			break
		}
	}

	if table == "" {
		return ErrorSqlError
	}

	var err error

	if where != "" {
		where = ""
		for i := pos; i < length; i++ {
			where += tokens[i].lit
		}

		filter := filter.NewFieldFilter(funcs)
		err = filter.Parse(where)
		if err != nil {
			return err
		}

		r.AddHandler(filter)
	}

	if len(fields) > 0 {
		mp := mapper.NewMapper(funcs)
		for _, field := range fields {
			err = mp.AddField(field.name, field.alias)
			if err != nil {
				return err
			}
		}
		r.AddHandler(mp)
	}

	r.name = table

	return nil
}
