package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"
)

type Color struct {
	R int `λ:"_0"`
	G int `λ:"_1"`
	B int `λ:"_2"`
}

type Image struct {
	Pixels [10][10]Color `λ:"(__0+__1)%256,__0%256,__1%256"`
}

type ExpressionKind string

const (
	CallExpression           ExpressionKind = "Call"
	LiteralExpression        ExpressionKind = "Literal"
	BinaryOperatorExpression ExpressionKind = "BinaryOperator"
	IdentifierExpression     ExpressionKind = "Identifier"
	TupleExpression          ExpressionKind = "Tuple"
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

type Tuple struct {
	Items []Expression
}

func (*Tuple) Kind() ExpressionKind { return TupleExpression }

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

func (p *Parser) ParseExpression() (Expression, error) {
	expr, err := p.ParseBasicExpression()
	if err != nil {
		return nil, err
	}
	tok := p.Peek(0)
	switch tok.String {
	case ",":
		return p.ParseTupleExpressiion(expr)
	}
	return expr, nil
}

func (p *Parser) ParseTupleExpressiion(first Expression) (Expression, error) {
	ret := []Expression{first}
	for p.Peek(0).Kind == ',' {
		p.Skip(1)
		item, err := p.ParseBasicExpression()
		if err != nil {
			return nil, err
		}
		ret = append(ret, item)
	}
	return &Tuple{Items: ret}, nil
}

func (p *Parser) ParseBasicExpression() (Expression, error) {
	expr, err := p.ParseSimpleExpression()
	if err != nil {
		return nil, err
	}
	tok := p.Peek(0)
	switch tok.String {
	case "+", "%":
		p.Skip(1)
		right, err := p.ParseBasicExpression()
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
				argExpr, err := p.ParseBasicExpression()
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
		expr, err := p.ParseExpression()
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

type Evaluator struct {
	scope map[string]interface{}
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		scope: map[string]interface{}{},
	}
}

func (ev *Evaluator) Eval(expr Expression) (interface{}, error) {
	switch ex := expr.(type) {
	case *Identifier:
		v, ok := ev.scope[ex.Tok.String]
		if !ok {
			return nil, fmt.Errorf("no %q in scope", ex.Tok.String)
		}
		return v, nil
	case *BinaryOperator:
		left, err := ev.Eval(ex.Left)
		if err != nil {
			return nil, err
		}
		right, err := ev.Eval(ex.Right)
		if err != nil {
			return nil, err
		}
		switch ex.Tok.String {
		case "%":
			return left.(int) % right.(int), nil
		case "+":
			return left.(int) + right.(int), nil
		default:
			panic(fmt.Sprintf("nyi - operator %s", ex.Tok.String))
		}
	case *Literal:
		return ex.Value, nil
	case *Tuple:
		var ret []interface{}
		for _, item := range ex.Items {
			i, err := ev.Eval(item)
			if err != nil {
				return nil, err
			}
			ret = append(ret, i)
		}
		return ret, nil
	case *Call:
		panic("nyi - eval call")
	default:
		panic(fmt.Sprintf("nyi - eval %s", expr.Kind()))
	}
}

func EvalType(ev *Evaluator, expr Expression, ty reflect.Type, depth int) (interface{}, error) {
	for {
		switch ty.Kind() {
		// An array is a `for i:=0i<n;i++ { ... }`
		case reflect.Array:
			var ret []interface{}
			for i := 0; i < ty.Len(); i++ {
				ev.scope[fmt.Sprintf("__%d", depth)] = i
				v, err := EvalType(ev, expr, ty.Elem(), depth+1)
				if err != nil {
					return nil, err
				}
				ret = append(ret, v)
			}
			return ret, nil
		// An struct is a `foo(...)`
		case reflect.Struct:
			v, err := ev.Eval(expr)
			if err != nil {
				return nil, err
			}
			arrV, ok := v.([]interface{})
			if !ok {
				arrV = []interface{}{v}
			}
			return EvalStruct(ty, arrV)
		// An int is a `(...).(int)`
		case reflect.Int:
			v, err := ev.Eval(expr)
			if err != nil {
				return nil, err
			}
			i, ok := v.(int)
			if !ok {
				return nil, fmt.Errorf("cannot convert %v to int", reflect.TypeOf(v))
			}
			return i, err
		default:
			panic(fmt.Sprintf("nyi - eval type %s", ty.Kind()))
		}
	}
}

func EvalStruct(ty reflect.Type, args []interface{}) (interface{}, error) {
	val := reflect.New(ty).Elem()
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)
		tag := field.Tag.Get("λ")
		parser := NewParser(tag, fmt.Sprintf("%s.%s", ty.Name(), field.Name))
		expr, err := parser.ParseExpression()
		if err != nil {
			return nil, err
		}

		ev := NewEvaluator()
		for i, a := range args {
			ev.scope[fmt.Sprintf("_%d", i)] = a
		}
		v, err := EvalType(ev, expr, field.Type, 0)
		if err != nil {
			return nil, err
		}

		x := reflect.ValueOf(v)
		Set(val.Field(i), x)
	}
	return val.Interface(), nil
}

func Set(dest reflect.Value, val reflect.Value) {
	dt := dest.Type()
	switch dt.Kind() {
	case reflect.Int:
		dest.Set(val)
	case reflect.Array:
		arrVal := reflect.ValueOf(val.Interface())
		for i := 0; i < dt.Len(); i++ {
			Set(dest.Index(i), arrVal.Index(i))
		}
	case reflect.Struct:
		structVal := reflect.ValueOf(val.Interface())
		for i := 0; i < dt.NumField(); i++ {
			dest.Field(i).Set(structVal.Field(i))
		}
	default:
		panic(fmt.Sprintf("nyi - set %s", dt.Kind()))
	}
}

func StructTagLang(v interface{}) error {
	t := reflect.TypeOf(v)
	res, err := EvalStruct(t, []interface{}{})
	if err != nil {
		return err
	}
	fmt.Printf("%v %v\n", reflect.TypeOf(res), res)
	return nil
}

func main() {
	err := StructTagLang(Image{})
	if err != nil {
		panic(err)
	}
}
