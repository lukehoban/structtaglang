package main

import (
	"fmt"
	"reflect"
)

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
	case reflect.Int, reflect.Uint8:
		v, err := ev.Eval(expr)
		if err != nil {
			return nil, err
		}
		i := reflect.ValueOf(v).Convert(ty).Interface()
		return i, err
	default:
		panic(fmt.Sprintf("nyi - eval type %s", ty.Kind()))
	}
}

func EvalStruct(ty reflect.Type, args []interface{}) (interface{}, error) {
	val := reflect.New(ty).Elem()
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)
		tag := field.Tag.Get("λ")
		if tag == "" {
			// If there is no tag, implicitly pass the i'th argument to the i'th struct field.  This ensures
			// that external struct types can also be used as constructors with positional arguments.
			tag = fmt.Sprintf("_%d", i)
		}
		parser := NewParser(tag, fmt.Sprintf("%s.%s", ty.Name(), field.Name))
		expr, err := parser.ParseExpression()
		if err != nil {
			return nil, err
		}

		ev := NewEvaluator()
		// All args are in scope
		for i, a := range args {
			ev.scope[fmt.Sprintf("_%d", i)] = a
		}
		// Also, all struct fields before this one
		for j := 0; j < i; j++ {
			ev.scope[ty.Field(j).Name] = val.Field(j).Interface()
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
	case reflect.Int, reflect.Uint8:
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
