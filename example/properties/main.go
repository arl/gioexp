package main

import (
	"image/color"
	"log"
	"math"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"

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
	th    *material.Theme
	plist *property.List

	prop5 *property.Uint
	dd    *property.DropDown

	btn widget.Clickable
}

var (
	aliceBlue = color.NRGBA{R: 240, G: 248, B: 255, A: 255}
)

func NewUI(theme *material.Theme) *UI {
	ui := &UI{
		th: theme,
		dd: property.NewDropDown([]string{"ciao", "bonjour", "hello", "hallo", "buongiorno", "buenos dias", "ola", "bom dia"}),
	}

	plist := property.NewList()

	plist.Add("int", property.NewInt(-10))
	plist.Add("uint", property.NewUInt(123))
	plist.Add("string", property.NewString("string property"))
	plist.Add("float64", property.NewFloat64(math.Pi))
	ui.prop5 = property.NewUInt(27)
	plist.Add("uint editable", ui.prop5)
	plist.Add("dropdown", ui.dd)
	plist.Add("float64(2)", property.NewFloat64(23564.32e12))

	ui.plist = plist
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
		ui.prop5.Editable = !ui.prop5.Editable
		ui.dd.Selected = 2
		ui.prop5.SetValue(234)
	}

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
					gtx.Constraints.Max.X = 400
					return ui.plist.Layout(ui.th, gtx)
				}),
				layout.Rigid(func(gtx C) D {
					return material.Button(ui.th, &ui.btn, "toggle editable").Layout(gtx)
				}),
			)
		}),
	)
}

var (
	red       = rgb(0xff0000)
	green     = rgb(0x00ff00)
	blue      = rgb(0x0000ff)
	lightGrey = rgb(0xd3d3d3)
	darkGrey  = rgb(0xa9a9a9)
)

func rgb(c uint32) color.NRGBA {
	return argb(0xff000000 | c)
}

func argb(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}
