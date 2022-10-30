package main

import (
	"image"
	"image/color"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	red       = color.NRGBA{R: 255, A: 255}
	blue      = color.NRGBA{B: 255, A: 255}
	green     = color.NRGBA{G: 255, A: 255}
	lightGrey = color.NRGBA{R: 211, G: 211, B: 211, A: 255}
	darkGrey  = color.NRGBA{R: 169, G: 169, B: 169, A: 255}
)

var (
	propertyHeight       = unit.Dp(22)
	propertyListWidth    = unit.Dp(200)
	propertyListBarWidth = unit.Dp(3)
)

type PropertyList struct {
	Properties []*StringProperty

	List  layout.List
	Width unit.Dp

	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32

	// Bar is the width for resizing the layout
	Bar unit.Dp

	drag   bool
	dragID pointer.ID
	dragX  float32
}

func NewPropertyList() *PropertyList {
	plist := &PropertyList{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
	return plist
}

func (plist *PropertyList) Add(prop *StringProperty) {
	plist.Properties = append(plist.Properties, prop)
}

// TODO(arl) add theme as first param
func (plist *PropertyList) Layout(gtx C) D {
	width := gtx.Metric.Dp(propertyListWidth)
	size := image.Point{
		X: width,
		Y: gtx.Constraints.Max.Y,
	}
	gtx.Constraints = layout.Exact(size)

	proportion := (plist.Ratio + 1) / 2

	bar := gtx.Dp(propertyListBarWidth)
	leftsize := int(proportion*float32(gtx.Constraints.Max.X) - float32(bar))
	rightoffset := leftsize + bar

	dim := widget.Border{
		Color:        darkGrey,
		CornerRadius: unit.Dp(2),
		Width:        unit.Dp(1),
	}.Layout(gtx, func(gtx C) D {
		return plist.List.Layout(gtx, len(plist.Properties), func(gtx C, i int) D {
			gtx.Constraints.Min.Y = int(propertyHeight)
			gtx.Constraints.Max.Y = int(propertyHeight)
			return plist.layoutProperty(plist.Properties[i], gtx)
		})
	})

	{
		// handle input
		for _, ev := range gtx.Events(plist) {
			e, ok := ev.(pointer.Event)
			if !ok {
				continue
			}

			switch e.Type {
			case pointer.Press:
				if plist.drag {
					break
				}

				plist.dragID = e.PointerID
				plist.dragX = e.Position.X

			case pointer.Drag:
				if plist.dragID != e.PointerID {
					break
				}

				deltaX := e.Position.X - plist.dragX
				plist.dragX = e.Position.X
				// TODO(arl) clamp drag position
				// plist.dragX = clamp(0, e.Position.X, float32(gtx.Constraints.Max.X))

				deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)
				plist.Ratio += deltaRatio

			case pointer.Release:
				fallthrough
			case pointer.Cancel:
				plist.drag = false
			}
		}

		// register for input
		barRect := image.Rect(leftsize, 0, rightoffset, gtx.Constraints.Max.X)
		area := clip.Rect(barRect).Push(gtx.Ops)
		pointer.InputOp{Tag: plist,
			Types: pointer.Press | pointer.Drag | pointer.Release,
			Grab:  plist.drag,
		}.Add(gtx.Ops)
		area.Pop()
	}

	return dim
}

func (plist *PropertyList) layoutProperty(prop *StringProperty, gtx C) D {
	size := gtx.Constraints.Max
	gtx.Constraints = layout.Exact(size)

	dim := splitLayout(plist.Ratio, gtx.Dp(propertyListBarWidth), gtx, prop.layoutLabel, prop.layoutValue)

	// Draw bottom border
	rect := clip.Rect{Min: image.Pt(0, size.Y-1), Max: size}.Op()
	paint.FillShape(gtx.Ops, darkGrey, rect)

	return dim
}

func splitLayout(ratio float32, barWidth int, gtx layout.Context, left, right layout.Widget) layout.Dimensions {
	proportion := (ratio + 1) / 2
	leftsize := int(proportion*float32(gtx.Constraints.Max.X) - float32(barWidth))

	rightoffset := leftsize + barWidth
	rightsize := gtx.Constraints.Max.X - rightoffset

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(leftsize, gtx.Constraints.Max.Y))
		left(gtx)
	}

	{
		// Draw split bar.
		gtx := gtx
		rect := clip.Rect{Min: image.Pt(leftsize, 0), Max: image.Pt(rightoffset, gtx.Constraints.Max.Y)}.Op()
		paint.FillShape(gtx.Ops, darkGrey, rect)
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

type StringProperty struct {
	Label string
	Value string

	Editable bool
	editor   widget.Editor

	Theme   *material.Theme // TODO(arl) theme should be passed to layout?
	BgColor color.NRGBA
}

func (prop *StringProperty) SetEditable(editable bool) {
	prop.Editable = editable
	if editable {
		prop.editor.SingleLine = true
	}
}

// TODO(arl) add them as first param?
func (prop *StringProperty) layoutLabel(gtx C) D {
	// Background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, prop.BgColor, rect)

	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(prop.Theme, unit.Sp(14), prop.Label)
		label.MaxLines = 1
		label.TextSize = unit.Sp(14)
		label.Font.Weight = 50
		label.Alignment = text.Start
		return label.Layout(gtx)
	})
}

// TODO(arl) add them as first param?
func (prop *StringProperty) layoutValue(gtx C) D {
	// Background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, prop.BgColor, rect)

	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	if prop.Editable {
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Editor(prop.Theme, &prop.editor, "")
			label.TextSize = unit.Sp(14)
			label.Font.Weight = 50
			return label.Layout(gtx)
		})
	} else {
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Label(prop.Theme, unit.Sp(14), prop.Value)
			label.MaxLines = 1
			label.TextSize = unit.Sp(14)
			label.Font.Weight = 50
			label.Alignment = text.Start
			return label.Layout(gtx)
		})
	}
}
