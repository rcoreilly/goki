// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	// "fmt"
	"github.com/rcoreilly/goki/gi/units"
	"github.com/rcoreilly/goki/ki"
	"image"
	"math"
)

// this is based on QtQuick layouts https://doc.qt.io/qt-5/qtquicklayouts-overview.html  https://doc.qt.io/qt-5/qml-qtquick-layouts-layout.html

// horizontal alignment type -- how to align items in the horizontal dimension
type AlignHoriz int32

const (
	AlignLeft AlignHoriz = iota
	AlignHCenter
	AlignRight
	AlignHJustify
	AlignHorizN
)

//go:generate stringer -type=AlignHoriz

var KiT_AlignHoriz = ki.Enums.AddEnumAltLower(AlignLeft, false, nil, "Align", int64(AlignHorizN))

// vertical alignment type -- how to align items in the vertical dimension -- must correspond with horizontal for layout
type AlignVert int32

const (
	AlignTop AlignVert = iota
	AlignVCenter
	AlignBottom
	AlignVJustify
	AlignBaseline
	AlignVertN
)

var KiT_AlignVert = ki.Enums.AddEnumAltLower(AlignTop, false, nil, "Align", int64(AlignVertN))

//go:generate stringer -type=AlignVert

// overflow type -- determines what happens when there is too much stuff in a layout
type Overflow int32

const (
	OverflowAuto Overflow = iota
	OverflowScroll
	OverflowVisible
	OverflowHidden
	OverflowN
)

var KiT_Overflow = ki.Enums.AddEnumAltLower(OverflowAuto, false, nil, "Overflow", int64(OverflowN))

//go:generate stringer -type=Overflow

// todo: for style
// Align = layouts
// Content -- enum of various options
// Items -- similar enum -- combine
// Self "
// Flex -- flexbox -- https://www.w3schools.com/css/css3_flexbox.asp -- key to look at further for layout ideas
// as is Position -- absolute, sticky, etc
// Resize: user-resizability
// vertical-align
// z-index

// style preferences on the layout of the element
type LayoutStyle struct {
	z_index        int           `xml:"z-index",desc:"ordering factor for rendering depth -- lower numbers rendered first -- sort children according to this factor"`
	AlignH         AlignHoriz    `xml:"align-horiz",desc:"horizontal alignment -- for widget layouts -- not a standard css property"`
	AlignV         AlignVert     `xml:"align-vert",desc:"vertical alignment -- for widget layouts -- not a standard css property"`
	PosX           units.Value   `xml:"x",desc:"horizontal position -- often superceded by layout but otherwise used"`
	PosY           units.Value   `xml:"y",desc:"vertical position -- often superceded by layout but otherwise used"`
	Width          units.Value   `xml:"width",desc:"specified size of element -- 0 if not specified"`
	Height         units.Value   `xml:"height",desc:"specified size of element -- 0 if not specified"`
	MaxWidth       units.Value   `xml:"max-width",desc:"specified maximum size of element -- 0  means just use other values, negative means stretch"`
	MaxHeight      units.Value   `xml:"max-height",desc:"specified maximum size of element -- 0 means just use other values, negative means stretch"`
	MinWidth       units.Value   `xml:"min-width",desc:"specified mimimum size of element -- 0 if not specified"`
	MinHeight      units.Value   `xml:"min-height",desc:"specified mimimum size of element -- 0 if not specified"`
	Offsets        []units.Value `xml:"{top,right,bottom,left}",desc:"specified offsets for each side"`
	Margin         units.Value   `xml:"margin",desc:"outer-most transparent space around box element -- todo: can be specified per side"`
	Overflow       Overflow      `xml:"overflow",desc:"what to do with content that overflows -- default is Auto add of scrollbars as needed -- todo: can have separate -x -y values"`
	ScrollBarWidth units.Value   `xml:"scrollbar-width",desc:"width of a layout scrollbar"`
}

func (ls *LayoutStyle) Defaults() {
	ls.MinWidth.Set(1.0, units.Em)
	ls.MinHeight.Set(1.0, units.Em)
	ls.Width.Set(1.0, units.Em)
	ls.Height.Set(1.0, units.Em)
	ls.ScrollBarWidth.Set(20.0, units.Px)
}

