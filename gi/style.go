// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"image/color"
	"log"
	"reflect"
	"strings"
	"unsafe"

	"github.com/rcoreilly/goki/gi/units"
	"github.com/rcoreilly/goki/ki"
	"github.com/rcoreilly/goki/ki/kit"
	"github.com/rcoreilly/prof"
)

////////////////////////////////////////////////////////////////////////////////////////
// Widget Styling

// using CSS style reference: https://www.w3schools.com/cssref/default.asp
// which are inherited: https://stackoverflow.com/questions/5612302/which-css-properties-are-inherited

// styling strategy:
// * indiv objects specify styles using property map -- good b/c it is fully open-ended
// * we process those properties dynamically when rendering (first pass only) into state
//   on objects that can be directly used during rendering
// * good for basic rendering -- lots of additional things that could be extended later..
// * todo: could we generalize this to not have to write the parsing code?  YES need to
//
// SVG Paint inheritance is probably NOT such a good idea for widgets??  fill = background?
// may need to figure that part out a bit more..

// todo: Animation

// Bottom = alignment too

// Clear -- no floating elements

// Clip -- clip images

// column- settings -- lots of those

// LayoutStyle is in layout.go
// FontStyle is in font.go
// TextStyle is in text.go

// List-style for lists

// Object-fit for videos

// visibility -- support more than just hidden  inherit:"true"

// Transform -- can throw in any 2D or 3D transform!  we support that!  sort of..

// transition -- animation of hover, etc

// use StylePropProps for any enum (not type -- types must have their own
// props) that is useful as a styling property -- use this for selecting types
// to add to Props
var StylePropProps = ki.Props{
	"style-prop": true,
}

// style parameters for backgrounds
type BackgroundStyle struct {
	Color Color `xml:"color" desc:"background color"`
	// todo: all the properties not yet implemented -- mostly about images
	// Image is like a PaintServer -- includes gradients etc
	// Attachment -- how the image moves
	// Clip -- how to clip the image
	// Origin
	// Position
	// Repeat
	// Size
}

func (b *BackgroundStyle) Defaults() {
	b.Color.SetColor(color.White)
}

// sides of a box -- some properties can be specified per each side (e.g., border) or not
type BoxSides int32

const (
	BoxTop BoxSides = iota
	BoxRight
	BoxBottom
	BoxLeft
	BoxN
)

//go:generate stringer -type=BoxSides

var KiT_BoxSides = kit.Enums.AddEnumAltLower(BoxN, false, StylePropProps, "Box")

// how to draw the border
type BorderDrawStyle int32

const (
	BorderSolid BorderDrawStyle = iota
	BorderDotted
	BorderDashed
	BorderDouble
	BorderGroove
	BorderRidge
	BorderInset
	BorderOutset
	BorderNone
	BorderHidden
	BorderN
)

//go:generate stringer -type=BorderDrawStyle

var KiT_BorderDrawStyle = kit.Enums.AddEnumAltLower(BorderN, false, StylePropProps, "Border")

// style parameters for borders
type BorderStyle struct {
	Style  BorderDrawStyle `xml:"style" desc:"how to draw the border"`
	Width  units.Value     `xml:"width" desc:"width of the border"`
	Radius units.Value     `xml:"radius" desc:"rounding of the corners"`
	Color  Color           `xml:"color" desc:"color of the border"`
}

// style parameters for shadows
type ShadowStyle struct {
	HOffset units.Value `xml:".h-offset" desc:"horizontal offset of shadow -- positive = right side, negative = left side"`
	VOffset units.Value `xml:".v-offset" desc:"vertical offset of shadow -- positive = below, negative = above"`
	Blur    units.Value `xml:".blur" desc:"blur radius -- higher numbers = more blurry"`
	Spread  units.Value `xml:".spread" desc:"spread radius -- positive number increases size of shadow, negative descreases size"`
	Color   Color       `xml:".color" desc:"color of the shadow"`
	Inset   bool        `xml:".inset" desc:"shadow is inset within box instead of outset outside of box"`
}

func (s *ShadowStyle) HasShadow() bool {
	return (s.HOffset.Dots > 0 || s.VOffset.Dots > 0)
}

