// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
	"github.com/go-gl/mathgl/mgl32"
	// "image"
)

// a 2D rectangle,
type GiRect struct {
	GiNode2D
	Pos    mgl32.Vec2 `svg:"{x,y}",desc:"position of top-left corner"`
	Size   mgl32.Vec2 `svg:"{width,height}",desc:"size of viewbox within parent Viewport2D"`
	Radius mgl32.Vec2 `svg:"{rx,ry}",desc:"radii for curved corners, as a proportion of width, height"`
}

func (g *GiRect) Render(vp *Viewport2D, xf *Transform2D) {

}