func (ls *LayoutStyle) SetStylePost() {
}

// return the alignment for given dimension, using horiz terminology (top = left, etc)
func (ls *LayoutStyle) AlignDim(d Dims2D) AlignHoriz {
	switch d {
	case X:
		return ls.AlignH
	default:
		return AlignHoriz(ls.AlignV)
	}
}

////////////////////////////////////////////////////////////////////////////////////////
// Layout Data for actually computing the layout

// size preferences
type SizePrefs struct {
	Need Vec2D `desc:"minimum size needed -- set to at least computed allocsize"`
	Pref Vec2D `desc:"preferred size -- start here for layout"`
	Max  Vec2D `desc:"maximum size -- will not be greater than this -- 0 = no constraint, neg = stretch"`
}

// return true if Max < 0 meaning can stretch infinitely along given dimension
func (sp SizePrefs) HasMaxStretch(d Dims2D) bool {
	return (sp.Max.Dim(d) < 0.0)
}

// return true if Pref > Need meaning can stretch more along given dimension
func (sp SizePrefs) CanStretchNeed(d Dims2D) bool {
	return (sp.Pref.Dim(d) > sp.Need.Dim(d))
}

// 2D margins
type Margins struct {
	left, right, top, bottom float64
}

// set a single margin for all items
func (m *Margins) SetMargin(marg float64) {
	m.left = marg
	m.right = marg
	m.top = marg
	m.bottom = marg
}

// LayoutData contains all the data needed to specify the layout of an item within a layout -- includes computed values of style prefs -- everything is concrete and specified here, whereas style may not be fully resolved
type LayoutData struct {
	Size         SizePrefs   `desc:"size constraints for this item -- from layout style"`
	Margins      Margins     `desc:"margins around this item"`
	GridPos      image.Point `desc:"position within a grid"`
	GridSpan     image.Point `desc:"number of grid elements that we take up in each direction"`
	AllocPos     Vec2D       `desc:"allocated relative position of this item, by the parent layout"`
	AllocSize    Vec2D       `desc:"allocated size of this item, by the parent layout"`
	AllocPosOrig Vec2D       `desc:"original copy of allocated relative position of this item, by the parent layout -- need for scrolling which can update AllocPos"`
}

func (ld *LayoutData) Defaults() {
	if ld.GridSpan.X < 1 {
		ld.GridSpan.X = 1
	}
	if ld.GridSpan.Y < 1 {
		ld.GridSpan.Y = 1
	}
}

func (ld *LayoutData) SetFromStyle(ls *LayoutStyle) {
	ld.Reset()
	// these are layout hints:
	ld.Size.Need = Vec2D{ls.MinWidth.Dots, ls.MinHeight.Dots}
	ld.Size.Pref = Vec2D{ls.Width.Dots, ls.Height.Dots}
	ld.Size.Max = Vec2D{ls.MaxWidth.Dots, ls.MaxHeight.Dots}

	// this is an actual initial desired setting
	ld.AllocPos = Vec2D{ls.PosX.Dots, ls.PosY.Dots}
	// not setting size, so we can keep that as a separate constraint
}

// called at start of layout process -- resets all values back to 0
func (ld *LayoutData) Reset() {
	ld.AllocPos = Vec2DZero
	ld.AllocPosOrig = Vec2DZero
	ld.AllocSize = Vec2DZero
}

// update our sizes based on AllocSize and Max constraints, etc
func (ld *LayoutData) UpdateSizes() {
	ld.Size.Need.SetMax(ld.AllocSize)   // min cannot be < alloc -- bare min
	ld.Size.Pref.SetMax(ld.Size.Need)   // pref cannot be < min
	ld.Size.Need.SetMinPos(ld.Size.Max) // min cannot be > max
	ld.Size.Pref.SetMinPos(ld.Size.Max) // pref cannot be > max
}

////////////////////////////////////////////////////////////////////////////////////////
//    Layout handles all major types of layout

// different types of layouts
type Layouts int32

