package main

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var (
	propertyHeight    = unit.Dp(30)
	propertyListWidth = unit.Dp(200)
)
var (
	red       = color.NRGBA{R: 255, A: 255}
	blue      = color.NRGBA{B: 255, A: 255}
	green     = color.NRGBA{G: 255, A: 255}
	lightGrey = color.NRGBA{R: 211, G: 211, B: 211, A: 255}
	darkGrey  = color.NRGBA{R: 169, G: 169, B: 169, A: 255}
)

type PropertyList struct {
	Properties []*StringProperty

	List  layout.List
	Width unit.Dp
}

func NewPropertyList() *PropertyList {
	plist := &PropertyList{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
	return plist
}

func (plist *PropertyList) Add(prop *StringProperty) {
	plist.Properties = append(plist.Properties, prop)
}

// TODO(arl) add theme as first param
func (plist *PropertyList) Layout(gtx C) D {
	width := gtx.Metric.Dp(propertyListWidth)
	size := image.Point{
		X: width,
		Y: gtx.Constraints.Max.Y,
	}
	gtx.Constraints = layout.Exact(size)

	drawProps := func(gtx C) D {
		return plist.List.Layout(gtx, len(plist.Properties), func(gtx C, i int) D {
			gtx.Constraints.Min.Y = int(propertyHeight)
			gtx.Constraints.Max.Y = int(propertyHeight)
			return plist.layoutProperty(plist.Properties[i], gtx)
		})
	}

	return drawProps(gtx)
}

func (plist *PropertyList) layoutProperty(prop *StringProperty, gtx C) D {
	size := gtx.Constraints.Max
	gtx.Constraints = layout.Exact(size)

	inset := layout.UniformInset(unit.Dp(1))
	dimensions := inset.Layout(gtx, func(gtx C) D {
		return widget.Border{
			Color:        lightGrey,
			CornerRadius: unit.Dp(2),
			Width:        unit.Dp(1),
		}.Layout(gtx, prop.Layout)
	})

	return dimensions
}

type StringProperty struct {
	Label string
	Value string

	Theme   *material.Theme // TODO(arl) theme should be passed to layout?
	BgColor color.NRGBA     // TODO(arl) this is just temporary while we test the split

	split Split
}

func (prop *StringProperty) Layout(gtx C) D {
	return prop.split.Layout(gtx, prop.layoutLabel, prop.layoutValue)
}

// TODO(arl) add them as first param?
func (prop *StringProperty) layoutLabel(gtx C) D {
	// TODO(arl) continuer ici, ajouter un split, basé sur une valeur passée en paramtere (car controlléé depuis la porpertyList)

	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, prop.BgColor, rect)
	body := material.Body1(prop.Theme, prop.Label)
	body.MaxLines = 1
	return layout.W.Layout(gtx, body.Layout)
}

// TODO(arl) add them as first param?
func (prop *StringProperty) layoutValue(gtx C) D {
	rect := clip.Rect{Max: gtx.Constraints.Max}.Op()
	paint.FillShape(gtx.Ops, prop.BgColor, rect)

	body := material.Body1(prop.Theme, prop.Label)
	body.MaxLines = 1
	return layout.E.Layout(gtx, body.Layout)
}
