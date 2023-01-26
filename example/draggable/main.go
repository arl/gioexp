package main

import (
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"

	"github.com/arl/gioexp/component/zoomable"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// Zoomable: experiment at making a widget which content can be zoomed/panned
// using the mouse wheel, while keeping the point under the cursor immobile
// (i.e. using the mouse cursor as scale origin).

func main() {
	ui := &UI{
		Theme: material.NewTheme(gofont.Collection()),
	}
	go func() {
		w := app.NewWindow(app.Title("Zoomable"))
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
	View  zoomable.View
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

var red = color.NRGBA{R: 200, A: 255}

func (ui *UI) Layout(gtx C) D {
	wrect := func(gtx C) D {
		return ui.View.Layout(gtx, func(gtx C) D {
			rect := clip.Rect{
				Min: image.Pt(100, 100),
				Max: image.Pt(300, 300),
			}
			paint.FillShape(gtx.Ops, red, rect.Op())
			d := image.Point{Y: 400}
			return D{Size: d}
		})
	}

	label := func(gtx C) D {
		gtx.Constraints = layout.Exact(gtx.Constraints.Max)
		return layout.Center.Layout(gtx,
			material.Label(ui.Theme, 40, "Drag and Scroll").Layout)
	}

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(wrect),
		layout.Stacked(label),
	)
}
