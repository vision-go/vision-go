package userinterface

import (
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/dustin/go-humanize"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"github.com/vision-go/vision-go/pkg/histogram"
	ourimage "github.com/vision-go/vision-go/pkg/ourImage"
)

type UI struct {
	App          fyne.App
	MainWindow   fyne.Window
	tabs         *container.DocTabs
	label        *widget.Label
	tabsElements []*ourimage.OurImage // Backend
	menu         *fyne.MainMenu
}

func (ui *UI) Init() {
	ui.tabs = container.NewDocTabs()
	ui.tabs.Hide()
	ui.tabs.CloseIntercept = func(tabItem *container.TabItem) {
		ui.tabs.Select(tabItem)
		dialog := dialog.NewConfirm("Close", "Are you sure you want to close "+tabItem.Text+" ?",
			func(choice bool) {
				if !choice {
					return
				}
				ui.removeImage(ui.tabs.SelectedIndex(), tabItem)
			},
			ui.MainWindow)
		dialog.Show()
	}

	ui.label = widget.NewLabel("")

	ui.menu = fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open", ui.openDialog),
			fyne.NewMenuItem("Save As...", ui.saveAsDialog),
		),
		fyne.NewMenu("Image",
			fyne.NewMenuItem("Negative", ui.negativeOp),
			fyne.NewMenuItem("Monochrome", ui.monochromeOp),
			fyne.NewMenuItem("Adjust Brightness/Contrast", ui.adjustBrightnessAndContrast),
			fyne.NewMenuItem("Linear Transformation", ui.linearTransformationOp),
		),
		fyne.NewMenu("View",
			fyne.NewMenuItem("Info", ui.infoView),
			fyne.NewMenuItem("Histogram", ui.histogram),
			fyne.NewMenuItem("Accumulative Histogram", ui.accumulativeHistogram),
			fyne.NewMenuItem("Normalized Histogram", ui.normalizedHistogram),
		),
	)

	ui.MainWindow.SetMainMenu(ui.menu)
	ui.MainWindow.Resize(fyne.NewSize(500, 500))
	ui.MainWindow.SetContent(container.NewBorder(nil, ui.label, nil, nil, ui.tabs))
	ui.MainWindow.ShowAndRun()
}

func (ui *UI) getCurrentImage() (*ourimage.OurImage, error) {
	if ui.tabs.SelectedIndex() == -1 {
		return nil, fmt.Errorf("no image selected")
	}
	return ui.tabsElements[ui.tabs.SelectedIndex()], nil
}

func (ui *UI) openDialog() {
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
			return
		}
		if reader == nil {
			return
		}
		img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
			ui.label, ui.MainWindow, ui.ROIcallback, ui.newImage)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(img)
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe"}))
	dialog.Show()
}

func (ui *UI) saveAsDialog() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	selectionWidget := widget.NewRadioGroup([]string{"png", "jpg"}, func(string) {})
	selectionWidget.SetSelected("png")
	dialog.ShowCustomConfirm("Select format", "Ok", "Cancel", selectionWidget,
		func(choice bool) {
			if !choice {
				return
			}
			dialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					dialog.ShowError(err, ui.MainWindow)
				}
				if writer == nil {
					return
				}
				outputFile, errFile := os.Create(writer.URI().Path())
				if errFile != nil {
					dialog.ShowError(errFile, ui.MainWindow)
				}
				defer outputFile.Close()

				img, _ := ui.getCurrentImage() // Already checked
				err = img.Save(outputFile, selectionWidget.Selected)
				if err != nil {
					dialog.ShowError(err, ui.MainWindow)
				}
			}, ui.MainWindow)
			formatedName := func(originalName, format string) string {
				pointIndex := strings.LastIndex(originalName, ".")
				if pointIndex == -1 {
					return originalName + "." + format
				}
				return originalName[:pointIndex+1] + format
			}(ui.tabs.Selected().Text, selectionWidget.Selected)
			dialog.SetFileName(formatedName)
			dialog.Show()
		},
		ui.MainWindow)
}

func (ui *UI) ROIcallback(cropped *ourimage.OurImage) {
	dialog.ShowCustomConfirm("Do you want this sub-image?", "Ok", "Cancel", container.NewCenter(cropped),
		func(choice bool) {
			if !choice {
				return
			}
			ui.newImage(cropped)
		},
		ui.MainWindow)
}

func (ui *UI) newImage(img *ourimage.OurImage) {
	ui.tabs.Append(container.NewTabItem(img.Name(), container.NewScroll(container.New(layout.NewCenterLayout(), img))))
	ui.tabs.SelectIndex(len(ui.tabs.Items) - 1) // Select the last one
	ui.tabsElements = append(ui.tabsElements, img)
	if len(ui.tabsElements) != 0 {
		ui.tabs.Show()
	}
}

func (ui *UI) removeImage(index int, tabItem *container.TabItem) {
	ui.tabsElements = append(ui.tabsElements[:index], ui.tabsElements[index+1:]...)
	ui.tabs.Remove(tabItem)
	if len(ui.tabsElements) == 0 {
		ui.tabs.Hide()
	}
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
	dialog := dialog.NewInformation("Information", message, ui.MainWindow)
	dialog.Show()
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
		GrayGraph = append(GrayGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalized.Values[i]))
		RedGraph = append(RedGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedR.Values[i]))
		GreenGraph = append(GreenGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedG.Values[i]))
		BlueGraph = append(BlueGraph, float64(ui.tabsElements[ui.tabs.SelectedIndex()].HistogramNormalizedB.Values[i]))
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
