// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gogi

import (
	"image"
	"image/color"
	"image/png"
	"io"
)

// Viewport2D provides an image and a stack of Paint contexts for drawing onto the image
// with a convenience forwarding of the Paint methods operating on the current Paint
type Viewport2D struct {
	GiNode
	ViewBox ViewBox2D `svg:"viewBox",desc:"viewbox within any parent Viewport2D"`
	Paints  []*Paint
	Pixels  *image.RGBA `desc:"pixels that we render into"`
}

// NewViewport2D creates a new image.RGBA with the specified width and height
// and prepares a context for rendering onto that image.
func NewViewport2D(width, height int) *Viewport2D {
	return NewViewport2DForRGBA(image.NewRGBA(image.Rect(0, 0, width, height)))
}

// NewViewport2DForImage copies the specified image into a new image.RGBA
// and prepares a context for rendering onto that image.
func NewViewport2DForImage(im image.Image) *Viewport2D {
	return NewViewport2DForRGBA(imageToRGBA(im))
}

// NewViewport2DForRGBA prepares a context for rendering onto the specified image.
// No copy is made.
func NewViewport2DForRGBA(im *image.RGBA) *Viewport2D {
	vp := &Viewport2D{
		ViewBox.Size.X: im.Bounds().Size().X,
		ViewBox.Size.y: im.Bounds().Size().Y,
		Pixels:         im,
	}
	vp.PushNewPaint()
}

func (vp *Viewport2D) CurPaint() *Paint {
	return Context[len(Context)-1]
}

func (vp *Viewport2D) PushNewContext() *Paint {
	c := &Paint{}
	if len(vp.Context) > 0 {
		*c = *CurPaint() // always copy current settings
	} else {
		c.Defaults()
	}
	vp.Context = append(vp.Context, c)
	return c
}

func (vp *Viewport2D) PopContext() {
	sz := len(vp.Context)
	vp.Context[sz-1] = nil
	vp.Context = vp.Context[:sz-1]
}

// SavePNG encodes the image as a PNG and writes it to disk.
func (vp *Viewport2D) SavePNG(path string) error {
	return SavePNG(path, vp.Pixels)
}

// EncodePNG encodes the image as a PNG and writes it to the provided io.Writer.
func (vp *Viewport2D) EncodePNG(w io.Writer) error {
	return png.Encode(w, vp.Pixels)
}

// Path Manipulation

// TransformPoint multiplies the specified point by the current matrix,
// returning a transformed position.
func (vp *Viewport2D2D) TransformPoint(x, y float64) Point2D {
	return Point2D{pc.XForm.TransformPoint(x, y)}
}

