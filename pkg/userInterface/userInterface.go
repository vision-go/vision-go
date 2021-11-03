package userinterface

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	humanize "github.com/dustin/go-humanize"

	"github.com/vision-go/vision-go/pkg/ourImage"
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
				if choice == false {
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
			fyne.NewMenuItem("Save", nil),
		),
		fyne.NewMenu("Image",
			fyne.NewMenuItem("Negative", ui.negativeOp),
			fyne.NewMenuItem("Monochrome", ui.monochromeOp),
      fyne.NewMenuItem("Linear Transformation", ui.linearTransformationOp),
		),
		fyne.NewMenu("View",
			fyne.NewMenuItem("Info", ui.infoView),
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
			return
		}
		if reader == nil {
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
	ui.tabsElements = append(ui.tabsElements, &img)
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
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	format := ui.tabsElements[ui.tabs.SelectedIndex()].Format()
	size := ui.tabsElements[ui.tabs.SelectedIndex()].Dimensions()
	message := fmt.Sprintf("Format: %v\n Size: %v bytes (%v x %v)", format, humanize.Bytes(uint64(size.X*size.Y)), size.X, size.Y)
	dialog := dialog.NewInformation("Information", message, ui.MainWindow)
	dialog.Show()
}
