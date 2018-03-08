// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package GoGi provides a complete Graphical Interface based on GoKi Tree Node structs

	The GiNode struct that implements the Ki interface, which
	can be used as an embedded type (or a struct field) in other structs to provide
	core tree functionality, including:
		* Parent / Child Tree structure -- each Node can ONLY have one parent
		* Paths for locating Nodes within the hierarchy -- key for many use-cases, including IO for pointers
		* Apply a function across nodes up or down a tree -- very flexible for tree walking
		* Generalized I/O -- can Save and Load the Tree as JSON, XML, etc -- including pointers which are saved using paths and automatically cached-out after loading
		* Event sending and receiving between Nodes (simlar to Qt Signals / Slots)
		* Robust updating state -- wrap updates in UpdateStart / End, and signals are blocked until the final end, at which point an update signal is sent -- works across levels
		* Properties (as a string-keyed map) with property inheritance -- css anyone!?
*/
package gogi

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/rcoreilly/goki/ki"
	"gopkg.in/go-playground/colors.v1"
	"log"
	"strconv"
)

// todo: not clear if we need any more interfaces??

// basic component node for GoGi
type GiNode struct {
	ki.Node
}

// standard css properties on nodes apply, including visible, etc.

// basic component node for 2D rendering
type GiNode2D struct {
	GiNode
	z_index int `svg:"z-index",desc:"ordering factor for rendering depth -- lower numbers rendered first -- sort children according to this factor"`
	// todo: do we want to cache any transforms or anything? maybe not?
}

// this is the primary interface for all 2D rendering nodes
type Renderer2D interface {
	// Render graphics into a 2D viewport, subject to overall rendering transform xf
	Render(vp *Viewport2D, xf Transform2D)
}

// basic component node for 3D rendering -- has a 3D transform
type GiNode3D struct {
	GiNode
}

// process properties and any css style sheets (todo) to get a length property of the given name -- returns false if property has not been set -- automatically deals with units such as px, em etc
func (g *GiNode) PropLength(name string) (float64, bool) {
	p := g.Prop(name, true) // true = inherit
	if p == nil {
		return 0, false
	}
	switch v := p.(type) {
	case string:
		// todo: need to parse units from string!
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Printf("GiNode %v PropLength convert from string err: %v", err)
			return 0, false
		}
		return f
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

// process properties and any css style sheets (todo) to get a number property of the given name -- returns false if property has not been set
func (g *GiNode) PropNumber(name string) (float64, bool) {
	p := g.Prop(name, true) // true = inherit
	if p == nil {
		return 0, false
	}
	switch v := p.(type) {
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Printf("GiNode %v PropNumber convert from string err: %v", err)
			return 0, false
		}
		return f
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

// process properties and any css style sheets (todo) to get an enumerated type as a string -- returns true if value is present
func (g *GiNode) PropEnum(name string) (string, bool) {
	p := g.Prop(name, true) // true = inherit
	if p == nil {
		return "", false
	}
	switch v := p.(type) {
	case string:
		return v, (len(v) > 0)
	default:
		return "", false
	}
}

// process properties and any css style sheets (todo) to get a color
func (g *GiNode) PropColor(name string) (color.Color, bool) {
	p := g.Prop(name, true) // true = inherit
	if p == nil {
		return nil, false
	}
	switch v := p.(type) {
	case string:
		cl, err := colors.Parse(v)
		if err != nil {
			log.Printf("GiNode %v PropColor convert from string err: %v", err)
			return nil, false
		}
		rgba := cl.ToRGBA()
		return color.RGBA(rgba.r, rgba.g, rgba.b, rgba.a), true
	default:
		return nil, false
	}
}
