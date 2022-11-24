package main

import (
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"

	"github.com/arl/gioexp/component/property"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

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

	PropertyList        *property.List
	prop1, prop2, prop3 *property.Property

	btn widget.Clickable

	// Modal can show widgets atop the rest of the ui.
	Modal component.ModalState
}

var (
	aliceBlue = color.NRGBA{R: 240, G: 248, B: 255, A: 255}
)

func NewUI(theme *material.Theme) *UI {
	prop1 := property.NewUInt(123456)
	prop1.Label = "Property 1"
	prop1.Editable = true

	prop2 := property.NewString("", &p2val)
	var p2val property.UInt = 123
	prop2.Label = "Property 1"
	prop2.Editable = true

	prop3 := property.NewFloat64(.2)
	prop3.Label = "Float64"
	prop3.Editable = true

	ui := &UI{
		Theme: theme,
		prop1: prop1,
		prop2: prop2,
		prop3: prop3,
	}

	ui.Modal.VisibilityAnimation.Duration = time.Millisecond * 250

	plist := property.NewList(&ui.Modal)
	plist.MaxHeight = 300
	plist.Add(ui.prop1)
	plist.Add(ui.prop2)
	plist.Add(ui.prop3)

	ui.PropertyList = plist

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
		ui.prop1.Editable = !ui.prop1.Editable
	}

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx C) D {
			gtx.Constraints.Min = gtx.Constraints.Max
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
		}),
		layout.Expanded(func(gtx C) D {
			return ui.layoutModal(gtx)
		}),
	)
}

func (ui *UI) layoutModal(gtx C) D {
	if ui.Modal.Clicked() {
		ui.Modal.ToggleVisibility(gtx.Now)
	}

	return component.Modal(ui.Theme, &ui.Modal).Layout(gtx)
}

func rgb(c uint32) color.NRGBA {
	return argb(0xff000000 | c)
}

func argb(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}