const (
	// arrange items horizontally across a row
	LayoutRow Layouts = iota
	// arrange items vertically in a column
	LayoutCol
	// arrange items according to a grid
	LayoutGrid
	// arrange items horizontally across a row, overflowing vertically as needed
	LayoutRowFlow
	// arrange items vertically within a column, overflowing horizontally as needed
	LayoutColFlow
	// arrange items stacked on top of each other -- Top index indicates which to show -- overall size accommodates largest in each dimension
	LayoutStacked
	LayoutsN
)

//go:generate stringer -type=Layouts

// note: Layout cannot be a Widget type because Controls in Widget is a Layout..

// Layout is the primary node type responsible for organizing the sizes and
// positions of child widgets -- all arbitrary collections of widgets should
// generally be contained within a layout -- otherwise the parent widget must
// take over responsibility for positioning.  The alignment is NOT inherited
// by default so must be specified per child, except that the parent alignment
// is used within the relevant dimension (e.g., align-horiz for a LayoutRow
// layout, to determine left, right, center, justified).  Layouts
// can automatically add scrollbars depending on the Overflow layout style
type Layout struct {
	Node2DBase
	Lay       Layouts    `xml:"lay",desc:"type of layout to use"`
	StackTop  ki.Ptr     `desc:"pointer to node to use as the top of the stack -- only node matching this pointer is rendered, even if this is nil"`
	ChildSize Vec2D      `xml:"-",desc:"total max size of children as laid out"`
	HScroll   *ScrollBar `xml:"-",desc:"horizontal scroll bar -- we fully manage this as needed"`
	VScroll   *ScrollBar `xml:"-",desc:"vertical scroll bar -- we fully manage this as needed"`
}

// must register all new types so type names can be looked up by name -- e.g., for json
var KiT_Layout = ki.Types.AddType(&Layout{}, nil)

// do we sum up elements along given dimension?  else max
func (ly *Layout) SumDim(d Dims2D) bool {
	if (d == X && ly.Lay == LayoutRow) || (d == Y && ly.Lay == LayoutCol) {
		return true
	}
	return false
}

// first depth-first pass: terminal concrete items compute their AllocSize
// we focus on Need: Max(Min, AllocSize), and Want: Max(Pref, AllocSize) -- Max is
// only used if we need to fill space, during final allocation
//
// second me-first pass: each layout allocates AllocSize for its children based on
// aggregated size data, and so on down the tree

// how much extra size we need in each dimension
func (ly *Layout) ExtraSize() Vec2D {
	lst := &ly.Style.Layout
	var es Vec2D
	es.SetVal(2.0 * lst.Margin.Dots)
	if ly.HScroll != nil {
		es.Y += lst.ScrollBarWidth.Dots
	}
	if ly.VScroll != nil {
		es.X += lst.ScrollBarWidth.Dots
	}
	return es
}

// first pass: gather the size information from the children
func (ly *Layout) GatherSizes() {
	if len(ly.Children) == 0 {
		return
	}

	var sumPref, sumNeed, maxPref, maxNeed Vec2D
	for _, c := range ly.Children {
		_, gi := KiToNode2D(c)
		if gi == nil {
			continue
		}
		gi.LayData.UpdateSizes()
		sumNeed = sumNeed.Add(gi.LayData.Size.Need)
		sumPref = sumPref.Add(gi.LayData.Size.Pref)
		maxNeed = maxNeed.Max(gi.LayData.Size.Need)
		maxPref = maxPref.Max(gi.LayData.Size.Pref)
	}

	for d := X; d <= Y; d++ {
		if ly.SumDim(d) { // our layout now updated to sum
			ly.LayData.Size.Need.SetMaxDim(d, sumNeed.Dim(d))
			ly.LayData.Size.Pref.SetMaxDim(d, sumPref.Dim(d))
		} else { // use max for other dir
			ly.LayData.Size.Need.SetMaxDim(d, maxNeed.Dim(d))
			ly.LayData.Size.Pref.SetMaxDim(d, maxPref.Dim(d))
		}
	}

	es := ly.ExtraSize()
	ly.LayData.Size.Need.SetAdd(es)
	ly.LayData.Size.Pref.SetAdd(es)

	// todo: something entirely different needed for grids..

	ly.LayData.UpdateSizes() // enforce max and normal ordering, etc

	// todo: here we need to also deal with -1 max stretch to give full alloc
	// in the "Max" dim if poss -- right now it is cutting that down based on
	// Pref size of childs.
}

