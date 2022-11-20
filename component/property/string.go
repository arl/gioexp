package property

import (
	"flag"
	"image/color"
	"strconv"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// NewString returns a Property that displays a string representation and is
// edited via a string editor. Filter is the list of allowed runes in the
// editor; if it is empty all runes are allowed..
func NewString(filter string, initial flag.Value) *Property {
	p := &Property{
		W: NewStringValue(initial, blue),
	}
	return p
}

type StringValue struct {
	hasFocus bool
	editor   widget.Editor

	bgcolor color.NRGBA
	val     flag.Value
}

func NewStringValue(initial flag.Value, bgcolor color.NRGBA) *StringValue {
	sv := &StringValue{
		bgcolor: bgcolor,
		val:     initial,
	}
	sv.SetValue(initial)
	return sv
}

// TODO(arl) add unit tests, check that SetValue sets the value to display.
func (sv *StringValue) SetValue(val any) error {
	sv.val = val.(flag.Value)
	sv.editor.SetText(sv.val.String())
	// Converting a non-nil Value to string can't fail.
	return nil
}

// TODO(arl) add unit tests, check that Value returns the currently displayed value.
func (sv *StringValue) Value() any {
	return sv.val
}

func (sv *StringValue) Layout(theme *material.Theme, editable bool, gtx C) D {
	// Draw background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, theme.Bg, rect)

	hadFocus := sv.hasFocus
	sv.hasFocus = sv.editor.Focused()
	if hadFocus && !sv.hasFocus {
		// We've just lost focus, it's the moment to check the
		// validity of the typed string.
		if err := sv.val.Set(sv.editor.Text()); err != nil {
			// TODO(arl) should we give the user a visual feedback in case of
			// validation error? maybe animate a red flash. or set a red
			// background that would quickly fade into the normal background
			// color

			// Revert the property text to the previous valid value.
			sv.SetValue(sv.val)
		}
	}

	// Draw value as an editor or a label depending on whether the property is editable or not.
	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	if editable {
		return FocusBorder(theme, sv.hasFocus).Layout(gtx, func(gtx C) D {
			return inset.Layout(gtx, func(gtx C) D {
				label := material.Editor(theme, &sv.editor, "")
				label.TextSize = unit.Sp(14)
				label.Font.Weight = 50
				return label.Layout(gtx)
			})
		})
	}

	return FocusBorder(theme, sv.hasFocus).Layout(gtx, func(gtx C) D {
		return inset.Layout(gtx, func(gtx C) D {
			label := material.Label(theme, unit.Sp(14), sv.val.String())
			label.MaxLines = 1
			label.TextSize = unit.Sp(14)
			label.Font.Weight = 50
			label.Alignment = text.Start
			label.Color = darkGrey
			return label.Layout(gtx)
		})
	})
}

func NewUInt(initial uint) *Property {
	return NewString("0123456789", (*UIntValue)(&initial))
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

func NewFloat64(initial float64) *Property {
	return NewString("0123456789.eE+-", (*Float64Value)(&initial))
}

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
