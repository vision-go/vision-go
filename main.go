package main

import (
	"fyne.io/fyne/v2/app"
	userinterface "github.com/vision-go/vision-go/pkg/userInterface"
)

func main() {
  a := app.New()
  w := a.NewWindow("vision-go")
  ui := userinterface.UI{App: a, MainWindow: w}
	
  ui.Init()
}
