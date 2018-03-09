// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
	"reflect"
	"testing"
)

func TestRect(t *testing.T) {
	parent := NewViewport2D(400, 400)
	parent.SetThisName(parent, "vp1")
	rect1 := parent.AddNewChildNamed(reflect.TypeOf(GiRect{}), "rect1").(*GiRect)
	rect1.SetProp("fill", "#008800")
	rect1.SetProp("stroke", "#0000FF")
	rect1.SetProp("stroke-width", 5.0)
	// rect1.SetProp("stroke-linejoin", "round")
	rect1.Pos = Point2D{10, 10}
	rect1.Size = Size2D{100, 100}
	parent.Clear()
	parent.RenderTopLevel()
	parent.SavePNG("test_rect.png")
}
