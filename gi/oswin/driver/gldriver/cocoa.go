// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin
// +build 386 amd64
// +build !ios

package gldriver

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework OpenGL
#include <OpenGL/gl3.h>
#import <Carbon/Carbon.h> // for HIToolbox/Events.h
#import <Cocoa/Cocoa.h>
#include <pthread.h>
#include <stdint.h>
#include <stdlib.h>

void startDriver();
void stopDriver();
void makeCurrentContext(uintptr_t ctx);
void flushContext(uintptr_t ctx);
uintptr_t doNewWindow(int width, int height, int left, int top, char* title, bool dialog, bool modal, bool tool, bool fullscreen);
void doShowWindow(uintptr_t id);
void doResizeWindow(uintptr_t id, int width, int height);
void doCloseWindow(uintptr_t id);
void getScreens();
uint64_t threadID();
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
	"unsafe"

	"github.com/rcoreilly/goki/gi/oswin"
	"github.com/rcoreilly/goki/gi/oswin/driver/internal/lifecycler"
	"github.com/rcoreilly/goki/gi/oswin/key"
	"github.com/rcoreilly/goki/gi/oswin/mouse"
	"github.com/rcoreilly/goki/gi/oswin/paint"
	"github.com/rcoreilly/goki/gi/oswin/window"
	"golang.org/x/mobile/gl"
)

const useLifecycler = true

var initThreadID C.uint64_t

func init() {
	// Lock the goroutine responsible for initialization to an OS thread.
	// This means the goroutine running main (and calling startDriver below)
	// is locked to the OS thread that started the program. This is
	// necessary for the correct delivery of Cocoa events to the process.
	//
	// A discussion on this topic:
	// https://groups.google.com/forum/#!msg/golang-nuts/IiWZ2hUuLDA/SNKYYZBelsYJ
	runtime.LockOSThread()
	initThreadID = C.threadID()
}

func newWindow(opts *oswin.NewWindowOptions) (uintptr, error) {
	dialog, modal, tool, fullscreen := oswin.WindowFlagsToBool(opts.Flags)

	title := C.CString(opts.GetTitle())
	defer C.free(unsafe.Pointer(title))

	return uintptr(C.doNewWindow(C.int(opts.Size.X), C.int(opts.Size.Y),
		C.int(opts.Pos.X), C.int(opts.Pos.Y), title,
		C.bool(dialog), C.bool(modal), C.bool(tool), C.bool(fullscreen))), nil
}

func initWindow(w *windowImpl) {
	w.glctx, w.worker = gl.NewContext()
}

func showWindow(w *windowImpl) {
	C.doShowWindow(C.uintptr_t(w.id))
}

func resizeWindow(w *windowImpl, sz image.Point) {
	C.doResizeWindow(C.uintptr_t(w.id), C.int(sz.X), C.int(sz.Y))
}

func getGeometry(w *windowImpl) {

}

//export preparedOpenGL
func preparedOpenGL(id, ctx, vba uintptr) {
	theApp.mu.Lock()
	w := theApp.windows[id]
	theApp.mu.Unlock()

	w.ctx = ctx
	go drawLoop(w, vba)
}

func closeWindow(id uintptr) {
	C.doCloseWindow(C.uintptr_t(id))
}

var mainCallback func(oswin.App)

func main(f func(oswin.App)) error {
	if tid := C.threadID(); tid != initThreadID {
		log.Fatalf("gldriver.Main called on thread %d, but gldriver.init ran on %d", tid, initThreadID)
	}

	mainCallback = f
	C.startDriver()
	return nil
}

//export driverStarted
func driverStarted() {
	oswin.TheApp = theApp
	C.getScreens()
	go func() {
		mainCallback(theApp)
		C.stopDriver()
	}()
}

//export drawgl
func drawgl(id uintptr) {
	theApp.mu.Lock()
	w := theApp.windows[id]
	theApp.mu.Unlock()

	if w == nil {
		return // closing window
	}

	// TODO: is this necessary?
	w.lifecycler.SetVisible(true)
	w.lifecycler.SendEvent(w, w.glctx)

	pe := &paint.Event{External: true}
	pe.Init()
	w.Send(pe)
	<-w.drawDone
}

