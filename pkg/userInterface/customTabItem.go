package userinterface

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type customTabItem struct {
	widget.BaseWidget
  tabItem *container.TabItem
}
func New(text string, content fyne.CanvasObject) *customTabItem {
  newTab := &customTabItem{tabItem: container.NewTabItem(text, content)}
  newTab.ExtendBaseWidget(newTab)
  return newTab
}

func (tab *customTabItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(tab.tabItem.Content)
}

func (tab *customTabItem) MouseIn(mouse *desktop.MouseEvent) {
  fmt.Println("MouseIn")
}

func (tab *customTabItem) MouseMoved(mouse *desktop.MouseEvent) {
  fmt.Println("MouseMoved")
}

func (tab *customTabItem) MouseOut(mouse *desktop.MouseEvent) {
  fmt.Println("MouseOut")
}
