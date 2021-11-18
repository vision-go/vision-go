package userinterface

import (
	"fmt"
	"image"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/dustin/go-humanize"
	"github.com/vision-go/vision-go/pkg/histogram"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	ourimage "github.com/vision-go/vision-go/pkg/ourImage"
)

func (ui *UI) equializationOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Equalization()) // TODO Improve name
}

func (ui *UI) negativeOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.Negative())
}

func (ui *UI) monochromeOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	ui.newImage(currentImage.Monochrome())
}

func (ui *UI) adjustBrightnessAndContrastOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	brightnessValue, contrastValue := binding.NewFloat(), binding.NewFloat()
	brightnessLabel, contrastLabel :=
		widget.NewLabelWithData(binding.FloatToStringWithFormat(brightnessValue, "%v")),
		widget.NewLabelWithData(binding.FloatToStringWithFormat(contrastValue, "%v"))
	brightnessSlider, contrastSlider :=
		widget.NewSliderWithData(0, 255, brightnessValue),
		widget.NewSliderWithData(0, 255, contrastValue)
	brightnessSlider.SetValue(currentImage.Brightness())
	contrastSlider.SetValue(currentImage.Contrast())
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
			ui.newImage(currentImage.BrightnessAndContrast(brightness, contrast))
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
		if valueFloat < 0.05 || valueFloat > 20 {
			return fmt.Errorf("gamma must be between values 0.05 and 20")
		}
		return nil
	}
	form := []*widget.FormItem{widget.NewFormItem("Gamma", entry)}
	dialog.ShowForm("Select gamma", "Ok", "Cancel", form,
		func(choice bool) {
			if !choice {
				return
			}
			gamma, _ := strconv.ParseFloat(entry.Text, 64) // No need to check thanks to validator
			ui.newImage(currentImage.GammaCorrection(gamma))
		},
		ui.MainWindow)
}

func (ui *UI) linearTransformationOp() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	entryForNumberOfPoints := widget.NewEntry()
	entryForNumberOfPoints.Validator = func(value string) error {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if valueInt < 2 {
			return fmt.Errorf("the number of points must greater than two")
		}
		return nil
	}
	dialog.ShowForm("How many points?", "OK", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("", entryForNumberOfPoints)},
		func(choice bool) {
			if !choice {
				return
			}
			var canvasPoints []fyne.CanvasObject
			var points []*histogram.Point
			pointsN, _ := strconv.Atoi(entryForNumberOfPoints.Text) // Error already checked in Validator
			for i := 0; i < pointsN; i++ {
				rawPoint, canvasPoint := histogram.NewPoint(i)
				points = append(points, rawPoint)
				canvasPoints = append(canvasPoints, canvasPoint)
			}
			showGraph := widget.NewButton("Graph", func() { // TODO what about duplicated?
				fmt.Println("Pim")
				validatedPoints := make([]*histogram.Point, 0, len(points))
				for _, point := range points {
					if err := point.Validate(); err != nil {
						continue
					}
					validatedPoints = append(validatedPoints, point)
				}
				fmt.Println("Validated points")
				fmt.Println(validatedPoints)
				fmt.Println("Pam")
			})
			content := container.NewVBox(container.NewVBox(canvasPoints...), showGraph)
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
					ui.newImage(currentImage.LinearTransformation(points))
				},
				ui.MainWindow)
		},
		ui.MainWindow)
}

func (ui *UI) infoView() {
	currentImage, err := ui.getCurrentImage()
	if err != nil {
		dialog.ShowError(err, ui.MainWindow)
		return
	}
	format := currentImage.Format()
	size := currentImage.Dimensions()
	message := fmt.Sprintf("Format: %v\n Size: %v bytes (%v x %v)\n", format, humanize.Bytes(uint64(size.X*size.Y)), size.X, size.Y)
	minColor, maxColor := currentImage.MinAndMaxColor()
	message += fmt.Sprintf("Range: [%v, %v]", minColor, maxColor)
	message += "\nBrightness: " + fmt.Sprintf("%f", currentImage.Brightness())
	message += "\nContrast: " + fmt.Sprintf("%f", currentImage.Contrast())
	entropy, numberOfColors := currentImage.EntropyAndNumberOfColors()
	message += "\nEntropy: " + strconv.Itoa(entropy) + " with " + strconv.Itoa(numberOfColors) + " diferent colors"
	dialog.ShowInformation("Information", message, ui.MainWindow)
}

