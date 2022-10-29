package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Zoomable: experiment at making a widget which content can be zoomed/panned
// using the mouse wheel, while keeping the point under the cursor immobile
// (i.e. using the mouse cursor as scale origin).

func main() {
	ui := NewUI(material.NewTheme(gofont.Collection()))

	go func() {
		w := app.NewWindow(app.Title("Property List"))
		if err := ui.Run(w); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}

type UI struct {
	Theme        *material.Theme
	PropertyList *PropertyList
}

func NewUI(theme *material.Theme) *UI {
	ui := &UI{
		Theme:        theme,
		PropertyList: NewPropertyList(),
	}
	prop1 := &StringProperty{
		Label:   "prop 1",
		Value:   "value2",
		BgColor: lightGrey,
		Theme:   theme, // TODO(arl) theme should be passed to layout?
	}
	prop2 := &StringProperty{
		Label:   "prop 2",
		Value:   "value2",
		BgColor: lightGrey,
		Theme:   theme, // TODO(arl) theme should be passed to layout?
	}
	ui.PropertyList.Add(prop1)
	ui.PropertyList.Add(prop2)
	return ui
}

func (ui *UI) Run(w *app.Window) error {
	var ops op.Ops
	for e := range w.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			ui.Layout(gtx)
			e.Frame(gtx.Ops)

		case key.Event:
			if e.Name == key.NameEscape {
				return nil
			}
		case system.DestroyEvent:
			return e.Err
		}
	}

	return nil
}

func (ui *UI) Layout(gtx C) D {
	return ui.PropertyList.Layout(gtx)
}