// all the CSS-based style elements -- used for widget-type objects
type Style struct {
	IsSet         bool            `desc:"has this style been set from object values yet?"`
	Display       bool            `xml:display" desc:"todo big enum of how to display item -- controls layout etc"`
	Visible       bool            `xml:visible" desc:"todo big enum of how to display item -- controls layout etc"`
	UnContext     units.Context   `xml:"-" desc:"units context -- parameters necessary for anchoring relative units"`
	Layout        LayoutStyle     `desc:"layout styles -- do not prefix with any xml"`
	Border        BorderStyle     `xml:"border" desc:"border around the box element -- todo: can have separate ones for different sides"`
	BoxShadow     ShadowStyle     `xml:"box-shadow" desc:"type of shadow to render around box"`
	Font          FontStyle       `xml:"font" desc:"font parameters"`
	Text          TextStyle       `desc:"text parameters -- no xml prefix"`
	Color         Color           `xml:"color" inherit:"true" desc:"text color"`
	Background    BackgroundStyle `xml:"background" desc:"background settings"`
	Opacity       float32         `xml:"opacity" desc:"alpha value to apply to all elements"`
	Outline       BorderStyle     `xml:"outline" desc:"draw an outline around an element -- mostly same styles as border -- default to none"`
	PointerEvents bool            `xml:"pointer-events" desc:"does this element respond to pointer events -- default is true"`
	// todo: also see above for more notes on missing style elements
}

func (s *Style) Defaults() {
	// mostly all the defaults are 0 initial values, except these..
	s.IsSet = false
	s.UnContext.Defaults()
	s.Opacity = 1.0
	s.Outline.Style = BorderNone
	s.PointerEvents = true
	s.Color.SetColor(color.Black)
	s.Background.Defaults()
	s.Layout.Defaults()
	s.Font.Defaults()
	s.Text.Defaults()
}

func NewStyle() Style {
	s := Style{}
	s.Defaults()
	return s
}

// SetStyle sets style values based on given property map (name: value pairs),
// inheriting elements as appropriate from parent
func (s *Style) SetStyle(parent *Style, props ki.Props) {
	if !s.IsSet && parent != nil { // first time
		StyleFields.Inherit(s, parent)
	}
	StyleFields.Style(s, parent, props)
	s.Text.AlignV = s.Layout.AlignV
	s.Layout.SetStylePost()
	s.Font.SetStylePost()
	s.Text.SetStylePost()
	s.IsSet = true
}

// SetUnitContext sets the unit context based on size of viewport and parent
// element (from bbox) and then cache everything out in terms of raw pixel
// dots for rendering -- call at start of render
func (s *Style) SetUnitContext(vp *Viewport2D, el Vec2D) {
	s.UnContext.Defaults()
	if vp != nil {
		if vp.Win != nil {
			s.UnContext.DPI = vp.Win.LogicalDPI()
			// fmt.Printf("set dpi: %v\n", s.UnContext.DPI)
			// } else {
			// 	fmt.Printf("No win for vp: %v\n", vp.PathUnique())
		}
		if vp.Render.Image != nil {
			sz := vp.Render.Image.Bounds().Size()
			s.UnContext.SetSizes(float32(sz.X), float32(sz.Y), el.X, el.Y)
		}
	}
	s.Font.SetUnitContext(&s.UnContext)
	s.ToDots()
}

// CopyUnitContext copies unit context from another, update with our font
// info, and then cache everything out in terms of raw pixel dots for
// rendering -- call at start of render
func (s *Style) CopyUnitContext(ctxt *units.Context) {
	s.UnContext = *ctxt
	s.Font.SetUnitContext(&s.UnContext)
	s.ToDots()
}

// ToDots calls ToDots on all units.Value fields in the style (recursively) --
// need to have set the UnContext first -- only after layout at render time is
// that possible
func (s *Style) ToDots() {
	StyleFields.ToDots(s, &s.UnContext)
}

// BoxSpace returns extra space around the central content in the box model,
// in dots -- todo: must complicate this if we want different spacing on
// different sides box outside-in: margin | border | padding | content
func (s *Style) BoxSpace() float32 {
	return s.Layout.Margin.Dots + s.Border.Width.Dots + s.Layout.Padding.Dots
}

// StyleDefault is default style can be used when property specifies "default"
var StyleDefault Style

// StyleFields contain the StyledFields for Style type
var StyleFields = initStyle()

func initStyle() *StyledFields {
	StyleDefault = NewStyle()
	sf := &StyledFields{}
	sf.Init(&StyleDefault)
	return sf
}

////////////////////////////////////////////////////////////////////////////////////////
//   StyledFields

