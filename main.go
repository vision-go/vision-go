package main

import (

	// "github.com/pkg/profile"

	"fyne.io/fyne/v2/app"
	userinterface "github.com/vision-go/vision-go/pkg/userInterface"
)

// const target = "bogart.tfe"
// const target = "animal.tfe"
// const target = "BUGS.tf"
//const target = "caras.tfe"
// const target = "IMPLANTE.tfe"
// const target = "MONTANIA.tfe"
// const target = "PLAYA2.tfe"

func main() {
  // defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
  // img, err := ourimage.NewImage("images/" + target)
  // if err != nil {
  //   log.Fatal(err)
  // }
  a := app.New()
  w := a.NewWindow("vision-go")
  ui := userinterface.UI{App: a, MainWindow: w}
	
  ui.Init()
}
