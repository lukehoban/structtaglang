package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"reflect"
)

type Color struct {
	R int `λ:"_0%4"`
	G int `λ:"_1%4"`
	B int `λ:"_2%4"`
}

type Image struct {
	Colors [256][256]Color       `λ:"__0+__1,__0,__1"`
	Pixels [256][256]color.NRGBA `λ:"__0+__1,__0,__1,255"`
}

func main() {
	res, err := EvalStruct(reflect.TypeOf(Image{}), []interface{}{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v %v\n", reflect.TypeOf(res), res)

	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 256; x++ {
			img.Set(x, y, res.(Image).Pixels[x][y])
		}
	}
	f, err := os.Create("image.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

}
