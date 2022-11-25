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
	ValueWidget

	Name     string
	Editable bool
}

func (prop *Property) LayoutValue(theme *material.Theme, gtx C) D {
	return prop.Layout(theme, prop.Editable, gtx)
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

// ValueWidget is the interface to hold, display and edit Property values.
//
// Well-behaved implementations guarantee Value always returns a value of the
// expected type. Also, implementations may panic if Value is called with an
// unexpected type.
type ValueWidget interface {
	// Layout lays out the property value using the given theme and handles user
	// interaction if the value is currently editable.
	Layout(theme *material.Theme, editable bool, gtx C) D

	// Value returns the property value.
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
