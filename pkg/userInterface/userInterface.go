package userinterface

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"

	"github.com/vision-go/vision-go/pkg/histogram"
	ourimage "github.com/vision-go/vision-go/pkg/ourImage"
)

type UI struct {
	App          fyne.App
	MainWindow   fyne.Window
	tabs         *container.AppTabs
	label        *widget.Label
	tabsElements []ourimage.OurImage // Backend
	menu         *fyne.MainMenu
}

func (ui *UI) Init() {
	ui.tabs = container.NewAppTabs()
	ui.label = widget.NewLabel("")

	ui.menu = fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open", ui.openDialog),
			fyne.NewMenuItem("Save As...", ui.saveAsDialog),
		),
		fyne.NewMenu("Image",
			fyne.NewMenuItem("Negative", ui.negativeOp),
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

func (ui *UI) openDialog() {
	dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		if err == nil && reader == nil {
			return
		}
		img, err := ourimage.NewImage(reader.URI().Path(), ui.label)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(img, reader.URI().Name())
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe"}))
	dialog.Show()
}

func (ui *UI) saveAsDialog() {
	dialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		if err == nil && writer == nil {
			return
		}
		if len(ui.tabs.Items) == 0 {
			dialog.ShowInformation("Error", "You must open atleast one image", ui.MainWindow)
			return
		}
		outputFile, errFile := os.Create(writer.URI().Path())
		if errFile != nil {
			dialog.ShowError(errFile, ui.MainWindow)
		}

		png.Encode(outputFile, ui.tabsElements[ui.tabs.SelectedIndex()].Image.Image)

		outputFile.Close()
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe"}))
	dialog.SetFileName(ui.tabs.Selected().Text)
	dialog.Show()
}

func (ui *UI) newImage(img ourimage.OurImage, name string) {
	ui.tabs.Append(container.NewTabItem(name, container.NewScroll(container.New(layout.NewCenterLayout(), &img))))
	ui.tabs.SelectIndex(len(ui.tabs.Items) - 1) // Select the last one
	ui.tabsElements = append(ui.tabsElements, img)
}

func (ui *UI) negativeOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ourimage.Negative((ui.tabsElements[ui.tabs.SelectedIndex()])), ui.tabs.Selected().Text+"(Negative)") // TODO Improve name
}

func (ui *UI) histogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(Histogram)")
	a.Resize(fyne.NewSize(500, 500))
	a.Show()

	image := ui.calculateHistogramGraph(&ui.tabsElements[ui.tabs.SelectedIndex()].Histogram,
		&ui.tabsElements[ui.tabs.SelectedInd5ex()].HistogramR,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramG,
		&ui.tabsElements[ui.tabs.SelectedIndex()].HistogramB)

	a.SetContent(canvas.NewImageFromImage(image))
}

func (ui *UI) accumulativeHistogram() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(Histogram)")
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
	a := ui.App.NewWindow(ui.tabs.Selected().Text + "(AccumulativeHistogram)")
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
