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
	"golang.org/x/exp/constraints"
)

var (
	red       = color.NRGBA{R: 255, A: 255}
	blue      = color.NRGBA{B: 255, A: 255}
	green     = color.NRGBA{G: 255, A: 255}
	lightGrey = color.NRGBA{R: 211, G: 211, B: 211, A: 255}
	darkGrey  = color.NRGBA{R: 169, G: 169, B: 169, A: 255}
	aliceBlue = color.NRGBA{R: 240, G: 248, B: 255, A: 255}
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

	// MaxHeight limits the property list height. If not set, the property list
	// takes all the vertical space it is given.
	MaxHeight unit.Dp

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

func (plist *PropertyList) Layout(theme *material.Theme, gtx C) D {
	var height int
	if plist.MaxHeight != 0 {
		height = int(plist.MaxHeight)
	} else {
		height = gtx.Constraints.Max.Y
	}
	width := gtx.Metric.Dp(propertyListWidth)
	size := image.Point{
		X: width,
		Y: height,
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
			return plist.layoutProperty(plist.Properties[i], theme, gtx)
		})
	})

	{
		// Handle input.
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

				// Clamp drag position so that the 'handle' remains always visible.
				minposx := int(propertyListBarWidth)
				maxposx := gtx.Constraints.Max.X - int(propertyListBarWidth)
				posx := float32(clamp(minposx, int(e.Position.X), maxposx))

				deltaX := posx - plist.dragX
				plist.dragX = posx

				deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)
				plist.Ratio += deltaRatio

			case pointer.Release, pointer.Cancel:
				plist.drag = false
			}
		}

		// Register for input.
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

func clamp[T constraints.Ordered](mn, val, mx T) T {
	if val < mn {
		return mn
	}
	if val > mx {
		return mx
	}
	return val
}

func (plist *PropertyList) layoutProperty(prop *StringProperty, theme *material.Theme, gtx C) D {
	size := gtx.Constraints.Max
	gtx.Constraints = layout.Exact(size)

	var dim layout.Dimensions
	{
		proportion := (plist.Ratio + 1) / 2
		barWidth := gtx.Dp(propertyListBarWidth)
		leftsize := int(proportion*float32(gtx.Constraints.Max.X) - float32(barWidth))

		rightoffset := leftsize + barWidth
		rightsize := gtx.Constraints.Max.X - rightoffset

		{
			// Draw label.
			gtx := gtx
			gtx.Constraints = layout.Exact(image.Pt(leftsize, gtx.Constraints.Max.Y))
			prop.layoutLabel(theme, gtx)
		}
		{
			// Draw split bar.
			gtx := gtx
			rect := clip.Rect{Min: image.Pt(leftsize, 0), Max: image.Pt(rightoffset, gtx.Constraints.Max.Y)}.Op()
			paint.FillShape(gtx.Ops, darkGrey, rect)
		}
		{
			// Draw value.
			off := op.Offset(image.Pt(rightoffset, 0)).Push(gtx.Ops)
			gtx := gtx
			gtx.Constraints = layout.Exact(image.Pt(rightsize, gtx.Constraints.Max.Y))
			prop.layoutValue(theme, gtx)
			off.Pop()
		}

		dim = layout.Dimensions{Size: gtx.Constraints.Max}
	}

	// Draw bottom border
	rect := clip.Rect{Min: image.Pt(0, size.Y-1), Max: size}.Op()
	paint.FillShape(gtx.Ops, darkGrey, rect)

	return dim
}

type StringProperty struct {
	Label string
	Value string

	editable bool
	editor   widget.Editor

	Background color.NRGBA
}

func (prop *StringProperty) SetEditable(editable bool) {
	prop.editable = editable
	if !editable {
		prop.Value = prop.editor.Text()
	} else {
		prop.editor.SingleLine = true
	}
}

func (prop *StringProperty) Editable() bool {
	return prop.editable
}

func (prop *StringProperty) layoutLabel(theme *material.Theme, gtx C) D {
	// Background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, prop.Background, rect)

	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(theme, unit.Sp(14), prop.Label)
		label.MaxLines = 1
		label.TextSize = unit.Sp(14)
		label.Font.Weight = 50
		label.Alignment = text.Start
		return label.Layout(gtx)
	})
}

func (prop *StringProperty) layoutValue(theme *material.Theme, gtx C) D {
	// Draw background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, prop.Background, rect)

	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	if prop.editable {
		return FocusBorder(theme, prop.editor.Focused()).Layout(gtx, func(gtx C) D {
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := material.Editor(theme, &prop.editor, "")
				label.TextSize = unit.Sp(14)
				label.Font.Weight = 50
				return label.Layout(gtx)
			})
		})
	}

	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := material.Label(theme, unit.Sp(14), prop.Value)
		label.MaxLines = 1
		label.TextSize = unit.Sp(14)
		label.Font.Weight = 50
		label.Alignment = text.Start
		return label.Layout(gtx)
	})
}
