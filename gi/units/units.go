// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package Units supports full range of CSS-style length units (em, px, dp, etc)

	The unit is stored along with a value, and can be converted at a later point into
	a raw display pixel value using the Context which contains all the necessary reference
	values to perform the conversion.  Typically the unit value is parsed early from a style
	and then converted later once the context is fully resolved.  The Value also holds the
	converted value (Dots) so it can be used directly without further re-conversion.

	'Dots' are used as term for underlying raw display pixels because "Pixel" and the px unit
	are actually not conventionally used as raw display pixels in the current HiDPI
	environment.  See https://developer.mozilla.org/en/docs/Web/CSS/length -- 1 px = 1/96 in
	Also supporting dp = density-independent pixel = 1/160 in
*/
package units

import (
	"fmt"
	"strings"

	"github.com/rcoreilly/goki/ki"
	"github.com/rcoreilly/goki/ki/kit"
	"golang.org/x/image/math/fixed"
)

// borrows from golang.org/x/exp/shiny/unit/ but extends with full range of css-based viewport-dependent factors

//

// standard conversion factors -- Px = DPI-independent pixel instead of actual "dot" raw pixel
const (
	PxPerInch = 96.0
	DpPerInch = 160.0
	MmPerInch = 25.4
	CmPerInch = 2.54
	PtPerInch = 72.0
	PcPerInch = 6.0
)

type Unit int32

const (
	// Px = pixels -- 1px = 1/96th of 1in -- these are NOT raw display pixels
	Px Unit = iota

	// Dp = density-independent pixels -- 1dp = 1/160th of 1in
	Dp

	// Pct = percentage of surrounding contextual element
	Pct

	// Rem = font size of the root element -- fallback to 12pt by default
	Rem

	// Em = font size of the element -- fallback to 12pt by default
	Em

	// Ex = x-height of the element's font -- fallback to 0.5em by default
	Ex

	// Ch = with of the '0' glyph in the element's font -- fallback to 0.5em by default
	Ch

	// Vw = 1% of the viewport's width
	Vw

	// Vh = 1% of the viewport's height
	Vh

	// Vmin = 1% of the viewport's smaller dimension
	Vmin

	// Vmax = 1% of the viewport's larger dimension
	Vmax

	// Cm = centimeters -- 1cm = 96px/2.54
	Cm

	// Mm = millimeters -- 1mm = 1/10th of cm
	Mm

	// Q = quarter-millimeters -- 1q = 1/40th of cm
	Q

	// In = inches -- 1in = 2.54cm = 96px
	In

	// Pc = picas -- 1pc = 1/6th of 1in
	Pc

	// Pt = points -- 1pt = 1/72th of 1in
	Pt

	// Dot = actual real display pixels -- generally only use internally
	Dot

	UnitN
)

//go:generate stringer -type=Unit

var KiT_Unit = kit.Enums.AddEnumAltLower(UnitN, false, nil, "")

func (ev Unit) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Unit) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

var UnitNames = [...]string{
	Px:   "px",
	Dp:   "dp",
	Pct:  "pct",
	Rem:  "rem",
	Em:   "em",
	Ex:   "ex",
	Ch:   "ch",
	Vw:   "vw",
	Vh:   "vh",
	Vmin: "vmin",
	Vmax: "vmax",
	Cm:   "cm",
	Mm:   "mm",
	Q:    "q",
	In:   "in",
	Pc:   "pc",
	Pt:   "pt",
	Dot:  "dot",
}

// Context specifies everything about the current context necessary for converting the number
// into specific display-dependent pixels
type Context struct {
	// dots-per-inch of the display
	DPI float32
	// point size of font (in points)
	FontEm float32
	// x-height of font in points
	FontEx float32
	// ch-size of font in points
	FontCh float32
	// rem-size of font in points
	FontRem float32
	// viewport width in dots
	VpW float32
	// viewport height in dots
	VpH float32
	// width of surrounding contextual element in dots
	ElW float32
	// height of surrounding contextual element in dots
	ElH float32
}

func (uc *Context) Defaults() {
	uc.DPI = PxPerInch
	uc.FontEm = 12.0
	uc.FontEx = 6.0
	uc.FontCh = 6.0
	uc.FontRem = 12.0
	uc.VpW = 800.0
	uc.VpH = 600.0
	uc.ElW = uc.VpW
	uc.ElH = uc.VpH
}

// set the context values
func (uc *Context) Set(em, ex, ch, rem, vpw, vph, elw, elh float32) {
	uc.SetSizes(vpw, vph, elw, elh)
	uc.SetFont(em, ex, ch, rem)
}

