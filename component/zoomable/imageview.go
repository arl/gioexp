package zoomable

import (
	"image"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// View is a zoomable and draggable view of a widget.
type View struct {
	scroll gesture.Scroll
	drag   gesture.Drag
	mouse  f32.Point

	dragOrg f32.Point
	dragCur f32.Point
	offset  f32.Point

	tr f32.Affine2D
}

// Layout makes w zoomable and draggable.
func (v *View) Layout(gtx C, w layout.Widget) D {
	r := clip.Rect{Max: gtx.Constraints.Max}
	r.Push(gtx.Ops)

	v.scroll.Add(gtx.Ops, image.Rect(0, -1, 0, 1))
	nscroll := v.scroll.Scroll(gtx.Metric, gtx, gtx.Now, gesture.Vertical)
	pointer.InputOp{Tag: v, Types: pointer.Move}.Add(gtx.Ops)

	for _, ev := range gtx.Events(v) {
		switch ev := ev.(type) {
		case pointer.Event:
			switch ev.Type {
			case pointer.Move:
				v.mouse = ev.Position
			}
		}
	}

	v.drag.Add(gtx.Ops)
	for _, ev := range v.drag.Events(gtx.Metric, gtx, gesture.Both) {
		switch ev.Type {
		case pointer.Press:
			v.dragOrg = ev.Position
		case pointer.Drag:
			v.dragCur = v.dragOrg.Sub(ev.Position)
		case pointer.Release:
			v.mouse = ev.Position
			v.offset = v.offset.Sub(v.dragOrg).Add(v.mouse)
			v.dragCur = f32.Point{}
		}
	}

	if nscroll != 0 {
		var change float32
		if nscroll > 0 {
			change = 0.9
		} else {
			change = 1.1
		}
		mouse := v.mouse.Sub(v.offset)
		v.tr = v.tr.Scale(mouse, f32.Pt(change, change))
	}

	op.Affine(v.tr).Add(gtx.Ops)

	// Adapt the offset to the scaling factor.
	sx, _, _, _, _, _ := v.tr.Elems()
	off := v.offset.Sub(v.dragCur).Div(sx)
	op.Offset(off.Round()).Add(gtx.Ops)

	return w(gtx)
}
