package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// This comes from gioui.org tutorial, nothing less nothing more.

type (
	C = layout.Context
	D = layout.Dimensions
)
type Split struct {
	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32
	// Bar is the width for resizing the layout
	Bar unit.Dp

	drag   bool
	dragID pointer.ID
	dragX  float32
}

const defaultBarWidth = unit.Dp(10)

func (s *Split) Layout(gtx layout.Context, left, right layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	proportion := (s.Ratio + 1) / 2
	leftsize := int(proportion*float32(gtx.Constraints.Max.X) - float32(bar))

	rightoffset := leftsize + bar
	rightsize := gtx.Constraints.Max.X - rightoffset

	{ // handle input
		for _, ev := range gtx.Events(s) {
			e, ok := ev.(pointer.Event)
			if !ok {
				continue
			}

			switch e.Type {
			case pointer.Press:
				if s.drag {
					break
				}

				s.dragID = e.PointerID
				s.dragX = e.Position.X

			case pointer.Drag:
				if s.dragID != e.PointerID {
					break
				}

				deltaX := e.Position.X - s.dragX
				s.dragX = e.Position.X

				deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)
				s.Ratio += deltaRatio

			case pointer.Release:
				fallthrough
			case pointer.Cancel:
				s.drag = false
			}
		}

		// register for input
		barRect := image.Rect(leftsize, 0, rightoffset, gtx.Constraints.Max.X)
		area := clip.Rect(barRect).Push(gtx.Ops)
		pointer.InputOp{Tag: s,
			Types: pointer.Press | pointer.Drag | pointer.Release,
			Grab:  s.drag,
		}.Add(gtx.Ops)
		area.Pop()
	}

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(leftsize, gtx.Constraints.Max.Y))
		left(gtx)
	}

	{
		off := op.Offset(image.Pt(rightoffset, 0)).Push(gtx.Ops)
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(rightsize, gtx.Constraints.Max.Y))
		right(gtx)
		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

var red = color.NRGBA{R: 255, A: 255}
var blue = color.NRGBA{B: 255, A: 255}

func FillWithLabel(gtx layout.Context, th *material.Theme, text string, backgroundColor color.NRGBA) D {
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	fmt.Println(text, gtx.Constraints.Min)
	paint.FillShape(gtx.Ops, backgroundColor, rect)
	return layout.Center.Layout(gtx, material.H3(th, text).Layout)
}

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

var split Split

func exampleSplit(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return split.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return FillWithLabel(gtx, th, "Left", red)
	}, func(gtx layout.Context) layout.Dimensions {
		return FillWithLabel(gtx, th, "Right", blue)
	})
}

func (u *ui) Run(w *app.Window) error {
	var ops op.Ops

	for e := range w.Events() {
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)

			exampleSplit(gtx, u.Theme)
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