// StyledFields contains fields of a struct that are styled -- create one
// instance of this for each type that has styled fields (Style, Paint, and a
// few with ad-hoc styled fields)
type StyledFields struct {
	Fields   map[string]*StyledField `desc:"the compiled stylable fields, mapped for the xml and alt tags for the field"`
	Inherits []*StyledField          `desc:"the compiled stylable fields that have inherit:"true" tags and should thus be inherited from parent objects"`
	Units    []*StyledField          `desc:"the compiled stylable fields of the unit.Value type, which should have ToDots run on them"`
	Default  interface{}             `desc:"points to the Default instance of this type, initialized with the default values used for 'initial' keyword"`
}

func (sf *StyledFields) Init(def interface{}) {
	sf.Default = def
	sf.CompileFields(def)
}

// get the full effective tag based on outer tag plus given tag
func StyleEffTag(tag, outerTag string) string {
	tagEff := tag
	if outerTag != "" && len(tag) > 0 {
		if tag[0] == '.' {
			tagEff = outerTag + tag
		} else {
			tagEff = outerTag + "-" + tag
		}
	}
	return tagEff
}

// AddField adds a single field -- must be a direct field on the object and
// not a field on an embedded type -- used for Widget objects where only one
// or a few fields are styled
func (sf *StyledFields) AddField(def interface{}, fieldName string) error {
	valtyp := reflect.TypeOf(units.Value{})

	if sf.Fields == nil {
		sf.Fields = make(map[string]*StyledField, 5)
		sf.Inherits = make([]*StyledField, 0, 5)
		sf.Units = make([]*StyledField, 0, 5)
	}
	otp := reflect.TypeOf(def)
	if otp.Kind() != reflect.Ptr {
		err := fmt.Errorf("gi.StyleFields.AddField: must pass pointers to the structs, not type: %v kind %v\n", otp, otp.Kind())
		log.Print(err)
		return err
	}
	ot := otp.Elem()
	if ot.Kind() != reflect.Struct {
		err := fmt.Errorf("gi.StyleFields.AddField: only works on structs, not type: %v kind %v\n", ot, ot.Kind())
		log.Print(err)
		return err
	}
	vo := reflect.ValueOf(def).Elem()
	struf, ok := ot.FieldByName(fieldName)
	if !ok {
		err := fmt.Errorf("gi.StyleFields.AddField: field name: %v not found in type %v\n", fieldName, ot.Name())
		log.Print(err)
		return err
	}

	vf := vo.FieldByName(fieldName)

	styf := &StyledField{Field: struf, NetOff: struf.Offset, Default: vf}
	tag := struf.Tag.Get("xml")
	sf.Fields[tag] = styf
	atags := struf.Tag.Get("alt")
	if atags != "" {
		atag := strings.Split(atags, ",")

		for _, tg := range atag {
			sf.Fields[tg] = styf
		}
	}
	inhs := struf.Tag.Get("inherit")
	if inhs == "true" {
		sf.Inherits = append(sf.Inherits, styf)
	}
	if vf.Kind() == reflect.Struct && vf.Type() == valtyp {
		sf.Units = append(sf.Units, styf)
	}
	return nil
}

// CompileFields gathers all the fields with xml tag != "-", plus those
// that are units.Value's for later optimized processing of styles
func (sf *StyledFields) CompileFields(def interface{}) {
	valtyp := reflect.TypeOf(units.Value{})

	sf.Fields = make(map[string]*StyledField, 50)
	sf.Inherits = make([]*StyledField, 0, 50)
	sf.Units = make([]*StyledField, 0, 50)

	WalkStyleStruct(def, "", uintptr(0),
		func(struf reflect.StructField, vf reflect.Value, outerTag string, baseoff uintptr) {
			styf := &StyledField{Field: struf, NetOff: baseoff + struf.Offset, Default: vf}
			tag := StyleEffTag(struf.Tag.Get("xml"), outerTag)
			if _, ok := sf.Fields[tag]; ok {
				fmt.Printf("gi.StyledFileds.CompileFields: ERROR redundant tag found -- please only use unique tags! %v\n", tag)
			}
			sf.Fields[tag] = styf
			atags := struf.Tag.Get("alt")
			if atags != "" {
				atag := strings.Split(atags, ",")

				for _, tg := range atag {
					tag = StyleEffTag(tg, outerTag)
					sf.Fields[tag] = styf
				}
			}
			inhs := struf.Tag.Get("inherit")
			if inhs == "true" {
				sf.Inherits = append(sf.Inherits, styf)
			}
			if vf.Kind() == reflect.Struct && vf.Type() == valtyp {
				sf.Units = append(sf.Units, styf)
			}
		})
	return
}

