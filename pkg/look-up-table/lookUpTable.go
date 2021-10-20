package lookUpTable

import (
	"image/color"
	"log"
)

const numberOfColors = 256

const (
	Negative = iota
)

// TODO lazyload
var NegativeTable [numberOfColors]color.Gray

func init() {
	negativeInit()
}

func RGBA(oldColour color.Color, op int) color.Color {
	r, g, b, a := oldColour.RGBA()
	r >>= 8
	g >>= 8
	b >>= 8
	a >>= 8
	switch op {
	case Negative:
		return color.RGBA{NegativeTable[r].Y, NegativeTable[g].Y, NegativeTable[b].Y, uint8(a)}
	default:
		log.Printf("Op: %v in lookUpTable-RGBA is not valid.", op)
		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	}
}

func negativeInit() {
	for i := range NegativeTable {
		NegativeTable[i] = color.Gray{uint8(numberOfColors - i)}
	}
}
