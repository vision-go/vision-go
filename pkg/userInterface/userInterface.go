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
	progessBar   *widget.ProgressBarInfinite
	tabsElements []*ourimage.OurImage // To avoid reflection on tabs
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
				if err := ui.removeImage(ui.tabs.SelectedIndex()); err != nil {
					dialog.ShowError(err, ui.MainWindow)
				}
			},
			ui.MainWindow)
		dialog.Show()
	}

	ui.label = widget.NewLabel("")
	ui.progessBar = widget.NewProgressBarInfinite()
	ui.progessBar.Hide()
	ui.progessBar.Stop()

	histograms := fyne.NewMenuItem("Histograms", nil)
	histograms.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Histogram", ui.histogram),
		fyne.NewMenuItem("Accumulative Histogram", ui.accumulativeHistogram),
		fyne.NewMenuItem("Normalized Histogram", ui.normalizedHistogram),
	)
	mirror := fyne.NewMenuItem("Mirror", nil)
	mirror.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Horizontal", ui.horizontal),
		fyne.NewMenuItem("Vertical", ui.vertical),
	)
	rotate := fyne.NewMenuItem("Rotate", nil)
	rotate.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Right", ui.rotateRight),
		fyne.NewMenuItem("Left", ui.rotateLeft),
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
			fyne.NewMenuItem("Difference", ui.imgDifference),
			fyne.NewMenuItem("Change Map From..", ui.imgChangeMap),
			fyne.NewMenuItem("Equalization", ui.equializationOp),
			fyne.NewMenuItem("Histogram Igualation", ui.histogramEqual),
		),
		fyne.NewMenu("View",
			fyne.NewMenuItem("Info", ui.infoView),
			histograms,
		),
		fyne.NewMenu("Transformation",
			mirror,
			rotate,
			fyne.NewMenuItem("Transpose", ui.transpose),
		),
	)

	ui.MainWindow.SetMainMenu(ui.menu)
	ui.MainWindow.Resize(fyne.NewSize(500, 500))
	ui.MainWindow.SetContent(container.NewBorder(nil, container.NewBorder(nil, nil, ui.label, ui.progessBar), nil, nil, ui.tabs))
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
		ui.progessBar.Start()
		ui.progessBar.Show()
		img, err := ourimage.NewFromPath(reader.URI().Path(), reader.URI().Name(),
			ui.label, ui.MainWindow, ui.ROIcallback, ui.closeTabsCallback)
		if err != nil {
			dialog.ShowError(err, ui.MainWindow)
		}
		ui.newImage(img)
		ui.progessBar.Hide()
		ui.progessBar.Stop()
	}, ui.MainWindow)
	dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe", ".tif"}))
	dialog.Show()
}

func (ui *UI) saveAsDialog() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	selectionWidget := widget.NewRadioGroup([]string{"png", "jpg", "tif"}, func(string) {})
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
			dialog.SetFilter(storage.NewExtensionFileFilter([]string{"." + selectionWidget.Selected}))
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

func (ui *UI) removeImage(index int) error {
	if index > len(ui.tabsElements) || index < 0 {
		return fmt.Errorf("you have tried to delete a image with index:%v\n. Number of images loaded:%v\n ", index, len(ui.tabsElements))
	}
	ui.tabsElements = append(ui.tabsElements[:index], ui.tabsElements[index+1:]...)
	ui.tabs.Remove(ui.tabs.Items[index])
	if len(ui.tabsElements) == 0 {
		ui.tabs.Hide()
	}
	return nil
}

func (ui *UI) closeTabsCallback(closeChoice int) {
	currentImageIndex := ui.tabs.SelectedIndex()
	if currentImageIndex == -1 {
		dialog.ShowError(fmt.Errorf("internal: no image selected"), ui.MainWindow)
		return
	}
	switch closeChoice {
	case ourimage.RightTabs: // TODO check for error
		for i := len(ui.tabsElements) - 1; i > currentImageIndex; i-- {
			if err := ui.removeImage(i); err != nil {
				dialog.ShowError(err, ui.MainWindow)
			}
		}
	case ourimage.AllTabs:
		for i := len(ui.tabsElements) - 1; i >= 0; i-- {
			if err := ui.removeImage(i); err != nil {
				dialog.ShowError(err, ui.MainWindow)
			}
		}
	case ourimage.OtherTabs:
		for i := len(ui.tabsElements) - 1; i >= 0; i-- {
			if i == currentImageIndex {
				continue
			}
			if err := ui.removeImage(i); err != nil {
				dialog.ShowError(err, ui.MainWindow)
			}
		}
	default:
		dialog.ShowError(fmt.Errorf("that is not a valid option in the pop-up menu"), ui.MainWindow)
		return
	}
}
