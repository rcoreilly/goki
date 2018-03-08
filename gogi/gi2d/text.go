// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
	//	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"image"
)

type PaintFont struct {
	Face   font.Face
	Height float64
}

func (p *PaintFont) Defaults() {
	p.Face = basicfont.Face7x13
	p.Height = 12
}
