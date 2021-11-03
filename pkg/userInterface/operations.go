package userinterface

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) negativeOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Negative()) // TODO Improve name
}

func (ui *UI) monochromeOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Monochrome()) // TODO Improve name
}

// Linear transformation operation
type point struct {
	x_         string
	y_         string
	validator_ func(value string) error
}

func newPoint(id int) (*point, fyne.CanvasObject) {
	p := &point{}
	p.validator_ = func(value string) error {
		if value == "" {
			return fmt.Errorf("Can't be null")
		}
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if valueInt < 0 || valueInt > 255 {
			return fmt.Errorf("The number must be in the range [0, 255]")
		}
		return nil
	}
	widgetForX := widget.NewEntryWithData(binding.BindString(&p.x_))
	widgetForX.Validator = p.validator_
	widgetForY := widget.NewEntryWithData(binding.BindString(&p.y_))
	widgetForY.Validator = p.validator_
	return p, container.NewAdaptiveGrid(3, widget.NewLabel("Point "+strconv.Itoa(id+1)+": "), widgetForX, widgetForY)
}

func (p point) Validate() error {
	errorX, errorY := p.validator_(p.x_), p.validator_(p.y_)
	if errorX != nil {
		return errorX
	}
	if errorY != nil {
		return errorY
	}
	return nil
}

func (ui *UI) linearTransformationOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	var linearTransformationUI func(string)

	// Ask for points
	entry := widget.NewEntry()
	entry.Validator = func(value string) error {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if valueInt <= 0 {
			return fmt.Errorf("The number of points must be positive")
		}
		return nil
	}
	entrySlice := []*widget.FormItem{widget.NewFormItem("", entry)}
	dialog.ShowForm("How many points?", "OK", "Cancel", entrySlice,
		func(choice bool) {
			if choice {
				linearTransformationUI(entrySlice[0].Widget.(*widget.Entry).Text)
				return
			}
			fmt.Println("Rechazó")
		}, ui.MainWindow)

	// linearTransformationUI
	linearTransformationUI = func(pointsNString string) {
		pointsN, _ := strconv.Atoi(pointsNString) // Error already checked in Validator
		var canvasPoints []fyne.CanvasObject
		var points []*point
		for i := 0; i < pointsN; i++ {
			rawPoint, canvasPoint := newPoint(i)
			points = append(points, rawPoint)
			canvasPoints = append(canvasPoints, canvasPoint)
		}
		content := container.NewVBox(canvasPoints...)
		dialog.ShowCustomConfirm("Linear Transformation", "OK", "Cancel", content,
			func(choice bool) {
				if !choice {
					return
				}
				for _, point := range points {
					if point.Validate() != nil {
						dialog.ShowError(point.Validate(), ui.MainWindow)
						return
					}
				}
			},
			ui.MainWindow)
	}
}
