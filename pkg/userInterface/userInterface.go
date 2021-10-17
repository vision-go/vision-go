package userinterface

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	ourimage "github.com/vision-go/vision-go/pkg/ourImage"
)

type UI struct {
  App fyne.App
  MainWindow fyne.Window
  tabs *container.AppTabs
  menu *fyne.MainMenu
}

func (ui *UI) Init() {
  ui.tabs = container.NewAppTabs()

  ui.menu = fyne.NewMainMenu(
    fyne.NewMenu("File", 
      fyne.NewMenuItem("Open", ui.openDialog),
      fyne.NewMenuItem("Save", nil),
    ),
  )

  ui.MainWindow.SetMainMenu(ui.menu)
  ui.MainWindow.Resize(fyne.NewSize(500, 500))
  ui.MainWindow.SetContent(ui.tabs)
  ui.MainWindow.ShowAndRun()
}

func (ui *UI) openDialog() {
  fmt.Println("Open file")
  dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
    if err != nil {
      dialog.ShowError(err, ui.MainWindow)
    }
    if err == nil && reader == nil {
      return
    }
    // fmt.Println(reader.URI().Path()) // TODO Does this work on Windows and Mac?
    img, err := ourimage.NewImage(reader.URI().Path())
    if err != nil {
      dialog.ShowError(err, ui.MainWindow)
    }
    ui.newImage(img, reader.URI().Name())
  }, ui.MainWindow)
  dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", "jpg"}))
  dialog.Show()
}

func (ui *UI) newImage(img image.Image, name string) {
  image := canvas.NewRasterFromImage(img)
  ui.tabs.Append(container.NewTabItem(name, image))
}