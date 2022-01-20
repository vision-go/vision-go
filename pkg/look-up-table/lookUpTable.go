package lookUpTable

import (
	"image/color"
	"log"
)

const (
	Negative = iota
)

var negativeTable [256]color.Gray

func init() {
	negativeInit()
}

func RGBA(oldColour color.Color, op int) color.Color {
	r, g, b, a := oldColour.RGBA()
	r, g, b = r>>8, g>>8, b>>8
	switch op {
	case Negative:
		return color.RGBA{negativeTable[r].Y, negativeTable[g].Y, negativeTable[b].Y, uint8(a)}
	default:
		log.Printf("Op: %v in lookUpTable-RGBA is not valid.", op)
		return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
	}
}

func negativeInit() {
	for i := range negativeTable {
		negativeTable[i] = color.Gray{uint8(255 - i)}
	}
}
