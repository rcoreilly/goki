// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"image"
	"reflect"

	"github.com/rcoreilly/goki/gi/oswin"
	"github.com/rcoreilly/goki/gi/oswin/key"
	"github.com/rcoreilly/goki/gi/units"
	"github.com/rcoreilly/goki/ki"
	"github.com/rcoreilly/goki/ki/bitflag"
	"github.com/rcoreilly/goki/ki/kit"
)

// state of the dialog
type DialogState int64

const (
	// existential state -- struct exists and is likely being constructed
	DialogExists DialogState = iota
	// dialog is open in a modal state, blocking all other input
	DialogOpenModal
	// dialog is open in a modeless state, allowing other input
	DialogOpenModeless
	// Ok was pressed -- dialog accepted
	DialogAccepted
	// Cancel was pressed -- button canceled
	DialogCanceled
	DialogStateN
)

//go:generate stringer -type=DialogState

// standard vertical space between elements in a dialog, in Em units
var StdDialogVSpace = float32(2.0)
var StdDialogVSpaceUnits = units.Value{StdDialogVSpace, units.Em, 0}

// Dialog supports dialog functionality -- based on a viewport that can either be rendered in a separate window or on top of an existing one
type Dialog struct {
	Viewport2D
	Title     string      `desc:"title text displayed at the top row of the dialog"`
	Prompt    string      `desc:"a prompt string displayed below the title"`
	Modal     bool        `desc:"open the dialog in a modal state, blocking all other input"`
	State     DialogState `desc:"state of the dialog"`
	DialogSig ki.Signal   `json:"-" xml:"-" desc:"signal for dialog -- sends a signal when opened, accepted, or canceled"`
}

var KiT_Dialog = kit.Types.AddType(&Dialog{}, DialogProps)

// Open this dialog, in given location (0 = middle of window), finding window from given viewport -- returns false if it fails for any reason
func (dlg *Dialog) Open(x, y int, avp *Viewport2D) bool {
	win := avp.Win
	if win == nil {
		return false
	}
	if x == 0 && y == 0 {
		x = win.Viewport.ViewBox.Size.X / 3
		y = win.Viewport.ViewBox.Size.Y / 3
	}

	bitflag.Set(&dlg.Flag, int(VpFlagPopup))
	// todo: deal with modeless -- need a separate window presumably -- not hard
	dlg.State = DialogOpenModal

	updt := dlg.UpdateStart()
	dlg.Win = win
	dlg.Init2DTree()
	dlg.Style2DTree()                                      // sufficient to get sizes
	dlg.LayData.AllocSize = win.Viewport.LayData.AllocSize // give it the whole vp initially
	dlg.Size2DTree()                                       // collect sizes
	dlg.Win = nil

	frame := dlg.ChildByName("frame", 0).(*Frame)
	vpsz := frame.LayData.Size.Pref.Min(win.Viewport.LayData.AllocSize).ToPoint()

	stw := int(dlg.Style.Layout.MinWidth.Dots)
	sth := int(dlg.Style.Layout.MinHeight.Dots)
	vpsz.X = kit.MaxInt(vpsz.X, stw)
	vpsz.Y = kit.MaxInt(vpsz.Y, sth)
	x = kit.MinInt(x, win.Viewport.ViewBox.Size.X-vpsz.X) // fit
	y = kit.MinInt(y, win.Viewport.ViewBox.Size.Y-vpsz.Y) // fit

	dlg.Resize(vpsz.X, vpsz.Y)
	dlg.ViewBox.Min = image.Point{x, y}
	dlg.UpdateEndNoSig(updt)

	// put window at the very end
	win.ReceiveEventType(dlg.This, oswin.KeyChordEvent, func(recv, send ki.Ki, sig int64, d interface{}) {
		kt := d.(*key.ChordEvent)
		ddlg, _ := recv.EmbeddedStruct(KiT_Dialog).(*Dialog)
		kf := KeyFun(kt.ChordString())
		switch kf {
		case KeyFunAbort:
			ddlg.Cancel()
		case KeyFunAccept:
			ddlg.Accept()
		}
	})

	win.PushPopup(dlg.This)
	return true
}

