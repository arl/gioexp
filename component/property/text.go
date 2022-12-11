package property

import (
	"image/color"
	"strconv"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var lightGrey = rgb(0xd3d3d3)

func rgb(c uint32) color.NRGBA {
	return argb(0xff000000 | c)
}

func argb(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}

// Stringer is the interface implemented by objects that can converted
// themselves to and from string.
type Stringer interface {
	String() string
	Set(string) error
}

// Text is a widget that holds, displays and edits a property shown converted to
// its textual representation. It's edited using a standard gio editor or laid
// out as a label when not editable.
type Text struct {
	val      Stringer
	editor   widget.Editor
	Editable bool
	hasFocus bool
}

// NewText creates a Text property and assigns it a value. filter is the list of
// characters allowed in the Editor. If empty all characters are allowed.
func NewText(val Stringer, filter string) *Text {
	t := &Text{
		Editable: true,
		editor: widget.Editor{
			Filter:     filter,
			SingleLine: true,
		},
	}
	t.setValue(val)
	return t
}

func (t *Text) setValue(val Stringer) {
	t.val = val
	t.editor.SetText(t.val.String())
}

func (t *Text) value() Stringer {
	return t.val
}

func (t *Text) Layout(th *material.Theme, _, gtx C) D {
	// Draw background color.
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	bgcol := th.Bg
	if !t.Editable {
		bgcol = lightGrey
	}
	paint.FillShape(gtx.Ops, bgcol, rect)

	hadFocus := t.hasFocus
	t.hasFocus = t.editor.Focused()
	if hadFocus && !t.hasFocus {
		// We've just lost focus, it's the moment to check the
		// validity of the typed string.
		if err := t.val.Set(t.editor.Text()); err != nil {
			// TODO(arl) should we give the user a visual feedback in case of
			// validation error? maybe animate a red flash. or set a red
			// background that would quickly fade into the normal background
			// color

			// Revert the property text to the previous valid value.
			t.setValue(t.val)
		}
	}

	// Draw value as an editor or a label depending on whether the property is
	// editable or not.
	inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}

	if !t.Editable {
		label := material.Label(th, th.TextSize, t.val.String())
		label.MaxLines = 1
		label.TextSize = th.TextSize
		label.Alignment = text.Start
		label.Color = th.Fg

		return FocusBorder(th, t.hasFocus).Layout(gtx, func(gtx C) D {
			return inset.Layout(gtx, label.Layout)
		})
	}

	ed := material.Editor(th, &t.editor, "")
	ed.TextSize = th.TextSize

	return FocusBorder(th, t.hasFocus).Layout(gtx, func(gtx C) D {
		return inset.Layout(gtx, ed.Layout)
	})
}

//
// UInt
//

type Uint struct {
	*Text
}

func NewUInt(val uint) *Uint {
	return &Uint{Text: NewText((*uintval)(&val), "0123456789")}
}

func (i *Uint) Value() uint {
	return uint(*(i.value().(*uintval)))
}

func (i *Uint) SetValue(val uint) {
	i.setValue((*uintval)(&val))
}

type uintval uint

func (i *uintval) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*i = uintval(v)
	return nil
}

func (i *uintval) Get() any       { return uint(*i) }
func (i *uintval) String() string { return strconv.FormatUint(uint64(*i), 10) }

//
// Int
//

type Int struct {
	*Text
}

func NewInt(val int) *Int {
	return &Int{Text: NewText((*intval)(&val), "-+0123456789")}
}

func (i *Int) Value() int {
	return int(*(i.value().(*intval)))
}

func (i *Int) SetValue(val int) {
	i.setValue((*intval)(&val))
}

type intval int

func (i *intval) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*i = intval(v)
	return nil
}

func (i *intval) Get() any       { return int(*i) }
func (i *intval) String() string { return strconv.FormatInt(int64(*i), 10) }

//
// Float64
//

type Float64 struct {
	*Text
}

func NewFloat64(val float64) *Float64 {
	return &Float64{Text: NewText((*f64val)(&val), "-+0123456789.eE")}
}

func (f *Float64) Value() float64 {
	return float64(*(f.value().(*f64val)))
}

func (f *Float64) SetValue(val float64) {
	f.setValue((*f64val)(&val))
}

type f64val float64

func (f *f64val) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*f = f64val(v)
	return nil
}

func (f *f64val) Get() any       { return uint(*f) }
func (f *f64val) String() string { return strconv.FormatFloat(float64(*f), 'g', 3, 64) }

//
// String
//

type String struct {
	*Text
}

func NewString(val string) *String {
	return &String{Text: NewText((*stringval)(&val), "")}
}

func NewStringWithFilter(val, filter string) *String {
	return &String{Text: NewText((*stringval)(&val), filter)}
}

func (s *String) Value() string {
	return string(*(s.value().(*stringval)))
}

func (s *String) SetValue(val string) {
	s.setValue((*stringval)(&val))
}

type stringval string

func (s *stringval) Set(str string) error {
	*s = (stringval)(str)
	return nil
}

func (s *stringval) Get() any       { return *s }
func (s *stringval) String() string { return string(*s) }
