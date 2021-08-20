package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	parser := NewParser("Color((_+__)%256,_%256,__%256)", "a")
	expr, err := parser.Parse()
	assert.NoError(t, err)
	assert.IsType(t, expr, &Call{})
	assert.IsType(t, (*(expr.(*Call))).Func, &Identifier{})
	assert.Len(t, (*(expr.(*Call))).Args, 3)
}