// Close requests that the dialog be closed -- it does not alter any state or send any signals
func (dlg *Dialog) Close() {
	win := dlg.Win
	if win != nil {
		win.ClosePopup(dlg.This)
	}
}

// Accept accepts the dialog, activated by the default Ok button
func (dlg *Dialog) Accept() {
	if dlg == nil {
		return
	}
	dlg.State = DialogAccepted
	dlg.DialogSig.Emit(dlg.This, int64(dlg.State), nil)
	dlg.Close()
}

// Cancel cancels the dialog, activated by the default Cancel button
func (dlg *Dialog) Cancel() {
	if dlg == nil {
		return
	}
	dlg.State = DialogCanceled
	dlg.DialogSig.Emit(dlg.This, int64(dlg.State), nil)
	dlg.Close()
}

////////////////////////////////////////////////////////////////////////////////////////
//  Configuration functions construct standard types of dialogs but anything can be done

var DialogProps = ki.Props{
	"#frame": ki.Props{
		"border-width":        units.NewValue(2, units.Px),
		"margin":              units.NewValue(8, units.Px),
		"padding":             units.NewValue(4, units.Px),
		"box-shadow.h-offset": units.NewValue(4, units.Px),
		"box-shadow.v-offset": units.NewValue(4, units.Px),
		"box-shadow.blur":     units.NewValue(4, units.Px),
		"box-shadow.color":    "#CCC",
	},
	"#title": ki.Props{
		// todo: add "bigger" font
		"max-width":        units.NewValue(-1, units.Px),
		"text-align":       AlignCenter,
		"vertical-align":   AlignTop,
		"background-color": "none",
	},
	"#prompt": ki.Props{
		"max-width":        units.NewValue(-1, units.Px),
		"text-align":       AlignLeft,
		"vertical-align":   AlignTop,
		"background-color": "none",
	},
}

// SetFrame creates a standard vertical column frame layout as first element of the dialog, named "frame"
func (dlg *Dialog) SetFrame() *Frame {
	frame := dlg.AddNewChild(KiT_Frame, "frame").(*Frame)
	frame.Lay = LayoutCol
	dlg.PartStyleProps(frame, DialogProps)
	return frame
}

// Frame returns the main frame for the dialog, assumed to be the first element in the dialog
func (dlg *Dialog) Frame() *Frame {
	return dlg.Child(0).(*Frame)
}

// SetTitle sets the title and adds a Label named "title" to the given frame layout if passed
func (dlg *Dialog) SetTitle(title string, frame *Frame) *Label {
	dlg.Title = title
	if frame != nil {
		lab := frame.AddNewChild(KiT_Label, "title").(*Label)
		lab.Text = title
		dlg.PartStyleProps(lab, DialogProps)
		return lab
	}
	return nil
}

// Title returns the title label widget, and its index, within frame -- nil, -1 if not found
func (dlg *Dialog) TitleWidget(frame *Frame) (*Label, int) {
	idx := frame.ChildIndexByName("title", 0)
	if idx < 0 {
		return nil, -1
	}
	return frame.Child(idx).(*Label), idx
}

// SetPrompt sets the prompt and adds a Label named "prompt" to the given frame layout if passed, with the given amount of space before it, sized in "Em"'s (units of font size), if > 0
func (dlg *Dialog) SetPrompt(prompt string, spaceBefore float32, frame *Frame) *Label {
	dlg.Prompt = prompt
	if frame != nil {
		if spaceBefore > 0 {
			spc := frame.AddNewChild(KiT_Space, "prompt-space").(*Space)
			spc.SetFixedHeight(units.NewValue(spaceBefore, units.Em))
		}
		lab := frame.AddNewChild(KiT_Label, "prompt").(*Label)
		lab.Text = prompt
		dlg.PartStyleProps(lab, DialogProps)
		return lab
	}
	return nil
}

