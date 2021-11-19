package histogram

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Point struct {
	X_ int
	Y_ int
}

func NewPoint(id int, onChanged func()) (*Point, fyne.CanvasObject) {
	p := &Point{}
	widgetForX := widget.NewEntry()
	widgetForX.OnChanged = func(change string) {
		changeInt, _ := strconv.Atoi(change) // TODO check error
		p.X_ = changeInt
		onChanged()
	}
	widgetForY := widget.NewEntry()
	widgetForY.OnChanged = func(change string) {
		changeInt, _ := strconv.Atoi(change) // TODO check error
		p.Y_ = changeInt
		onChanged()
	}
	return p, container.NewAdaptiveGrid(3, widget.NewLabel("Point "+strconv.Itoa(id+1)+": "), widgetForX, widgetForY)
}

func (p Point) Validate() error {
	if p.X_ < 0 || p.X_ > 255 || p.Y_ < 0 || p.Y_ > 255 {
		return fmt.Errorf("the number must be in the range [0, 255]")
	}
	return nil
}
