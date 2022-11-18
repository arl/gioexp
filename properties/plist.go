package main

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
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
	Properties []*Property

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

	// The modal plane on which properties temorarily needing more space can lay
	// out themselves.
	modal *component.ModalState
}

func NewPropertyList(modal *component.ModalState) *PropertyList {
	plist := &PropertyList{
		List: layout.List{
			Axis: layout.Vertical,
		},
		modal: modal,
	}
	return plist
}

func (plist *PropertyList) Add(prop *Property) {
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
			return plist.layoutProperty(plist.Properties[i], theme, plist.modal, gtx)
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

func (plist *PropertyList) layoutProperty(prop *Property, theme *material.Theme, modal *component.ModalState, gtx C) D {
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
			paint.FillShape(gtx.Ops, darkGrey, rect)
		}
		{
			// Draw value.
			off := op.Offset(image.Pt(rightoffset, 0)).Push(gtx.Ops)
			gtx := gtx
			gtx.Constraints = layout.Exact(image.Pt(rightsize, gtx.Constraints.Max.Y))
			prop.LayoutValue(theme, gtx)
			off.Pop()
		}

		// Test, draw modal
		modal.Show(gtx.Now, func(gtx C) D {
			// Draw a red rectangle
			size := gtx.Constraints.Max
			rect := clip.Rect{Min: image.Pt(100, 100), Max: size}.Op()
			paint.FillShape(gtx.Ops, red, rect)

			fmt.Println("in modal, max", gtx.Constraints.Max)
			return D{Size: size}
		})

		dim = layout.Dimensions{Size: gtx.Constraints.Max}
	}

	// Draw bottom border
	rect := clip.Rect{Min: image.Pt(0, size.Y-1), Max: size}.Op()
	paint.FillShape(gtx.Ops, darkGrey, rect)

	return dim
}

type Value interface {
	String() string
	Set(string) error
}

type Property struct {
	Label string

	editable bool
	w        ValueWidget

	Background color.NRGBA
}

func NewFloat64Property(initial float64) *Property {
	return NewTypedProperty("0123456789.eE+-", (*Float64Value)(&initial))
}

func NewUIntProperty(initial uint) *Property {
	return NewTypedProperty("0123456789", (*UIntValue)(&initial))
}

// TODO(arl) comment
//
// NewTypedProperty...
//
// filter is the list of characters allowed in the Editor. If Filter is empty,
// all characters are allowed.
func NewTypedProperty(filter string, initial Value) *Property {
	p := &Property{
		w: newTypedValueWidget(initial, blue),
	}
	return p
}

func (prop *Property) LayoutValue(theme *material.Theme, gtx C) D {
	return prop.w.LayoutValue(theme, prop.editable, gtx)
}

func (prop *Property) SetEditable(editable bool) {
	prop.editable = editable
}

func (prop *Property) Editable() bool {
	return prop.editable
}

func (prop *Property) LayoutLabel(theme *material.Theme, gtx C) D {
	// Background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, theme.Bg, rect)

	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	return inset.Layout(gtx, func(gtx C) D {
		label := material.Label(theme, unit.Sp(14), prop.Label)
		label.MaxLines = 1
		label.TextSize = unit.Sp(14)
		label.Font.Weight = 50
		label.Alignment = text.Start
		return label.Layout(gtx)
	})
}

type ValueWidget interface {
	LayoutValue(theme *material.Theme, editable bool, gtx C) D
	Value() Value
	SetValue(Value) error
}

type TypedValueWidget struct {
	hasFocus bool
	editor   widget.Editor

	bgcolor color.NRGBA
	val     Value
}

func newTypedValueWidget(initial Value, bgcolor color.NRGBA) *TypedValueWidget {
	tv := &TypedValueWidget{
		bgcolor: bgcolor,
		val:     initial,
	}
	tv.SetValue(initial)
	return tv
}

// TODO(arl) add unit tests, check that SetValue sets the value to display.
func (tv *TypedValueWidget) SetValue(val Value) error {
	tv.val = val
	tv.editor.SetText(tv.val.String())
	// Converting a non-nil Value to string can't fail.
	return nil
}

// TODO(arl) add unit tests, check that Value returns the currently displayed value.
func (tv *TypedValueWidget) Value() Value {
	return tv.val
}

func (tv *TypedValueWidget) LayoutValue(theme *material.Theme, editable bool, gtx C) D {
	// Draw background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, theme.Bg, rect)

	hadFocus := tv.hasFocus
	tv.hasFocus = tv.editor.Focused()
	if hadFocus && !tv.hasFocus {
		// We've just lost focus, it's the moment to check the
		// validity of the typed string.
		if err := tv.val.Set(tv.editor.Text()); err != nil {
			// TODO(arl) should we give the user a visual feedback in case of
			// validation error? maybe animate a red flash. or set a red
			// background that would quickly fade into the normal background
			// color

			// Revert the property text to the previous valid value.
			tv.SetValue(tv.val)
		}
	}

	// Draw value as an editor or a label depending on whether the property is editable or not.
	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	if editable {
		return FocusBorder(theme, tv.hasFocus).Layout(gtx, func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				label := material.Editor(theme, &tv.editor, "")
				label.TextSize = unit.Sp(14)
				label.Font.Weight = 50
				return label.Layout(gtx)
			})
		})
	}

	return FocusBorder(theme, tv.hasFocus).Layout(gtx, func(gtx C) D {
		return inset.Layout(gtx, func(gtx C) D {
			label := material.Label(theme, unit.Sp(14), tv.val.String())
			label.MaxLines = 1
			label.TextSize = unit.Sp(14)
			label.Font.Weight = 50
			label.Alignment = text.Start
			label.Color = darkGrey
			return label.Layout(gtx)
		})
	})
}

type UIntValue uint

func (i *UIntValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*i = UIntValue(v)
	return nil
}

func (i *UIntValue) Get() any { return uint(*i) }

func (i *UIntValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

type Float64Value float64

func (i *Float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*i = Float64Value(v)
	return nil
}

func (i *Float64Value) Get() any { return uint(*i) }

func (i *Float64Value) String() string { return strconv.FormatFloat(float64(*i), 'g', 3, 64) }
