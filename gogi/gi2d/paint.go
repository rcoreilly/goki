// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
	"github.com/golang/freetype/raster"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/f64"
	"image"
	"image/color"
)

/*
This is mostly just restructured version of: https://github.com/fogleman/gg

Copyright (C) 2016 Michael Fogleman

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// The Paint object provides the full context (parameters) and functions for
// painting onto an image -- image is always passed as an argument so it can be
// applied to anything
type Paint struct {
	Stroke     PaintStroke
	Fill       PaintFill
	Font       PaintFont
	StrokePath raster.Path
	FillPath   raster.Path
	Start      Point2D
	Current    Point2D
	HasCurrent bool
	XForm      XFormMatrix2D
	Mask       *image.Alpha
}

func (p *Paint) Defaults() {
	p.Stroke.Defaults()
	p.Fill.Defaults()
	p.Font.Defaults()
	p.XForm.Identity()
}

// Path Manipulation

// TransformPoint multiplies the specified point by the current transform matrix,
// returning a transformed position.
func (pc *Paint2D) TransformPoint(x, y float64) Point2D {
	return Point2D{pc.XForm.TransformPoint(x, y)}
}

// MoveTo starts a new subpath within the current path starting at the
// specified point.
func (pc *Paint) MoveTo(x, y float64) {
	if pc.HasCurrent {
		pc.fillPath.Add1(pc.start.Fixed())
	}
	p := pc.TransformPoint(x, y)
	pc.StrokePath.Start(p.Fixed())
	pc.FillPath.Start(p.Fixed())
	pc.Start = p
	pc.Current = p
	pc.HasCurrent = true
}

// LineTo adds a line segment to the current path starting at the current
// point. If there is no current point, it is equivalent to MoveTo(x, y)
func (pc *Paint) LineTo(x, y float64) {
	if !pc.HasCurrent {
		pc.MoveTo(x, y)
	} else {
		p := pc.TransformPoint(x, y)
		pc.StrokePath.Add1(p.Fixed())
		pc.FillPath.Add1(p.Fixed())
		pc.Current = p
	}
}

// QuadraticTo adds a quadratic bezier curve to the current path starting at
// the current point. If there is no current point, it first performs
// MoveTo(x1, y1)
func (pc *Paint) QuadraticTo(x1, y1, x2, y2 float64) {
	if !pc.HasCurrent {
		pc.MoveTo(x1, y1)
	}
	p1 := pc.TransformPoint(x1, y1)
	p2 := pc.TransformPoint(x2, y2)
	pc.StrokePath.Add2(p1.Fixed(), p2.Fixed())
	pc.FillPath.Add2(p1.Fixed(), p2.Fixed())
	pc.Current = p2
}

// CubicTo adds a cubic bezier curve to the current path starting at the
// current point. If there is no current point, it first performs
// MoveTo(x1, y1). Because freetype/raster does not support cubic beziers,
// this is emulated with many small line segments.
func (pc *Paint) CubicTo(x1, y1, x2, y2, x3, y3 float64) {
	if !pc.HasCurrent {
		pc.MoveTo(x1, y1)
	}
	x0, y0 := pc.Current.X, pc.Current.Y
	x1, y1 = pc.XForm.TransformPoint(x1, y1)
	x2, y2 = pc.XForm.TransformPoint(x2, y2)
	x3, y3 = pc.XForm.TransformPoint(x3, y3)
	points := CubicBezier(x0, y0, x1, y1, x2, y2, x3, y3)
	previous := pc.Current.Fixed()
	for _, p := range points[1:] {
		f := p.Fixed()
		if f == previous {
			// TODO: this fixes some rendering issues but not all
			continue
		}
		previous = f
		pc.StrokePath.Add1(f)
		pc.FillPath.Add1(f)
		pc.Current = p
	}
}

// ClosePath adds a line segment from the current point to the beginning
// of the current subpath. If there is no current point, this is a no-op.
func (pc *Paint) ClosePath() {
	if pc.HasCurrent {
		pc.StrokePath.Add1(pc.Start.Fixed())
		pc.FillPath.Add1(pc.Start.Fixed())
		pc.Current = pc.Start
	}
}

// ClearPath clears the current path. There is no current point after this
// operation.
func (pc *Paint) ClearPath() {
	pc.StrokePath.Clear()
	pc.FillPath.Clear()
	pc.HasCurrent = false
}

// NewSubPath starts a new subpath within the current path. There is no current
// point after this operation.
func (pc *Paint) NewSubPath() {
	if pc.HasCurrent {
		pc.FillPath.Add1(pc.Start.Fixed())
	}
	pc.HasCurrent = false
}

// Path Drawing

func (pc *Paint) capper() raster.Capper {
	switch pc.lineCap {
	case LineCapButt:
		return raster.ButtCapper
	case LineCapRound:
		return raster.Rounvpapper
	case LineCapSquare:
		return raster.SquareCapper
	}
	return nil
}

func (pc *Paint) joiner() raster.Joiner {
	switch pc.lineJoin {
	case LineJoinBevel:
		return raster.BevelJoiner
	case LineJoinRound:
		return raster.RoundJoiner
	}
	return nil
}

func (pc *Paint) stroke(painter raster.Painter) {
	pc := pc.CurContext()
	path := pc.StrokePath
	if len(pc.Stroke.Dashes) > 0 {
		path = dashed(path, pc.Stroke.Dashes)
	} else {
		// TODO: this is a temporary workaround to remove tiny segments
		// that result in rendering issues
		path = rasterPath(flattenPath(path))
	}
	r := raster.NewRasterizer(pc.ViewBox.Size.X, pc.ViewBox.Size.Y)
	r.UseNonZeroWinding = true
	r.AddStroke(path, fix(pc.lineWidth), pc.capper(), pc.joiner())
	r.Rasterize(painter)
}

func (pc *Paint) fill(painter raster.Painter) {
	pc := pc.CurContext()
	path := pc.FillPath
	if pc.HasCurrent {
		path = make(raster.Path, len(pc.FillPath))
		copy(path, pc.FillPath)
		path.Add1(pc.Start.Fixed())
	}
	r := raster.NewRasterizer(pc.ViewBox.Size.X, pc.ViewBox.Size.Y)
	r.UseNonZeroWinding = (pc.Fill.FillRule == FillRuleNonZero)
	r.AddPath(path)
	r.Rasterize(painter)
}

// StrokePreserve strokes the current path with the current color, line width,
// line cap, line join and dash settings. The path is preserved after this
// operation.
func (pc *Paint) StrokePreserve() {
	pc := pc.CurContext()
	painter := newPatternPainter(pc.Pixels, pc.Mask, pc.Stroke.Pattern)
	pc.stroke(painter)
}

// Stroke strokes the current path with the current color, line width,
// line cap, line join and dash settings. The path is cleared after this
// operation.
func (pc *Paint) Stroke() {
	pc.StrokePreserve()
	pc.ClearPath()
}

// FillPreserve fills the current path with the current color. Open subpaths
// are implicity closed. The path is preserved after this operation.
func (pc *Paint) FillPreserve() {
	painter := newPatternPainter(pc.Pixels, pc.Mask, pc.fillPattern)
	pc.fill(painter)
}

// Fill fills the current path with the current color. Open subpaths
// are implicity closed. The path is cleared after this operation.
func (pc *Paint) Fill() {
	pc.FillPreserve()
	pc.ClearPath()
}

// ClipPreserve updates the clipping region by intersecting the current
// clipping region with the current path as it would be filled by pc.Fill().
// The path is preserved after this operation.
func (pc *Paint) ClipPreserve() {
	clip := image.NewAlpha(image.Rect(0, 0, pc.ViewBox.Size.X, pc.ViewBox.Size.Y))
	painter := raster.NewAlphaOverPainter(clip)
	pc.fill(painter)
	if pc.Mask == nil {
		pc.Mask = clip
	} else {
		mask := image.NewAlpha(image.Rect(0, 0, pc.ViewBox.Size.X, pc.ViewBox.Size.Y))
		draw.DrawMask(mask, mask.Bounds(), clip, image.ZP, pc.Mask, image.ZP, draw.Over)
		pc.Mask = mask
	}
}

// SetMask allows you to directly set the *image.Alpha to be used as a clipping
// mask. It must be the same size as the context, else an error is returned
// and the mask is unchanged.
func (pc *Paint) SetMask(mask *image.Alpha) error {
	if mask.Bounds().Size() != pc.Pixels.Bounds().Size() {
		return errors.New("mask size must match context size")
	}
	pc.Mask = mask
	return nil
}

// AsMask returns an *image.Alpha representing the alpha channel of this
// context. This can be useful for advanced clipping operations where you first
// render the mask geometry and then use it as a mask.
func (pc *Paint) AsMask() *image.Alpha {
	mask := image.NewAlpha(pc.Pixels.Bounds())
	draw.Draw(mask, pc.Pixels.Bounds(), pc.Pixels, image.ZP, draw.Src)
	return mask
}

// Clip updates the clipping region by intersecting the current
// clipping region with the current path as it would be filled by pc.Fill().
// The path is cleared after this operation.
func (pc *Paint) Clip() {
	pc.ClipPreserve()
	pc.ClearPath()
}

// ResetClip clears the clipping region.
func (pc *Paint) ResetClip() {
	pc.Mask = nil
}

// Convenient Drawing Functions

// Clear fills the entire image with the current color.
func (pc *Paint) Clear() {
	src := image.NewUniform(pc.color)
	draw.Draw(pc.Pixels, pc.Pixels.Bounds(), src, image.ZP, draw.Src)
}

// SetPixel sets the color of the specified pixel using the current color.
func (pc *Paint) SetPixel(x, y int) {
	pc.Pixels.Set(x, y, pc.color)
}

// DrawPoint is like DrawCircle but ensures that a circle of the specified
// size is drawn regardless of the current transformation matrix. The position
// is still transformed, but not the shape of the point.
func (pc *Paint) DrawPoint(x, y, r float64) {
	pc.Push()
	tx, ty := pc.TransformPoint(x, y)
	pc.Identity()
	pc.DrawCircle(tx, ty, r)
	pc.Pop()
}

func (pc *Paint) DrawLine(x1, y1, x2, y2 float64) {
	pc.MoveTo(x1, y1)
	pc.LineTo(x2, y2)
}

func (pc *Paint) DrawRectangle(x, y, w, h float64) {
	pc.NewSubPath()
	pc.MoveTo(x, y)
	pc.LineTo(x+w, y)
	pc.LineTo(x+w, y+h)
	pc.LineTo(x, y+h)
	pc.ClosePath()
}

func (pc *Paint) DrawRoundedRectangle(x, y, w, h, r float64) {
	x0, x1, x2, x3 := x, x+r, x+w-r, x+w
	y0, y1, y2, y3 := y, y+r, y+h-r, y+h
	pc.NewSubPath()
	pc.MoveTo(x1, y0)
	pc.LineTo(x2, y0)
	pc.DrawArc(x2, y1, r, Radians(270), Radians(360))
	pc.LineTo(x3, y2)
	pc.DrawArc(x2, y2, r, Radians(0), Radians(90))
	pc.LineTo(x1, y3)
	pc.DrawArc(x1, y2, r, Radians(90), Radians(180))
	pc.LineTo(x0, y1)
	pc.DrawArc(x1, y1, r, Radians(180), Radians(270))
	pc.ClosePath()
}

func (pc *Paint) DrawEllipticalArc(x, y, rx, ry, angle1, angle2 float64) {
	const n = 16
	for i := 0; i < n; i++ {
		p1 := float64(i+0) / n
		p2 := float64(i+1) / n
		a1 := angle1 + (angle2-angle1)*p1
		a2 := angle1 + (angle2-angle1)*p2
		x0 := x + rx*math.Cos(a1)
		y0 := y + ry*math.Sin(a1)
		x1 := x + rx*math.Cos(a1+(a2-a1)/2)
		y1 := y + ry*math.Sin(a1+(a2-a1)/2)
		x2 := x + rx*math.Cos(a2)
		y2 := y + ry*math.Sin(a2)
		cx := 2*x1 - x0/2 - x2/2
		cy := 2*y1 - y0/2 - y2/2
		if i == 0 && !pc.HasCurrent {
			pc.MoveTo(x0, y0)
		}
		pc.QuadraticTo(cx, cy, x2, y2)
	}
}

func (pc *Paint) DrawEllipse(x, y, rx, ry float64) {
	pc.NewSubPath()
	pc.DrawEllipticalArc(x, y, rx, ry, 0, 2*math.Pi)
	pc.ClosePath()
}

func (pc *Paint) DrawArc(x, y, r, angle1, angle2 float64) {
	pc.DrawEllipticalArc(x, y, r, r, angle1, angle2)
}

func (pc *Paint) DrawCircle(x, y, r float64) {
	pc.NewSubPath()
	pc.DrawEllipticalArc(x, y, r, r, 0, 2*math.Pi)
	pc.ClosePath()
}

func (pc *Paint) DrawRegularPolygon(n int, x, y, r, rotation float64) {
	angle := 2 * math.Pi / float64(n)
	rotation -= math.Pi / 2
	if n%2 == 0 {
		rotation += angle / 2
	}
	pc.NewSubPath()
	for i := 0; i < n; i++ {
		a := rotation + angle*float64(i)
		pc.LineTo(x+r*math.Cos(a), y+r*math.Sin(a))
	}
	pc.ClosePath()
}

// DrawImage draws the specified image at the specified point.
func (pc *Paint) DrawImage(im image.Image, x, y int) {
	pc.DrawImageAnchored(im, x, y, 0, 0)
}

// DrawImageAnchored draws the specified image at the specified anchor point.
// The anchor point is x - w * ax, y - h * ay, where w, h is the size of the
// image. Use ax=0.5, ay=0.5 to center the image at the specified point.
func (pc *Paint) DrawImageAnchored(im image.Image, x, y int, ax, ay float64) {
	s := im.Bounds().Size()
	x -= int(ax * float64(s.X))
	y -= int(ay * float64(s.Y))
	transformer := draw.BiLinear
	fx, fy := float64(x), float64(y)
	m := pc.XForm.Translate(fx, fy)
	s2d := f64.Aff3{m.XX, m.XY, m.X0, m.YX, m.YY, m.Y0}
	if pc.Mask == nil {
		transformer.Transform(pc.Pixels, s2d, im, im.Bounds(), draw.Over, nil)
	} else {
		transformer.Transform(pc.Pixels, s2d, im, im.Bounds(), draw.Over, &draw.Options{
			DstMask:  pc.Mask,
			DstMaskP: image.ZP,
		})
	}
}

// Text Functions

func (pc *Paint) SetFontFace(fontFace font.Face) {
	pc.fontFace = fontFace
	pc.fontHeight = float64(fontFace.Metrics().Height) / 64
}

func (pc *Paint) LoadFontFace(path string, points float64) error {
	face, err := LoadFontFace(path, points)
	if err == nil {
		pc.fontFace = face
		pc.fontHeight = points * 72 / 96
	}
	return err
}

func (pc *Paint) FontHeight() float64 {
	return pc.fontHeight
}

func (pc *Paint) drawString(im *image.RGBA, s string, x, y float64) {
	d := &font.Drawer{
		Dst:  im,
		Src:  image.NewUniform(pc.color),
		Face: pc.fontFace,
		Dot:  fixp(x, y),
	}
	// based on Drawer.DrawString() in golang.org/x/image/font/font.go
	prevC := rune(-1)
	for _, c := range s {
		if prevC >= 0 {
			d.Dot.X += d.Face.Kern(prevC, c)
		}
		dr, mask, maskp, advance, ok := d.Face.Glyph(d.Dot, c)
		if !ok {
			// TODO: is falling back on the U+FFFD glyph the responsibility of
			// the Drawer or the Face?
			// TODO: set prevC = '\ufffd'?
			continue
		}
		sr := dr.Sub(dr.Min)
		transformer := draw.BiLinear
		fx, fy := float64(dr.Min.X), float64(dr.Min.Y)
		m := pc.XForm.Translate(fx, fy)
		s2d := f64.Aff3{m.XX, m.XY, m.X0, m.YX, m.YY, m.Y0}
		transformer.Transform(d.Dst, s2d, d.Src, sr, draw.Over, &draw.Options{
			SrcMask:  mask,
			SrcMaskP: maskp,
		})
		d.Dot.X += advance
		prevC = c
	}
}

// DrawString draws the specified text at the specified point.
func (pc *Paint) DrawString(s string, x, y float64) {
	pc.DrawStringAnchored(s, x, y, 0, 0)
}

// DrawStringAnchored draws the specified text at the specified anchor point.
// The anchor point is x - w * ax, y - h * ay, where w, h is the size of the
// text. Use ax=0.5, ay=0.5 to center the text at the specified point.
func (pc *Paint) DrawStringAnchored(s string, x, y, ax, ay float64) {
	w, h := pc.MeasureString(s)
	x -= ax * w
	y += ay * h
	if pc.Mask == nil {
		pc.drawString(pc.Pixels, s, x, y)
	} else {
		im := image.NewRGBA(image.Rect(0, 0, pc.ViewBox.Size.X, pc.ViewBox.Size.Y))
		pc.drawString(im, s, x, y)
		draw.DrawMask(pc.Pixels, pc.Pixels.Bounds(), im, image.ZP, pc.Mask, image.ZP, draw.Over)
	}
}

// DrawStringWrapped word-wraps the specified string to the given max width
// and then draws it at the specified anchor point using the given line
// spacing and text alignment.
func (pc *Paint) DrawStringWrapped(s string, x, y, ax, ay, width, lineSpacing float64, align Align) {
	lines := pc.WordWrap(s, width)
	h := float64(len(lines)) * pc.fontHeight * lineSpacing
	h -= (lineSpacing - 1) * pc.fontHeight
	x -= ax * width
	y -= ay * h
	switch align {
	case AlignLeft:
		ax = 0
	case AlignCenter:
		ax = 0.5
		x += width / 2
	case AlignRight:
		ax = 1
		x += width
	}
	ay = 1
	for _, line := range lines {
		pc.DrawStringAnchored(line, x, y, ax, ay)
		y += pc.fontHeight * lineSpacing
	}
}

// MeasureString returns the rendered width and height of the specified text
// given the current font face.
func (pc *Paint) MeasureString(s string) (w, h float64) {
	d := &font.Drawer{
		Face: pc.fontFace,
	}
	a := d.MeasureString(s)
	return float64(a >> 6), pc.fontHeight
}

// WordWrap wraps the specified string to the given max width and current
// font face.
func (pc *Paint) WordWrap(s string, w float64) []string {
	return wordWrap(vp, s, w)
}

// Transformation Matrix Operations

// Identity resets the current transformation matrix to the identity matrix.
// This results in no translating, scaling, rotating, or shearing.
func (pc *Paint) Identity() {
	pc.XForm = Identity()
}

// Translate updates the current matrix with a translation.
func (pc *Paint) Translate(x, y float64) {
	pc.XForm = pc.XForm.Translate(x, y)
}

// Scale updates the current matrix with a scaling factor.
// Scaling occurs about the origin.
func (pc *Paint) Scale(x, y float64) {
	pc.XForm = pc.XForm.Scale(x, y)
}

// ScaleAbout updates the current matrix with a scaling factor.
// Scaling occurs about the specified point.
func (pc *Paint) ScaleAbout(sx, sy, x, y float64) {
	pc.Translate(x, y)
	pc.Scale(sx, sy)
	pc.Translate(-x, -y)
}

// Rotate updates the current matrix with a clockwise rotation.
// Rotation occurs about the origin. Angle is specified in radians.
func (pc *Paint) Rotate(angle float64) {
	pc.XForm = pc.XForm.Rotate(angle)
}

// RotateAbout updates the current matrix with a clockwise rotation.
// Rotation occurs about the specified point. Angle is specified in radians.
func (pc *Paint) RotateAbout(angle, x, y float64) {
	pc.Translate(x, y)
	pc.Rotate(angle)
	pc.Translate(-x, -y)
}

// Shear updates the current matrix with a shearing angle.
// Shearing occurs about the origin.
func (pc *Paint) Shear(x, y float64) {
	pc.XForm = pc.XForm.Shear(x, y)
}

// ShearAbout updates the current matrix with a shearing angle.
// Shearing occurs about the specified point.
func (pc *Paint) ShearAbout(sx, sy, x, y float64) {
	pc.Translate(x, y)
	pc.Shear(sx, sy)
	pc.Translate(-x, -y)
}

// InvertY flips the Y axis so that Y grows from bottom to top and Y=0 is at
// the bottom of the image.
func (pc *Paint) InvertY() {
	pc.Translate(0, float64(pc.ViewBox.Size.Y))
	pc.Scale(1, -1)
}
