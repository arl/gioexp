package property

import (
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

func NewDropDown(items []string) *DropDown {
	return &DropDown{items: items}
}

type DropDown struct {
	Selected int

	items      []string
	area       component.ContextArea
	menu       component.MenuState
	clickables []*widget.Clickable
	focused    bool
}

func (a *DropDown) Layout(th *material.Theme, pgtx, gtx C) D {
	a.menu.Options = a.menu.Options[:0]
	for len(a.clickables) <= len(a.items) {
		a.clickables = append(a.clickables, &widget.Clickable{})
	}
	for i := range a.items {
		click := a.clickables[i]
		if click.Clicked() {
			a.Selected = i
		}
		a.menu.Options = append(a.menu.Options, component.MenuItem(th, click, a.items[i]).Layout)
	}
	a.area.Activation = pointer.ButtonPrimary
	a.area.AbsolutePosition = true

	// Register an key op to get focus events.
	key.InputOp{Tag: a, Hint: key.HintAny}.Add(gtx.Ops)
	for _, e := range gtx.Events(a) {
		switch e := e.(type) {
		case key.FocusEvent:
			a.focused = e.Focus
		}
	}

	wgtx := gtx
	return layout.Stack{}.Layout(pgtx,
		layout.Stacked(func(gtx C) D {
			gtx.Constraints = layout.Exact(wgtx.Constraints.Max)
			inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
			label := material.Label(th, th.TextSize, a.items[a.Selected])
			label.MaxLines = 1
			label.TextSize = th.TextSize
			label.Alignment = text.Start
			label.Color = th.Fg

			return FocusBorder(th, a.focused).Layout(gtx, func(gtx C) D {
				return inset.Layout(gtx, label.Layout)
			})
		}),
		layout.Expanded(func(gtx C) D {
			gtx.Constraints = layout.Exact(gtx.Constraints.Max)
			return a.area.Layout(gtx, component.Menu(th, &a.menu).Layout)
		}),
	)
}
