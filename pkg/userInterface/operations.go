package userinterface

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (ui *UI) negativeOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Negative(), ui.tabs.Selected().Text+"(Negative)") // TODO Improve name
}

func (ui *UI) monochromeOp() {
	if ui.tabs.SelectedIndex() == -1 {
		dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
		return
	}
	ui.newImage(ui.tabsElements[ui.tabs.SelectedIndex()].Monochrome(), ui.tabs.Selected().Text+"(Monochrome)") // TODO Improve name
}

func (ui *UI) linearTransformationOp() {
  // if ui.tabs.SelectedIndex() == -1 {
  //   dialog.ShowError(fmt.Errorf("no image selected"), ui.MainWindow)
  //   return
  // }
  point1 := container.NewGridWithColumns(2, widget.NewLabel("Point1"), widget.NewEntry())
  points := []fyne.CanvasObject{point1}
  var addButton *widget.Button
  withAddButton := func(points []fyne.CanvasObject) []fyne.CanvasObject {
    return append(points, addButton)
  }
  pointsContainer := container.NewVBox(withAddButton(points)...)
  addButton = widget.NewButtonWithIcon("", theme.ContentAddIcon(), 
    func() {
      points = append(points, point1)
      pointsContainer = container.NewVBox(withAddButton(points)...)
    })
  // content := container.NewGridWithRows(2, point, addButton)
  dialog.ShowCustomConfirm("Linear Transformation", "OK", "Cancel", pointsContainer, func(bool) {

  }, ui.MainWindow)
}
