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
	"fyne.io/fyne/v2/widget"

	"github.com/dustin/go-humanize"
	"github.com/vision-go/vision-go/pkg/histogram"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

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

// Operations

func (ui *UI) histogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(Histogram)")
	a.Resize(fyne.NewSize(500, 500))
	a.Show()

	image := ui.calculateHistogramGraph(&ui.tabsElements[ui.tabs.SelectedIndex()].Histogram,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramR,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramG,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramB)

	a.SetContent(canvas.NewImageFromImage(image))
}

func (ui *UI) accumulativeHistogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(AccumulativeHistogram)")
	a.Resize(fyne.NewSize(500, 500))
	a.Show()

	image := ui.calculateHistogramGraph(&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulative,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeR,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeG,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramAccumulativeB)

	a.SetContent(canvas.NewImageFromImage(image))
}

func (ui *UI) normalizedHistogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(Normalized Histogram)")
	a.Resize(fyne.NewSize(500, 500))
	a.Show()

	var RedGraph []float64
	var GreenGraph []float64
	var BlueGraph []float64

	var GrayGraph []float64
	var indexValues []float64

	for i := 0; i < 256; i++ {
		indexValues = append(indexValues, float64(i))
		GrayGraph = append(GrayGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalized[i]))
		RedGraph = append(RedGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedR[i]))
		GreenGraph = append(GreenGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedG[i]))
		BlueGraph = append(BlueGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedB[i]))
	}
	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: drawing.Color{
						R: 255,
						G: 0,
						B: 0,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 255,
						G: 0,
						B: 0,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: RedGraph,
			},
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: drawing.Color{
						R: 0,
						G: 255,
						B: 0,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 0,
						G: 255,
						B: 0,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: GreenGraph,
			},
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeWidth: 2.0,
					StrokeColor: drawing.Color{
						R: 0,
						G: 0,
						B: 255,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 0,
						G: 0,
						B: 255,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: BlueGraph,
			},
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: drawing.Color{
						R: 0,
						G: 0,
						B: 0,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 0,
						G: 0,
						B: 0,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: GrayGraph,
			},
		},
	}
	collector := &chart.ImageWriter{}
	graph.Render(chart.PNG, collector)

	image, err := collector.Image()
	if err != nil {
		log.Fatal(err)
	}
	a.SetContent(canvas.NewImageFromImage(image))
}

func (ui *UI) calculateHistogramGraph(grey *histogram.Histogram, red *histogram.Histogram, green *histogram.Histogram, blue *histogram.Histogram) image.Image {

	var RedGraph []float64
	var GreenGraph []float64
	var BlueGraph []float64

	var GrayGraph []float64
	var indexValues []float64

	for i := 0; i < 256; i++ {
		indexValues = append(indexValues, float64(i))
		GrayGraph = append(GrayGraph, float64(grey.At(i)))
		RedGraph = append(RedGraph, float64(red.At(i)))
		GreenGraph = append(GreenGraph, float64(green.At(i)))
		BlueGraph = append(BlueGraph, float64(blue.At(i)))
	}
	graph := chart.Chart{
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: drawing.Color{
						R: 255,
						G: 0,
						B: 0,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 255,
						G: 0,
						B: 0,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: RedGraph,
			},
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: drawing.Color{
						R: 0,
						G: 255,
						B: 0,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 0,
						G: 255,
						B: 0,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: GreenGraph,
			},
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeWidth: 2.0,
					StrokeColor: drawing.Color{
						R: 0,
						G: 0,
						B: 255,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 0,
						G: 0,
						B: 255,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: BlueGraph,
			},
			chart.ContinuousSeries{
				Style: chart.Style{
					StrokeColor: drawing.Color{
						R: 0,
						G: 0,
						B: 0,
						A: 255,
					},
					FillColor: drawing.Color{
						R: 0,
						G: 0,
						B: 0,
						A: 127,
					},
				},
				XValues: indexValues,
				YValues: GrayGraph,
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

// TODO Improve this PLS
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
            continue;
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
