// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oswin

import (
	"image"
	"time"

	"github.com/rcoreilly/goki/ki/kit"
)

// GoGi event structure is dervied from go.wde and golang/x/mobile/event
//
// GoGi requires event type enum for widgets to request what events to
// receive, and we add an overall interface with base support for time and
// marking events as processed, which is critical for simplifying logic and
// preventing unintended multiple effects
//
// OSWin deals exclusively in raw "dot" pixel integer coordinates (as in
// go.wde) -- abstraction to different DPI etc takes place higher up in the
// system

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
   Copyright 2012 the go.wde authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// EventType determines which type of GUI event is being sent -- need this for
// indexing into different event signalers based on event type, and sending
// event type in signals -- critical to break up different event types into
// the right categories needed for different types of widgets -- e.g., most do
// not need move or scroll events, so those are separated
type EventType int64

const (
	// MouseEvent includes all mouse button actions, but not move or drag
	MouseEvent EventType = iota
	// MouseMoveEvent is when the mouse is moving but no button is down
	MouseMoveEvent
	// MouseDragEvent is when the mouse is moving and there is a button down
	MouseDragEvent
	// MouseScrollEvent is for mouse scroll wheel events
	MouseScrollEvent
	// MouseFocusEvent is for mouse focus (enter / exit of widget area) --
	// generated by gi.Window based on mouse move events
	MouseFocusEvent

	// KeyEvent for key pressed or released -- fine-grained data about each key as it happens
	KeyEvent
	// KeyChordEvent is only generated when a non-modifier key is released,
	// and it also contains a string representation of the full chord,
	// suitable for translation into keyboard commands, emacs-style etc
	KeyChordEvent

	// touch and gesture events
	TouchEvent
	MagnifyEvent
	RotateEvent

	// WindowEvent reports any changes in the window size, orientation, iconify, close, open
	WindowEvent

	// LifeCycleEvent for app window activation, focus, and startup / termination
	LifeCycleEvent

	// PaintEvent for notifications about painting windows
	PaintEvent

	// number of event types
	EventTypeN
)

//go:generate stringer -type=EventType

var KiT_EventType = kit.Enums.AddEnum(EventTypeN, false, nil)

// interface for GUI events
type Event interface {
	// Type returns the type of event associated with given event
	Type() EventType

	// HasPos returns true if the event has a window position where it takes place
	HasPos() bool

	// Pos returns the position in raw display dots (pixels) where event took place -- needed for sending events to the right place
	Pos() image.Point

	// OnFocus returns true if the event operates only on focus item (e.g., keyboard events)
	OnFocus() bool

	// Window returns the oswin window that the event applies to
	Window() Window

	// Time returns the time at which the event was generated
	Time() time.Time

	// IsProcessed returns whether this event has already been processed
	IsProcessed() bool

	// SetProcessed marks the event as having been processed
	SetProcessed()

	// Init sets the window and the time to now -- done just prior to event delivery
	Init(win Window)

	// SetTime sets the event time to Now
	SetTime()
}

// base type for events -- records time and whether event has been processed by a receiver of the event -- in which case it is skipped
type EventBase struct {
	// Window is the oswin window that the event applies to
	OSWin Window

	// GenTime records the time when the event was first generated
	GenTime time.Time

	// Processed indicates if the event has been processed by an end receiver,
	// and thus should no longer be processed by other possible receivers
	Processed bool
}

// SetTime sets the event time to Now
func (ev *EventBase) SetTime() {
	ev.GenTime = time.Now()
}

func (ev *EventBase) Init(win Window) {
	ev.SetTime()
	ev.OSWin = win
}

func (ev EventBase) Window() Window {
	return ev.OSWin
}

func (ev EventBase) Time() time.Time {
	return ev.GenTime
}

func (ev EventBase) IsProcessed() bool {
	return ev.Processed
}

func (ev *EventBase) SetProcessed() {
	ev.Processed = true
}

// EventDeque is an infinitely buffered double-ended queue of events.
type EventDeque interface {
	// Send adds an event to the end of the deque. They are returned by
	// NextEvent in FIFO order.
	Send(event Event)

	// SendFirst adds an event to the start of the deque. They are returned by
	// NextEvent in LIFO order, and have priority over events sent via Send.
	SendFirst(event Event)

	// NextEvent returns the next event in the deque. It blocks until such an
	// event has been sent.
	NextEvent() Event

	// TODO: LatestLifecycleEvent? Is that still worth it if the
	// lifecycle.Event struct type loses its DrawContext field?

	// TODO: LatestSizeEvent?
}
