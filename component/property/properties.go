package property

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// A Property is a generic widget handling a single item in a List. Its name,
// drawn with as a material.Label, is typically constant. Layout and user
// interaction to edit the property value are defined and delegated to a
// ValueWidget.
type Property struct {
	Name     string
	W        ValueWidget
	Editable bool
}

func (prop *Property) Value() any {
	return prop.W.Value()
}

func (prop *Property) SetValue(val any) {
	prop.W.SetValue(val)
}

func (prop *Property) LayoutValue(theme *material.Theme, gtx C) D {
	return prop.W.Layout(theme, prop.Editable, gtx)
}

func (prop *Property) LayoutName(theme *material.Theme, gtx C) D {
	// Background color.
	paint.FillShape(gtx.Ops, theme.Bg, clip.Rect{Max: gtx.Constraints.Max}.Op())

	label := material.Label(theme, unit.Sp(14), prop.Name)
	label.MaxLines = 1
	label.TextSize = unit.Sp(14)
	label.Font.Weight = 50
	label.Alignment = text.Start

	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
	return inset.Layout(gtx, label.Layout)
}

// A ValueWidget allows to get, set and lay out the value of a Property, as well
// as handle user interaction for value edition..
type ValueWidget interface {

	// Layout lays out the property value with respect to the given theme and
	// the boolean indicating whether the property is editable.
	Layout(theme *material.Theme, editable bool, gtx C) D

	// Value returns property value.
	Value() any

	// SetValue defines the property value. Panics if the type of any is not of
	// the expected type. Returns an error if the value is invalid
	//
	// TODO(arl) do we really need to return an error here?
	//
	// TODO(arl) can we have a type-safe API with generics if we manage to avoid
	// the generic type to 'leak' in to property.List?
	SetValue(any) error
}
