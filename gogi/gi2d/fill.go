// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
	//	"github.com/go-gl/mathgl/mgl32"
	"image"
)

type FillRule int

const (
	FillRuleNonZero FillRule = iota
	FillRuleEvenOdd
)

// contrary to some docs, apparently need to run go generate manually
//go:generate stringer -type=FillRule

// PaintFill contains all the properties specific to filling a region
type PaintFill struct {
	Color color.Color `svg:"fill",desc:"color to fill in"`
	Rule  FillRule    `svg:"fill-rule",desc:"rule for how to fill more complex shapes with crossing lines"`
	Pat   Pattern     `desc:"pattern for the stroke -- not clear if this is in svg"`
}

// initialize default values for paint fill
func (p *PaintFill) Defaults() {
	Color = color.Transparent
}

// todo: figure out more elemental, generic de-stringer kind of thing

// update the fill settings from the style info on the node
func (s *PaintFill) FillStyle(g *GiNode2D) {
	// always check if property has been set before setting -- otherwise defaults to empty -- true = inherit props
	// todo: need to be able to process colors!

	if c, got := g.PropColor("fill"); got {
		s.Color = c
	}
	if o, got := g.PropNumber("fill-opacity"); got {
		// todo: need to set the color alpha according to value
	}
	if ps, got := g.PropEnum("fill-rule", true); got {
		var fr FillRule = -1
		switch ps {
		case "nonzero":
			fr = FillRuleNonZero
		case "evenodd":
			fr = FillRuleEvenOdd
		}
		if fr == -1 {
			i, err := StringToFillRule(ps) // stringer gen
			if err != nil {
				s.Rule = i
			} else {
				log.Print(err)
			}
		}
	}
}