// Inherit copies all the values from par to obj for fields marked
// as "inherit" -- inherited by default
func (sf *StyledFields) Inherit(obj, par interface{}) {
	pr := prof.Start("StyleFields.Inherit")
	for _, fld := range sf.Inherits {
		vf := fld.FieldValue(obj)
		pf := fld.FieldValue(par)
		vf.Elem().Set(pf.Elem()) // copy
	}
	pr.End()
}

// Style applies styles to the fields from given properties for given object
func (sf *StyledFields) Style(obj, par interface{}, props ki.Props) {
	pr := prof.Start("StyleFields.Style")
	hasPar := (par != nil)
	// fewer props than fields, esp with alts!
	for key, val := range props {
		if key[0] == '#' || key[0] == '.' || key[0] == ':' {
			continue
		}
		if vstr, ok := val.(string); ok {
			if len(vstr) > 0 && vstr[0] == '$' { // special case to use other value
				nkey := vstr[1:] // e.g., border-color has "$background-color" value
				if vfld, nok := sf.Fields[nkey]; nok {
					nval := vfld.FieldValue(obj).Elem().Interface()
					if fld, fok := sf.Fields[key]; fok {
						fld.FromProps(sf.Fields, obj, par, nval, hasPar)
						continue
					}
				}
				fmt.Printf("gi.StyledFields.Style: redirect field not found: %v for key: %v\n", nkey, key)
			}
		}
		fld, ok := sf.Fields[key]
		if !ok {
			// note: props can apply to Paint or Style and not easy to keep those
			// precisely separated, so there will be mismatch..
			// log.Printf("SetStyleFields: Property key: %v not among xml or alt field tags for styled obj: %T\n", key, obj)
			continue
		}
		fld.FromProps(sf.Fields, obj, par, val, hasPar)
	}
	pr.End()
}

// ToDots runs ToDots on unit values, to compile down to raw pixels
func (sf *StyledFields) ToDots(obj interface{}, uc *units.Context) {
	pr := prof.Start("StyleFields.ToDots")
	for _, fld := range sf.Units {
		uv := fld.UnitsValue(obj)
		uv.ToDots(uc)
	}
	pr.End()
}

////////////////////////////////////////////////////////////////////////////////////////
//   StyledField

// StyledField contains the relevant data for a given stylable field in a struct
type StyledField struct {
	Field   reflect.StructField
	NetOff  uintptr       `desc:"net accumulated offset from the overall main type, e.g., Style"`
	Default reflect.Value `desc:"value of default value of this field"`
}

// FieldValue returns a reflect.Value for a given object, computed from NetOff
// -- this is VERY expensive time-wise -- need to figure out a better solution..
func (sf *StyledField) FieldValue(obj interface{}) reflect.Value {
	ov := reflect.ValueOf(obj)
	f := unsafe.Pointer(ov.Pointer() + sf.NetOff)
	nw := reflect.NewAt(sf.Field.Type, f)
	return kit.UnhideIfaceValue(nw).Elem()
}

// UnitsValue returns a units.Value for a field, which must be of that type..
func (sf *StyledField) UnitsValue(obj interface{}) *units.Value {
	ov := reflect.ValueOf(obj)
	uv := (*units.Value)(unsafe.Pointer(ov.Pointer() + sf.NetOff))
	return uv
}