// in case we don't have any explicit allocsize set for us -- go up parents
// until we find one -- typically a viewport
func (ly *Layout) AllocFromParent() {
	if !ly.LayData.AllocSize.IsZero() {
		return
	}

	// todo: take into account position within parent size??

	ly.FunUpParent(0, ly.This, func(k ki.Ki, level int, d interface{}) bool {
		_, pg := KiToNode2D(k)
		if pg == nil {
			return false
		}
		if !pg.LayData.AllocSize.IsZero() {
			ly.LayData.AllocSize = pg.LayData.AllocSize
			// fmt.Printf("layout got parent alloc: %v from %v\n", ly.LayData.AllocSize,
			// 	pg.Name)
			return false
		}
		return true
	})
}

// calculations to layout a single-element dimension, returns pos and size
func (ly *Layout) LayoutSingleImpl(avail, need, pref, max float64, al AlignHoriz) (pos, size float64) {
	usePref := true
	targ := pref
	extra := avail - targ
	if extra < -0.1 { // not fitting in pref, go with min
		usePref = false
		targ = need
		extra = avail - targ
	}
	extra = math.Max(extra, 0.0) // no negatives

	stretchNeed := false // stretch relative to need
	stretchMax := false  // only stretch Max = neg

	if usePref && extra >= 0.0 { // have some stretch extra
		if max < 0.0 {
			stretchMax = true // only stretch those marked as infinitely stretchy
		}
	} else if extra >= 0.0 { // extra relative to Need
		stretchNeed = true // stretch relative to need
	}

	pos = 0.0
	size = need
	if usePref {
		size = pref
	}
	if stretchMax || stretchNeed {
		size += extra
	} else {
		switch al {
		case AlignLeft:
			pos = 0.0
		case AlignHCenter:
			pos = 0.5 * extra
		case AlignRight:
			pos = extra
		case AlignHJustify: // treat justify as stretch!
			pos = 0.0
			size += extra
		}
	}
	return
}

// layout item in single-dimensional case -- e.g., orthogonal dimension from LayoutRow / Col
func (ly *Layout) LayoutSingle(dim Dims2D) {
	es := ly.ExtraSize()
	avail := ly.LayData.AllocSize.Dim(dim) - es.Dim(dim)
	for _, c := range ly.Children {
		_, gi := KiToNode2D(c)
		if gi == nil {
			continue
		}
		al := gi.Style.Layout.AlignDim(dim)
		pref := gi.LayData.Size.Pref.Dim(dim)
		need := gi.LayData.Size.Need.Dim(dim)
		max := gi.LayData.Size.Max.Dim(dim)
		pos, size := ly.LayoutSingleImpl(avail, need, pref, max, al)
		gi.LayData.AllocSize.SetDim(dim, size)
		gi.LayData.AllocPos.SetDim(dim, pos)
	}
}

