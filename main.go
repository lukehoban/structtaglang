package main

import (
	"fmt"
	"reflect"
)

type Color struct {
	R int `λ:"_0%256"`
	G int `λ:"_1%256"`
	B int `λ:"_2%256"`
}

type Image struct {
	Pixels [600][600]Color `λ:"__0+__1,__0,__1"`
}

func main() {
	res, err := EvalStruct(reflect.TypeOf(Image{}), []interface{}{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v %v\n", reflect.TypeOf(res), res)
}