// FromProps styles given field from property value val, with optional parent object obj
func (fld *StyledField) FromProps(fields map[string]*StyledField, obj, par, val interface{}, hasPar bool) {
	vf := fld.FieldValue(obj)
	var pf reflect.Value
	if hasPar {
		pf = fld.FieldValue(par)
	}
	prstr := ""
	switch prtv := val.(type) {
	case string:
		prstr = prtv
		if prtv == "inherit" {
			if hasPar {
				vf.Set(pf)
				fmt.Printf("StyleField %v set to inherited value: %v\n", fld.Field.Name, pf.Interface())
			}
			return
		}
		if prtv == "initial" {
			vf.Set(fld.Default)
			// fmt.Printf("StyleField set tag: %v to initial default value: %v\n", tag, df)
			return
		}
	}

	// todo: support keywords such as auto, normal, which should just set to 0

	npvf := kit.NonPtrValue(vf)

	vk := npvf.Kind()
	vt := npvf.Type()

	if vk == reflect.Struct { // only a few types -- todo: could make an interface if needed
		if vt == reflect.TypeOf(Color{}) {
			vc := vf.Interface().(*Color)
			switch prtv := val.(type) {
			case string:
				if idx := strings.Index(prtv, "$"); idx > 0 {
					oclr := prtv[idx+1:]
					prtv = prtv[:idx]
					if vfld, nok := fields[oclr]; nok {
						nclr, nok := vfld.FieldValue(obj).Interface().(*Color)
						if nok {
							vc.SetColor(nclr) // init from color
							fmt.Printf("StyleField %v initialized to other color: %v val: %v\n", fld.Field.Name, oclr, vc)
						}
					}
				}
				err := vc.SetString(prtv, nil)
				if err != nil {
					log.Printf("StyleField: %v\n", err)
				}
			case color.Color:
				vc.SetColor(prtv)
			}
			return
		} else if vt == reflect.TypeOf(units.Value{}) {
			uv := vf.Interface().(*units.Value)
			switch prtv := val.(type) {
			case string:
				uv.SetFromString(prtv)
			case units.Value:
				*uv = prtv
			default: // assume Px as an implicit default
				valflt := reflect.ValueOf(val).Convert(reflect.TypeOf(float32(0.0))).Interface().(float32)
				uv.Set(valflt, units.Px)
			}
			return
		}
		return // no can do any struct otherwise
	} else if vk >= reflect.Int && vk <= reflect.Uint64 { // some kind of int
		if prstr != "" {
			tn := kit.FullTypeName(fld.Field.Type)
			if kit.Enums.Enum(tn) != nil {
				kit.Enums.SetEnumValueFromStringAltFirst(vf, prstr)
			} else {
				fmt.Printf("gi.StyleField: enum name not found %v for field %v\n", tn, fld.Field.Name)
			}
			return
		} else {
			// somehow this doesn't work:
			// vf.Set(reflect.ValueOf(val))
			ival, ok := kit.ToInt(val)
			if !ok {
				log.Printf("gi.StyledField.FromProps: for field: %v could not convert property to int: %v %T\n", fld.Field.Name, val, val)
			} else {
				kit.Enums.SetEnumValueFromInt64(vf, ival)
			}
			return
		}
	}
	// again, this should work but does not:
	// vf.Set(reflect.ValueOf(val).Convert(reflect.TypeOf(vt)))
	kit.SetRobust(vf.Interface(), val)
}

////////////////////////////////////////////////////////////////////////////////////////
//   WalkStyleStruct

// this is the function to process a given field when walking the style
type WalkStyleFieldFun func(struf reflect.StructField, vf reflect.Value, tag string, baseoff uintptr)

// WalkStyleStruct walks through a struct, calling a function on fields with
// xml tags that are not "-", recursively through all the fields
func WalkStyleStruct(obj interface{}, outerTag string, baseoff uintptr, fun WalkStyleFieldFun) {
	otp := reflect.TypeOf(obj)
	if otp.Kind() != reflect.Ptr {
		log.Printf("gi.WalkStyleStruct -- you must pass pointers to the structs, not type: %v kind %v\n", otp, otp.Kind())
		return
	}
	ot := otp.Elem()
	if ot.Kind() != reflect.Struct {
		log.Printf("gi.WalkStyleStruct -- only works on structs, not type: %v kind %v\n", ot, ot.Kind())
		return
	}
	vo := reflect.ValueOf(obj).Elem()
	for i := 0; i < ot.NumField(); i++ {
		struf := ot.Field(i)
		if struf.PkgPath != "" { // skip unexported fields
			continue
		}
		tag := struf.Tag.Get("xml")
		if tag == "-" {
			continue
		}
		ft := struf.Type
		// note: need Addrs() to pass pointers to fields, not fields themselves
		// fmt.Printf("processing field named: %v\n", struf.Nm)
		vf := vo.Field(i)
		vfi := vf.Addr().Interface()
		if ft.Kind() == reflect.Struct && ft.Name() != "Value" && ft.Name() != "Color" {
			WalkStyleStruct(vfi, tag, baseoff+struf.Offset, fun)
		} else {
			if tag == "" { // non-struct = don't process
				continue
			}
			fun(struf, vf, outerTag, baseoff)
		}
	}
}

// todo:
// * need to be able to process entire chunks at a time: box-shadow: val val val

// manual method for getting a units value directly
func StyleUnitsValue(tag string, uv *units.Value, props ki.Props) bool {
	val, got := props[tag]
	if !got {
		return false
	}
	switch v := val.(type) {
	case string:
		uv.SetFromString(v)
	case float64:
		uv.Set(float32(v), units.Px) // assume px
	case float32:
		uv.Set(v, units.Px) // assume px
	case int:
		uv.Set(float32(v), units.Px) // assume px
	}
	return true
}
