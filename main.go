package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"image/png"
	// "github.com/pkg/profile"
	"github.com/vision-go/vision-go/pkg/ourImage"
)

// const target = "bogart.tfe"
// const target = "animal.tfe"
// const target = "BUGS.tf"
// const target = "caras.tfe"
// const target = "IMPLANTE.tfe"
// const target = "MONTANIA.tfe"
// const target = "PLAYA2.tfe"
const target = "pumba.png"

func main() {
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	image, err := ourimage.NewImage(target)
	if err != nil {
		log.Fatal(err)
	}
  newImage := ourimage.Negative(image)
  f, err := os.Create(strings.ReplaceAll(target, ".tfe", ".png"))
  if err != nil {
    log.Fatal(err)
  }
	err = png.Encode(f, newImage)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Pixel at {40,50}: r:%[1]v, g:%[1]v, b:%[1]v\n", image.At(40, 50))
}
