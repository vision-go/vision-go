package histogram

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Point struct {
	X int
	Y int
}

func NewPoint(id int, onChanged func()) (*Point, fyne.CanvasObject) {
	p := &Point{X: -1, Y: -1}
	widgetForX := widget.NewEntry()
	widgetForX.OnChanged = func(change string) {
		changeInt, err := strconv.Atoi(change)
		if err != nil {
			changeInt = -1
		}
		p.X = changeInt
		onChanged()
	}
	widgetForY := widget.NewEntry()
	widgetForY.OnChanged = func(change string) {
		changeInt, err := strconv.Atoi(change)
		if err != nil {
			changeInt = -1
		}
		p.Y = changeInt
		onChanged()
	}
	return p, container.NewAdaptiveGrid(3, widget.NewLabel("Point "+strconv.Itoa(id+1)+": "), widgetForX, widgetForY)
}

func (p Point) Validate() error {
	if p.X < 0 || p.X > 255 || p.Y < 0 || p.Y > 255 {
		return fmt.Errorf("the values must be integers in the range [0, 255]")
	}
	return nil
}