// Prompt returns the prompt label widget, and its index, within frame -- if nil returns the title widget (flexible if prompt is nil)
func (dlg *Dialog) PromptWidget(frame *Frame) (*Label, int) {
	idx := frame.ChildIndexByName("prompt", 0)
	if idx < 0 {
		return dlg.TitleWidget(frame)
	}
	return frame.Child(idx).(*Label), idx
}

// AddButtonBox adds a button box (Row Layout) named "buttons" to given frame,
// with optional fixed space and stretch elements before it
func (dlg *Dialog) AddButtonBox(spaceBefore float32, stretchBefore bool, frame *Frame) *Layout {
	if frame == nil {
		return nil
	}

	if spaceBefore > 0 {
		spc := frame.AddNewChild(KiT_Space, "button-space").(*Space)
		spc.SetFixedHeight(units.NewValue(spaceBefore, units.Em))
	}
	if stretchBefore {
		frame.AddNewChild(KiT_Stretch, "button-stretch")
	}
	bb := frame.AddNewChild(KiT_Layout, "buttons").(*Layout)
	bb.Lay = LayoutRow
	bb.SetProp("max-width", -1)
	return bb
}

// ButtonBox returns the ButtonBox layout widget, and its index, within frame -- nil, -1 if not found
func (dlg *Dialog) ButtonBox(frame *Frame) (*Layout, int) {
	idx := frame.ChildIndexByName("buttons", 0)
	if idx < 0 {
		return nil, -1
	}
	return frame.Child(idx).(*Layout), idx
}

// StdButtonConfig returns a kit.TypeAndNameList for calling on ConfigChildren of a button box, to create standard Ok, Cancel buttons (if true), optionally starting with a Stretch element that will cause the buttons to be arranged on the right -- a space element is added between buttons if more than one
func (dlg *Dialog) StdButtonConfig(stretch, ok, cancel bool) kit.TypeAndNameList {
	config := kit.TypeAndNameList{} // note: slice is already a pointer
	if stretch {
		config.Add(KiT_Stretch, "stretch")
	}
	if ok {
		config.Add(KiT_Button, "ok")
	}
	if cancel {
		if ok {
			config.Add(KiT_Space, "space")
		}
		config.Add(KiT_Button, "cancel")
	}
	return config
}

// StdButtonConnnect connects standard buttons in given button box layout to Accept / Cancel actions
func (dlg *Dialog) StdButtonConnect(ok, cancel bool, bb *Layout) {
	if ok {
		okb := bb.ChildByName("ok", 0).EmbeddedStruct(KiT_Button).(*Button)
		okb.SetText("Ok")
		okb.ButtonSig.Connect(dlg.This, func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(ButtonClicked) {
				dlg := recv.EmbeddedStruct(KiT_Dialog).(*Dialog)
				dlg.Accept()
			}
		})
	}
	if cancel {
		canb := bb.ChildByName("cancel", 0).EmbeddedStruct(KiT_Button).(*Button)
		canb.SetText("Cancel")
		canb.ButtonSig.Connect(dlg.This, func(recv, send ki.Ki, sig int64, data interface{}) {
			if sig == int64(ButtonClicked) {
				dlg := recv.EmbeddedStruct(KiT_Dialog).(*Dialog)
				dlg.Cancel()
			}
		})
	}
}

// StdDialog configures a basic standard dialog with a title, prompt, and ok /
// cancel buttons -- any empty text will not be added
func (dlg *Dialog) StdDialog(title, prompt string, ok, cancel bool) {
	frame := dlg.SetFrame()
	pspc := float32(0.0)
	if title != "" {
		dlg.SetTitle(title, frame)
		pspc = StdDialogVSpace
	}
	if prompt != "" {
		dlg.SetPrompt(prompt, pspc, frame)
	}
	bb := dlg.AddButtonBox(StdDialogVSpace, true, frame)
	bbc := dlg.StdButtonConfig(true, ok, cancel)
	mods, updt := bb.ConfigChildren(bbc, false) // not unique names
	dlg.StdButtonConnect(ok, cancel, bb)
	bitflag.Set(&dlg.Flag, int(VpFlagPopupDestroyAll)) // std is disposable
	if mods {
		bb.UpdateEnd(updt)
	}
}

