package userinterface

import (
	"fmt"

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
			fyne.NewMenuItem("Save", nil),
		),
		fyne.NewMenu("Image",
			fyne.NewMenuItem("Negative", ui.negativeOp),
			fyne.NewMenuItem("Monochrome", ui.monochromeOp),
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
	ui.newImage(ourimage.Negative(ui.tabsElements[ui.tabs.SelectedIndex()]), ui.tabs.Selected().Text+"(Negative)") // TODO Improve name
}

func (ui *UI) monochromeOp() {
  if ui.tabs.SelectedIndex() == -1 {
    dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
    return
  }
	ui.newImage(ourimage.Monochrome(ui.tabsElements[ui.tabs.SelectedIndex()]), ui.tabs.Selected().Text+"(Monochrome)") // TODO Improve name
}
