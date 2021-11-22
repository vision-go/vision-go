package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/vision-go/vision-go/pkg/userInterface"
)

func main() {
	a := app.New()
	w := a.NewWindow("vision-go")
	w.SetOnClosed(a.Quit)
	ui := userinterface.UI{App: a, MainWindow: w}

	ui.Init()
}
