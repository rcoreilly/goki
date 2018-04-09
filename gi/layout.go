// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/rcoreilly/goki/gi/units"
	"github.com/rcoreilly/goki/ki"
	"github.com/rcoreilly/goki/ki/kit"
)

// all different types of alignment -- only some are applicable to different
// contexts, but there is also so much overlap that it makes sense to have
// them all in one list -- some are not standard CSS and used by layout
type Align int32

const (
	AlignLeft Align = iota
	AlignTop
	AlignCenter
	// middle = vertical version of center
	AlignMiddle
	AlignRight
	AlignBottom
	AlignBaseline
	// same as CSS space-between
	AlignJustify
	AlignSpaceAround
	AlignFlexStart
	AlignFlexEnd
	AlignTextTop
	AlignTextBottom
	// align to subscript
	AlignSub
	// align to superscript
	AlignSuper
	AlignN
)

//go:generate stringer -type=Align

var KiT_Align = kit.Enums.AddEnumAltLower(AlignN, false, nil, "Align")

// is this a generalized alignment to start of container?
func IsAlignStart(a Align) bool {
	return (a == AlignLeft || a == AlignTop || a == AlignFlexStart || a == AlignTextTop)
}

// is this a generalized alignment to middle of container?
func IsAlignMiddle(a Align) bool {
	return (a == AlignCenter || a == AlignMiddle)
}

// is this a generalized alignment to end of container?
func IsAlignEnd(a Align) bool {
	return (a == AlignRight || a == AlignBottom || a == AlignFlexEnd || a == AlignTextBottom)
}

// overflow type -- determines what happens when there is too much stuff in a layout
type Overflow int32

const (
	// automatically add scrollbars as needed -- this is pretty much the only sensible option, and is the default here, but Visible is default in html
	OverflowAuto Overflow = iota
	// pretty much the same as auto -- we treat it as such
	OverflowScroll
	// make the overflow visible -- this is generally unsafe and not very feasible and will be ignored as long as possible -- currently falls back on auto, but could go to Hidden if that works better overall
	OverflowVisible
	// hide the overflow and don't present scrollbars (supported)
	OverflowHidden
	OverflowN
)

var KiT_Overflow = kit.Enums.AddEnumAltLower(OverflowN, false, nil, "Overflow")

//go:generate stringer -type=Overflow

// todo: for style
// Align = layouts
// Flex -- flexbox -- https://www.w3schools.com/css/css3_flexbox.asp -- key to look at further for layout ideas
// as is Position -- absolute, sticky, etc
// Resize: user-resizability
// z-index

// CSS vs. Layout alignment
//
// CSS has align-self, align-items (for a container, provides a default for
// items) and align-content which only applies to lines in a flex layout (akin
// to a flow layout) -- there is a presumed horizontal aspect to these, except
// align-content, so they are subsumed in the AlignH parameter in this style.
// Vertical-align works as expected, and Text.Align uses left/center/right
//
// LayoutRow, Col both allow explicit Top/Left Center/Middle, Right/Bottom alignment
// along with Justify and SpaceAround -- they use IsAlign functions

// style preferences on the layout of the element
type LayoutStyle struct {
	z_index   int           `xml:"z-index" desc:"ordering factor for rendering depth -- lower numbers rendered first -- sort children according to this factor"`
	AlignH    Align         `xml:"align-self" alt:"horiz-align,align-horiz" desc:"horizontal alignment -- for widget layouts -- not a standard css property"`
	AlignV    Align         `xml:"vertical-align" alt:"vert-align,align-vert" desc:"vertical alignment -- for widget layouts -- not a standard css property"`
	PosX      units.Value   `xml:"x" desc:"horizontal position -- often superceded by layout but otherwise used"`
	PosY      units.Value   `xml:"y" desc:"vertical position -- often superceded by layout but otherwise used"`
	Width     units.Value   `xml:"width" desc:"specified size of element -- 0 if not specified"`
	Height    units.Value   `xml:"height" desc:"specified size of element -- 0 if not specified"`
	MaxWidth  units.Value   `xml:"max-width" desc:"specified maximum size of element -- 0  means just use other values, negative means stretch"`
	MaxHeight units.Value   `xml:"max-height" desc:"specified maximum size of element -- 0 means just use other values, negative means stretch"`
	MinWidth  units.Value   `xml:"min-width" desc:"specified mimimum size of element -- 0 if not specified"`
	MinHeight units.Value   `xml:"min-height" desc:"specified mimimum size of element -- 0 if not specified"`
	Offsets   []units.Value `xml:"{top,right,bottom,left}" desc:"specified offsets for each side"`
	Margin    units.Value   `xml:"margin" desc:"outer-most transparent space around box element -- todo: can be specified per side"`
	Padding   units.Value   `xml:"padding" desc:"transparent space around central content of box -- todo: if 4 values it is top, right, bottom, left; 3 is top, right&left, bottom; 2 is top & bottom, right and left"`
	Overflow  Overflow      `xml:"overflow" desc:"what to do with content that overflows -- default is Auto add of scrollbars as needed -- todo: can have separate -x -y values"`
	Columns   int           `xml:"columns" alt:"grid-cols" desc:"number of columns to use in a grid layout -- used as a constraint in layout if individual elements do not specify their row, column positions"`
	Row       int           `xml:"row" desc:"specifies the row that this element should appear within a grid layout"`
	Col       int           `xml:"col" desc:"specifies the column that this element should appear within a grid layout"`
	RowSpan   int           `xml:"row-span" desc:"specifies the number of sequential rows that this element should occupy within a grid layout (todo: not currently supported)"`
	ColSpan   int           `xml:"col-span" desc:"specifies the number of sequential columns that this element should occupy within a grid layout"`

	ScrollBarWidth units.Value `xml:"scrollbar-width" desc:"width of a layout scrollbar"`
}

