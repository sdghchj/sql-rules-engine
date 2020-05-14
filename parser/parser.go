package parser

import (
	"errors"
	"github.com/sdghchj/sql-rules-engine/function"
	"go/parser"
	"go/scanner"
	"go/token"
	"strings"
)

type Parser interface {
	Parse(text string, funcs function.Functions) (Resolver, error)
}

type goParser int

var DefaultGoParser goParser

var ErrTypeError = errors.New("type error")

type sqlToken struct {
	pos token.Pos
	tok token.Token
	lit string
}

func (goParser) translateSqlWhere(where string) string {
	var tokens []*sqlToken
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(where))
	s.Init(file, []byte(where), nil, 0)
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		} else if tok != token.SEMICOLON { //SEMICOLON means \n
			tokens = append(tokens, &sqlToken{pos, tok, lit})
		}
	}
	if tokens == nil {
		return ""
	}

	length := len(tokens)

	for i := 0; i < length; i++ {
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
		}
	}

	var need = false
	for i := 0; i < length; i++ {
		switch tokens[i].tok {
		case token.IDENT:
			if i > 0 && tokens[i-1].tok == token.PERIOD || i < length-1 && tokens[i+1].tok == token.PERIOD {
				continue
			}

			if strings.EqualFold(tokens[i].lit, "and") {
				tokens[i].tok = token.LAND
				tokens[i].lit = "&&"
				need = true
			} else if strings.EqualFold(tokens[i].lit, "or") {
				tokens[i].tok = token.LOR
				tokens[i].lit = "||"
				need = true
			} else if strings.EqualFold(tokens[i].lit, "not") {
				tokens[i].tok = token.NOT
				tokens[i].lit = "!"
				need = true
			} else if strings.EqualFold(tokens[i].lit, "null") {
				tokens[i].lit = "nil"
				need = true
			} else if i > 0 && i+1 < length && strings.EqualFold(tokens[i].lit, "in") && tokens[i-1].tok >= token.IDENT && tokens[i-1].tok <= token.STRING && tokens[i+1].tok == token.LPAREN {
				// val in (comma separated array)
				temp := tokens[i-1]
				tokens[i-1] = tokens[i]
				tokens[i] = tokens[i+1]
				tokens[i+1] = temp
				tokens[i+1].lit += ","
				need = true
			}
			break
		case token.MUL:
			if i > 0 && i+1 < length && tokens[i-1].tok == token.LBRACK && tokens[i+1].tok == token.RBRACK {
				tokens[i].tok = token.INT
				tokens[i].lit = "-1"
			}
		case token.ASSIGN:
			tokens[i].lit = "=="
			need = true
			break
		default:
			break
		}
	}

	if need {
		where = ""
		for i := 0; i < length; i++ {
			where += tokens[i].lit
		}
	}

	return where
}

func (p goParser) Parse(text string, funcs function.Functions) (Resolver, error) {
	text = p.translateSqlWhere(text)
	exp, err := parser.ParseExpr(text)
	if err != nil {
		return nil, err
	}
	return NewGoResolver(exp, funcs), nil
}
