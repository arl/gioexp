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

var (
	// TODO(arl) document and export?
	listHeight   = unit.Dp(22)
	listWidth    = unit.Dp(200)
	listBarWidth = unit.Dp(4)
)

// A List holds and presents a vertical, scrollable list of properties.
type List struct {
	props []*Property

	// TODO(arl) unexport
	list  layout.List
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
	return &List{
		modal: modal,
		list: layout.List{
			Axis: layout.Vertical,
		},
	}
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
	width := gtx.Metric.Dp(listWidth)
	size := image.Point{X: width, Y: height}

	proportion := (plist.Ratio + 1) / 2
	bar := gtx.Dp(listBarWidth)
	lsize := int(proportion*float32(size.X)) - bar
	roff := lsize + bar

	gtx.Constraints = layout.Exact(size)

	border := widget.Border{
		Color:        theme.Fg,
		CornerRadius: unit.Dp(2),
		Width:        unit.Dp(1),
	}
	dim := border.Layout(gtx, func(gtx C) D {
		return plist.list.Layout(gtx, len(plist.props), func(gtx C, i int) D {
			gtx.Constraints.Min.Y = int(listHeight)
			gtx.Constraints.Max.Y = int(listHeight)
			return plist.layoutProperty(i, theme, plist.modal, gtx)
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
				minposx := int(listBarWidth)
				maxposx := gtx.Constraints.Max.X - int(listBarWidth)
				posx := float32(clamp(minposx, int(e.Position.X), maxposx))
				deltaX := posx - plist.dragX
				deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)

				plist.dragX = posx
				plist.Ratio += deltaRatio

			case pointer.Release, pointer.Cancel:
				plist.drag = false
			}
		}

		// Register for receving input in the bar rect.
		barRect := image.Rect(lsize, 0, roff, gtx.Constraints.Max.X)
		defer clip.Rect(barRect).Push(gtx.Ops).Pop()
		pointer.InputOp{
			Tag:   plist,
			Types: pointer.Press | pointer.Drag | pointer.Release,
			Grab:  plist.drag,
		}.Add(gtx.Ops)
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

// layoutProperty lays out the property at index i from the list.
func (plist *List) layoutProperty(idx int, theme *material.Theme, modal *component.ModalState, gtx C) D {
	size := gtx.Constraints.Max
	gtx.Constraints = layout.Exact(size)

	var dim layout.Dimensions
	{
		proportion := (plist.Ratio + 1) / 2
		barw := gtx.Dp(listBarWidth)
		lsize := int(proportion*float32(gtx.Constraints.Max.X) - float32(barw))

		roff := lsize + barw
		rsize := gtx.Constraints.Max.X - roff

		{
			// Draw property name.
			gtx := gtx
			size := image.Pt(lsize, gtx.Constraints.Max.Y)
			gtx.Constraints = layout.Exact(size)
			plist.props[idx].LayoutName(theme, gtx)
		}
		{
			// Draw property split bar.
			gtx := gtx
			max := image.Pt(roff, gtx.Constraints.Max.Y)
			rect := clip.Rect{Min: image.Pt(lsize, 0), Max: max}.Op()
			paint.FillShape(gtx.Ops, theme.ContrastBg, rect)
		}
		{
			// Draw property value.
			gtx := gtx
			off := op.Offset(image.Pt(roff, 0)).Push(gtx.Ops)
			size := image.Pt(rsize, gtx.Constraints.Max.Y)
			gtx.Constraints = layout.Exact(size)
			plist.props[idx].LayoutValue(theme, gtx)
			off.Pop()
		}

		dim = layout.Dimensions{Size: gtx.Constraints.Max}
	}

	// Draw bottom border
	rect := clip.Rect{Min: image.Pt(0, size.Y-1), Max: size}.Op()
	paint.FillShape(gtx.Ops, theme.Fg, rect)

	return dim
}
