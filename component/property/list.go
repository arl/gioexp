package property

import (
	"image"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/constraints"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var ( // TODO(arl) document and export?
	propertyHeight       = unit.Dp(22)
	propertyListWidth    = unit.Dp(200)
	propertyListBarWidth = unit.Dp(3)
)

// A List holds and presents a vertical, scrollable list of properties.
type List struct {
	props []*Property

	// TODO(arl) unexport
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

	// modal plane the properties need more space can use to lay out themselves.
	modal *component.ModalState
}

// NewList creates a new List.
func NewList(modal *component.ModalState) *List {
	plist := &List{
		List: layout.List{
			Axis: layout.Vertical,
		},
		modal: modal,
	}
	return plist
}

// Add adds a new property to the list.
func (plist *List) Add(prop *Property) {
	plist.props = append(plist.props, prop)
}

func (plist *List) Layout(theme *material.Theme, gtx C) D {
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
		Color:        theme.Fg,
		CornerRadius: unit.Dp(2),
		Width:        unit.Dp(1),
	}.Layout(gtx, func(gtx C) D {
		return plist.List.Layout(gtx, len(plist.props), func(gtx C, i int) D {
			gtx.Constraints.Min.Y = int(propertyHeight)
			gtx.Constraints.Max.Y = int(propertyHeight)
			return plist.layoutProperty(plist.props[i], theme, plist.modal, gtx)
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

func (plist *List) layoutProperty(prop *Property, theme *material.Theme, modal *component.ModalState, gtx C) D {
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
			prop.LayoutLabel(theme, gtx)
		}
		{
			// Draw split bar.
			gtx := gtx
			rect := clip.Rect{Min: image.Pt(leftsize, 0), Max: image.Pt(rightoffset, gtx.Constraints.Max.Y)}.Op()
			paint.FillShape(gtx.Ops, theme.Fg, rect)
		}
		{
			// Draw value.
			off := op.Offset(image.Pt(rightoffset, 0)).Push(gtx.Ops)
			gtx := gtx
			gtx.Constraints = layout.Exact(image.Pt(rightsize, gtx.Constraints.Max.Y))
			prop.LayoutValue(theme, gtx)
			off.Pop()
		}

		dim = layout.Dimensions{Size: gtx.Constraints.Max}
	}

	// Draw bottom border
	rect := clip.Rect{Min: image.Pt(0, size.Y-1), Max: size}.Op()
	paint.FillShape(gtx.Ops, darkGrey, rect)

	return dim
}