// MoveTo starts a new subpath within the current path starting at the
// specified point.
func (vp *Viewport2D) MoveTo(x, y float64) {
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
func (vp *Viewport2D) LineTo(x, y float64) {
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
func (vp *Viewport2D) QuadraticTo(x1, y1, x2, y2 float64) {
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
func (vp *Viewport2D) CubicTo(x1, y1, x2, y2, x3, y3 float64) {
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
func (vp *Viewport2D) ClosePath() {
	if pc.HasCurrent {
		pc.StrokePath.Add1(pc.Start.Fixed())
		pc.FillPath.Add1(pc.Start.Fixed())
		pc.Current = pc.Start
	}
}

// ClearPath clears the current path. There is no current point after this
// operation.
func (vp *Viewport2D) ClearPath() {
	pc.StrokePath.Clear()
	pc.FillPath.Clear()
	pc.HasCurrent = false
}

// NewSubPath starts a new subpath within the current path. There is no current
// point after this operation.
func (vp *Viewport2D) NewSubPath() {
	if pc.HasCurrent {
		pc.FillPath.Add1(pc.Start.Fixed())
	}
	pc.HasCurrent = false
}

// StrokePreserve strokes the current path with the current color, line width,
// line cap, line join and dash settings. The path is preserved after this
// operation.
func (vp *Viewport2D) StrokePreserve() {
	pc := vp.CurPaint()
	painter := newPatternPainter(vp.Pixels, vp.Mask, vp.Stroke.Pattern)
	vp.stroke(painter)
}

// Stroke strokes the current path with the current color, line width,
// line cap, line join and dash settings. The path is cleared after this
// operation.
func (vp *Viewport2D) Stroke() {
	vp.StrokePreserve()
	vp.ClearPath()
}

// FillPreserve fills the current path with the current color. Open subpaths
// are implicity closed. The path is preserved after this operation.
func (vp *Viewport2D) FillPreserve() {
	painter := newPatternPainter(vp.Pixels, vp.Mask, vp.fillPattern)
	vp.fill(painter)
}

// Fill fills the current path with the current color. Open subpaths
// are implicity closed. The path is cleared after this operation.
func (vp *Viewport2D) Fill() {
	vp.FillPreserve()
	vp.ClearPath()
}

// ClipPreserve updates the clipping region by intersecting the current
// clipping region with the current path as it would be filled by vp.Fill().
// The path is preserved after this operation.
func (vp *Viewport2D) ClipPreserve() {
	clip := image.NewAlpha(image.Rect(0, 0, vp.ViewBox.Size.X, vp.ViewBox.Size.Y))
	painter := raster.NewAlphaOverPainter(clip)
	vp.fill(painter)
	if vp.Mask == nil {
		vp.Mask = clip
	} else {
		mask := image.NewAlpha(image.Rect(0, 0, vp.ViewBox.Size.X, vp.ViewBox.Size.Y))
		draw.DrawMask(mask, mask.Bounds(), clip, image.ZP, vp.Mask, image.ZP, draw.Over)
		vp.Mask = mask
	}
}

// SetMask allows you to directly set the *image.Alpha to be used as a clipping
// mask. It must be the same size as the context, else an error is returned
// and the mask is unchanged.
func (vp *Viewport2D) SetMask(mask *image.Alpha) error {
	if mask.Bounds().Size() != vp.Pixels.Bounds().Size() {
		return errors.New("mask size must match context size")
	}
	vp.Mask = mask
	return nil
}

// AsMask returns an *image.Alpha representing the alpha channel of this
// context. This can be useful for advanced clipping operations where you first
// render the mask geometry and then use it as a mask.
func (vp *Viewport2D) AsMask() *image.Alpha {
	mask := image.NewAlpha(vp.Pixels.Bounds())
	draw.Draw(mask, vp.Pixels.Bounds(), vp.Pixels, image.ZP, draw.Src)
	return mask
}

// Clip updates the clipping region by intersecting the current
// clipping region with the current path as it would be filled by vp.Fill().
// The path is cleared after this operation.
func (vp *Viewport2D) Clip() {
	vp.ClipPreserve()
	vp.ClearPath()
}

// ResetClip clears the clipping region.
func (vp *Viewport2D) ResetClip() {
	vp.Mask = nil
}

// Convenient Drawing Functions

// Clear fills the entire image with the current color.
func (vp *Viewport2D) Clear() {
	src := image.NewUniform(vp.color)
	draw.Draw(vp.Pixels, vp.Pixels.Bounds(), src, image.ZP, draw.Src)
}

// SetPixel sets the color of the specified pixel using the current color.
func (vp *Viewport2D) SetPixel(x, y int) {
	vp.Pixels.Set(x, y, vp.color)
}

// DrawPoint is like DrawCircle but ensures that a circle of the specified
// size is drawn regardless of the current transformation matrix. The position
// is still transformed, but not the shape of the point.
func (vp *Viewport2D) DrawPoint(x, y, r float64) {
	vp.Push()
	tx, ty := vp.TransformPoint(x, y)
	vp.Identity()
	vp.DrawCircle(tx, ty, r)
	vp.Pop()
}

func (vp *Viewport2D) DrawLine(x1, y1, x2, y2 float64) {
	vp.MoveTo(x1, y1)
	vp.LineTo(x2, y2)
}

func (vp *Viewport2D) DrawRectangle(x, y, w, h float64) {
	vp.NewSubPath()
	vp.MoveTo(x, y)
	vp.LineTo(x+w, y)
	vp.LineTo(x+w, y+h)
	vp.LineTo(x, y+h)
	vp.ClosePath()
}

func (vp *Viewport2D) DrawRoundedRectangle(x, y, w, h, r float64) {
	x0, x1, x2, x3 := x, x+r, x+w-r, x+w
	y0, y1, y2, y3 := y, y+r, y+h-r, y+h
	vp.NewSubPath()
	vp.MoveTo(x1, y0)
	vp.LineTo(x2, y0)
	vp.DrawArc(x2, y1, r, Radians(270), Radians(360))
	vp.LineTo(x3, y2)
	vp.DrawArc(x2, y2, r, Radians(0), Radians(90))
	vp.LineTo(x1, y3)
	vp.DrawArc(x1, y2, r, Radians(90), Radians(180))
	vp.LineTo(x0, y1)
	vp.DrawArc(x1, y1, r, Radians(180), Radians(270))
	vp.ClosePath()
}

func (vp *Viewport2D) DrawEllipticalArc(x, y, rx, ry, angle1, angle2 float64) {
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
		if i == 0 && !vp.HasCurrent {
			vp.MoveTo(x0, y0)
		}
		vp.QuadraticTo(cx, cy, x2, y2)
	}
}

func (vp *Viewport2D) DrawEllipse(x, y, rx, ry float64) {
	vp.NewSubPath()
	vp.DrawEllipticalArc(x, y, rx, ry, 0, 2*math.Pi)
	vp.ClosePath()
}

func (vp *Viewport2D) DrawArc(x, y, r, angle1, angle2 float64) {
	vp.DrawEllipticalArc(x, y, r, r, angle1, angle2)
}

func (vp *Viewport2D) DrawCircle(x, y, r float64) {
	vp.NewSubPath()
	vp.DrawEllipticalArc(x, y, r, r, 0, 2*math.Pi)
	vp.ClosePath()
}

func (vp *Viewport2D) DrawRegularPolygon(n int, x, y, r, rotation float64) {
	angle := 2 * math.Pi / float64(n)
	rotation -= math.Pi / 2
	if n%2 == 0 {
		rotation += angle / 2
	}
	vp.NewSubPath()
	for i := 0; i < n; i++ {
		a := rotation + angle*float64(i)
		vp.LineTo(x+r*math.Cos(a), y+r*math.Sin(a))
	}
	vp.ClosePath()
}

// DrawImage draws the specified image at the specified point.
func (vp *Viewport2D) DrawImage(im image.Image, x, y int) {
	vp.DrawImageAnchored(im, x, y, 0, 0)
}

// DrawImageAnchored draws the specified image at the specified anchor point.
// The anchor point is x - w * ax, y - h * ay, where w, h is the size of the
// image. Use ax=0.5, ay=0.5 to center the image at the specified point.
func (vp *Viewport2D) DrawImageAnchored(im image.Image, x, y int, ax, ay float64) {
	s := im.Bounds().Size()
	x -= int(ax * float64(s.X))
	y -= int(ay * float64(s.Y))
	transformer := draw.BiLinear
	fx, fy := float64(x), float64(y)
	m := vp.matrix.Translate(fx, fy)
	s2d := f64.Aff3{m.XX, m.XY, m.X0, m.YX, m.YY, m.Y0}
	if vp.Mask == nil {
		transformer.Transform(vp.Pixels, s2d, im, im.Bounds(), draw.Over, nil)
	} else {
		transformer.Transform(vp.Pixels, s2d, im, im.Bounds(), draw.Over, &draw.Options{
			DstMask:  vp.Mask,
			DstMaskP: image.ZP,
		})
	}
}

// Text Functions

func (vp *Viewport2D) SetFontFace(fontFace font.Face) {
	vp.fontFace = fontFace
	vp.fontHeight = float64(fontFace.Metrics().Height) / 64
}

func (vp *Viewport2D) LoadFontFace(path string, points float64) error {
	face, err := LoadFontFace(path, points)
	if err == nil {
		vp.fontFace = face
		vp.fontHeight = points * 72 / 96
	}
	return err
}

func (vp *Viewport2D) FontHeight() float64 {
	return vp.fontHeight
}

func (vp *Viewport2D) drawString(im *image.RGBA, s string, x, y float64) {
	d := &font.Drawer{
		Dst:  im,
		Src:  image.NewUniform(vp.color),
		Face: vp.fontFace,
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
		m := vp.matrix.Translate(fx, fy)
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
func (vp *Viewport2D) DrawString(s string, x, y float64) {
	vp.DrawStringAnchored(s, x, y, 0, 0)
}

// DrawStringAnchored draws the specified text at the specified anchor point.
// The anchor point is x - w * ax, y - h * ay, where w, h is the size of the
// text. Use ax=0.5, ay=0.5 to center the text at the specified point.
func (vp *Viewport2D) DrawStringAnchored(s string, x, y, ax, ay float64) {
	w, h := vp.MeasureString(s)
	x -= ax * w
	y += ay * h
	if vp.Mask == nil {
		vp.drawString(vp.Pixels, s, x, y)
	} else {
		im := image.NewRGBA(image.Rect(0, 0, vp.ViewBox.Size.X, vp.ViewBox.Size.Y))
		vp.drawString(im, s, x, y)
		draw.DrawMask(vp.Pixels, vp.Pixels.Bounds(), im, image.ZP, vp.Mask, image.ZP, draw.Over)
	}
}

// DrawStringWrapped word-wraps the specified string to the given max width
// and then draws it at the specified anchor point using the given line
// spacing and text alignment.
func (vp *Viewport2D) DrawStringWrapped(s string, x, y, ax, ay, width, lineSpacing float64, align Align) {
	lines := vp.WordWrap(s, width)
	h := float64(len(lines)) * vp.fontHeight * lineSpacing
	h -= (lineSpacing - 1) * vp.fontHeight
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
		vp.DrawStringAnchored(line, x, y, ax, ay)
		y += vp.fontHeight * lineSpacing
	}
}

// MeasureString returns the rendered width and height of the specified text
// given the current font face.
func (vp *Viewport2D) MeasureString(s string) (w, h float64) {
	d := &font.Drawer{
		Face: vp.fontFace,
	}
	a := d.MeasureString(s)
	return float64(a >> 6), vp.fontHeight
}

// WordWrap wraps the specified string to the given max width and current
// font face.
func (vp *Viewport2D) WordWrap(s string, w float64) []string {
	return wordWrap(vp, s, w)
}

// Transformation Matrix Operations

// Identity resets the current transformation matrix to the identity matrix.
// This results in no translating, scaling, rotating, or shearing.
func (vp *Viewport2D) Identity() {
	vp.matrix = Identity()
}

// Translate updates the current matrix with a translation.
func (vp *Viewport2D) Translate(x, y float64) {
	vp.matrix = vp.matrix.Translate(x, y)
}

// Scale updates the current matrix with a scaling factor.
// Scaling occurs about the origin.
func (vp *Viewport2D) Scale(x, y float64) {
	vp.matrix = vp.matrix.Scale(x, y)
}

// ScaleAbout updates the current matrix with a scaling factor.
// Scaling occurs about the specified point.
func (vp *Viewport2D) ScaleAbout(sx, sy, x, y float64) {
	vp.Translate(x, y)
	vp.Scale(sx, sy)
	vp.Translate(-x, -y)
}

// Rotate updates the current matrix with a clockwise rotation.
// Rotation occurs about the origin. Angle is specified in radians.
func (vp *Viewport2D) Rotate(angle float64) {
	vp.matrix = vp.matrix.Rotate(angle)
}

// RotateAbout updates the current matrix with a clockwise rotation.
// Rotation occurs about the specified point. Angle is specified in radians.
func (vp *Viewport2D) RotateAbout(angle, x, y float64) {
	vp.Translate(x, y)
	vp.Rotate(angle)
	vp.Translate(-x, -y)
}

// Shear updates the current matrix with a shearing angle.
// Shearing occurs about the origin.
func (vp *Viewport2D) Shear(x, y float64) {
	vp.matrix = vp.matrix.Shear(x, y)
}

// ShearAbout updates the current matrix with a shearing angle.
// Shearing occurs about the specified point.
func (vp *Viewport2D) ShearAbout(sx, sy, x, y float64) {
	vp.Translate(x, y)
	vp.Shear(sx, sy)
	vp.Translate(-x, -y)
}

// InvertY flips the Y axis so that Y grows from bottom to top and Y=0 is at
// the bottom of the image.
func (vp *Viewport2D) InvertY() {
	vp.Translate(0, float64(vp.ViewBox.Size.Y))
	vp.Scale(1, -1)
}