func (ls *LayoutStyle) Defaults() {
	ls.MinWidth.Set(2.0, units.Px)
	ls.MinHeight.Set(2.0, units.Px)
	ls.ScrollBarWidth.Set(16.0, units.Px)
}

func (ls *LayoutStyle) SetStylePost() {
}

// return the alignment for given dimension
func (ls *LayoutStyle) AlignDim(d Dims2D) Align {
	switch d {
	case X:
		return ls.AlignH
	default:
		return ls.AlignV
	}
}

// position settings, in dots
func (ls *LayoutStyle) PosDots() Vec2D {
	return NewVec2D(ls.PosX.Dots, ls.PosY.Dots)
}

// size settings, in dots
func (ls *LayoutStyle) SizeDots() Vec2D {
	return NewVec2D(ls.Width.Dots, ls.Height.Dots)
}

// size max settings, in dots
func (ls *LayoutStyle) MaxSizeDots() Vec2D {
	return NewVec2D(ls.MaxWidth.Dots, ls.MaxHeight.Dots)
}

// size min settings, in dots
func (ls *LayoutStyle) MinSizeDots() Vec2D {
	return NewVec2D(ls.MinWidth.Dots, ls.MinHeight.Dots)
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
	ld.Size.Need = ls.MinSizeDots()
	ld.Size.Pref = ls.SizeDots()
	ld.Size.Max = ls.MaxSizeDots()

	// this is an actual initial desired setting
	ld.AllocPos = ls.PosDots()
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
	Lay          Layouts      `xml:"lay" desc:"type of layout to use"`
	StackTop     ki.Ptr       `desc:"pointer to node to use as the top of the stack -- only node matching this pointer is rendered, even if this is nil"`
	ChildSize    Vec2D        `xml:"-" desc:"total max size of children as laid out"`
	ExtraSize    Vec2D        `xml:"-" desc:"extra size in each dim due to scrollbars we add"`
	HasHScroll   bool         `desc:"horizontal scrollbar is used, at bottom of layout"`
	HasVScroll   bool         `desc:"vertical scrollbar is used, at right of layout"`
	HScroll      *ScrollBar   `xml:"-" desc:"horizontal scroll bar -- we fully manage this as needed"`
	VScroll      *ScrollBar   `xml:"-" desc:"vertical scroll bar -- we fully manage this as needed"`
	GridSize     image.Point  `desc:"computed size of a grid layout based on all the constraints -- computed during Size2D pass"`
	GridDataRows []LayoutData `json:"-" xml:"-" desc:"grid data for rows"`
	GridDataCols []LayoutData `json:"-" xml:"-" desc:"grid data for cols"`
}

var KiT_Layout = kit.Types.AddType(&Layout{}, nil)

// do we sum up elements along given dimension?  else max
func (ly *Layout) SumDim(d Dims2D) bool {
	if (d == X && ly.Lay == LayoutRow) || (d == Y && ly.Lay == LayoutCol) {
		return true
	}
	return false
}

