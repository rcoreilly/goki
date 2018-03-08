// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
	//	"github.com/go-gl/mathgl/mgl64"
	"image"
)

// end-cap of a line: stroke-linecap property in SVG
type LineCap int

const (
	LineCapRound LineCap = iota
	LineCapButt
	LineCapSquare
)

// contrary to some docs, apparently need to run go generate manually
//go:generate stringer -type=LineCap

// the way in which lines are joined together: stroke-linejoin property in SVG
type LineJoin int

const (
	LineJoinMiter     LineJoin = iota
	LineJoinMiterClip          // SVG2 -- not yet supported
	LineJoinRound
	LineJoinBevel
	LineJoinArcs // SVG2 -- not yet supported
)

// contrary to some docs, apparently need to run go generate manually
//go:generate stringer -type=LineJoin

// PaintStroke contains all the properties specific to painting a line -- the svg elements define the corresponding SVG style attributes, which are processed in StrokeStyle
type PaintStroke struct {
	Color      color.Color `svg:"stroke",desc:"color of the stroke"`
	Width      float64     `svg:"stroke-width",desc:"line width"`
	Dashes     []float64   `svg:"stroke-dasharray",desc:"dash pattern"`
	Cap        LineCap     `svg:"stroke-linecap",desc:"how to draw the end cap of lines"`
	Join       LineJoin    `svg:"stroke-linejoin",desc:"how to join line segments"`
	MiterLimit float64     `svg:"stroke-miterlimit,min:"1",desc:"limit of how far to miter -- must be 1 or larger"`
	Pat        Pattern     `desc:"pattern for the stroke -- not clear if this is in svg"`
}

// initialize default values for paint stroke
func (p *PaintStroke) Defaults() {
	Color = color.Black
	Width = 1.0
}

// todo: figure out more elemental, generic de-stringer kind of thing

// update the stroke settings from the style info on the node
func (s *PaintStroke) StrokeStyle(g *GiNode2D) {
	// always check if property has been set before setting -- otherwise defaults to empty -- true = inherit props
	if c, got := g.PropColor("stroke"); got {
		s.Color = c
	}
	if w, got := g.PropLength("stroke-width"); got {
		s.Width = w
	}
	if o, got := g.PropNumber("stroke-opacity"); got {
		// todo: need to set the color alpha according to value
	}
	if ps, got := g.PropEnum("stroke-linecap", true); got {
		var lc LineCap = -1
		switch ps { // first go through short-hand codes
		case "round":
			lc = LineCapRound
		case "butt":
			lc = LineCapButt
		case "square":
			lc = LineCapSquare
		}
		if lc == -1 {
			i, err := StringToLineCap(ps) // stringer gen
			if err != nil {
				s.Cap = i
			} else {
				log.Print(err)
			}
		} else {
			s.Cap = lc
		}
	}
	if ps, got := g.Prop("stroke-linejoin", true); got {
		var lc LineJoin = -1
		switch ps { // first go through short-hand codes
		case "miter":
			lc = LineJoinMiter
		case "miter-clip":
			lc = LineJoinMiterClip
		case "round":
			lc = LineJoinRound
		case "bevel":
			lc = LineJoinBevel
		case "arcs":
			lc = LineJoinArcs
		}
		if lc == -1 {
			i, err := StringToLineJoin(ps) // stringer gen
			if err != nil {
				s.Join = i
			} else {
				log.Print(err)
			}
		} else {
			s.Join = lc
		}
	}
	if l, got := g.PropNumber("miter-limit"); got {
		s.MiterLimit = l
	}
}