// layout all children along given dim -- only affects that dim -- e.g., use
// LayoutSingle for other dim
func (ly *Layout) LayoutAll(dim Dims2D) {
	sz := len(ly.Children)
	if sz == 0 {
		return
	}

	al := ly.Style.Layout.AlignDim(dim)
	es := ly.ExtraSize()
	marg := es.Dim(dim)
	avail := ly.LayData.AllocSize.Dim(dim) - marg
	pref := ly.LayData.Size.Pref.Dim(dim) - marg
	need := ly.LayData.Size.Need.Dim(dim) - marg

	targ := pref
	usePref := true
	extra := avail - targ
	if extra < -0.1 { // not fitting in pref, go with need
		usePref = false
		targ = need
		extra = avail - targ
	}
	extra = math.Max(extra, 0.0) // no negatives

	nstretch := 0
	stretchTot := 0.0
	stretchNeed := false         // stretch relative to need
	stretchMax := false          // only stretch Max = neg
	addSpace := false            // apply extra toward spacing -- for justify
	if usePref && extra >= 0.0 { // have some stretch extra
		for _, c := range ly.Children {
			_, gi := KiToNode2D(c)
			if gi == nil {
				continue
			}
			if gi.LayData.Size.HasMaxStretch(dim) { // negative = stretch
				nstretch++
				stretchTot += gi.LayData.Size.Pref.Dim(dim)
			}
		}
		if nstretch > 0 {
			stretchMax = true // only stretch those marked as infinitely stretchy
		}
	} else if extra >= 0.0 { // extra relative to Need
		for _, c := range ly.Children {
			_, gi := KiToNode2D(c)
			if gi == nil {
				continue
			}
			if gi.LayData.Size.HasMaxStretch(dim) || gi.LayData.Size.CanStretchNeed(dim) {
				nstretch++
				stretchTot += gi.LayData.Size.Pref.Dim(dim)
			}
		}
		if nstretch > 0 {
			stretchNeed = true // stretch relative to need
		}
	}

	extraSpace := 0.0
	if sz > 1 && extra > 0.0 && al == AlignHJustify && !stretchNeed && !stretchMax {
		addSpace = true
		// if neither, then just distribute as spacing for justify
		extraSpace = extra / float64(sz-1)
	}

	// now arrange everyone
	pos := ly.Style.Layout.Margin.Dots

	// todo: need a direction setting too
	if al == AlignRight && !stretchNeed && !stretchMax {
		pos = extra
	}

	for i, c := range ly.Children {
		_, gi := KiToNode2D(c)
		if gi == nil {
			continue
		}
		size := gi.LayData.Size.Need.Dim(dim)
		if usePref {
			size = gi.LayData.Size.Pref.Dim(dim)
		}
		if stretchMax { // negative = stretch
			if gi.LayData.Size.HasMaxStretch(dim) { // in proportion to pref
				size += extra * (gi.LayData.Size.Pref.Dim(dim) / stretchTot)
			}
		} else if stretchNeed {
			if gi.LayData.Size.HasMaxStretch(dim) || gi.LayData.Size.CanStretchNeed(dim) {
				size += extra * (gi.LayData.Size.Pref.Dim(dim) / stretchTot)
			}
		} else if addSpace { // implies align justify
			if i > 0 {
				pos += extraSpace
			}
		}

		gi.LayData.AllocSize.SetDim(dim, size)
		gi.LayData.AllocPos.SetDim(dim, pos)
		pos += size
	}
}

// final pass through children to finalize the layout, capturing original
// positions and computing summary size stats
func (ly *Layout) FinalizeLayout() {
	ly.ChildSize = Vec2DZero
	for _, c := range ly.Children {
		_, gi := KiToNode2D(c)
		if gi == nil {
			continue
		}
		gi.LayData.AllocPosOrig = gi.LayData.AllocPos
		ly.ChildSize.SetMax(gi.LayData.AllocPos.Add(gi.LayData.AllocSize))
	}
}

func (ly *Layout) SetHScroll() {
	if ly.HScroll == nil {
		ly.HScroll = &ScrollBar{}
		ly.HScroll.SetThisName(ly.HScroll, "Lay_HScroll")
		ly.HScroll.SetParent(ly.This)
		ly.HScroll.Horiz = true
		ly.HScroll.InitNode2D()
		ly.HScroll.Defaults()
	}
	sc := ly.HScroll
	sc.SetFixedHeight(units.NewValue(ly.Style.Layout.ScrollBarWidth.Dots, units.Px))
	sc.SetFixedWidth(units.NewValue(ly.LayData.AllocSize.X, units.Px))
	sc.Style2D()
	sc.Min = 0.0
	sc.Max = ly.ChildSize.X
	sc.Step = ly.Style.Font.Size.Dots // step by lines
	sc.PageStep = 10.0 * sc.Step      // todo: more dynamic
	sc.ThumbVal = ly.LayData.AllocSize.X
	// fmt.Printf("Setting up HScroll: max: %v  thumb: %v sz: %v\n", sc.Max, sc.ThumbVal, sc.ThumbSize)
	sc.Tracking = true
	sc.SliderSig.Connect(ly.This, func(rec, send ki.Ki, sig int64, data interface{}) {
		// ss, _ := send.(*ScrollBar)
		// fmt.Printf("HScroll to %v\n", ss.Value)
		li, _ := KiToNode2D(rec) // note: avoid using closures
		ls := li.AsLayout2D()
		ls.UpdateStart()
		ls.UpdateEnd()
	})
}