// drawLoop is the primary drawing loop.
//
// After Cocoa has created an NSWindow and called prepareOpenGL,
// it starts drawLoop on a locked goroutine to handle OpenGL calls.
//
// The screen is drawn every time a paint.Event is received, which can be
// triggered either by the user or by Cocoa via drawgl (for example, when
// the window is resized).
func drawLoop(w *windowImpl, vba uintptr) {
	runtime.LockOSThread()
	C.makeCurrentContext(C.uintptr_t(w.ctx.(uintptr)))

	// Starting in OS X 10.11 (El Capitan), the vertex array is
	// occasionally getting unbound when the context changes threads.
	//
	// Avoid this by binding it again.
	C.glBindVertexArray(C.GLuint(vba))
	if errno := C.glGetError(); errno != 0 {
		panic(fmt.Sprintf("gldriver: glBindVertexArray failed: %d", errno))
	}

	workAvailable := w.worker.WorkAvailable()

	// TODO(crawshaw): exit this goroutine on Release.
	for {
		select {
		case <-workAvailable:
			w.worker.DoWork()
		case <-w.publish:
		loop:
			for {
				select {
				case <-workAvailable:
					w.worker.DoWork()
				default:
					break loop
				}
			}
			C.flushContext(C.uintptr_t(w.ctx.(uintptr)))
			w.publishDone <- oswin.PublishResult{}
		}
	}
}

//export setGeom
func setGeom(id uintptr, scrno int, dpi float32, widthPx, heightPx, leftPx, topPx int) {
	theApp.mu.Lock()
	w := theApp.windows[id]
	theApp.mu.Unlock()

	if w == nil {
		return // closing window
	}

	ldpi := oswin.LogicalFmPhysicalDPI(dpi)

	act := window.ActionN

	sz := image.Point{widthPx, heightPx}
	ps := image.Point{leftPx, topPx}

	if w.Sz != sz || w.PhysDPI != dpi || w.LogDPI != ldpi {
		act = window.Resize
	} else if w.Pos != ps {
		act = window.Move
	} else {
		act = window.Resize // todo: for now safer to default to resize -- to catch the filtering
	}

	w.sizeMu.Lock()
	w.Sz = sz
	w.Pos = ps
	w.PhysDPI = dpi
	w.LogDPI = ldpi

	if scrno > 0 && len(theApp.screens) > scrno {
		w.Scrn = theApp.screens[scrno]
	}

	w.sizeMu.Unlock()

	winEv := window.Event{
		Size:       sz,
		LogicalDPI: ldpi,
		Action:     act,
	}
	winEv.Init()
	w.Send(&winEv)
}

//export resetScreens
func resetScreens() {
	theApp.mu.Lock()
	theApp.screens = make([]*oswin.Screen, 0)
	theApp.mu.Unlock()
}

//export setScreen
func setScreen(scrIdx int, dpi, pixratio float32, widthPx, heightPx, widthMM, heightMM, depth int) {
	theApp.mu.Lock()
	var sc *oswin.Screen
	if scrIdx < len(theApp.screens) {
		sc = theApp.screens[scrIdx]
	} else {
		sc = &oswin.Screen{}
		theApp.screens = append(theApp.screens, sc)
	}
	theApp.mu.Unlock()

	sc.ScreenNumber = scrIdx
	sc.Geometry = image.Rectangle{Min: image.ZP, Max: image.Point{widthPx, heightPx}}
	sc.Depth = depth
	sc.LogicalDPI = oswin.LogicalFmPhysicalDPI(dpi)
	sc.PhysicalDPI = dpi
	sc.DevicePixelRatio = pixratio
	sc.PhysicalSize = image.Point{widthMM, heightMM}
	// todo: rest of the fields
}

//export windowClosing
func windowClosing(id uintptr) {
	winEv := window.Event{
		Action: window.Close,
	}
	sendWindowEvent(id, &winEv)
	sendLifecycle(id, (*lifecycler.State).SetDead, true)
}

func sendWindowEvent(id uintptr, e oswin.Event) {
	theApp.mu.Lock()
	w := theApp.windows[id]
	theApp.mu.Unlock()

	if w == nil {
		return // closing window
	}
	e.Init()
	w.Send(e)
}

