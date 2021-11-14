package ourimage

import (
	"strconv"
	"math"
	"image"

	"fyne.io/fyne/v2"
  "fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Events //
func (self *OurImage) MouseIn(mouse *desktop.MouseEvent) {
	if self.statusBar != nil {
		r, g, b, a := self.canvasImage.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
		self.statusBar.SetText("R: " + strconv.Itoa(int(r>>8)) + " || G: " + strconv.Itoa(int(g>>8)) + " || B: " + strconv.Itoa(int(b>>8)) + " || A: " + strconv.Itoa(int(a>>8)))
	}
}

// MouseMoved is a hook that is called if the mouse pointer moved over the element.
func (self *OurImage) MouseMoved(mouse *desktop.MouseEvent) {
	if self.statusBar != nil {
		r, g, b, a := self.canvasImage.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
		self.statusBar.SetText("R: " + strconv.Itoa(int(r>>8)) + " || G: " + strconv.Itoa(int(g>>8)) + " || B: " + strconv.Itoa(int(b>>8)) + " || A: " + strconv.Itoa(int(a>>8)))
	}
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (ourimage *OurImage) MouseOut() {
	if ourimage.statusBar != nil {
		ourimage.statusBar.SetText("")
	}
}

// desktop.Mouseable
func (ourimage *OurImage) MouseDown(mouseEvent *desktop.MouseEvent) {
	if mouseEvent.Button == desktop.MouseButtonSecondary {
		popUp := widget.NewPopUpMenu(
			fyne.NewMenu("PopUp",
				fyne.NewMenuItem("Duplicate",
					func() {
						ourimage.newImageCallback(ourimage) // TODO am I really duplicating the image?
					},
				),
			),
			ourimage.mainWindow.Canvas(),
		)
		popUp.ShowAtPosition(mouseEvent.AbsolutePosition)
	}
	ourimage.rectangle.Min = image.Pt(int(math.Round(float64(mouseEvent.Position.X))), int(math.Round(float64(mouseEvent.Position.Y))))
}

func (ourimage *OurImage) MouseUp(mouseEvent *desktop.MouseEvent) {
	ourimage.rectangle.Max = image.Pt(int(math.Round(float64(mouseEvent.Position.X))), int(math.Round(float64(mouseEvent.Position.Y))))
	if ourimage.rectangle.Dx() > 10 && ourimage.rectangle.Dy() > 10 {
		ourimage.ROIcallback(ourimage.ROI(ourimage.rectangle))
	}
}