// NewStdDialog returns a basic standard dialog with a name, title, prompt,
// and ok / cancel buttons -- any empty text will not be added -- returns with
// UpdateStart started but NOT ended -- must call UpdateEnd(true) once done
// configuring!
func NewStdDialog(name, title, prompt string, ok, cancel bool) *Dialog {
	dlg := Dialog{}
	dlg.InitName(&dlg, name)
	bitflag.Set(&dlg.Flag, int(VpFlagPopup))
	dlg.UpdateStart() // guaranteed to be true
	dlg.StdDialog(title, prompt, ok, cancel)
	return &dlg
}

// Prompt opens a basic standard dialog with a title, prompt, and ok / cancel buttons -- any empty text will not be added -- optionally connects to given signal receiving object and function for dialog signals (nil to ignore)
func PromptDialog(avp *Viewport2D, title, prompt string, ok, cancel bool, recv ki.Ki, fun ki.RecvFunc) {
	dlg := NewStdDialog("prompt", title, prompt, ok, cancel)
	if recv != nil && fun != nil {
		dlg.DialogSig.Connect(recv, fun)
	}
	dlg.UpdateEndNoSig(true) // going to be shown
	dlg.Open(0, 0, avp)
}

////////////////////////////////////////////////////////////////////////////////////////
// Node2D interface

func (dlg *Dialog) Init2D() {
	dlg.Viewport2D.Init2D()
	bitflag.Set(&dlg.Flag, int(VpFlagPopup))
}

func (dlg *Dialog) HasFocus2D() bool {
	return true // dialog ALWAYS gets all the events!
}

// check for interface implementation
var _ Node2D = &Dialog{}

////////////////////////////////////////////////////////////////////////////////////////
// more specialized types of dialogs

// New Ki item(s) of type dialog, showing types that implement given interface -- use construct of form: reflect.TypeOf((*gi.Node2D)(nil)).Elem() to get the interface type -- optionally connects to given signal receiving object and function for dialog signals (nil to ignore)
func NewKiDialog(avp *Viewport2D, iface reflect.Type, title, prompt string, recv ki.Ki, fun ki.RecvFunc) *Dialog {
	dlg := NewStdDialog("new-ki", title, prompt, true, true)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	nspc := frame.InsertNewChild(KiT_Space, prIdx+1, "n-space").(*Space)
	nspc.SetFixedHeight(StdDialogVSpaceUnits)

	nrow := frame.InsertNewChild(KiT_Layout, prIdx+2, "n-row").(*Layout)
	nrow.Lay = LayoutRow

	nlbl := nrow.AddNewChild(KiT_Label, "n-label").(*Label)
	nlbl.Text = "Number:  "

	nsb := nrow.AddNewChild(KiT_SpinBox, "n-field").(*SpinBox)
	nsb.Defaults()
	nsb.SetMin(1)
	nsb.Value = 1
	nsb.Step = 1

	tspc := frame.InsertNewChild(KiT_Space, prIdx+3, "type-space").(*Space)
	tspc.SetFixedHeight(units.NewValue(0.5, units.Em))

	trow := frame.InsertNewChild(KiT_Layout, prIdx+4, "t-row").(*Layout)
	trow.Lay = LayoutRow

	tlbl := trow.AddNewChild(KiT_Label, "t-label").(*Label)
	tlbl.Text = "Type:    "

	typs := trow.AddNewChild(KiT_ComboBox, "types").(*ComboBox)
	typs.ItemsFromTypes(kit.Types.AllImplementersOf(iface, false), true, true, 50)
	// typs.ComboSig.Connect(rec.This, func(recv, send ki.Ki, sig int64, data interface{}) {
	// 	fmt.Printf("ComboBox %v selected index: %v data: %v\n", send.Name(), sig, data)
	// })

	if recv != nil && fun != nil {
		dlg.DialogSig.Connect(recv, fun)
	}
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp)
	return dlg
}