func (ly *Layout) LayoutScrolls() {
	sw := ly.Style.Layout.ScrollBarWidth.Dots
	if ly.HScroll != nil {
		sc := ly.HScroll
		sc.Layout2D(0)
		sc.LayData.AllocPos.X = ly.LayData.AllocPos.X
		sc.LayData.AllocPos.Y = ly.LayData.AllocPos.Y + ly.LayData.AllocSize.Y - sw
		sc.LayData.AllocPosOrig = sc.LayData.AllocPos
		sc.LayData.AllocSize.X = ly.LayData.AllocSize.X
		sc.LayData.AllocSize.Y = sw
		sc.Layout2D(1)
	}
	if ly.VScroll != nil {
		sc := ly.VScroll
		sc.Layout2D(0)
		sc.LayData.AllocPos.X = ly.LayData.AllocPos.X + ly.LayData.AllocSize.X - sw
		sc.LayData.AllocPos.Y = ly.LayData.AllocPos.Y
		sc.LayData.AllocPosOrig = sc.LayData.AllocPos
		sc.LayData.AllocSize.Y = ly.LayData.AllocSize.Y
		sc.LayData.AllocSize.X = sw
		sc.Layout2D(1)
	}
}

func (ly *Layout) RenderScrolls() {
	if ly.HScroll != nil {
		ly.HScroll.Render2D()
	}
	if ly.VScroll != nil {
		ly.VScroll.Render2D()
	}
}

func (ly *Layout) DeleteHScroll() {
	if ly.HScroll == nil {
		return
	}
	// todo: disconnect from events, call pointer cut function on ki
	ly.HScroll = nil
}

func (ly *Layout) SetVScroll() {
	if ly.VScroll == nil {
		ly.VScroll = &ScrollBar{}
		ly.VScroll.SetThisName(ly.VScroll, "Lay_VScroll")
		ly.VScroll.SetParent(ly.This)
		ly.VScroll.InitNode2D()
		ly.VScroll.Defaults()
	}
	sc := ly.VScroll
	sc.SetFixedWidth(units.NewValue(ly.Style.Layout.ScrollBarWidth.Dots, units.Px))
	sc.SetFixedHeight(units.NewValue(ly.LayData.AllocSize.Y, units.Px))
	sc.Style2D()
	sc.Min = 0.0
	sc.Max = ly.ChildSize.Y
	sc.Step = ly.Style.Font.Size.Dots // step by lines
	sc.PageStep = 10.0 * sc.Step      // todo: more dynamic
	sc.ThumbVal = ly.LayData.AllocSize.Y
	sc.Tracking = true
	sc.SliderSig.Connect(ly.This, func(rec, send ki.Ki, sig int64, data interface{}) {
		// ss, _ := send.(*ScrollBar)
		// fmt.Printf("VScroll to %v\n", ss.Value)
		li, _ := KiToNode2D(rec) // note: avoid using closures
		ls := li.AsLayout2D()
		ls.UpdateStart()
		ls.UpdateEnd()
	})
}

func (ly *Layout) DeleteVScroll() {
	if ly.VScroll == nil {
		return
	}
	// todo: disconnect from events, call pointer cut function on ki
	ly.VScroll = nil
}

func (ly *Layout) ManageOverflow() {
	lst := &ly.Style.Layout
	if lst.Overflow == OverflowVisible {
		return
	}
	if ly.ChildSize.X > ly.LayData.AllocSize.X { // overflowing
		if lst.Overflow != OverflowHidden {
			ly.SetHScroll()
		}
	} else {
		ly.DeleteHScroll()
	}
	if ly.ChildSize.Y > ly.LayData.AllocSize.Y { // overflowing
		if lst.Overflow != OverflowHidden {
			ly.SetVScroll()
		}
	} else {
		ly.DeleteVScroll()
	}
}

