package main

import (
	// "fmt"
	"fmt"
	"image"

	// "log"

	// "strings"

	// "github.com/pkg/profile"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	// "fyne.io/fyne/v2/widget"
	"github.com/vision-go/vision-go/pkg/ourImage"
)

// const target = "bogart.tfe"
// const target = "animal.tfe"
// const target = "BUGS.tf"
const target = "caras.tfe"
// const target = "IMPLANTE.tfe"
// const target = "MONTANIA.tfe"
// const target = "PLAYA2.tfe"

type UI struct {
  app fyne.App
  mainWindow fyne.Window
  tabs *container.AppTabs
  menu *fyne.MainMenu
}

func (ui *UI) init() {
  ui.tabs = container.NewAppTabs()

  ui.menu = fyne.NewMainMenu(
    fyne.NewMenu("File", 
      fyne.NewMenuItem("Open", ui.openDialog),
      fyne.NewMenuItem("Save", nil),
    ),
  )

  ui.mainWindow.SetMainMenu(ui.menu)
  ui.mainWindow.Resize(fyne.NewSize(500, 500))
  ui.mainWindow.SetContent(ui.tabs)
  ui.mainWindow.ShowAndRun()
}

func (ui *UI) openDialog() {
  fmt.Println("Open file")
  dialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
    if err != nil {
      dialog.ShowError(err, ui.mainWindow)
    }
    if err == nil && reader == nil {
      return
    }
    // fmt.Println(reader.URI().Path()) // TODO Does this work on Windows and Mac?
    img, err := ourimage.NewImage(reader.URI().Path())
    if err != nil {
      dialog.ShowError(err, ui.mainWindow)
    }
    ui.newImage(img, reader.URI().Name())
  }, ui.mainWindow)
  dialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpeg", "jpg"}))
  dialog.Show()
}

func (ui *UI) newImage(img image.Image, name string) {
  image := canvas.NewImageFromImage(img)
  image.FillMode = canvas.ImageFillOriginal
  ui.tabs.Append(container.NewTabItem(name, image))
}


func main() {
  // defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
  // img, err := ourimage.NewImage("images/" + target)
  // if err != nil {
  //   log.Fatal(err)
  // }
  a := app.New()
  w := a.NewWindow("vision-go")
  ui := UI{app: a, mainWindow: w}

  ui.init()
}
