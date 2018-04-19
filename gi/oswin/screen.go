// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// based on golang.org/x/exp/shiny:
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oswin

import (
	"image"

	"github.com/rcoreilly/goki/ki/kit"
)

// note: fields obtained from QScreen in Qt

// Screen contains data about each physical and / or logical screen
type Screen struct {
	// ScreenNumber is the index of this screen in the list of screens
	// maintained under Screen
	ScreenNumber int

	// Geometry contains the geometry of the screen in raw pixels -- all physical screens start at 0,0
	Geometry image.Rectangle

	// Color depth of the screen, in bits
	Depth int

	// LogicalDPI is the logical dots per inch of the window, which is used for all
	// rendering -- subject to zooming effects etc -- see the gi/units package
	// for translating into various other units
	LogicalDPI float32

	// PhysicalDPI is the physical dots per inch of the window, for generating
	// true-to-physical-size output, for example -- see the gi/units package for
	// translating into various other units
	PhysicalDPI float32

	// PhysicalSize is the actual physical size of the screen, in mm
	PhysicalSize image.Point

	// DevicePixelRatio is a multiplier factor that scales the screen's
	// "natural" pixel coordinates into actual device pixels.
	//
	// On OS-X  it is backingScaleFactor, which is 2.0 on "retina" displays
	DevicePixelRatio float32

	RefreshRate float32

	AvailableGeometry        image.Rectangle
	VirtualGeometry          image.Rectangle
	AvailableVirtualGeometry image.Rectangle

	Orientation        ScreenOrientation
	NativeOrientation  ScreenOrientation
	PrimaryOrientation ScreenOrientation

	Name         string
	Manufacturer string
	Model        string
	SerialNumber string
}

// ScreenOrientation is the orientation of the device screen.
type ScreenOrientation int32

const (
	// OrientationUnknown means device orientation cannot be determined.
	//
	// Equivalent on Android to Configuration.ORIENTATION_UNKNOWN
	// and on iOS to:
	//	UIDeviceOrientationUnknown
	//	UIDeviceOrientationFaceUp
	//	UIDeviceOrientationFaceDown
	OrientationUnknown ScreenOrientation = iota

	// Portrait is a device oriented so it is tall and thin.
	//
	// Equivalent on Android to Configuration.ORIENTATION_PORTRAIT
	// and on iOS to:
	//	UIDeviceOrientationPortrait
	//	UIDeviceOrientationPortraitUpsideDown
	Portrait

	// Landscape is a device oriented so it is short and wide.
	//
	// Equivalent on Android to Configuration.ORIENTATION_LANDSCAPE
	// and on iOS to:
	//	UIDeviceOrientationLandscapeLeft
	//	UIDeviceOrientationLandscapeRight
	Landscape

	ScreenOrientationN
)

//go:generate stringer -type=ScreenOrientation

var KiT_ScreenOrientation = kit.Enums.AddEnum(ScreenOrientationN, false, nil)