// get the user-set values from a NewKiDialog
func NewKiDialogValues(dlg *Dialog) (int, reflect.Type) {
	frame := dlg.Frame()
	nrow := frame.ChildByName("n-row", 0).(*Layout)
	ntf := nrow.ChildByName("n-field", 0).(*SpinBox)
	n := int(ntf.Value)
	trow := frame.ChildByName("t-row", 0).(*Layout)
	typs := trow.ChildByName("types", 0).(*ComboBox)
	typ := typs.CurVal.(reflect.Type)
	return n, typ
}

// Struct View dialog for editing fields of a structure using a StructView -- optionally connects to given signal receiving object and function for dialog signals (nil to ignore)
func StructViewDialog(avp *Viewport2D, stru interface{}, tmpSave ValueView, title, prompt string, recv ki.Ki, fun ki.RecvFunc) *Dialog {
	dlg := NewStdDialog("struct-view", title, prompt, true, false)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	nspc := frame.InsertNewChild(KiT_Space, prIdx+1, "view-space").(*Space)
	nspc.SetFixedHeight(StdDialogVSpaceUnits)

	sv := frame.InsertNewChild(KiT_StructView, prIdx+2, "struct-view").(*StructView)
	sv.SetStruct(stru, tmpSave)

	if recv != nil && fun != nil {
		dlg.DialogSig.Connect(recv, fun)
	}
	dlg.SetProp("min-width", units.NewValue(60, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp)
	return dlg
}

// Map View dialog for editing fields of a map using a MapView -- optionally connects to given signal receiving object and function for dialog signals (nil to ignore)
func MapViewDialog(avp *Viewport2D, mp interface{}, tmpSave ValueView, title, prompt string, recv ki.Ki, fun ki.RecvFunc) *Dialog {
	dlg := NewStdDialog("map-view", title, prompt, true, false)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	nspc := frame.InsertNewChild(KiT_Space, prIdx+1, "view-space").(*Space)
	nspc.SetFixedHeight(StdDialogVSpaceUnits)

	sv := frame.InsertNewChild(KiT_MapView, prIdx+2, "map-view").(*MapView)
	sv.SetMap(mp, tmpSave)

	if recv != nil && fun != nil {
		dlg.DialogSig.Connect(recv, fun)
	}
	dlg.SetProp("min-width", units.NewValue(60, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp)
	return dlg
}

// Slice View dialog for editing fields of a slice using a SliceView -- optionally connects to given signal receiving object and function for dialog signals (nil to ignore)
func SliceViewDialog(avp *Viewport2D, mp interface{}, tmpSave ValueView, title, prompt string, recv ki.Ki, fun ki.RecvFunc) *Dialog {
	dlg := NewStdDialog("slice-view", title, prompt, true, false)

	frame := dlg.Frame()
	_, prIdx := dlg.PromptWidget(frame)

	nspc := frame.InsertNewChild(KiT_Space, prIdx+1, "view-space").(*Space)
	nspc.SetFixedHeight(StdDialogVSpaceUnits)

	sv := frame.InsertNewChild(KiT_SliceView, prIdx+2, "slice-view").(*SliceView)
	sv.SetSlice(mp, tmpSave)

	if recv != nil && fun != nil {
		dlg.DialogSig.Connect(recv, fun)
	}
	dlg.SetProp("min-width", units.NewValue(60, units.Em))
	dlg.SetProp("min-height", units.NewValue(30, units.Em))
	dlg.UpdateEndNoSig(true)
	dlg.Open(0, 0, avp)
	return dlg
}
