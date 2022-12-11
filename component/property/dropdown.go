package property

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
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

	wsize := gtx.Constraints.Max

	return layout.Stack{}.Layout(pgtx,
		layout.Stacked(func(gtx C) D {
			return component.Surface(th).Layout(gtx, func(gtx C) D {
				return FocusBorder(th, false).Layout(gtx, func(gtx C) D {

					inset := layout.Inset{Top: 1, Right: 4, Bottom: 1, Left: 4}
					label := material.Label(th, th.TextSize, a.items[a.Selected])
					dim := inset.Layout(gtx, label.Layout)
					dim.Size.X = wsize.X - 6
					return dim
				})
			})
		}),
		layout.Expanded(func(gtx C) D {
			return a.area.Layout(gtx, component.Menu(th, &a.menu).Layout)
		}),
	)
}