// set the context values for non-font sizes -- el is ignored if zero
func (uc *Context) SetSizes(vpw, vph, elw, elh float32) {
	uc.VpW = vpw
	uc.VpH = vph
	if elw != 0 {
		uc.ElW = elw
	}
	if elh != 0 {
		uc.ElH = elh
	}
}

// set the context values for fonts
func (uc *Context) SetFont(em, ex, ch, rem float32) {
	uc.FontEm = em
	uc.FontEx = ex
	uc.FontCh = ch
	uc.FontRem = rem
}

// factor needed to convert given unit into raw pixels (dots in DPI)
func (uc *Context) ToDotsFactor(un Unit) float32 {
	if uc.DPI == 0 {
		// log.Printf("gi/units Context was not initialized -- falling back on defaults\n")
		uc.Defaults()
	}
	switch un {
	case Pct:
		return uc.ElW // todo: height should be in terms of Elh.. sheesh
	case Em:
		return uc.DPI / (PtPerInch / uc.FontEm)
	case Ex:
		return uc.DPI / (PtPerInch / uc.FontEx)
	case Ch:
		return uc.DPI / (PtPerInch / uc.FontCh)
	case Rem:
		return uc.DPI / (PtPerInch / uc.FontRem)
	case Vw:
		return uc.VpW
	case Vh:
		return uc.VpH
	case Vmin:
		return kit.Min32(uc.VpW, uc.VpH)
	case Vmax:
		return kit.Max32(uc.VpW, uc.VpH)
	case Cm:
		return uc.DPI / CmPerInch
	case Mm:
		return uc.DPI / MmPerInch
	case Q:
		return uc.DPI / (4.0 * MmPerInch)
	case In:
		return uc.DPI
	case Pc:
		return uc.DPI / PcPerInch
	case Pt:
		return uc.DPI / PtPerInch
	case Px:
		return uc.DPI / PxPerInch
	case Dp:
		return uc.DPI / DpPerInch
	case Dot:
		return 1.0
	}
	return uc.DPI
}

// convert value in given units into raw display pixels (dots in DPI)
func (uc *Context) ToDots(val float32, un Unit) float32 {
	return val * uc.ToDotsFactor(un)
}

// Value and units, and converted value into raw pixels (dots in DPI)
type Value struct {
	Val  float32
	Un   Unit
	Dots float32
}

var KiT_Value = kit.Types.AddType(&Value{}, ValueProps)

var ValueProps = ki.Props{
	"style-prop": true,
}

// convenience for not having to specify the Dots member
func NewValue(val float32, un Unit) Value {
	return Value{val, un, 0.0}
}

func (v *Value) Set(val float32, un Unit) {
	v.Val = val
	v.Un = un
}

// Convert value to raw display pixels (dots as in DPI), setting also the Dots field
func (v *Value) ToDots(ctxt *Context) float32 {
	v.Dots = ctxt.ToDots(v.Val, v.Un)
	return v.Dots
}

// Convert value to raw display pixels (dots in DPI) in fixed-point 26.6 format for rendering
func (v *Value) ToDotsFixed(ctxt *Context) fixed.Int26_6 {
	return fixed.Int26_6(v.ToDots(ctxt))
}

// Convert converts value to the given units, given unit context
func (v *Value) Convert(to Unit, ctxt *Context) Value {
	dots := v.ToDots(ctxt)
	return Value{dots / ctxt.ToDotsFactor(to), to, dots}
}

// String implements the fmt.Stringer interface.
func (v *Value) String() string {
	return fmt.Sprintf("%f%s", v.Val, UnitNames[v.Un])
}

// parse string into a value
func (v *Value) SetFromString(str string) {
	trstr := strings.TrimSpace(str)
	sz := len(trstr)
	if sz < 2 {
		v.Set(0, Px)
		return
	}
	var ends [4]string
	ends[0] = strings.ToLower(trstr[sz-1:])
	ends[1] = strings.ToLower(trstr[sz-2:])
	if sz > 3 {
		ends[2] = strings.ToLower(trstr[sz-3:])
	}
	if sz > 4 {
		ends[3] = strings.ToLower(trstr[sz-4:])
	}

	var numstr string
	var un Unit = Px // default to pixels
	for i, nm := range UnitNames {
		unsz := len(nm)
		if ends[unsz-1] == nm {
			numstr = trstr[:sz-unsz]
			un = Unit(i)
			break
		}
	}
	if len(numstr) == 0 { // no units
		numstr = trstr
	}
	var val float32
	fmt.Sscanf(strings.TrimSpace(numstr), "%g", &val)
	v.Set(val, un)
}

func StringToValue(str string) Value {
	var v Value
	v.SetFromString(str)
	return v
}