// Mac virtual keys (kVK_) defined here:
// /System/Library/Frameworks/Carbon.framework/Frameworks/HIToolbox.framework/Headers/Events.h
// and NSEventModifierFlags are here:
// /System/Library/Frameworks/AppKit.framework/Headers/NSEvent.h
var mods = [...]struct {
	flags uint32
	code  uint16
	mod   key.Modifiers
}{
	{C.NSEventModifierFlagShift, C.kVK_Shift, key.Shift},
	{C.NSEventModifierFlagShift, C.kVK_RightShift, key.Shift},
	{C.NSEventModifierFlagControl, C.kVK_Control, key.Control},
	{C.NSEventModifierFlagControl, C.kVK_RightControl, key.Control},
	{C.NSEventModifierFlagOption, C.kVK_Option, key.Alt},
	{C.NSEventModifierFlagOption, C.kVK_RightOption, key.Alt},
	{C.NSEventModifierFlagCommand, C.kVK_Command, key.Meta},
	{C.NSEventModifierFlagCommand, C.kVK_RightCommand, key.Meta},
}

func cocoaMods(flags uint32) (m int32) {
	for _, mod := range mods {
		if flags&mod.flags != 0 {
			m |= 1 << uint32(mod.mod)
		}
	}
	return m
}

func cocoaMouseAct(ty int32) mouse.Action {
	switch ty {
	case C.NSLeftMouseDown, C.NSRightMouseDown, C.NSOtherMouseDown:
		return mouse.Press
	case C.NSLeftMouseUp, C.NSRightMouseUp, C.NSOtherMouseUp:
		return mouse.Release
	default:
		return mouse.NoAction
	}
}

func cocoaMouseButton(button int32) mouse.Button {
	switch button {
	case 0:
		return mouse.Left
	case 1:
		return mouse.Right
	case 2:
		return mouse.Middle
	default:
		return mouse.NoButton
	}
}

var lastMouseClickEvent oswin.Event
var lastMouseEvent oswin.Event

//export mouseEvent
func mouseEvent(id uintptr, x, y, dx, dy float32, ty, button int32, flags uint32) {
	cmButton := cocoaMouseButton(button)
	where := image.Point{int(x), int(y)}
	from := image.ZP
	if lastMouseEvent != nil {
		from = lastMouseEvent.Pos()
	}
	mods := cocoaMods(flags)
	var event oswin.Event
	switch ty {
	case C.NSMouseMoved:
		event = &mouse.MoveEvent{
			Event: mouse.Event{
				Where:     where,
				Button:    cmButton,
				Action:    mouse.Move,
				Modifiers: mods,
			},
			From: from,
		}
	case C.NSLeftMouseDragged, C.NSRightMouseDragged, C.NSOtherMouseDragged:
		event = &mouse.DragEvent{
			MoveEvent: mouse.MoveEvent{
				Event: mouse.Event{
					Where:     where,
					Button:    cmButton,
					Action:    mouse.Drag,
					Modifiers: mods,
				},
				From: from,
			},
		}
	case C.NSScrollWheel:
		// Note that the direction of scrolling is inverted by default
		// on OS X by the "natural scrolling" setting. At the Cocoa
		// level this inversion is applied to trackpads and mice behind
		// the scenes, and the value of dy goes in the direction the OS
		// wants scrolling to go.
		//
		// This means the same trackpad/mouse motion on OS X and Linux
		// can produce wheel events in opposite directions, but the
		// direction matches what other programs on the OS do.
		//
		// If we wanted to expose the phsyical device motion in the
		// event we could use [NSEvent isDirectionInvertedFromDevice]
		// to know if "natural scrolling" is enabled.
		event = &mouse.ScrollEvent{
			Event: mouse.Event{
				Where:     where,
				Button:    cmButton,
				Action:    mouse.Scroll,
				Modifiers: mods,
			},
			Delta: image.Point{int(dx), int(dy)},
		}
	default:
		act := cocoaMouseAct(ty)

		// todo: process DoubleClickWait option -- requires caching the
		// event and delivering conditional on a global timer... probably
		// don't want to delay things here.. some kind of go routine with a
		// timer delay on it or something like that

		if act == mouse.Press && lastMouseClickEvent != nil {
			interval := time.Now().Sub(lastMouseClickEvent.Time())
			// fmt.Printf("interval: %v\n", interval)
			if (interval / time.Millisecond) < time.Duration(mouse.DoubleClickMSec) {
				act = mouse.DoubleClick
			}
		}
		event = &mouse.Event{
			Where:     where,
			Button:    cmButton,
			Action:    act,
			Modifiers: mods,
		}
		if act == mouse.Press {
			event.SetTime()
			lastMouseClickEvent = event
		}
	}
	event.SetTime()
	lastMouseEvent = event
	sendWindowEvent(id, event)
}

