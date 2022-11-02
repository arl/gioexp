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
	"gioui.org/widget"
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
	Theme *material.Theme

	PropertyList *PropertyList
	prop1, prop2 *StringProperty

	btn widget.Clickable
}

func NewUI(theme *material.Theme) *UI {
	ui := &UI{
		Theme:        theme,
		PropertyList: NewPropertyList(),
		prop1: &StringProperty{
			Label: "prop 1",
			Value: "value 1",
			// BgColor: lightGrey,
			Theme: theme, // TODO(arl) theme should be passed to layout?
		},
		prop2: &StringProperty{
			Label: "prop 2",
			Value: "value 2",
			// BgColor: lightGrey,
			Theme: theme, // TODO(arl) theme should be passed to layout?
		},
	}
	ui.PropertyList.MaxHeight = 300
	ui.prop1.SetEditable(true)
	ui.prop2.SetEditable(true)
	ui.PropertyList.Add(ui.prop1)
	ui.PropertyList.Add(ui.prop2)
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
	if ui.btn.Clicked() {
		ui.prop1.SetEditable(!ui.prop1.IsEditable())
	}

	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceEnd,
			}.Layout(gtx,
				layout.Rigid(ui.PropertyList.Layout),
				layout.Rigid(func(gtx C) D {
					return material.Button(ui.Theme, &ui.btn, "toggle editable").Layout(gtx)
				}),
			)
		}),
	)
}
