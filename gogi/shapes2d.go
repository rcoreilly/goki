// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
// "fmt"
)

// a 2D rectangle,
type GiRect struct {
	GiNode2D
	Pos    Point2D `svg:"{x,y}",desc:"position of top-left corner"`
	Size   Size2D  `svg:"{width,height}",desc:"size of viewbox within parent Viewport2D"`
	Radius Point2D `svg:"{rx,ry}",desc:"radii for curved corners, as a proportion of width, height"`
}

func (g *GiRect) Render2DNode() *GiNode2D {
	return &g.GiNode2D
}

// viewport render has already handled the SetPaintFromNode call, and also looked for disabled
func (g *GiRect) Render2D(vp *Viewport2D) bool {
	if vp.HasNoStrokeOrFill() {
		return true
	}
	if g.Radius.X == 0 && g.Radius.Y == 0 {
		vp.DrawRectangle(g.Pos.X, g.Pos.Y, g.Size.X, g.Size.Y)
	} else {
		// todo: only supports 1 radius right now -- easy to add another
		vp.DrawRoundedRectangle(g.Pos.X, g.Pos.Y, g.Size.X, g.Size.Y, g.Radius.X)
	}
	if vp.HasFill() {
		vp.FillPreserve()
	}
	if vp.HasStroke() {
		vp.StrokePreserve()
	}
	vp.ClearPath()
	return true
}
