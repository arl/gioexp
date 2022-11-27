package property

import (
	"image"

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

type (
	C = layout.Context
	D = layout.Dimensions
)

const (
	DefaultPropertyHeight  = unit.Dp(30)
	DefaultHandleBarWidth  = unit.Dp(3)
	DefaultHandleBarHeight = unit.Dp(35)
)

// A List holds and presents a vertical, scrollable list of properties. A List
// is divided into into 2 columns: property names on the left and widgets for
// property values on the right. These 2 sections can be resized thanks to a
// divider, which can be dragged.
type List struct {
	widgets []Widget
	names   []string

	// PropertyHeight is the height of a single property. All properties have
	// the same dimensions. The width depends of the horizontal space available
	// for the list
	PropertyHeight unit.Dp

	// HandleBarWidth is the width of the handlebar used to resize the columns.
	HandleBarWidth unit.Dp

	// HandleBarHeight is the width of the handlebar.
	HandleBarHeight unit.Dp

	list layout.List

	// ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	ratio float32

	drag   bool
	dragID pointer.ID
	dragX  float32
}

// NewList creates a new List.
func NewList() *List {
	return &List{
		PropertyHeight:  DefaultPropertyHeight,
		HandleBarWidth:  DefaultHandleBarWidth,
		HandleBarHeight: DefaultHandleBarHeight,
		list: layout.List{
			Axis: layout.Vertical,
		},
	}
}

// Add adds a new property to the list.
func (plist *List) Add(name string, widget Widget) {
	plist.widgets = append(plist.widgets, widget)
	plist.names = append(plist.names, name)
}

func (plist *List) visibleHeight(gtx C) int {
	return min(gtx.Dp(plist.PropertyHeight)*len(plist.widgets), gtx.Constraints.Max.Y)
}

func (plist *List) Layout(th *material.Theme, gtx C) D {
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Max.Y}
	gtx.Constraints = layout.Exact(size)

	proportion := (plist.ratio + 1) / 2
	whandle := gtx.Dp(plist.HandleBarWidth)
	lsize := int(proportion*float32(size.X)) - whandle
	roff := lsize + whandle

	barHeight := gtx.Dp(plist.HandleBarHeight)
	vh := plist.visibleHeight(gtx)
	barRect := image.Rect(lsize, (vh-barHeight)/2, roff, (vh+barHeight)/2)

	dim := widget.Border{
		Color:        th.Fg,
		CornerRadius: unit.Dp(2),
		Width:        unit.Dp(1),
	}.Layout(gtx, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx C) D {
				gtx.Constraints.Min = gtx.Constraints.Max
				pgtx := gtx
				return plist.list.Layout(gtx, len(plist.widgets), func(gtx C, i int) D {
					gtx.Constraints.Min.Y = gtx.Dp(plist.PropertyHeight)
					gtx.Constraints.Max.Y = gtx.Dp(plist.PropertyHeight)
					return plist.layoutProperty(i, th, pgtx, gtx)
				})
			}),
			layout.Stacked(func(gtx C) D {
				// Draw divider line
				xdiv := lsize + whandle/2
				paint.FillShape(gtx.Ops, th.ContrastBg, clip.Rect{
					Min: image.Pt(xdiv, 0),
					Max: image.Pt(xdiv+1, vh),
				}.Op())
				// Draw handlebar
				paint.FillShape(gtx.Ops, th.ContrastBg, clip.Rect(barRect).Op())
				return D{}
			}),
		)
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
				minposx := whandle
				maxposx := gtx.Constraints.Max.X - minposx
				posx := float32(clamp(minposx, int(e.Position.X), maxposx))
				deltaX := posx - plist.dragX
				deltaRatio := deltaX * 2 / float32(gtx.Constraints.Max.X)

				plist.dragX = posx
				plist.ratio += deltaRatio

			case pointer.Release, pointer.Cancel:
				plist.drag = false
			}
		}

		// Register for receving input in the handlebar rect.
		defer clip.Rect(barRect).Push(gtx.Ops).Pop()
		pointer.CursorColResize.Add(gtx.Ops)
		pointer.InputOp{
			Tag:   plist,
			Types: pointer.Press | pointer.Drag | pointer.Release,
			Grab:  plist.drag,
		}.Add(gtx.Ops)
	}

	return dim
}

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
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
func (plist *List) layoutProperty(idx int, th *material.Theme, pgtx, gtx C) D {
	size := gtx.Constraints.Max
	gtx.Constraints = layout.Exact(size)

	proportion := (plist.ratio + 1) / 2
	whandle := gtx.Dp(plist.HandleBarWidth)
	lsize := int(proportion*float32(gtx.Constraints.Max.X) - float32(whandle))

	roff := lsize + whandle
	rsize := gtx.Constraints.Max.X - roff

	{
		// Draw property name.
		gtx := gtx
		size := image.Pt(lsize, gtx.Constraints.Max.Y)
		gtx.Constraints = layout.Exact(size)
		plist.LayoutName(idx, th, gtx)
	}
	{
		// Draw property value.
		gtx := gtx
		off := op.Offset(image.Pt(roff, 0)).Push(gtx.Ops)
		size := image.Pt(rsize, gtx.Constraints.Max.Y)
		gtx.Constraints = layout.Exact(size)
		plist.widgets[idx].Layout(th, pgtx, gtx)
		off.Pop()
	}

	// Draw bottom border.
	paint.FillShape(gtx.Ops, th.Fg, clip.Rect{
		Min: image.Pt(0, size.Y-1),
		Max: size,
	}.Op())

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (plist *List) LayoutName(idx int, th *material.Theme, gtx C) D {
	paint.FillShape(gtx.Ops, th.Bg, clip.Rect{Max: gtx.Constraints.Max}.Op())

	label := material.Label(th, th.TextSize, plist.names[idx])
	label.MaxLines = 1
	label.TextSize = th.TextSize
	label.Font.Weight = 50
	label.Alignment = text.Start

	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	return inset.Layout(gtx, label.Layout)
}

// Widget shows the value of a property and handles user actions to edit it.
type Widget interface {
	// Layout lays out the property widget using gtx which is the
	// property-specific context, and pgtx which is the parent context (useful
	// for properties that require more space during edition).
	Layout(th *material.Theme, pgtx, gtx layout.Context) D
}