// all child nodes call this during their Render2DCheck() -- if we return
// false, child is not rendered -- we can also update the AllocPos of the
// child based on scrolling
func (ly *Layout) RenderChild(gi *Node2DBase) bool {
	// always display our scrollbars!
	if ly.HScroll != nil && gi.This == ly.HScroll.This {
		return true
	}
	if ly.VScroll != nil && gi.This == ly.VScroll.This {
		return true
	}
	if ly.Lay == LayoutStacked && ly.StackTop.Ptr != gi.This {
		return false
	}
	gi.LayData.AllocPos = gi.LayData.AllocPosOrig
	if ly.HScroll != nil {
		off := ly.HScroll.Value
		gi.LayData.AllocPos.X -= off
	}
	if ly.VScroll != nil {
		off := ly.VScroll.Value
		gi.LayData.AllocPos.Y -= off
	}
	gi.GeomFromLayout()
	if !gi.WinBBox.Overlaps(ly.WinBBox) { // out of view
		return false
	}
	// todo: need appropriate clipping at this point!  proably put our bbox in
	// the laydata of all the children, and they clip using that?
	return true
}

// convenience for LayoutStacked to show child node at a given index
func (ly *Layout) ShowChildAtIndex(idx int) error {
	ch, err := ly.KiChild(idx)
	if err != nil {
		return err
	}
	ly.StackTop.Ptr = ch
	return nil
}

///////////////////////////////////////////////////
//   Standard Node2D interface

func (ly *Layout) AsNode2D() *Node2DBase {
	return &ly.Node2DBase
}

func (ly *Layout) AsViewport2D() *Viewport2D {
	return nil
}

func (ly *Layout) AsLayout2D() *Layout {
	return ly
}

func (ly *Layout) InitNode2D() {
	ly.InitNode2DBase()
}

func (ly *Layout) Node2DBBox() image.Rectangle {
	return ly.WinBBoxFromAlloc()
}

func (ly *Layout) Style2D() {
	ly.Style2DWidget()
}

// need multiple iterations?
func (ly *Layout) Layout2D(iter int) {

	if iter == 0 {
		ly.InitLayout2D()
		ly.GatherSizes()
	} else {
		ly.AllocFromParent() // in case we didn't get anything
		switch ly.Lay {
		case LayoutRow:
			ly.LayoutAll(X)
			ly.LayoutSingle(Y)
		case LayoutCol:
			ly.LayoutAll(Y)
			ly.LayoutSingle(X)
		case LayoutStacked:
			ly.LayoutSingle(X)
			ly.LayoutSingle(Y)
		}
		ly.FinalizeLayout()
		ly.GeomFromLayout()
		ly.ManageOverflow()
		ly.LayoutScrolls()
	}
	// todo: test if this is needed -- if there are any el-relative settings anyway
	ly.Style.SetUnitContext(&ly.Viewport.Render, 0)
}

func (ly *Layout) Render2D() {
	ly.RenderScrolls()
}

func (ly *Layout) CanReRender2D() bool {
	return true
}

func (ly *Layout) FocusChanged2D(gotFocus bool) {
}

// check for interface implementation
var _ Node2D = &Layout{}

///////////////////////////////////////////////////////////
//    Frame -- generic container that is also a Layout

// Frame is a basic container for widgets -- a layout that renders the
// standard box model
type Frame struct {
	Layout
}

// must register all new types so type names can be looked up by name -- e.g., for json
var KiT_Frame = ki.Types.AddType(&Frame{}, nil)

func (g *Frame) AsNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *Frame) AsViewport2D() *Viewport2D {
	return nil
}

func (g *Frame) AsLayout2D() *Layout {
	return &g.Layout
}

func (g *Frame) InitNode2D() {
	g.InitNode2DBase()
}

var FrameProps = map[string]interface{}{
	"border-width":     "1px",
	"border-radius":    "0px",
	"border-color":     "black",
	"border-style":     "solid",
	"padding":          "2px",
	"margin":           "2px",
	"color":            "black",
	"background-color": "#FFF",
}

func (g *Frame) Style2D() {
	// first do our normal default styles
	g.Style.SetStyle(nil, &StyleDefault, FrameProps)
	// then style with user props
	g.Style2DWidget()
}

