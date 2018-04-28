// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"reflect"

	"github.com/rcoreilly/goki/gi"
	"github.com/rcoreilly/goki/gi/oswin"
	"github.com/rcoreilly/goki/gi/oswin/driver"
	"github.com/rcoreilly/goki/gi/units"
	"github.com/rcoreilly/goki/ki"
	"github.com/rcoreilly/goki/ki/kit"
)

func main() {
	driver.Main(func(app oswin.App) {
		mainrun()
	})
}

func mainrun() {
	width := 1024
	height := 768

	// turn this on to see a trace of the rendering
	// gi.Update2DTrace = true
	// gi.Render2DTrace = true
	// gi.Layout2DTrace = true

	rec := ki.Node{}          // receiver for events
	rec.InitName(&rec, "rec") // this is essential for root objects not owned by other Ki tree nodes

	win := gi.NewWindow2D("GoGi Widgets Window", width, height, true) // pixel sizes

	icnm := "widget-wedge-down"
	wdicon := gi.IconByName(icnm)

	vp := win.WinViewport2D()
	updt := vp.UpdateStart()
	vp.Fill = true

	// style sheet!
	var css = ki.Props{
		"button": ki.Props{
			"background-color": "#FEE",
		},
		"#combo": ki.Props{
			"background-color": "#EFE",
		},
		".hslides": ki.Props{
			"background-color": "#EDF",
		},
	}
	vp.CSS = css

	vlay := vp.AddNewChild(gi.KiT_Frame, "vlay").(*gi.Frame)
	vlay.Lay = gi.LayoutCol

	trow := vlay.AddNewChild(gi.KiT_Layout, "trow").(*gi.Layout)
	trow.Lay = gi.LayoutRow
	trow.SetStretchMaxWidth()

	trow.AddNewChild(gi.KiT_Stretch, "str1")
	title := trow.AddNewChild(gi.KiT_Label, "title").(*gi.Label)
	title.Text = "This is a demonstration of the various GoGi Widgets\nShortcuts: Control+Alt+P = Preferences, Control+Alt+E = Editor, Command +/- = zoom"
	title.SetProp("text-align", gi.AlignTop)
	title.SetProp("align-vert", gi.AlignTop)
	title.SetProp("word-wrap", true)
	// title.SetMinPrefWidth(units.NewValue(30, units.Ch))
	trow.AddNewChild(gi.KiT_Stretch, "str2")

	//////////////////////////////////////////
	//      Buttons

	vlay.AddNewChild(gi.KiT_Space, "blspc")
	blrow := vlay.AddNewChild(gi.KiT_Layout, "blrow").(*gi.Layout)
	blab := blrow.AddNewChild(gi.KiT_Label, "blab").(*gi.Label)
	blab.Text = "Buttons:"

	brow := vlay.AddNewChild(gi.KiT_Layout, "brow").(*gi.Layout)
	brow.Lay = gi.LayoutRow
	brow.SetProp("align-horiz", gi.AlignLeft)
	// brow.SetProp("align-horiz", gi.AlignJustify)
	brow.SetStretchMaxWidth()

	button1 := brow.AddNewChild(gi.KiT_Button, "button1").(*gi.Button)
	button1.SetIcon(wdicon)
	button1.ButtonSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received button signal: %v from button: %v\n", gi.ButtonSignals(sig), send.Name())
		if sig == int64(gi.ButtonClicked) { // note: 3 diff ButtonSig sig's possible -- important to check
			gi.PromptDialog(vp, "Button1 Dialog", "This is a dialog!  Various specific types of dialogs are available.", true, true, nil, nil)
		}
	})

	button2 := brow.AddNewChild(gi.KiT_Button, "button2").(*gi.Button)
	button2.SetText("Open GoGiEditor")
	// button2.SetProp("background-color", "#EDF")
	button2.ButtonSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received button signal: %v from button: %v\n", gi.ButtonSignals(sig), send.Name())
		if sig == int64(gi.ButtonClicked) {
			gi.GoGiEditorOf(vp)
		}
	})

	checkbox := brow.AddNewChild(gi.KiT_CheckBox, "checkbox").(*gi.CheckBox)
	checkbox.Text = "Toggle"

	mb1 := brow.AddNewChild(gi.KiT_MenuButton, "menubutton1").(*gi.MenuButton)
	mb1.SetText("Menu Button")
	mb1.AddMenuText("Menu Item 1", rec.This, 1, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received menu action data: %v from menu action: %v\n", data, send.Name())
	})

	mb1.AddMenuText("Menu Item 2", rec.This, 2, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received menu action data: %v from menu action: %v\n", data, send.Name())
	})

	mb1.AddMenuText("Menu Item 3", rec.This, 3, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received menu action data: %v from menu action: %v\n", data, send.Name())
	})

	brow.SetPropChildren("vertical-align", gi.AlignMiddle) // align all..
	brow.SetPropChildren("margin", units.NewValue(2, units.Ex))

	//////////////////////////////////////////
	//      Sliders

	vlay.AddNewChild(gi.KiT_Space, "slspc")
	slrow := vlay.AddNewChild(gi.KiT_Layout, "slrow").(*gi.Layout)
	slab := slrow.AddNewChild(gi.KiT_Label, "slab").(*gi.Label)
	slab.Text = "Sliders:"

	srow := vlay.AddNewChild(gi.KiT_Layout, "srow").(*gi.Layout)
	srow.Lay = gi.LayoutRow
	srow.SetProp("align-horiz", "left")
	srow.SetStretchMaxWidth()

	slider1 := srow.AddNewChild(gi.KiT_Slider, "slider1").(*gi.Slider)
	slider1.Dim = gi.X
	slider1.Class = "hslides"
	slider1.Defaults()
	slider1.SetMinPrefWidth(units.NewValue(20, units.Em))
	slider1.SetMinPrefHeight(units.NewValue(2, units.Em))
	slider1.SetValue(0.5)
	slider1.Snap = true
	slider1.Tracking = true
	slider1.Icon = gi.IconByName("widget-circlebutton-on")

	slider2 := srow.AddNewChild(gi.KiT_Slider, "slider2").(*gi.Slider)
	slider2.Dim = gi.Y
	slider2.Defaults()
	slider2.SetMinPrefHeight(units.NewValue(10, units.Em))
	slider2.SetMinPrefWidth(units.NewValue(1, units.Em))
	slider2.SetStretchMaxHeight()
	slider2.SetValue(0.5)

	slider1.SliderSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received slider signal: %v from slider: %v with data: %v\n", gi.SliderSignals(sig), send.Name(), data)
	})

	slider2.SliderSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received slider signal: %v from slider: %v with data: %v\n", gi.SliderSignals(sig), send.Name(), data)
	})

	scrollbar1 := srow.AddNewChild(gi.KiT_ScrollBar, "scrollbar1").(*gi.ScrollBar)
	scrollbar1.Dim = gi.X
	scrollbar1.Class = "hslides"
	scrollbar1.Defaults()
	scrollbar1.SetMinPrefWidth(units.NewValue(20, units.Em))
	scrollbar1.SetMinPrefHeight(units.NewValue(1, units.Em))
	scrollbar1.SetThumbValue(0.25)
	scrollbar1.SetValue(0.25)
	// scrollbar1.Snap = true
	scrollbar1.Tracking = true
	scrollbar1.SliderSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received scrollbar signal: %v from scrollbar: %v with data: %v\n", gi.SliderSignals(sig), send.Name(), data)
	})

	scrollbar2 := srow.AddNewChild(gi.KiT_ScrollBar, "scrollbar2").(*gi.ScrollBar)
	scrollbar2.Dim = gi.Y
	scrollbar2.Defaults()
	scrollbar2.SetMinPrefHeight(units.NewValue(10, units.Em))
	scrollbar2.SetMinPrefWidth(units.NewValue(1, units.Em))
	scrollbar2.SetStretchMaxHeight()
	scrollbar2.SetThumbValue(0.1)
	scrollbar2.SetValue(0.5)
	scrollbar2.SliderSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received scrollbar signal: %v from scrollbar: %v with data: %v\n", gi.SliderSignals(sig), send.Name(), data)
	})

	//////////////////////////////////////////
	//      Text Widgets

	vlay.AddNewChild(gi.KiT_Space, "tlspc")
	txlrow := vlay.AddNewChild(gi.KiT_Layout, "txlrow").(*gi.Layout)
	txlab := txlrow.AddNewChild(gi.KiT_Label, "txlab").(*gi.Label)
	txlab.Text = "Text Widgets:"
	txrow := vlay.AddNewChild(gi.KiT_Layout, "txrow").(*gi.Layout)
	txrow.Lay = gi.LayoutRow
	// txrow.SetProp("align-horiz", gi.AlignJustify)
	txrow.SetStretchMaxWidth()
	// txrow.SetStretchMaxHeight()

	edit1 := txrow.AddNewChild(gi.KiT_TextField, "edit1").(*gi.TextField)
	edit1.Text = "Edit this text"
	edit1.SetProp("min-width", "20em")
	edit1.TextFieldSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("Received line edit signal: %v from edit: %v with data: %v\n", gi.TextFieldSignals(sig), send.Name(), data)
	})
	// edit1.SetProp("read-only", true)

	sb := txrow.AddNewChild(gi.KiT_SpinBox, "spin").(*gi.SpinBox)
	sb.HasMin = true
	sb.Min = 0.0
	sb.SpinBoxSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("SpinBox %v value changed: %v\n", send.Name(), data)
	})

	cb := txrow.AddNewChild(gi.KiT_ComboBox, "combo").(*gi.ComboBox)
	cb.ItemsFromTypes(kit.Types.AllImplementersOf(reflect.TypeOf((*gi.Node2D)(nil)).Elem(), false), true, true, 50)
	cb.ComboSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		fmt.Printf("ComboBox %v selected index: %v data: %v\n", send.Name(), sig, data)
	})

	txrow.SetPropChildren("align-vert", gi.AlignMiddle)
	txrow.SetPropChildren("margin", units.NewValue(2, units.Ex))

	vp.UpdateEndNoSig(updt)

	win.StartEventLoop()

	fmt.Printf("ending\n")
}
