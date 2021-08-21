package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	parser := NewParser("Color((_+__)%256,_%256,__%256)", "a")
	expr, err := parser.ParseExpression()
	assert.NoError(t, err)
	assert.IsType(t, expr, &Call{})
	assert.IsType(t, (*(expr.(*Call))).Func, &Identifier{})
	assert.Len(t, (*(expr.(*Call))).Args, 3)
}

func TestImage(t *testing.T) {
	res, err := EvalStruct(reflect.TypeOf(Image{}), []interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, res)
}
