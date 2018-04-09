// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"image"
	"strconv"
	"strings"
	"unicode"

	"github.com/rcoreilly/goki/ki/kit"
)

// shapes2d contains all the SVG-based objects for drawing shapes, paths, etc

////////////////////////////////////////////////////////////////////////////////////////
// Rect

// 2D rectangle, optionally with rounded corners
type Rect struct {
	Node2DBase
	Pos    Vec2D `xml:"{x,y}" desc:"position of the top-left of the rectangle"`
	Size   Vec2D `xml:"{width,height}" desc:"size of the rectangle"`
	Radius Vec2D `xml:"{rx,ry}" desc:"radii for curved corners, as a proportion of width, height"`
}

var KiT_Rect = kit.Types.AddType(&Rect{}, nil)

func (g *Rect) BBox2D() image.Rectangle {
	rs := &g.Viewport.Render
	return g.Paint.BoundingBox(rs, g.Pos.X, g.Pos.Y, g.Pos.X+g.Size.X, g.Pos.Y+g.Size.Y)
}

func (g *Rect) Render2D() {
	if g.PushBounds() {
		pc := &g.Paint
		rs := &g.Viewport.Render
		rs.PushXForm(pc.XForm)
		if g.Radius.X == 0 && g.Radius.Y == 0 {
			pc.DrawRectangle(rs, g.Pos.X, g.Pos.Y, g.Size.X, g.Size.Y)
		} else {
			// todo: only supports 1 radius right now -- easy to add another
			pc.DrawRoundedRectangle(rs, g.Pos.X, g.Pos.Y, g.Size.X, g.Size.Y, g.Radius.X)
		}
		pc.FillStrokeClear(rs)
		g.Render2DChildren()
		g.PopBounds()
		rs.PopXForm()
	}
}

