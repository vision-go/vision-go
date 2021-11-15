package userinterface

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

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

  histograms := fyne.NewMenuItem("Histograms", nil)
  histograms.ChildMenu = fyne.NewMenu("",
			fyne.NewMenuItem("Histogram", ui.histogram),
			fyne.NewMenuItem("Accumulative Histogram", ui.accumulativeHistogram),
			fyne.NewMenuItem("Normalized Histogram", ui.normalizedHistogram),
  )
	ui.menu = fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open", ui.openDialog),
			fyne.NewMenuItem("Save As...", ui.saveAsDialog),
		),
		fyne.NewMenu("Image",
			fyne.NewMenuItem("Negative", ui.negativeOp),
			fyne.NewMenuItem("Monochrome", ui.monochromeOp),
			fyne.NewMenuItem("Adjust Brightness/Contrast", ui.adjustBrightnessAndContrastOp),
			fyne.NewMenuItem("Linear Transformation", ui.linearTransformationOp),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Gamma Correction", ui.gammaCorrectionOp),
		),
		fyne.NewMenu("View",
			fyne.NewMenuItem("Info", ui.infoView),
      histograms,
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
