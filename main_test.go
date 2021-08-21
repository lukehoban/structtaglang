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

type Vector struct {
	X float64
	Y float64
	Z float64
}

type VectorTimes struct {
	Vector `λ:"_0.X*_1,_0.Y*_1,_0.Z*_1"`
}

type VectorPlus struct {
	Vector Vector `λ:"_0.X+_1.X,_0.Y+_1.Y,_0.Z+_1.Z"`
}

type TracePixel struct {
	RecenterX      float64     `λ:"(128.0-_1)/512"`
	RecenterY      float64     `λ:"(_2-128.0)/512"`
	Right          VectorTimes `λ:"_0.Right,RecenterX"`
	Up             VectorTimes `λ:"_0.Up,RecenterY"`
	RightUp        VectorPlus  `λ:"Up,Right"`
	RightUpForward VectorPlus  `λ:"RightUp.Vector,_0.Forward"`
	// TODO: Create ray and trace it
}

type Raytracer struct {
	Pixels [256][256]TracePixel `λ:"_0,__0,__1"`
}

type Camera struct {
	Up      Vector `λ:"_0.X,_0.Y,_0.Z"`
	Right   Vector `λ:"_1.X,_1.Y,_1.Z"`
	Forward Vector `λ:"_2.X,_2.Y,_2.Z"`
}

type Main struct {
	Up        Vector    `λ:"0,0,1"`
	Right     Vector    `λ:"0,1,0"`
	Forward   Vector    `λ:"1,0,0"`
	Camera    Camera    `λ:"Up,Right,Forward"`
	Raytracer Raytracer `λ:"Camera"`
}

func TestRaytracer(t *testing.T) {
	res, err := EvalStruct(reflect.TypeOf(Main{}), []interface{}{nil})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1.0, res.(Main).Raytracer.Pixels[0][0].RightUpForward.Vector.X)
	assert.Equal(t, 128.0, res.(Main).Raytracer.Pixels[0][0].RightUpForward.Vector.Y)
	assert.Equal(t, -128.0, res.(Main).Raytracer.Pixels[0][0].RightUpForward.Vector.Z)
}
