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
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Negative())
}

func (ui *UI) monochromeOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Monochrome())
}

func (ui *UI) adjustBrightnessAndContrastOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	newImage := *currentImage
	brightnessValue, contrastValue := binding.NewFloat(), binding.NewFloat()
	brightnessLabel, contrastLabel :=
		widget.NewLabelWithData(binding.FloatToStringWithFormat(brightnessValue, "%v")),
		widget.NewLabelWithData(binding.FloatToStringWithFormat(contrastValue, "%v"))
	brightnessSlider, contrastSlider :=
		widget.NewSliderWithData(0, 255, brightnessValue),
		widget.NewSliderWithData(0, 255, contrastValue)
	brightnessSlider.SetValue(newImage.Brightness())
	contrastSlider.SetValue(newImage.Contrast())
	brightnessSlider.OnChanged = func(value float64) {
		brightnessValue.Set(value)
	}
	contrastSlider.OnChanged = func(value float64) {
		contrastValue.Set(value)
	}
	content := container.NewGridWithRows(4, brightnessLabel, brightnessSlider, contrastLabel, contrastSlider)
	dialog.ShowCustomConfirm("Adjust Brightness and Contrast", "Ok", "Cancel", content,
		func(choice bool) {
			if !choice {
				return
			}
			brightness, err := brightnessValue.Get()
			if err != nil {
				dialog.ShowError(err, ui.MainWindow)
			}
			contrast, err := contrastValue.Get()
			if err != nil {
				dialog.ShowError(err, ui.MainWindow)
			}
			ui.newImage(newImage.BrightnessAndContrast(brightness, contrast))
		},
		ui.MainWindow)
}

func (ui *UI) gammaCorrectionOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entry := widget.NewEntry()
	entry.Validator = func(value string) error {
		valueFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		if valueFloat < 0 || valueFloat > 20 {
			return fmt.Errorf("Gamma must be between values 0 and 20")
		}
		return nil
	}
	form := []*widget.FormItem{widget.NewFormItem("Gamma", entry)}
	dialog.ShowForm("Select gamma", "Ok", "Cancel", form,
		func(choice bool) {
			if !choice {
				return
			}
			// No need to check thanks to validator
			gamma, _ := strconv.ParseFloat(entry.Text, 64)
			ui.newImage(currentImage.GammaCorrection(gamma))
		},
		ui.MainWindow)
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
