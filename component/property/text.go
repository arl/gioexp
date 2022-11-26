package property

import (
	"strconv"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// TextValue is the interface implemented by objects that can converted
// themselves to and from string.
type TextValue interface {
	String() string
	Set(string) error
}

// TextWidget is a widget that holds, displays and edits a property shown
// converted to its textual representation. It's edited using a standard gio
// editor or laid out as a label when not editable.
type TextWidget struct {
	hasFocus bool
	editor   widget.Editor
	val      TextValue
}

// TODO(arl) add unit tests, check that SetValue sets the value to display.
func (sv *TextWidget) SetValue(val any) error {
	sv.val = val.(TextValue)
	sv.editor.SetText(sv.val.String())
	return nil // Converting a non-nil Value to string can't fail.
}

// TODO(arl) add unit tests, check that Value returns the currently displayed value.
func (sv *TextWidget) Value() any {
	return sv.val
}

// TODO(arl) show ellipsis if the text can't be shown entirely

func (sv *TextWidget) Layout(theme *material.Theme, editable bool, gtx C) D {
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
		ed := material.Editor(theme, &sv.editor, "")
		ed.TextSize = unit.Sp(14)
		ed.Font.Weight = 50

		return FocusBorder(theme, sv.hasFocus).Layout(gtx, func(gtx C) D {
			return inset.Layout(gtx, ed.Layout)
		})
	}

	label := material.Label(theme, unit.Sp(14), sv.val.String())
	label.MaxLines = 1
	label.TextSize = unit.Sp(14)
	label.Font.Weight = 50
	label.Alignment = text.Start
	label.Color = theme.Fg

	return FocusBorder(theme, sv.hasFocus).Layout(gtx, func(gtx C) D {
		return inset.Layout(gtx, label.Layout)
	})
}

// NewText creates a Property with v as initial value and the type of value as
// underlying type. filter is the list of characters allowed in the Editor. If
// empty all characters are allowed.
func NewText(v TextValue, filter string) *Property {
	w := &TextWidget{val: v}
	w.editor.Filter = filter
	w.SetValue(v)
	return &Property{ValueWidget: w}
}

func NewString(v string) *Property {
	return NewText((*StringValue)(&v), "")
}

type StringValue string

func (s *StringValue) Set(str string) error {
	*s = (StringValue)(str)
	return nil
}
func (s *StringValue) Get() any       { return *s }
func (s *StringValue) String() string { return string(*s) }

func NewUInt(v uint) *Property {
	return NewText((*UIntValue)(&v), "0123456789")
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

func (i *UIntValue) Get() any       { return uint(*i) }
func (i *UIntValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

func NewFloat64(v float64) *Property {
	return NewText((*Float64Value)(&v), "0123456789.eE+-")
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

func (i *Float64Value) Get() any       { return uint(*i) }
func (i *Float64Value) String() string { return strconv.FormatFloat(float64(*i), 'g', 3, 64) }
