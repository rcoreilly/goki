// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	// "math"
	"github.com/rcoreilly/goki/ki"
	"image"
)

// this is based on QtQuick layouts https://doc.qt.io/qt-5/qtquicklayouts-overview.html  https://doc.qt.io/qt-5/qml-qtquick-layouts-layout.html

// horizontal alignment type -- how to align items in the horizontal dimension
type AlignHorizontal int32

const (
	AlignLeft AlignHorizontal = iota
	AlignHCenter
	AlignRight
	AlignJustify
)

//go:generate stringer -type=AlignHorizontal

// vertical alignment type -- how to align items in the vertical dimension
type AlignVertical int32

const (
	AlignTop AlignVertical = iota
	AlignVCenter
	AlignBottom
	AlignBaseline
)

//go:generate stringer -type=AlignVertical

// size preferences -- a value of 0 indicates no preference
type SizePrefs struct {
	Min  Size2D `desc:"minimum size -- will not be less than this"`
	Pref Size2D `desc:"preferred size -- start here"`
	Max  Size2D `desc:"maximum size -- will not be greater than this -- 0 = max size"`
}

// want is the maximum across any of our prefs
func (sp *SizePrefs) Want() Size2D {
	want := sp.Max.Max(sp.Pref)
	want = want.Max(sp.Min)
	return want
}

// need is the minimum across any of our prefs
func (sp *SizePrefs) Need() Size2D {
	need := sp.Max.Min(sp.Pref)
	need = need.Min(sp.Min)
	return need
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

// all the data needed to specify the layout of an item within a layout
type LayoutData struct {
	AlignH    AlignHorizontal `desc:"horizontal alignment"`
	AlignV    AlignVertical   `desc:"vertical alignment"`
	Size      SizePrefs       `desc:"size constraints for this item"`
	Margins   Margins         `desc:"margins around this item"`
	GridPos   image.Point     `desc:"position within a grid"`
	GridSpan  image.Point     `desc:"number of grid elements that we take up in each direction"`
	AllocPos  Point2D         `desc:"allocated relative position of this item, by the parent layout"`
	AllocSize Size2D          `desc:"allocated size of this item, by the parent layout"`
}

func (ld *LayoutData) Defaults() {
	if ld.GridSpan.X < 1 {
		ld.GridSpan.X = 1
	}
	if ld.GridSpan.Y < 1 {
		ld.GridSpan.Y = 1
	}
}

// want is max across prefs and existing allocsize
func (ld *LayoutData) WantSize() Size2D {
	want := ld.Size.Want()
	want = want.Max(ld.AllocSize)
	return want
}

// need is min across prefs and existing allocsize
func (ld *LayoutData) NeedSize() Size2D {
	need := ld.Size.Need()
	need = need.Max(ld.AllocSize)
	return need
}

// RowLayout arranges its elements in a horizontal fashion
type RowLayout struct {
	Node2DBase
	Layout LayoutData
}

// must register all new types so type names can be looked up by name -- e.g., for json
var KiT_RowLayout = ki.KiTypes.AddType(&RowLayout{})

// need multiple iterations..
func (rl *RowLayout) Layout2D(iter int) {
	if len(rl.Children) == 0 {
		rl.Layout.AllocSize = rl.Layout.Size.Min
		return
	}

	if iter == 0 { // start out with nothing
		rl.Layout.AllocSize = Size2D{0, 0}
	}

	// todo: need to include margins in all this!  do we use our margins or items?

	var sumWant, sumNeed, sumMin, maxWant, maxNeed, maxMin Size2D
	for _, c := range rl.Children {
		gii, ok := c.(Node2D)
		if !ok {
			continue
		}
		gi := gii.GiNode2D()
		want := gi.Layout.WantSize()
		need := gi.Layout.NeedSize()
		min := gi.Layout.Size.Need() // ignoring current allocations
		sumWant = sumWant.Sum(want)
		sumNeed = sumNeed.Sum(need)
		sumMin = sumMin.Sum(min)
		maxWant = maxWant.Max(want)
		maxNeed = maxNeed.Max(need)
		maxMin = maxMin.Max(min)
	}

	minAvail := rl.Layout.Size.Need().X
	maxAvail := rl.Layout.Size.Want().X
	curAlloc := rl.Layout.AllocSize.X
	if curAlloc > 0 && curAlloc < maxAvail {
		maxAvail = curAlloc
	}
	if curAlloc > 0 && curAlloc < minAvail {
		minAvail = curAlloc
	}
	extra := 0.0
	avail := maxAvail // start with that
	targ := sumWant.X
	useWant := true
	useMin := false
	if avail == 0 { // no limits -- size to fit
		extra = 0.0
	} else {
		extra = avail - targ // go big first
		if extra < 0 {
			useWant = false
			targ = sumNeed.X
			extra = avail - targ
			if extra < 0 {
				useMin = true
				targ = sumMin.X
				extra = avail - targ
			}
		} else if avail-extra > minAvail { // lots of extra
			avail = minAvail
			extra = avail - targ
		}
	}

	// todo: vertical too!

	pos := 0.0
	per := extra / float64(len(rl.Children))
	for _, c := range rl.Children {
		gii, ok := c.(Node2D)
		if !ok {
			continue
		}
		gi := gii.GiNode2D()
		want := gi.Layout.WantSize()
		need := gi.Layout.NeedSize()
		min := gi.Layout.Size.Need() // ignoring current allocations
		base := need
		if useWant {
			base = want
		} else if useMin {
			base = min
		}
		gi.Layout.AllocSize = base
		gi.Layout.AllocSize.X += per
		gi.Layout.AllocPos.X = pos
		gi.Layout.AllocPos.Y = 0 // todo: alignment
		pos += gi.Layout.AllocSize.X
	}
}
