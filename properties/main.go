package main

import (
	"fmt"
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

	PropertyList        *PropertyList
	prop1, prop2, prop3 *Property

	btn widget.Clickable
}

var p1 UIntValue = 123456
var p2 UIntValue = 1234
var p3 Float64Value = .2

func NewUI(theme *material.Theme) *UI {
	prop1 := NewProperty("0123456789", &p1)
	prop1.Label = "Property 1"
	prop1.Background = aliceBlue
	prop1.SetEditable(true)

	prop2 := NewProperty("", &p2)
	prop2.Label = "Property 1"
	prop2.Background = aliceBlue
	prop2.SetEditable(true)

	prop3 := NewProperty("", &p3)
	prop3.Label = "Float64"
	prop3.Background = aliceBlue
	prop3.SetEditable(true)

	ui := &UI{
		Theme:        theme,
		PropertyList: NewPropertyList(),
		prop1:        prop1,
		prop2:        prop2,
		prop3:        prop3,
	}
	ui.PropertyList.MaxHeight = 300

	ui.PropertyList.Add(ui.prop1)
	ui.PropertyList.Add(ui.prop2)
	ui.PropertyList.Add(ui.prop3)
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

			fmt.Println(p1, p2)

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
		ui.prop1.SetEditable(!ui.prop1.Editable())
	}

	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceEnd,
			}.Layout(gtx,
				layout.Rigid(func(gtx C) D {
					return ui.PropertyList.Layout(ui.Theme, gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return material.Button(ui.Theme, &ui.btn, "toggle editable").Layout(gtx)
				}),
			)
		}),
	)
}
