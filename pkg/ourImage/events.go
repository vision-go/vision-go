package ourimage

import (
	"image"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	RightTabs = iota
	AllTabs
	OtherTabs
)

func (ourimage *OurImage) MouseIn(mouse *desktop.MouseEvent) {
	if ourimage.statusBar != nil {
		r, g, b, a := ourimage.canvasImage.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
		ourimage.statusBar.SetText("x=" + strconv.Itoa(int(math.Round(float64(mouse.Position.X)))) + ", y=" + strconv.Itoa(int(math.Round(float64(mouse.Position.Y)))) +
			", R: " + strconv.Itoa(int(r>>8)) + " || G: " + strconv.Itoa(int(g>>8)) + " || B: " + strconv.Itoa(int(b>>8)) + " || A: " + strconv.Itoa(int(a>>8)))
	}
}

// MouseMoved is a hook that is called if the mouse pointer moved over the element.
func (ourimage *OurImage) MouseMoved(mouse *desktop.MouseEvent) {
	if ourimage.statusBar != nil {
		r, g, b, a := ourimage.canvasImage.Image.At(int(mouse.Position.X), int(mouse.Position.Y)).RGBA()
		ourimage.statusBar.SetText("x=" + strconv.Itoa(int(math.Round(float64(mouse.Position.X)))) + ", y=" + strconv.Itoa(int(math.Round(float64(mouse.Position.Y)))) +
			", R: " + strconv.Itoa(int(r>>8)) + " || G: " + strconv.Itoa(int(g>>8)) + " || B: " + strconv.Itoa(int(b>>8)) + " || A: " + strconv.Itoa(int(a>>8)))
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
				fyne.NewMenuItem("Close tabs to right",
					func() {
						ourimage.closeTabsCallback(RightTabs)
					},
				),
				fyne.NewMenuItem("Close all tabs",
					func() {
						ourimage.closeTabsCallback(AllTabs)
					},
				),
				fyne.NewMenuItem("Close other tabs",
					func() {
						ourimage.closeTabsCallback(OtherTabs)
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
func (ourimage *OurImage) Cursor() desktop.Cursor {
	return desktop.CrosshairCursor
}
