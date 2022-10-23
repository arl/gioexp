package main

import (
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
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
	ui := &ui{
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

type ui struct {
	Theme *material.Theme
}

func (u *ui) Run(w *app.Window) error {
	var ops op.Ops
	z := zoomable{}

	for e := range w.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			z.Layout(gtx, func(gtx C) D {
				rect := clip.Rect{
					Min: image.Pt(100, 100),
					Max: image.Pt(300, 300),
				}.Op()
				rect.Push(gtx.Ops)
				color := color.NRGBA{R: 200, A: 255}
				paint.FillShape(gtx.Ops, color, rect)
				d := image.Point{Y: 400}
				return D{Size: d}
			})
			e.Frame(gtx.Ops)

		case key.Event:
			switch e.Name {
			case key.NameEscape:
				return nil
			}

		case system.DestroyEvent:
			return e.Err
		}
	}

	return nil
}

type zoomable struct {
	scroll gesture.Scroll
	mouse  f32.Point
	tr     f32.Affine2D
}

func (z *zoomable) Layout(gtx C, zoomed layout.Widget) D {
	{
		stack := clip.Rect{
			Max: gtx.Constraints.Max,
		}.Push(gtx.Ops)

		z.scroll.Add(gtx.Ops, image.Rect(0, -1, 0, 1))
		nscroll := z.scroll.Scroll(gtx.Metric, gtx, gtx.Now, gesture.Vertical)
		pointer.InputOp{Tag: z, Types: pointer.Move}.Add(gtx.Ops)

		for _, ev := range gtx.Events(z) {
			switch ev := ev.(type) {
			case pointer.Event:
				switch ev.Type {
				case pointer.Move:
					z.mouse = ev.Position
				}
			}
		}

		if nscroll != 0 {
			var change float32
			if nscroll > 0 {
				change = 1.1
			} else {
				change = 0.9
			}

			z.tr = z.tr.Scale(z.mouse, f32.Pt(change, change))
		}

		op.Affine(z.tr).Add(gtx.Ops)
		stack.Pop()
	}
	return zoomed(gtx)
}