func (ui *UI) histogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(Histogram)")
	a.Resize(fyne.NewSize(500, 500))
	a.Show()

	image := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].Histogram[:]), drawing.ColorBlack)
	imageR := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramR[:]), drawing.ColorRed)
	imageG := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramG[:]), drawing.ColorGreen)
	imageB := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramB[:]), drawing.ColorBlue)

	image1 := canvas.NewImageFromImage(image)
	image2 := canvas.NewImageFromImage(imageR)
	image3 := canvas.NewImageFromImage(imageG)
	image4 := canvas.NewImageFromImage(imageB)

	content := container.New(layout.NewAdaptiveGridLayout(2), image1, image2, image3, image4)

	a.SetContent(content)
}

func (ui *UI) accumulativeHistogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(AccumulativeHistogram)")
	a.Resize(fyne.NewSize(500, 500))
	a.Show()

	image := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulative[:]), drawing.ColorBlack)
	imageR := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeR[:]), drawing.ColorRed)
	imageG := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeG[:]), drawing.ColorGreen)
	imageB := ui.calculateHistogramGraph(convertToFloat(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeB[:]), drawing.ColorBlue)

	image1 := canvas.NewImageFromImage(image)
	image2 := canvas.NewImageFromImage(imageR)
	image3 := canvas.NewImageFromImage(imageG)
	image4 := canvas.NewImageFromImage(imageB)

	content := container.New(layout.NewAdaptiveGridLayout(2), image1, image2, image3, image4)

	a.SetContent(content)
}

func (ui *UI) normalizedHistogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(NormalizedHistogram)")
	a.Resize(fyne.NewSize(500, 500))
	a.Show()

	image := ui.calculateHistogramGraph(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalized[:], drawing.ColorBlack)
	imageR := ui.calculateHistogramGraph(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedR[:], drawing.ColorRed)
	imageG := ui.calculateHistogramGraph(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedG[:], drawing.ColorGreen)
	imageB := ui.calculateHistogramGraph(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedB[:], drawing.ColorBlue)

	image1 := canvas.NewImageFromImage(image)
	image2 := canvas.NewImageFromImage(imageR)
	image3 := canvas.NewImageFromImage(imageG)
	image4 := canvas.NewImageFromImage(imageB)

	content := container.New(layout.NewAdaptiveGridLayout(2), image1, image2, image3, image4)

	a.SetContent(content)
}

func (ui *UI) calculateHistogramGraph(valuesY []float64, color drawing.Color) image.Image {
	var indexValues []float64
	strokeColor := color
	fillColor := color
	fillColor.A = 128

	for i := 0; i < 256; i++ {
		indexValues = append(indexValues, float64(i))
	}
	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: strokeColor,
					FillColor:   fillColor,
				},
				XValues: indexValues,
				YValues: valuesY,
			},
		},
	}
	collector := &chart.ImageWriter{}
	graph.Render(chart.PNG, collector)

	image, err := collector.Image()
	if err != nil {
		log.Fatal(err)
	}
	return image
}

func (ui *UI) histogramEqual() {
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
			return
		}
		if reader == nil {
			return
		}
		img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
			ui.label, ui.MainWindow, ui.ROIcallback, ui.closeTabsCallback)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramIgualation(img))
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe"}))
	dialog.Show()

}

func (ui *UI) imgDifference() {
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
			return
		}
		if reader == nil {
			return
		}
		img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
			ui.label, ui.MainWindow, ui.ROIcallback, ui.closeTabsCallback)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].ImageDiference(img))
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe"}))
	dialog.Show()
}

func (ui *UI) imgChangeMap() {
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
			return
		}
		if reader == nil {
			return
		}
		img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
			ui.label, ui.MainWindow, ui.ROIcallback, ui.closeTabsCallback)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].ChangeMap(img))
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe"}))
	dialog.Show()

}

func convertToFloat(f32 []int) []float64 {
	f64 := make([]float64, len(f32))
	for i, f := range f32 {
		f64[i] = float64(f)
	}
	return f64
}