//export keyEvent
func keyEvent(id uintptr, runeVal rune, act uint8, code uint16, flags uint32) {
	er := cocoaRune(runeVal)
	ea := key.Action(act)
	ec := cocoaKeyCode(code)
	em := cocoaMods(flags)

	event := &key.Event{
		Rune:      er,
		Action:    ea,
		Code:      ec,
		Modifiers: em,
	}

	sendWindowEvent(id, event)

	// do ChordEvent -- only for non-modifier key presses -- call
	// key.ChordString to convert the event into a parsable string for GUI
	// events
	if ea == key.Press && !key.CodeIsModifier(ec) {
		che := &key.ChordEvent{Event: *event}
		sendWindowEvent(id, che)
	}
}

//export flagEvent
func flagEvent(id uintptr, flags uint32) {
	for _, mod := range mods {
		if flags&mod.flags != 0 && lastFlags&mod.flags == 0 {
			keyEvent(id, -1, C.NSKeyDown, mod.code, flags)
		}
		if lastFlags&mod.flags != 0 && flags&mod.flags == 0 {
			keyEvent(id, -1, C.NSKeyUp, mod.code, flags)
		}
	}
	lastFlags = flags
}

var lastFlags uint32

func sendLifecycle(id uintptr, setter func(*lifecycler.State, bool), val bool) {
	theApp.mu.Lock()
	w := theApp.windows[id]
	theApp.mu.Unlock()

	if w == nil {
		return
	}
	setter(&w.lifecycler, val)
	w.lifecycler.SendEvent(w, w.glctx)
}

func sendLifecycleAll(dead bool) {
	windows := []*windowImpl{}

	theApp.mu.Lock()
	for _, w := range theApp.windows {
		windows = append(windows, w)
	}
	theApp.mu.Unlock()

	for _, w := range windows {
		w.lifecycler.SetFocused(false)
		w.lifecycler.SetVisible(false)
		if dead {
			w.lifecycler.SetDead(true)
		}
		w.lifecycler.SendEvent(w, w.glctx)
	}
}

//export lifecycleDeadAll
func lifecycleDeadAll() { sendLifecycleAll(true) }

//export lifecycleHideAll
func lifecycleHideAll() { sendLifecycleAll(false) }

//export lifecycleVisible
func lifecycleVisible(id uintptr, val bool) {
	sendLifecycle(id, (*lifecycler.State).SetVisible, val)
}

//export lifecycleFocused
func lifecycleFocused(id uintptr, val bool) {
	// winEv := window.Event{
	// 	Action: window.Enter,
	// }
	// sendWindowEvent(id, &winEv)
	sendLifecycle(id, (*lifecycler.State).SetFocused, val)
}

// cocoaRune marks the Carbon/Cocoa private-range unicode rune representing
// a non-unicode key event to -1, used for Rune in the key package.
//
// http://www.unicode.org/Public/MAPPINGS/VENDORS/APPLE/CORPCHAR.TXT
func cocoaRune(r rune) rune {
	if '\uE000' <= r && r <= '\uF8FF' {
		return -1
	}
	return r
}