// first depth-first Size2D pass: terminal concrete items compute their AllocSize
// we focus on Need: Max(Min, AllocSize), and Want: Max(Pref, AllocSize) -- Max is
// only used if we need to fill space, during final allocation
//
// second me-first Layout2D pass: each layout allocates AllocSize for its
// children based on aggregated size data, and so on down the tree

// first pass: gather the size information from the children
func (ly *Layout) GatherSizes() {
	if len(ly.Kids) == 0 {
		return
	}

	var sumPref, sumNeed, maxPref, maxNeed Vec2D
	for _, c := range ly.Kids {
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

	spc := ly.Style.BoxSpace()
	ly.LayData.Size.Need.SetAddVal(2.0 * spc)
	ly.LayData.Size.Pref.SetAddVal(2.0 * spc)

	// todo: something entirely different needed for grids..

	ly.LayData.UpdateSizes() // enforce max and normal ordering, etc
	if Layout2DTrace {
		fmt.Printf("Size:   %v gather sizes need: %v, pref: %v\n", ly.PathUnique(), ly.LayData.Size.Need, ly.LayData.Size.Pref)
	}
}

// first pass: gather the size information from the children, grid version
func (ly *Layout) GatherSizesGrid() {
	if len(ly.Kids) == 0 {
		return
	}

	cols := ly.Style.Layout.Columns
	rows := 0

	sz := len(ly.Kids)
	// collect overal size
	for _, c := range ly.Kids {
		_, gi := KiToNode2D(c)
		if gi == nil {
			continue
		}
		lst := gi.Style.Layout
		if lst.Col > 0 {
			cols = kit.MaxInt(cols, lst.Col+lst.ColSpan)
		}
		if lst.Row > 0 {
			rows = kit.MaxInt(rows, lst.Row+lst.RowSpan)
		}
	}

	if cols == 0 {
		cols = int(math.Sqrt(float64(sz))) // whatever -- not well defined
	}
	if rows == 0 {
		rows = sz / cols
	}
	for rows*cols < sz { // not defined to have multiple items per cell -- make room for everyone
		rows++
	}

	ly.GridSize.X = cols
	ly.GridSize.Y = rows

	if len(ly.GridDataRows) != rows {
		ly.GridDataRows = make([]LayoutData, rows)
	}
	if len(ly.GridDataCols) != cols {
		ly.GridDataCols = make([]LayoutData, cols)
	}

	for _, ld := range ly.GridDataRows {
		ld.Size.Need.Set(0, 0)
		ld.Size.Pref.Set(0, 0)
	}
	for _, ld := range ly.GridDataCols {
		ld.Size.Need.Set(0, 0)
		ld.Size.Pref.Set(0, 0)
	}

	col := 0
	row := 0
	// var sumPref, sumNeed, maxPref, maxNeed Vec2D
	for _, c := range ly.Kids {
		_, gi := KiToNode2D(c)
		if gi == nil {
			continue
		}
		gi.LayData.UpdateSizes()
		lst := gi.Style.Layout
		if lst.Col > 0 {
			col = lst.Col
		}
		if lst.Row > 0 {
			row = lst.Row
		}
		// r   0   1   col X = max(all in col), Y = sum of all in col
		//   +--+---+
		// 0 |  |   |  Y = max(all in row), X = sum of all in row
		//   +--+---+
		// 1 |  |   |
		//   +--+---+

		// todo: need to deal with span in sums..
		ly.GridDataRows[row].Size.Need.SetMaxDim(Y, gi.LayData.Size.Need.Y)
		ly.GridDataRows[row].Size.Pref.SetMaxDim(Y, gi.LayData.Size.Pref.Y)
		ly.GridDataRows[col].Size.Need.SetMaxDim(X, gi.LayData.Size.Need.X)
		ly.GridDataRows[col].Size.Pref.SetMaxDim(X, gi.LayData.Size.Pref.X)

		col++
		if col >= cols { // todo: really only works if NO items specify row,col or ALL do..
			col = 0
			row++
			if row >= rows { // wrap-around.. no other good option
				row = 0
			}
		}
	}

	lst := ly.Style.Layout

	// Y = sum across rows which have max's
	var sumPref, sumNeed Vec2D
	for _, ld := range ly.GridDataRows {
		sumNeed.SetAddDim(Y, ld.Size.Pref.Y)
		sumPref.SetAddDim(Y, ld.Size.Pref.Y)
	}
	// X = sum across cols which have max's
	for _, ld := range ly.GridDataCols {
		sumNeed.SetAddDim(X, ld.Size.Pref.X)
		sumPref.SetAddDim(X, ld.Size.Pref.X)
	}

	sumNeed.SetAddDim(Y, float64(len(ly.GridDataRows)-1)*lst.Margin.Dots)
	sumPref.SetAddDim(Y, float64(len(ly.GridDataRows)-1)*lst.Margin.Dots)
	sumNeed.SetAddDim(X, float64(len(ly.GridDataCols)-1)*lst.Margin.Dots)
	sumPref.SetAddDim(X, float64(len(ly.GridDataCols)-1)*lst.Margin.Dots)

	ly.LayData.Size.Need.SetMax(sumNeed)
	ly.LayData.Size.Pref.SetMax(sumPref)

	spc := ly.Style.BoxSpace()
	ly.LayData.Size.Need.SetAddVal(2.0 * spc)
	ly.LayData.Size.Pref.SetAddVal(2.0 * spc)

	ly.LayData.UpdateSizes() // enforce max and normal ordering, etc
	if Layout2DTrace {
		fmt.Printf("Size:   %v gather sizes need: %v, pref: %v\n", ly.PathUnique(), ly.LayData.Size.Need, ly.LayData.Size.Pref)
	}
}

// if we are not a child of a layout, then get allocation from a parent obj that
// has a layout size
func (ly *Layout) AllocFromParent() {
	if ly.Par == nil || !ly.LayData.AllocSize.IsZero() {
		return
	}
	pgi, _ := KiToNode2D(ly.Par)
	lyp := pgi.AsLayout2D()
	if lyp == nil {
		ly.FuncUpParent(0, ly.This, func(k ki.Ki, level int, d interface{}) bool {
			_, pg := KiToNode2D(k)
			if pg == nil {
				return false
			}
			if !pg.LayData.AllocSize.IsZero() {
				ly.LayData.AllocPos = pg.LayData.AllocPos
				ly.LayData.AllocSize = pg.LayData.AllocSize
				if Layout2DTrace {
					fmt.Printf("Layout: %v got parent alloc: %v from %v\n", ly.PathUnique(), ly.LayData.AllocSize, pg.PathUnique())
				}
				return false
			}
			return true
		})
	}
}

// calculations to layout a single-element dimension, returns pos and size
func (ly *Layout) LayoutSingleImpl(avail, need, pref, max float64, al Align) (pos, size float64) {
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

	if usePref && extra > 0.0 { // have some stretch extra
		if max < 0.0 {
			stretchMax = true // only stretch those marked as infinitely stretchy
		}
	} else if extra > 0.0 { // extra relative to Need
		stretchNeed = true // stretch relative to need
	}

	pos = ly.Style.BoxSpace()
	size = need
	if usePref {
		size = pref
	}
	if stretchMax || stretchNeed {
		size += extra
	} else {
		if IsAlignMiddle(al) {
			pos += 0.5 * extra
		} else if IsAlignEnd(al) {
			pos += extra
		} else if al == AlignJustify { // treat justify as stretch
			size += extra
		}
	}

	// if ly.IsField() {
	// 	fmt.Printf("ly %v avail: %v targ: %v, extra %v, strMax: %v, strNeed: %v, pos: %v size: %v\n", ly.Nm, avail, targ, extra, stretchMax, stretchNeed, pos, size)
	// }

	return
}

// layout item in single-dimensional case -- e.g., orthogonal dimension from LayoutRow / Col
func (ly *Layout) LayoutSingle(dim Dims2D) {
	spc := ly.Style.BoxSpace()
	avail := ly.LayData.AllocSize.Dim(dim) - 2.0*spc
	for _, c := range ly.Kids {
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
	sz := len(ly.Kids)
	if sz == 0 {
		return
	}

	al := ly.Style.Layout.AlignDim(dim)
	spc := ly.Style.BoxSpace()
	avail := ly.LayData.AllocSize.Dim(dim) - 2.0*spc
	pref := ly.LayData.Size.Pref.Dim(dim)
	need := ly.LayData.Size.Need.Dim(dim)

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
	stretchNeed := false        // stretch relative to need
	stretchMax := false         // only stretch Max = neg
	addSpace := false           // apply extra toward spacing -- for justify
	if usePref && extra > 0.0 { // have some stretch extra
		for _, c := range ly.Kids {
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
	} else if extra > 0.0 { // extra relative to Need
		for _, c := range ly.Kids {
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
	if sz > 1 && extra > 0.0 && al == AlignJustify && !stretchNeed && !stretchMax {
		addSpace = true
		// if neither, then just distribute as spacing for justify
		extraSpace = extra / float64(sz-1)
	}

	// now arrange everyone
	pos := spc

	// todo: need a direction setting too
	if IsAlignEnd(al) && !stretchNeed && !stretchMax {
		pos += extra
	}

	if Layout2DTrace {
		fmt.Printf("Layout: %v All on dim %v, avail: %v need: %v pref: %v targ: %v, extra %v, strMax: %v, strNeed: %v, nstr %v, strTot %v\n", ly.PathUnique(), dim, avail, need, pref, targ, extra, stretchMax, stretchNeed, nstretch, stretchTot)
	}

	for i, c := range ly.Kids {
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
		// if ly.IsField() {
		// 	fmt.Printf("child: %v, pos: %v, size: %v\n", gi.Nm, pos, size)
		// }
		pos += size
	}
}

// final pass through children to finalize the layout, capturing original
// positions and computing summary size stats
func (ly *Layout) FinalizeLayout() {
	ly.ChildSize = Vec2DZero
	for _, c := range ly.Kids {
		_, gi := KiToNode2D(c)
		if gi == nil {
			continue
		}
		gi.LayData.AllocPosOrig = gi.LayData.AllocPos
		ly.ChildSize.SetMax(gi.LayData.AllocPos.Add(gi.LayData.AllocSize))
	}
}

// process any overflow according to overflow settings
func (ly *Layout) ManageOverflow() {
	if len(ly.Kids) == 0 {
		return
	}
	spc := ly.Style.BoxSpace()
	avail := ly.LayData.AllocSize.SubVal(spc)

	ly.ExtraSize.SetVal(0.0)
	ly.HasHScroll = false
	ly.HasVScroll = false

	if ly.Style.Layout.Overflow != OverflowHidden {
		sbw := ly.Style.Layout.ScrollBarWidth.Dots
		if ly.ChildSize.X > avail.X { // overflowing
			ly.HasHScroll = true
			ly.ExtraSize.Y += sbw
		}
		if ly.ChildSize.Y > avail.Y { // overflowing
			ly.HasVScroll = true
			ly.ExtraSize.X += sbw
		}

		if ly.HasHScroll {
			ly.SetHScroll()
			// } else {
			// todo: probably don't need to delete hscroll - just keep around
		}
		if ly.HasVScroll {
			ly.SetVScroll()
		}
		ly.LayoutScrolls()
	}
}

func (ly *Layout) SetHScroll() {
	if ly.HScroll == nil {
		ly.HScroll = &ScrollBar{}
		ly.HScroll.InitName(ly.HScroll, "Lay_HScroll")
		ly.HScroll.SetParent(ly.This)
		ly.HScroll.Horiz = true
		ly.HScroll.Init2D()
		ly.HScroll.Defaults()
	}
	spc := ly.Style.BoxSpace()
	sc := ly.HScroll
	sc.SetFixedHeight(ly.Style.Layout.ScrollBarWidth)
	sc.SetFixedWidth(units.NewValue(ly.LayData.AllocSize.X, units.Dot))
	sc.Style2D()
	sc.Min = 0.0
	sc.Max = ly.ChildSize.X + ly.ExtraSize.X // only scrollbar
	sc.Step = ly.Style.Font.Size.Dots        // step by lines
	sc.PageStep = 10.0 * sc.Step             // todo: more dynamic
	sc.ThumbVal = ly.LayData.AllocSize.X - spc
	sc.Tracking = true
	sc.SliderSig.Connect(ly.This, func(rec, send ki.Ki, sig int64, data interface{}) {
		if sig != int64(SliderValueChanged) {
			return
		}
		li, _ := KiToNode2D(rec) // note: avoid using closures
		ls := li.AsLayout2D()
		ls.Move2DTree()
		ls.Viewport.ReRender2DNode(li)
	})
}

// todo: we are leaking the scrollbars..
func (ly *Layout) DeleteHScroll() {
	if ly.HScroll == nil {
		return
	}
	sc := ly.HScroll
	win := ly.ParentWindow()
	if win != nil {
		sc.DisconnectAllEvents(win)
	}
	sc.Destroy()
	ly.HScroll = nil
}

func (ly *Layout) SetVScroll() {
	if ly.VScroll == nil {
		ly.VScroll = &ScrollBar{}
		ly.VScroll.InitName(ly.VScroll, "Lay_VScroll")
		ly.VScroll.SetParent(ly.This)
		ly.VScroll.Init2D()
		ly.VScroll.Defaults()
	}
	spc := ly.Style.BoxSpace()
	sc := ly.VScroll
	sc.SetFixedWidth(ly.Style.Layout.ScrollBarWidth)
	sc.SetFixedHeight(units.NewValue(ly.LayData.AllocSize.Y, units.Dot))
	sc.Style2D()
	sc.Min = 0.0
	sc.Max = ly.ChildSize.Y + ly.ExtraSize.Y // only scrollbar
	sc.Step = ly.Style.Font.Size.Dots        // step by lines
	sc.PageStep = 10.0 * sc.Step             // todo: more dynamic
	sc.ThumbVal = ly.LayData.AllocSize.Y - spc
	sc.Tracking = true
	sc.SliderSig.Connect(ly.This, func(rec, send ki.Ki, sig int64, data interface{}) {
		if sig != int64(SliderValueChanged) {
			return
		}
		li, _ := KiToNode2D(rec) // note: avoid using closures
		ls := li.AsLayout2D()
		ls.Move2DTree()
		ls.Viewport.ReRender2DNode(li)
	})
}

func (ly *Layout) DeleteVScroll() {
	if ly.VScroll == nil {
		return
	}
	sc := ly.VScroll
	win := ly.ParentWindow()
	if win != nil {
		sc.DisconnectAllEvents(win)
	}
	sc.Destroy() // this resets all signals and connections
	ly.VScroll = nil
}

func (ly *Layout) DeactivateScroll(sc *ScrollBar) {
	sc.LayData.AllocPos = Vec2DZero
	sc.LayData.AllocSize = Vec2DZero
	sc.VpBBox = image.ZR
	sc.WinBBox = image.ZR
}

func (ly *Layout) LayoutScrolls() {
	sbw := ly.Style.Layout.ScrollBarWidth.Dots
	if ly.HasHScroll {
		sc := ly.HScroll
		sc.Size2D()
		sc.LayData.AllocPos.X = ly.LayData.AllocPos.X
		sc.LayData.AllocPos.Y = ly.LayData.AllocPos.Y + ly.LayData.AllocSize.Y - sbw - 2.0
		sc.LayData.AllocPosOrig = sc.LayData.AllocPos
		sc.LayData.AllocSize.X = ly.LayData.AllocSize.X
		if ly.HasVScroll { // make room for V
			sc.LayData.AllocSize.X -= sbw
		}
		sc.LayData.AllocSize.Y = sbw
		sc.Layout2D(ly.VpBBox)
	} else {
		if ly.HScroll != nil {
			ly.DeactivateScroll(ly.HScroll)
		}
	}
	if ly.HasVScroll {
		sc := ly.VScroll
		sc.Size2D()
		sc.LayData.AllocPos.X = ly.LayData.AllocPos.X + ly.LayData.AllocSize.X - sbw - 2.0
		sc.LayData.AllocPos.Y = ly.LayData.AllocPos.Y
		sc.LayData.AllocPosOrig = sc.LayData.AllocPos
		sc.LayData.AllocSize.Y = ly.LayData.AllocSize.Y
		if ly.HasHScroll { // make room for H
			sc.LayData.AllocSize.Y -= sbw
		}
		sc.LayData.AllocSize.X = sbw
		sc.Layout2D(ly.VpBBox)
	} else {
		if ly.VScroll != nil {
			ly.DeactivateScroll(ly.VScroll)
		}
	}
}

func (ly *Layout) RenderScrolls() {
	if ly.HasHScroll {
		ly.HScroll.Render2D()
	}
	if ly.HasVScroll {
		ly.VScroll.Render2D()
	}
}

// render the children
func (ly *Layout) Render2DChildren() {
	if ly.Lay == LayoutStacked {
		if ly.StackTop.Ptr == nil {
			return
		}
		gii, _ := KiToNode2D(ly.StackTop.Ptr)
		gii.Render2D()
		return
	}
	for _, kid := range ly.Kids {
		gii, _ := KiToNode2D(kid)
		if gii != nil {
			gii.Render2D()
		}
	}
}

// convenience for LayoutStacked to show child node at a given index
func (ly *Layout) ShowChildAtIndex(idx int) error {
	idx, err := ly.Kids.ValidIndex(idx)
	if err != nil {
		return err
	}
	ly.StackTop.Ptr = ly.Child(idx)
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

func (g *Layout) AsLayout2D() *Layout {
	return g
}

func (ly *Layout) Init2D() {
	ly.Init2DBase()
}

func (ly *Layout) BBox2D() image.Rectangle {
	return ly.BBoxFromAlloc()
}

func (ly *Layout) ComputeBBox2D(parBBox image.Rectangle) {
	ly.ComputeBBox2DBase(parBBox)
}

func (ly *Layout) ChildrenBBox2D() image.Rectangle {
	nb := ly.ChildrenBBox2DWidget()
	nb.Max.X -= int(ly.ExtraSize.X)
	nb.Max.Y -= int(ly.ExtraSize.Y)
	return nb
}

func (ly *Layout) Style2D() {
	ly.Style2DWidget(nil)
}

func (ly *Layout) Size2D() {
	ly.InitLayout2D()
	ly.GatherSizes()
}

func (ly *Layout) Layout2D(parBBox image.Rectangle) {
	ly.AllocFromParent()           // in case we didn't get anything
	ly.Layout2DBase(parBBox, true) // init style
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
	ly.ManageOverflow()
	ly.Layout2DChildren() // layout done with canonical positions

	delta := ly.Move2DDelta(Vec2DZero)
	if !delta.IsZero() {
		ly.Move2DChildren(delta) // move is a separate step
	}
}

// we add our own offset here
func (ly *Layout) Move2DDelta(delta Vec2D) Vec2D {
	if ly.HasHScroll {
		off := ly.HScroll.Value
		delta.X -= off
	}
	if ly.HasVScroll {
		off := ly.VScroll.Value
		delta.Y -= off
	}
	return delta
}

func (ly *Layout) Move2D(delta Vec2D, parBBox image.Rectangle) {
	ly.Move2DBase(delta, parBBox)
	delta = ly.Move2DDelta(delta) // add our offset
	ly.Move2DChildren(delta)
}

func (ly *Layout) Render2D() {
	if ly.PushBounds() {
		ly.RenderScrolls()
		ly.Render2DChildren()
		ly.PopBounds()
	}
}

func (ly *Layout) ReRender2D() (node Node2D, layout bool) {
	node = ly.This.(Node2D)
	layout = true
	return
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

var KiT_Frame = kit.Types.AddType(&Frame{}, nil)

func (g *Frame) AsNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *Frame) AsViewport2D() *Viewport2D {
	return nil
}

func (g *Frame) AsLayout2D() *Layout {
	return &g.Layout
}

func (g *Frame) Init2D() {
	g.Init2DBase()
}

var FrameProps = map[string]interface{}{
	"border-width":     units.NewValue(2, units.Px),
	"border-radius":    units.NewValue(0, units.Px),
	"border-color":     color.Black,
	"border-style":     BorderSolid,
	"padding":          units.NewValue(2, units.Px),
	"margin":           units.NewValue(2, units.Px),
	"color":            color.Black,
	"background-color": color.White,
}

func (g *Frame) Style2D() {
	g.Style2DWidget(FrameProps)
}

func (g *Frame) Size2D() {
	g.Layout.Size2D()
}

func (g *Frame) Layout2D(parBBox image.Rectangle) {
	g.Layout.Layout2D(parBBox)
}

func (g *Frame) BBox2D() image.Rectangle {
	return g.Layout.BBox2D()
}

func (g *Frame) ComputeBBox2D(parBBox image.Rectangle) {
	g.Layout.ComputeBBox2D(parBBox)
}

func (g *Frame) Move2D(delta Vec2D, parBBox image.Rectangle) {
	g.Layout.Move2D(delta, parBBox)
}

func (g *Frame) Render2D() {
	if g.PushBounds() {
		pc := &g.Paint
		st := &g.Style
		rs := &g.Viewport.Render
		// first draw a background rectangle in our full area
		pc.StrokeStyle.SetColor(nil)
		pc.FillStyle.SetColor(&st.Background.Color)
		pos := g.LayData.AllocPos
		sz := g.LayData.AllocSize
		pc.DrawRectangle(rs, pos.X, pos.Y, sz.X, sz.Y)
		pc.FillStrokeClear(rs)

		rad := st.Border.Radius.Dots
		pos = pos.AddVal(st.Layout.Margin.Dots).SubVal(0.5 * st.Border.Width.Dots)
		sz = sz.SubVal(2.0 * st.Layout.Margin.Dots).AddVal(st.Border.Width.Dots)

		// then any shadow
		if st.BoxShadow.HasShadow() {
			spos := pos.Add(Vec2D{st.BoxShadow.HOffset.Dots, st.BoxShadow.VOffset.Dots})
			pc.StrokeStyle.SetColor(nil)
			pc.FillStyle.SetColor(&st.BoxShadow.Color)
			if rad == 0.0 {
				pc.DrawRectangle(rs, spos.X, spos.Y, sz.X, sz.Y)
			} else {
				pc.DrawRoundedRectangle(rs, spos.X, spos.Y, sz.X, sz.Y, rad)
			}
			pc.FillStrokeClear(rs)
		}

		pc.FillStyle.SetColor(&st.Background.Color)
		pc.StrokeStyle.SetColor(&st.Border.Color)
		pc.StrokeStyle.Width = st.Border.Width
		if rad == 0.0 {
			pc.DrawRectangle(rs, pos.X, pos.Y, sz.X, sz.Y)
		} else {
			pc.DrawRoundedRectangle(rs, pos.X, pos.Y, sz.X, sz.Y, rad)
		}
		pc.FillStrokeClear(rs)

		g.Layout.Render2D()
		g.PopBounds()
	}
}

func (g *Frame) ReRender2D() (node Node2D, layout bool) {
	node = g.This.(Node2D)
	layout = true
	return
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

var KiT_Stretch = kit.Types.AddType(&Stretch{}, nil)

func (g *Stretch) AsNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *Stretch) AsViewport2D() *Viewport2D {
	return nil
}

func (g *Stretch) AsLayout2D() *Layout {
	return nil
}

func (g *Stretch) Init2D() {
	g.Init2DBase()
}

var StretchProps = map[string]interface{}{
	"max-width":  -1.0,
	"max-height": -1.0,
}

func (g *Stretch) Style2D() {
	g.Style2DWidget(StretchProps)
}

func (g *Stretch) Size2D() {
	g.InitLayout2D()
}

func (g *Stretch) Layout2D(parBBox image.Rectangle) {
	g.Layout2DBase(parBBox, true) // init style
	g.Layout2DChildren()
}

func (g *Stretch) BBox2D() image.Rectangle {
	return g.BBoxFromAlloc()
}

func (g *Stretch) ComputeBBox2D(parBBox image.Rectangle) {
	g.ComputeBBox2DBase(parBBox)
}

func (g *Stretch) ChildrenBBox2D() image.Rectangle {
	return g.VpBBox
}

func (g *Stretch) Move2D(delta Vec2D, parBBox image.Rectangle) {
	g.Move2DBase(delta, parBBox)
	g.Move2DChildren(delta)
}

func (g *Stretch) Render2D() {
	if g.PushBounds() {
		g.Render2DChildren()
		g.PopBounds()
	}
}

func (g *Stretch) ReRender2D() (node Node2D, layout bool) {
	node = g.This.(Node2D)
	layout = false
	return
}

func (g *Stretch) FocusChanged2D(gotFocus bool) {
}

// check for interface implementation
var _ Node2D = &Stretch{}

// Space adds a fixed sized (1 em by default) blank space to a layout -- set width / height property to change
type Space struct {
	Node2DBase
}

var KiT_Space = kit.Types.AddType(&Space{}, nil)

func (g *Space) AsNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *Space) AsViewport2D() *Viewport2D {
	return nil
}

func (g *Space) AsLayout2D() *Layout {
	return nil
}

func (g *Space) Init2D() {
	g.Init2DBase()
}

var SpaceProps = map[string]interface{}{
	"width":  units.NewValue(1, units.Em),
	"height": units.NewValue(1, units.Em),
}

func (g *Space) Style2D() {
	g.Style2DWidget(SpaceProps)
}

func (g *Space) Size2D() {
	g.InitLayout2D()
}

func (g *Space) Layout2D(parBBox image.Rectangle) {
	g.Layout2DBase(parBBox, true) // init style
	g.Layout2DChildren()
}

func (g *Space) BBox2D() image.Rectangle {
	return g.BBoxFromAlloc()
}

func (g *Space) ComputeBBox2D(parBBox image.Rectangle) {
	g.ComputeBBox2DBase(parBBox)
}

func (g *Space) ChildrenBBox2D() image.Rectangle {
	return g.VpBBox
}

func (g *Space) Move2D(delta Vec2D, parBBox image.Rectangle) {
	g.Move2DBase(delta, parBBox)
	g.Move2DChildren(delta)
}

func (g *Space) Render2D() {
	if g.PushBounds() {
		g.Render2DChildren()
		g.PopBounds()
	}
}

func (g *Space) ReRender2D() (node Node2D, layout bool) {
	node = g.This.(Node2D)
	layout = false
	return
}

func (g *Space) FocusChanged2D(gotFocus bool) {
}

// check for interface implementation
var _ Node2D = &Space{}
