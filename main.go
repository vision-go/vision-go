package main

import (
	// "fmt"
	"fmt"
	"log"
	"strings"

	// "github.com/pkg/profile"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	// "fyne.io/fyne/v2/theme"
	"github.com/vision-go/vision-go/pkg/ourImage"
)

// const target = "bogart.tfe"
// const target = "animal.tfe"
// const target = "BUGS.tf"
const target = "caras.tfe"
// const target = "IMPLANTE.tfe"
// const target = "MONTANIA.tfe"
// const target = "PLAYA2.tfe"

func main() {
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	image, err := ourimage.NewImage("images/" + target)
	if err != nil {
		log.Fatal(err)
	}
	err = image.Encode(strings.ReplaceAll(target, ".tfe", ".png"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(image.At(40, 50).RGBA())
	myApp := app.New()
	w := myApp.NewWindow("Vision")
  w.Resize(fyne.NewSize(float32(image.Bounds().Dx()), float32(image.Bounds().Dy())))
  imageCanvas := canvas.NewImageFromImage(image)
	imageCanvas.FillMode = canvas.ImageFillContain
	w.SetContent(imageCanvas)
	w.ShowAndRun()
}