// cocoaKeyCode converts a Carbon/Cocoa virtual key code number
// into the standard keycodes used by the key package.
//
// To get a sense of the key map, see the diagram on
//	http://boredzo.org/blog/archives/2007-05-22/virtual-key-codes
func cocoaKeyCode(vkcode uint16) key.Code {
	switch vkcode {
	case C.kVK_ANSI_A:
		return key.CodeA
	case C.kVK_ANSI_B:
		return key.CodeB
	case C.kVK_ANSI_C:
		return key.CodeC
	case C.kVK_ANSI_D:
		return key.CodeD
	case C.kVK_ANSI_E:
		return key.CodeE
	case C.kVK_ANSI_F:
		return key.CodeF
	case C.kVK_ANSI_G:
		return key.CodeG
	case C.kVK_ANSI_H:
		return key.CodeH
	case C.kVK_ANSI_I:
		return key.CodeI
	case C.kVK_ANSI_J:
		return key.CodeJ
	case C.kVK_ANSI_K:
		return key.CodeK
	case C.kVK_ANSI_L:
		return key.CodeL
	case C.kVK_ANSI_M:
		return key.CodeM
	case C.kVK_ANSI_N:
		return key.CodeN
	case C.kVK_ANSI_O:
		return key.CodeO
	case C.kVK_ANSI_P:
		return key.CodeP
	case C.kVK_ANSI_Q:
		return key.CodeQ
	case C.kVK_ANSI_R:
		return key.CodeR
	case C.kVK_ANSI_S:
		return key.CodeS
	case C.kVK_ANSI_T:
		return key.CodeT
	case C.kVK_ANSI_U:
		return key.CodeU
	case C.kVK_ANSI_V:
		return key.CodeV
	case C.kVK_ANSI_W:
		return key.CodeW
	case C.kVK_ANSI_X:
		return key.CodeX
	case C.kVK_ANSI_Y:
		return key.CodeY
	case C.kVK_ANSI_Z:
		return key.CodeZ
	case C.kVK_ANSI_1:
		return key.Code1
	case C.kVK_ANSI_2:
		return key.Code2
	case C.kVK_ANSI_3:
		return key.Code3
	case C.kVK_ANSI_4:
		return key.Code4
	case C.kVK_ANSI_5:
		return key.Code5
	case C.kVK_ANSI_6:
		return key.Code6
	case C.kVK_ANSI_7:
		return key.Code7
	case C.kVK_ANSI_8:
		return key.Code8
	case C.kVK_ANSI_9:
		return key.Code9
	case C.kVK_ANSI_0:
		return key.Code0
	// TODO: move the rest of these codes to constants in key.go
	// if we are happy with them.
	case C.kVK_Return:
		return key.CodeReturnEnter
	case C.kVK_Escape:
		return key.CodeEscape
	case C.kVK_Delete:
		return key.CodeDeleteBackspace
	case C.kVK_Tab:
		return key.CodeTab
	case C.kVK_Space:
		return key.CodeSpacebar
	case C.kVK_ANSI_Minus:
		return key.CodeHyphenMinus
	case C.kVK_ANSI_Equal:
		return key.CodeEqualSign
	case C.kVK_ANSI_LeftBracket:
		return key.CodeLeftSquareBracket
	case C.kVK_ANSI_RightBracket:
		return key.CodeRightSquareBracket
	case C.kVK_ANSI_Backslash:
		return key.CodeBackslash
	// 50: Keyboard Non-US "#" and ~
	case C.kVK_ANSI_Semicolon:
		return key.CodeSemicolon
	case C.kVK_ANSI_Quote:
		return key.CodeApostrophe
	case C.kVK_ANSI_Grave:
		return key.CodeGraveAccent
	case C.kVK_ANSI_Comma:
		return key.CodeComma
	case C.kVK_ANSI_Period:
		return key.CodeFullStop
	case C.kVK_ANSI_Slash:
		return key.CodeSlash
	case C.kVK_CapsLock:
		return key.CodeCapsLock
	case C.kVK_F1:
		return key.CodeF1
	case C.kVK_F2:
		return key.CodeF2
	case C.kVK_F3:
		return key.CodeF3
	case C.kVK_F4:
		return key.CodeF4
	case C.kVK_F5:
		return key.CodeF5
	case C.kVK_F6:
		return key.CodeF6
	case C.kVK_F7:
		return key.CodeF7
	case C.kVK_F8:
		return key.CodeF8
	case C.kVK_F9:
		return key.CodeF9
	case C.kVK_F10:
		return key.CodeF10
	case C.kVK_F11:
		return key.CodeF11
	case C.kVK_F12:
		return key.CodeF12
	// 70: PrintScreen
	// 71: Scroll Lock
	// 72: Pause
	// 73: Insert
	case C.kVK_Home:
		return key.CodeHome
	case C.kVK_PageUp:
		return key.CodePageUp
	case C.kVK_ForwardDelete:
		return key.CodeDeleteForward
	case C.kVK_End:
		return key.CodeEnd
	case C.kVK_PageDown:
		return key.CodePageDown
	case C.kVK_RightArrow:
		return key.CodeRightArrow
	case C.kVK_LeftArrow:
		return key.CodeLeftArrow
	case C.kVK_DownArrow:
		return key.CodeDownArrow
	case C.kVK_UpArrow:
		return key.CodeUpArrow
	case C.kVK_ANSI_KeypadClear:
		return key.CodeKeypadNumLock
	case C.kVK_ANSI_KeypadDivide:
		return key.CodeKeypadSlash
	case C.kVK_ANSI_KeypadMultiply:
		return key.CodeKeypadAsterisk
	case C.kVK_ANSI_KeypadMinus:
		return key.CodeKeypadHyphenMinus
	case C.kVK_ANSI_KeypadPlus:
		return key.CodeKeypadPlusSign
	case C.kVK_ANSI_KeypadEnter:
		return key.CodeKeypadEnter
	case C.kVK_ANSI_Keypad1:
		return key.CodeKeypad1
	case C.kVK_ANSI_Keypad2:
		return key.CodeKeypad2
	case C.kVK_ANSI_Keypad3:
		return key.CodeKeypad3
	case C.kVK_ANSI_Keypad4:
		return key.CodeKeypad4
	case C.kVK_ANSI_Keypad5:
		return key.CodeKeypad5
	case C.kVK_ANSI_Keypad6:
		return key.CodeKeypad6
	case C.kVK_ANSI_Keypad7:
		return key.CodeKeypad7
	case C.kVK_ANSI_Keypad8:
		return key.CodeKeypad8
	case C.kVK_ANSI_Keypad9:
		return key.CodeKeypad9
	case C.kVK_ANSI_Keypad0:
		return key.CodeKeypad0
	case C.kVK_ANSI_KeypadDecimal:
		return key.CodeKeypadFullStop
	case C.kVK_ANSI_KeypadEquals:
		return key.CodeKeypadEqualSign
	case C.kVK_F13:
		return key.CodeF13
	case C.kVK_F14:
		return key.CodeF14
	case C.kVK_F15:
		return key.CodeF15
	case C.kVK_F16:
		return key.CodeF16
	case C.kVK_F17:
		return key.CodeF17
	case C.kVK_F18:
		return key.CodeF18
	case C.kVK_F19:
		return key.CodeF19
	case C.kVK_F20:
		return key.CodeF20
	// 116: Keyboard Execute
	case C.kVK_Help:
		return key.CodeHelp
	// 118: Keyboard Menu
	// 119: Keyboard Select
	// 120: Keyboard Stop
	// 121: Keyboard Again
	// 122: Keyboard Undo
	// 123: Keyboard Cut
	// 124: Keyboard Copy
	// 125: Keyboard Paste
	// 126: Keyboard Find
	case C.kVK_Mute:
		return key.CodeMute
	case C.kVK_VolumeUp:
		return key.CodeVolumeUp
	case C.kVK_VolumeDown:
		return key.CodeVolumeDown
	// 130: Keyboard Locking Caps Lock
	// 131: Keyboard Locking Num Lock
	// 132: Keyboard Locking Scroll Lock
	// 133: Keyboard Comma
	// 134: Keyboard Equal Sign
	// ...: Bunch of stuff
	case C.kVK_Control:
		return key.CodeLeftControl
	case C.kVK_Shift:
		return key.CodeLeftShift
	case C.kVK_Option:
		return key.CodeLeftAlt
	case C.kVK_Command:
		return key.CodeLeftGUI
	case C.kVK_RightControl:
		return key.CodeRightControl
	case C.kVK_RightShift:
		return key.CodeRightShift
	case C.kVK_RightOption:
		return key.CodeRightAlt
	// TODO key.CodeRightGUI
	default:
		return key.CodeUnknown
	}
}

func surfaceCreate() error {
	return errors.New("gldriver: surface creation not implemented on darwin")
}

func (app *appImpl) PrefsDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Print(err)
		return "/tmp"
	}
	return filepath.Join(usr.HomeDir, "Library")
}

func (app *appImpl) GoGiPrefsDir() string {
	pdir := filepath.Join(app.PrefsDir(), "GoGi")
	os.MkdirAll(pdir, 0755)
	return pdir
}

func (app *appImpl) AppPrefsDir() string {
	pdir := filepath.Join(app.PrefsDir(), app.Name())
	os.MkdirAll(pdir, 0755)
	return pdir
}

func (app *appImpl) FontPaths() []string {
	return []string{"/Library/Fonts"}
}
