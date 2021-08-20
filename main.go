package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"
)

type Color struct {
	R int `plus:""`
	G int `plus:""`
	B int `plus:""`
}

type Image struct {
	pixels [600][600]Color `init:"Color((_+__)%256,_%256,__%256)"`
}

const tagName = "init"

type ExpressionKind string

const (
	CallExpression           ExpressionKind = "Call"
	LiteralExpression        ExpressionKind = "Literal"
	BinaryOperatorExpression ExpressionKind = "BinaryOperator"
	IdentifierExpression     ExpressionKind = "Identifier"
)

type Expression interface {
	Kind() ExpressionKind
}

type Call struct {
	Func Expression
	Args []Expression
}

func (*Call) Kind() ExpressionKind { return CallExpression }

type Literal struct {
	Value interface{}
}

func (*Literal) Kind() ExpressionKind { return LiteralExpression }

type BinaryOperator struct {
	Tok   Token
	Left  Expression
	Right Expression
}

func (*BinaryOperator) Kind() ExpressionKind { return BinaryOperatorExpression }

type Identifier struct {
	Tok Token
}

func (*Identifier) Kind() ExpressionKind { return IdentifierExpression }

type Token struct {
	Kind     rune
	String   string
	Position scanner.Position
}

type Parser struct {
	Tokens []Token
	Index  int
}

func NewParser(src string, name string) Parser {
	var s scanner.Scanner
	s.Init(strings.NewReader(src))
	s.Filename = name
	parser := Parser{}
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		parser.Tokens = append(parser.Tokens, Token{Kind: tok, String: s.TokenText(), Position: s.Pos()})
		fmt.Printf("%s: %d %s\n", s.Position, tok, s.TokenText())
	}
	return parser
}

func (p *Parser) Peek(n int) Token {
	if n+p.Index >= len(p.Tokens) {
		return Token{Kind: scanner.EOF}
	}
	return p.Tokens[n+p.Index]
}

func (p *Parser) Skip(n int) bool {
	if n+p.Index > len(p.Tokens) {
		return false
	}
	p.Index += n
	return true
}

func (p *Parser) Take() Token {
	if p.Index >= len(p.Tokens) {
		return Token{Kind: scanner.EOF}
	}
	tok := p.Tokens[p.Index]
	p.Index++
	return tok
}

func (p *Parser) MustTake(t rune) error {
	tok := p.Take()
	if tok.Kind != t {
		return fmt.Errorf("expected %d(%s) got %d(%s)", t, string(t), tok.Kind, string(tok.Kind))
	}
	return nil
}

func (p *Parser) Parse() (Expression, error) {
	expr, err := p.ParseSimpleExpression()
	if err != nil {
		return nil, err
	}
	tok := p.Peek(0)
	switch tok.String {
	case "+", "%":
		p.Skip(1)
		right, err := p.Parse()
		if err != nil {
			return nil, err
		}
		expr = &BinaryOperator{Tok: tok, Left: expr, Right: right}
	case "(":
		p.Skip(1)
		var args []Expression
	L:
		for {
			switch p.Peek(0).String {
			case ")":
				break L
			default:
				argExpr, err := p.Parse()
				if err != nil {
					return nil, err
				}
				args = append(args, argExpr)
				if p.Peek(0).String == "," {
					p.Skip(1)
				}
			}
		}
		expr = &Call{Func: expr, Args: args}
	}
	return expr, nil
}

func (p *Parser) ParseSimpleExpression() (Expression, error) {
	tok := p.Take()
	switch tok.Kind {
	case scanner.Ident:
		return &Identifier{Tok: tok}, nil
	case scanner.Int:
		i, err := strconv.Atoi(tok.String)
		if err != nil {
			return nil, err
		}
		return &Literal{Value: i}, nil
	case '(':
		expr, err := p.Parse()
		if err != nil {
			return nil, err
		}
		if err := p.MustTake(')'); err != nil {
			return nil, err
		}
		return expr, nil
	default:
		return nil, fmt.Errorf("simple expression: unexpected token %v", tok)
	}
}

func StructTagLang(v interface{}) error {
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagName)
		parser := NewParser(tag, fmt.Sprintf("%s.%s", t.Name(), field.Name))
		expr, err := parser.Parse()
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", expr)
	}
	return nil
}

func main() {
	err := StructTagLang(Image{})
	if err != nil {
		panic(err)
	}
}