func (g *Frame) Layout2D(iter int) {
	g.Layout.Layout2D(iter) // use the layout version
}

func (g *Frame) Node2DBBox() image.Rectangle {
	return g.WinBBoxFromAlloc()
}

func (g *Frame) Render2D() {
	pc := &g.Paint
	st := &g.Style
	rs := &g.Viewport.Render
	pc.StrokeStyle.SetColor(&st.Border.Color)
	pc.StrokeStyle.Width = st.Border.Width
	pc.FillStyle.SetColor(&st.Background.Color)
	pos := g.LayData.AllocPos.AddVal(st.Layout.Margin.Dots).SubVal(st.Border.Width.Dots)
	sz := g.LayData.AllocSize.SubVal(2.0 * st.Layout.Margin.Dots).AddVal(2.0 * st.Border.Width.Dots)
	// pos := g.LayData.AllocPos
	// sz := g.LayData.AllocSize
	rad := st.Border.Radius.Dots
	if rad == 0.0 {
		pc.DrawRectangle(rs, pos.X, pos.Y, sz.X, sz.Y)
	} else {
		pc.DrawRoundedRectangle(rs, pos.X, pos.Y, sz.X, sz.Y, rad)
	}
	pc.FillStrokeClear(rs)

	g.Layout.Render2D()
}

func (g *Frame) CanReRender2D() bool {
	return true
}

func (g *Frame) FocusChanged2D(gotFocus bool) {
}

// check for interface implementation
var _ Node2D = &Frame{}

///////////////////////////////////////////////////////////
//    Stretch and Space -- dummy elements for layouts

// Stretch adds an infinitely stretchy element for spacing out layouts
// (max-size = -1) set the width / height property to determine how much it
// takes relative to other stretchy elements
type Stretch struct {
	Node2DBase
}

// must register all new types so type names can be looked up by name -- e.g., for json
var KiT_Stretch = ki.Types.AddType(&Stretch{}, nil)

func (g *Stretch) AsNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *Stretch) AsViewport2D() *Viewport2D {
	return nil
}

func (g *Stretch) AsLayout2D() *Layout {
	return nil
}

func (g *Stretch) InitNode2D() {
	g.InitNode2DBase()
}

var StretchProps = map[string]interface{}{
	"max-width":  -1.0,
	"max-height": -1.0,
}

func (g *Stretch) Style2D() {
	// first do our normal default styles
	g.Style.SetStyle(nil, &StyleDefault, StretchProps)
	// then style with user props
	g.Style2DWidget()
}

func (g *Stretch) Layout2D(iter int) {
	g.BaseLayout2D(iter)
}

func (g *Stretch) Node2DBBox() image.Rectangle {
	return g.WinBBoxFromAlloc()
}

func (g *Stretch) Render2D() {
}

func (g *Stretch) CanReRender2D() bool {
	return true
}

func (g *Stretch) FocusChanged2D(gotFocus bool) {
}

// check for interface implementation
var _ Node2D = &Stretch{}

// Space adds an infinitely stretchy element for spacing out layouts
// (max-size = -1) set the width / height property to determine how much it
// takes relative to other stretchy elements
type Space struct {
	Node2DBase
}

// must register all new types so type names can be looked up by name -- e.g., for json
var KiT_Space = ki.Types.AddType(&Space{}, nil)

func (g *Space) AsNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *Space) AsViewport2D() *Viewport2D {
	return nil
}

func (g *Space) AsLayout2D() *Layout {
	return nil
}

func (g *Space) InitNode2D() {
	g.InitNode2DBase()
}

func (g *Space) Style2D() {
	// // first do our normal default styles
	// g.Style.SetStyle(nil, &StyleDefault, SpaceProps)
	// then style with user props
	g.Style2DWidget()
}

func (g *Space) Layout2D(iter int) {
	g.BaseLayout2D(iter)
}

func (g *Space) Node2DBBox() image.Rectangle {
	return g.WinBBoxFromAlloc()
}

func (g *Space) Render2D() {
}

func (g *Space) CanReRender2D() bool {
	return true
}

func (g *Space) FocusChanged2D(gotFocus bool) {
}

// check for interface implementation
var _ Node2D = &Space{}