func (g *Rect) ReRender2D() (node Node2D, layout bool) {
	svg := g.ParentSVG()
	if svg != nil {
		node = svg
	} else {
		node = g.This.(Node2D) // no other option..
	}
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Rect{}

////////////////////////////////////////////////////////////////////////////////////////
// Viewport2DFill

// todo: for ViewportFill support an option to insert a HiDPI correction scaling factor at the top!

// viewport fill fills entire viewport -- just a rect that automatically sets size to viewport
type Viewport2DFill struct {
	Rect
}

var KiT_Viewport2DFill = kit.Types.AddType(&Viewport2DFill{}, nil)

func (g *Viewport2DFill) Init2D() {
	g.Init2DBase()
	vp := g.Viewport
	g.Pos = Vec2DZero
	g.Size = Vec2D{float64(vp.ViewBox.Size.X), float64(vp.ViewBox.Size.Y)} // assuming no transforms..
}

func (g *Viewport2DFill) Style2D() {
	g.Style2DSVG(nil)
}

func (g *Viewport2DFill) BBox2D() image.Rectangle {
	g.Init2D() // keep up-to-date -- cheap
	rs := &g.Viewport.Render
	return g.Paint.BoundingBox(rs, g.Pos.X, g.Pos.Y, g.Pos.X+g.Size.X, g.Pos.Y+g.Size.Y)
}

func (g *Viewport2DFill) ReRender2D() (node Node2D, layout bool) {
	node = g.This.(Node2D)
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Viewport2DFill{}

////////////////////////////////////////////////////////////////////////////////////////
// Circle

// 2D circle
type Circle struct {
	Node2DBase
	Pos    Vec2D   `xml:"{cx,cy}" desc:"position of the center of the circle"`
	Radius float64 `xml:"r" desc:"radius of the circle"`
}

var KiT_Circle = kit.Types.AddType(&Circle{}, nil)

func (g *Circle) BBox2D() image.Rectangle {
	rs := &g.Viewport.Render
	return g.Paint.BoundingBox(rs, g.Pos.X-g.Radius, g.Pos.Y-g.Radius, g.Pos.X+g.Radius, g.Pos.Y+g.Radius)
}

func (g *Circle) Render2D() {
	if g.PushBounds() {
		pc := &g.Paint
		rs := &g.Viewport.Render
		rs.PushXForm(pc.XForm)
		pc.DrawCircle(rs, g.Pos.X, g.Pos.Y, g.Radius)
		pc.FillStrokeClear(rs)
		g.Render2DChildren()
		g.PopBounds()
		rs.PopXForm()
	}
}

func (g *Circle) ReRender2D() (node Node2D, layout bool) {
	svg := g.ParentSVG()
	if svg != nil {
		node = svg
	} else {
		node = g.This.(Node2D) // no other option..
	}
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Circle{}

////////////////////////////////////////////////////////////////////////////////////////
// Ellipse

// 2D ellipse
type Ellipse struct {
	Node2DBase
	Pos   Vec2D `xml:"{cx,cy}" desc:"position of the center of the ellipse"`
	Radii Vec2D `xml:"{rx, ry}" desc:"radii of the ellipse in the horizontal, vertical axes"`
}

var KiT_Ellipse = kit.Types.AddType(&Ellipse{}, nil)

func (g *Ellipse) BBox2D() image.Rectangle {
	rs := &g.Viewport.Render
	return g.Paint.BoundingBox(rs, g.Pos.X-g.Radii.X, g.Pos.Y-g.Radii.Y, g.Pos.X+g.Radii.X, g.Pos.Y+g.Radii.Y)
}

func (g *Ellipse) Render2D() {
	if g.PushBounds() {
		pc := &g.Paint
		rs := &g.Viewport.Render
		rs.PushXForm(pc.XForm)
		pc.DrawEllipse(rs, g.Pos.X, g.Pos.Y, g.Radii.X, g.Radii.Y)
		pc.FillStrokeClear(rs)
		g.Render2DChildren()
		g.PopBounds()
		rs.PopXForm()
	}
}

func (g *Ellipse) ReRender2D() (node Node2D, layout bool) {
	svg := g.ParentSVG()
	if svg != nil {
		node = svg
	} else {
		node = g.This.(Node2D) // no other option..
	}
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Ellipse{}

////////////////////////////////////////////////////////////////////////////////////////
// Line

// a 2D line
type Line struct {
	Node2DBase
	Start Vec2D `xml:"{x1,y1}" desc:"position of the start of the line"`
	End   Vec2D `xml:"{x2, y2}" desc:"position of the end of the line"`
}

var KiT_Line = kit.Types.AddType(&Line{}, nil)

func (g *Line) BBox2D() image.Rectangle {
	rs := &g.Viewport.Render
	return g.Paint.BoundingBox(rs, g.Start.X, g.Start.Y, g.End.X, g.End.Y).Canon()
}

func (g *Line) Render2D() {
	if g.PushBounds() {
		pc := &g.Paint
		rs := &g.Viewport.Render
		rs.PushXForm(pc.XForm)
		pc.DrawLine(rs, g.Start.X, g.Start.Y, g.End.X, g.End.Y)
		pc.Stroke(rs)
		g.Render2DChildren()
		g.PopBounds()
		rs.PopXForm()
	}
}

func (g *Line) ReRender2D() (node Node2D, layout bool) {
	svg := g.ParentSVG()
	if svg != nil {
		node = svg
	} else {
		node = g.This.(Node2D) // no other option..
	}
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Line{}

////////////////////////////////////////////////////////////////////////////////////////
// Polyline

// 2D Polyline
type Polyline struct {
	Node2DBase
	Points []Vec2D `xml:"points" desc:"the coordinates to draw -- does a moveto on the first, then lineto for all the rest"`
}

var KiT_Polyline = kit.Types.AddType(&Polyline{}, nil)

func (g *Polyline) BBox2D() image.Rectangle {
	rs := &g.Viewport.Render
	return g.Paint.BoundingBoxFromPoints(rs, g.Points)
}

func (g *Polyline) Render2D() {
	if len(g.Points) < 2 {
		return
	}
	if g.PushBounds() {
		pc := &g.Paint
		rs := &g.Viewport.Render
		rs.PushXForm(pc.XForm)
		pc.DrawPolyline(rs, g.Points)
		pc.FillStrokeClear(rs)
		g.Render2DChildren()
		g.PopBounds()
		rs.PopXForm()
	}
}

func (g *Polyline) ReRender2D() (node Node2D, layout bool) {
	svg := g.ParentSVG()
	if svg != nil {
		node = svg
	} else {
		node = g.This.(Node2D) // no other option..
	}
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Polyline{}

////////////////////////////////////////////////////////////////////////////////////////
// Polygon

// 2D Polygon
type Polygon struct {
	Node2DBase
	Points []Vec2D `xml:"points" desc:"the coordinates to draw -- does a moveto on the first, then lineto for all the rest, then does a closepath at the end"`
}

var KiT_Polygon = kit.Types.AddType(&Polygon{}, nil)

func (g *Polygon) BBox2D() image.Rectangle {
	rs := &g.Viewport.Render
	return g.Paint.BoundingBoxFromPoints(rs, g.Points)
}

func (g *Polygon) Render2D() {
	if len(g.Points) < 2 {
		return
	}
	if g.PushBounds() {
		pc := &g.Paint
		rs := &g.Viewport.Render
		rs.PushXForm(pc.XForm)
		pc.DrawPolygon(rs, g.Points)
		pc.FillStrokeClear(rs)
		g.Render2DChildren()
		g.PopBounds()
		rs.PopXForm()
	}
}

func (g *Polygon) ReRender2D() (node Node2D, layout bool) {
	svg := g.ParentSVG()
	if svg != nil {
		node = svg
	} else {
		node = g.This.(Node2D) // no other option..
	}
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Polygon{}

////////////////////////////////////////////////////////////////////////////////////////
// Path

// the commands within the path SVG drawing data type
type PathCmds byte

const (
	// move pen, abs coords
	PcM PathCmds = iota
	// move pen, rel coords
	Pcm
	// lineto, abs
	PcL
	// lineto, rel
	Pcl
	// horizontal lineto, abs
	PcH
	// relative lineto, rel
	Pch
	// vertical lineto, abs
	PcV
	// vertical lineto, rel
	Pcv
	// Bezier curveto, abs
	PcC
	// Bezier curveto, rel
	Pcc
	// smooth Bezier curveto, abs
	PcS
	// smooth Bezier curveto, rel
	Pcs
	// quadratic Bezier curveto, abs
	PcQ
	// quadratic Bezier curveto, rel
	Pcq
	// smooth quadratic Bezier curveto, abs
	PcT
	// smooth quadratic Bezier curveto, rel
	Pct
	// elliptical arc, abs
	PcA
	// elliptical arc, rel
	Pca
	// close path
	PcZ
	// close path
	Pcz
)

// to encode the data, we use 32-bit floats which are converted into int32 for
// path commands, which contain the number of data points following the path
// command to interpret as numbers, in the lower and upper 2 bytes of the
// converted int32 number we don't need that many bits, but keeping 32-bit
// alignment is probably good and really these things don't need to be crazy
// compact as it is unlikely to make a relevant diff in size or perf to pack
// down further
type PathData float32

// decode path data as a command and a number of subsequent values for that command
func (pd PathData) Cmd() (PathCmds, int) {
	iv := int32(pd)
	cmd := PathCmds(iv & 0xFF)   // only the lowest byte for cmd
	n := int((iv & 0xFF00) >> 8) // extract the n from next highest byte
	return cmd, n
}

// encode command and n into PathData
func (pc PathCmds) EncCmd(n int) PathData {
	nb := int32(n << 8) // n up-shifted
	pd := PathData(int32(pc) | nb)
	return pd
}

// 2D Path, using SVG-style data that can render just about anything
type Path struct {
	Node2DBase
	Data []PathData `xml:"d" desc:"the path data to render -- path commands and numbers are serialized, with each command specifying the number of floating-point coord data points that follow"`
}

var KiT_Path = kit.Types.AddType(&Path{}, nil)

func (g *Path) BBox2D() image.Rectangle {
	// todo -- this is somewhat expensive -- probably better to compute earlier and save?
	// psz := g.Viewport.VpBBox.Size()
	return image.Rect(0, 0, 100, 100)
	// return g.Paint.BoundingBoxFromPoints(g.Points)
}

// get the next path data element, incrementing the index -- ++ not an
// expression so its clunky -- hopefully this is inlined..
func NextPathData(data []PathData, i *int) PathData {
	pd := data[*i]
	(*i)++
	return pd
}

// this traverses the path data and renders it using paint and render state --
// we assume all the data has been validated and that n's are sufficient, etc
func RenderPathData(data []PathData, pc *Paint, rs *RenderState) {
	sz := len(data)
	if sz == 0 {
		return
	}
	var cx, cy, x1, y1, x2, y2 PathData
	for i := 0; i < sz; {
		cmd, n := NextPathData(data, &i).Cmd()
		switch cmd {
		case PcM:
			cx = NextPathData(data, &i)
			cy = NextPathData(data, &i)
			pc.MoveTo(rs, float64(cx), float64(cy))
			for np := 1; np < n/2; np++ {
				cx = NextPathData(data, &i)
				cy = NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case Pcm:
			cx += NextPathData(data, &i)
			cy += NextPathData(data, &i)
			pc.MoveTo(rs, float64(cx), float64(cy))
			for np := 1; np < n/2; np++ {
				cx += NextPathData(data, &i)
				cy += NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case PcL:
			for np := 0; np < n/2; np++ {
				cx = NextPathData(data, &i)
				cy = NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case Pcl:
			for np := 0; np < n/2; np++ {
				cx += NextPathData(data, &i)
				cy += NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case PcH:
			for np := 0; np < n; np++ {
				cx = NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case Pch:
			for np := 0; np < n; np++ {
				cx += NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case PcV:
			for np := 0; np < n; np++ {
				cy = NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case Pcv:
			for np := 0; np < n; np++ {
				cy += NextPathData(data, &i)
				pc.LineTo(rs, float64(cx), float64(cy))
			}
		case PcC:
			for np := 0; np < n/6; np++ {
				x1 = NextPathData(data, &i)
				y1 = NextPathData(data, &i)
				x2 = NextPathData(data, &i)
				y2 = NextPathData(data, &i)
				cx = NextPathData(data, &i)
				cy = NextPathData(data, &i)
				pc.CubicTo(rs, float64(x1), float64(y1), float64(x2), float64(y2), float64(cx), float64(cy))
			}
		case Pcc:
			for np := 0; np < n/6; np++ {
				x1 = cx + NextPathData(data, &i)
				y1 = cy + NextPathData(data, &i)
				x2 = cx + NextPathData(data, &i)
				y2 = cy + NextPathData(data, &i)
				cx += NextPathData(data, &i)
				cy += NextPathData(data, &i)
				pc.CubicTo(rs, float64(x1), float64(y1), float64(x2), float64(y2), float64(cx), float64(cy))
			}
		case PcS:
			for np := 0; np < n/4; np++ {
				x1 = 2*cx - x2 // this is a reflection -- todo: need special case where x2 no existe
				y1 = 2*cy - y2
				x2 = NextPathData(data, &i)
				y2 = NextPathData(data, &i)
				cx = NextPathData(data, &i)
				cy = NextPathData(data, &i)
				pc.CubicTo(rs, float64(x1), float64(y1), float64(x2), float64(y2), float64(cx), float64(cy))
			}
		case Pcs:
			for np := 0; np < n/4; np++ {
				x1 = 2*cx - x2 // this is a reflection -- todo: need special case where x2 no existe
				y1 = 2*cy - y2
				x2 = cx + NextPathData(data, &i)
				y2 = cy + NextPathData(data, &i)
				cx += NextPathData(data, &i)
				cy += NextPathData(data, &i)
				pc.CubicTo(rs, float64(x1), float64(y1), float64(x2), float64(y2), float64(cx), float64(cy))
			}
		case PcQ:
			for np := 0; np < n/4; np++ {
				x1 = NextPathData(data, &i)
				y1 = NextPathData(data, &i)
				cx = NextPathData(data, &i)
				cy = NextPathData(data, &i)
				pc.QuadraticTo(rs, float64(x1), float64(y1), float64(cx), float64(cy))
			}
		case Pcq:
			for np := 0; np < n/4; np++ {
				x1 = cx + NextPathData(data, &i)
				y1 = cy + NextPathData(data, &i)
				cx += NextPathData(data, &i)
				cy += NextPathData(data, &i)
				pc.QuadraticTo(rs, float64(x1), float64(y1), float64(cx), float64(cy))
			}
		case PcT:
			for np := 0; np < n/2; np++ {
				x1 = 2*cx - x1 // this is a reflection
				y1 = 2*cy - y1
				cx = NextPathData(data, &i)
				cy = NextPathData(data, &i)
				pc.QuadraticTo(rs, float64(x1), float64(y1), float64(cx), float64(cy))
			}
		case Pct:
			for np := 0; np < n/2; np++ {
				x1 = 2*cx - x1 // this is a reflection
				y1 = 2*cy - y1
				cx += NextPathData(data, &i)
				cy += NextPathData(data, &i)
				pc.QuadraticTo(rs, float64(x1), float64(y1), float64(cx), float64(cy))
			}
		case PcA:
			for np := 0; np < n/7; np++ {
				rx := NextPathData(data, &i)
				ry := NextPathData(data, &i)
				ang := NextPathData(data, &i)
				_ = NextPathData(data, &i) // large-arc-flag
				_ = NextPathData(data, &i) // sweep-flag
				cx = NextPathData(data, &i)
				cy = NextPathData(data, &i)
				/// https://www.w3.org/TR/SVG/paths.html#PathDataEllipticalArcCommands
				// todo: paint expresses in terms of 2 angles, SVG has these flags.. how to map?
				pc.DrawEllipticalArc(rs, float64(cx), float64(cy), float64(rx), float64(ry), float64(ang), 0)
			}
		case Pca:
			for np := 0; np < n/7; np++ {
				rx := NextPathData(data, &i)
				ry := NextPathData(data, &i)
				ang := NextPathData(data, &i)
				_ = NextPathData(data, &i) // large-arc-flag
				_ = NextPathData(data, &i) // sweep-flag
				cx += NextPathData(data, &i)
				cy += NextPathData(data, &i)
				/// https://www.w3.org/TR/SVG/paths.html#PathDataEllipticalArcCommands
				// todo: paint expresses in terms of 2 angles, SVG has these flags.. how to map?
				pc.DrawEllipticalArc(rs, float64(cx), float64(cy), float64(rx), float64(ry), float64(ang), 0)
			}
		case PcZ:
			pc.ClosePath(rs)
		case Pcz:
			pc.ClosePath(rs)
		}
	}
}

func ParsePathData(d string) []PathData {
	dt := strings.Replace(d, ",", " ", -1) // replace commas with spaces
	ds := strings.Fields(dt)               // split by whitespace
	pd := make([]PathData, 0, 20)
	sz := len(ds)
	cmd := PcM
	cmdIdx := 0 // last command index
	for i := 0; i < sz; {
		cf := ds[i]
		c := cf[0]
		mn := 0 // minimum n associated with current cmd
		switch c {
		case 'M':
			cmd = PcM
			mn = 2
		case 'm':
			cmd = Pcm
			mn = 2
		case 'L':
			cmd = PcL
			mn = 2
		case 'l':
			cmd = Pcl
			mn = 2
		case 'H':
			cmd = PcH
			mn = 1
		case 'h':
			cmd = Pch
			mn = 1
		case 'V':
			cmd = PcV
			mn = 1
		case 'v':
			cmd = Pcv
			mn = 1
		case 'C':
			cmd = PcC
			mn = 6
		case 'c':
			cmd = Pcc
			mn = 6
		case 'S':
			cmd = PcS
			mn = 4
		case 's':
			cmd = Pcs
			mn = 4
		case 'Q':
			cmd = PcQ
			mn = 4
		case 'q':
			cmd = Pcq
			mn = 4
		case 'T':
			cmd = PcT
			mn = 2
		case 't':
			cmd = Pct
			mn = 2
		case 'A':
			cmd = PcA
			mn = 7
		case 'a':
			cmd = Pca
			mn = 7
		case 'Z':
			cmd = PcZ
			mn = 0
		case 'z':
			cmd = Pcz
			mn = 0
		}
		pc := cmd.EncCmd(mn) // start with mn
		cmdIdx = len(pd)
		pd = append(pd, pc) // push on

		if mn == 0 {
			if i >= sz-1 {
				break
			}
			continue
		}

		if len(cf) > 1 {
			cf = cf[1:]
		} else {
			i++
			cf = ds[i]
		}
		vl, _ := strconv.ParseFloat(cf, 32)
		pd = append(pd, PathData(vl)) // push on

		// get rest of numbers
		for np := 1; np < mn; np++ {
			i++
			cf = ds[i]
			vl, _ := strconv.ParseFloat(cf, 32)
			pd = append(pd, PathData(vl)) // push on
		}
		if i >= sz-1 {
			break
		}

		ntot := mn
		for {
			i++
			cf = ds[i]
			if unicode.IsLetter(rune(cf[0])) {
				break
			}
			i--
			for np := 0; np < mn; np++ {
				i++
				cf = ds[i]
				vl, _ := strconv.ParseFloat(cf, 32)
				pd = append(pd, PathData(vl)) // push on
			}
			ntot += mn
			if i >= sz-1 {
				break
			}
		}
		if ntot > mn {
			pc = cmd.EncCmd(ntot)
			pd[cmdIdx] = pc
		}
	}
	return pd
}

func (g *Path) Render2D() {
	if len(g.Data) < 2 {
		return
	}
	if g.PushBounds() {
		pc := &g.Paint
		rs := &g.Viewport.Render
		rs.PushXForm(pc.XForm)
		RenderPathData(g.Data, pc, rs)
		pc.FillStrokeClear(rs)
		g.Render2DChildren()
		g.PopBounds()
		rs.PopXForm()
	}
}

func (g *Path) ReRender2D() (node Node2D, layout bool) {
	svg := g.ParentSVG()
	if svg != nil {
		node = svg
	} else {
		node = g.This.(Node2D) // no other option..
	}
	layout = false
	return
}

// check for interface implementation
var _ Node2D = &Path{}

// todo: new in SVG2: mesh
