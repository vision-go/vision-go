package userinterface

import (
	"fmt"
	"image"
	"image/draw"
	"reflect"

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
    fyne.NewMenu("Image", 
      fyne.NewMenuItem("Negative", ui.negativeOp),
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
  dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", ".jpg", ".tfe"}))
  dialog.Show()
}

func (ui *UI) newImage(img image.Image, name string) {
  // image := canvas.NewRasterFromImage(img)
  image := canvas.NewImageFromImage(img) // Zeus check this out :p
  image.FillMode = canvas.ImageFillContain
  ui.tabs.Append(container.NewTabItem(name, image))
  fmt.Println(reflect.TypeOf(ui.tabs.Selected()))
}

func (ui *UI) negativeOp() {
  if ui.tabs.SelectedIndex() == -1 {
    dialog.ShowError(fmt.Errorf("No image selected"), ui.MainWindow)
    return
  }
  canvasImage, ok := ui.tabs.Selected().Content.(*canvas.Image)
  if !ok {
    dialog.ShowError(fmt.Errorf("The content in this tab is not an canvas.Image"), ui.MainWindow)
    return
  }
  img, ok := canvasImage.Image.(draw.Image)
  ui.newImage(ourimage.Negative(img), ui.tabs.CurrentTab().Text + "(Negative)")
}
