// Copyright (c) 2018, Randall C. O'Reilly. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"fmt"
	"github.com/rcoreilly/goki/ki"
	"image"
	// "log"
	// "reflect"
)

////////////////////////////////////////////////////////////////////////////////////////
//  Node Widget

// signals that buttons can send
type NodeWidgetSignalType int64

const (
	// node was selected -- data is the node widget
	NodeSelected NodeWidgetSignalType = iota
	// node widget unselected
	NodeUnselected
	// collapsed node widget was opened
	NodeOpened
	// open node widget was collapsed -- children not visible
	NodeCollapsed
	NodeWidgetSignalTypeN
)

//go:generate stringer -type=NodeWidgetSignalType

// these extend NodeBase NodeFlags
const (
	// node is collapsed
	NodeFlagCollapsed NodeFlags = NodeFlagsN + iota
	// node is selected
	NodeFlagSelected
	// a full re-render is required due to nature of update event -- otherwise default is local re-render
	NodeFlagFullReRender
)

// mutually-exclusive button states -- determines appearance
type NodeWidgetStates int32

const (
	// normal state -- there but not being interacted with
	NodeWidgetNormalState NodeWidgetStates = iota
	// selected
	NodeWidgetSelState
	// in focus -- will respond to keyboard input
	NodeWidgetFocusState
	NodeWidgetStatesN
)

//go:generate stringer -type=NodeWidgetStates

// NodeWidget represents one node in the tree -- fully recursive -- creates
//  sub-nodes etc
type NodeWidget struct {
	WidgetBase
	SrcNode       ki.Ptr                   `desc:"Ki Node that this widget is viewing in the tree -- the source"`
	NodeWidgetSig ki.Signal                `json:"-",desc:"signal for node widget -- see NodeWidgetSignalType for the types"`
	StateStyles   [NodeWidgetStatesN]Style `desc:"styles for different states of the widget -- everything inherits from the base Style which is styled first according to the user-set styles, and then subsequent style settings can override that"`
	WidgetSize    Size2D                   `desc:"just the size of our widget -- our alloc includes all of our children, but we only draw us"`
}

// must register all new types so type names can be looked up by name -- e.g., for json
var KiT_NodeWidget = ki.Types.AddType(&NodeWidget{}, nil)

// important: do NOT assume kid is a NodeWidget unless absolutely necessary -- otherwise
// treat as generic gi.Node or Node2D, so others could subclass -- can make interface if needed

// set the source node that we are viewing
func (g *NodeWidget) SetSrcNode(k ki.Ki) {
	g.UpdateStart()
	if len(g.Children) > 0 {
		g.DeleteChildren(true) // todo: later deal with destroyed
	}
	g.SrcNode.Ptr = k
	k.NodeSignal().Connect(g.This, SrcNodeSignal) // we recv signals from source
	nm := "ViewOf_" + k.KiUniqueName()
	if g.Name != nm {
		g.SetName(nm)
	}
	kids := k.KiChildren()
	// breadth first -- first make all our kids, then have them make their kids
	for _, kid := range kids {
		g.AddNewChildNamed(nil, "ViewOf_"+kid.KiUniqueName()) // our name is view of ki unique name
	}
	for i, kid := range kids {
		vki, _ := g.KiChild(i)
		vk, ok := vki.(*NodeWidget)
		if !ok {
			continue // shouldn't happen
		}
		vk.SetSrcNode(kid)
	}
	g.UpdateEnd()
}

// function for receiving node signals from our SrcNode
func SrcNodeSignal(nwki, send ki.Ki, sig int64, data interface{}) {
	// todo: need a *node* deleted signal!  and children etc
	// track changes in source node
}

// is this node itself collapsed?
func (g *NodeWidget) IsCollapsed() bool {
	return ki.HasBitFlag64(g.NodeFlags, int(NodeFlagCollapsed))
}

// does this node have a collapsed parent? if so, don't render!
func (g *NodeWidget) HasCollapsedParent() bool {
	pcol := false
	g.FunUpParent(0, g.This, func(k ki.Ki, level int, d interface{}) bool {
		_, pg := KiToNode2D(k)
		if pg != nil {
			if ki.HasBitFlag64(pg.NodeFlags, int(NodeFlagCollapsed)) {
				pcol = true
				return false
			}
		}
		return true
	})
	return pcol
}

// is this node selected?
func (g *NodeWidget) IsSelected() bool {
	return ki.HasBitFlag64(g.NodeFlags, int(NodeFlagSelected))
}

func (g *NodeWidget) GetLabel() string {
	label := ""
	if g.IsCollapsed() { // todo: temp hack
		label = "> "
	} else {
		label = "v "
	}
	label += g.SrcNode.Ptr.KiName()
	return label
}

// todo mutex unselect all other nodes
func (g *NodeWidget) SelectNode() {
	if !g.IsSelected() {
		g.UpdateStart()
		ki.SetBitFlag64(&g.NodeFlags, int(NodeFlagSelected))
		g.NodeWidgetSig.Emit(g.This, int64(NodeSelected), nil)
		fmt.Printf("selected node: %v\n", g.Name)
		g.UpdateEnd()
	}
}

func (g *NodeWidget) UnselectNode() {
	if g.IsSelected() {
		g.UpdateStart()
		ki.ClearBitFlag64(&g.NodeFlags, int(NodeFlagSelected))
		g.NodeWidgetSig.Emit(g.This, int64(NodeUnselected), nil)
		fmt.Printf("unselectednode: %v\n", g.Name)
		g.UpdateEnd()
	}
}

func (g *NodeWidget) CollapseNode() {
	if !g.IsCollapsed() {
		g.UpdateStart()
		ki.SetBitFlag64(&g.NodeFlags, int(NodeFlagFullReRender))
		ki.SetBitFlag64(&g.NodeFlags, int(NodeFlagCollapsed))
		g.NodeWidgetSig.Emit(g.This, int64(NodeCollapsed), nil)
		fmt.Printf("collapsed node: %v\n", g.Name)
		g.UpdateEnd()
	}
}

func (g *NodeWidget) OpenNode() {
	if g.IsCollapsed() {
		g.UpdateStart()
		ki.SetBitFlag64(&g.NodeFlags, int(NodeFlagFullReRender))
		ki.ClearBitFlag64(&g.NodeFlags, int(NodeFlagCollapsed))
		g.NodeWidgetSig.Emit(g.This, int64(NodeOpened), nil)
		fmt.Printf("opened node: %v\n", g.Name)
		g.UpdateEnd()
	}
}

func (g *NodeWidget) AsNode2D() *Node2DBase {
	return &g.Node2DBase
}

func (g *NodeWidget) AsViewport2D() *Viewport2D {
	return nil
}

func (g *NodeWidget) InitNode2D() {
	g.ReceiveEventType(MouseDownEventType, func(recv, send ki.Ki, sig int64, d interface{}) {
		_, ok := recv.(*NodeWidget)
		if !ok {
			return
		}
		// todo: specifically on down?  needed this for emergent
	})
	g.ReceiveEventType(MouseUpEventType, func(recv, send ki.Ki, sig int64, d interface{}) {
		fmt.Printf("button %v pressed!\n", recv.PathUnique())
		ab, ok := recv.(*NodeWidget)
		if !ok {
			return
		}
		ab.SelectNode()
	})
	g.ReceiveEventType(KeyTypedEventType, func(recv, send ki.Ki, sig int64, d interface{}) {
		ab, ok := recv.(*NodeWidget)
		if !ok {
			return
		}
		kt, ok := d.(KeyTypedEvent)
		if ok {
			fmt.Printf("node widget key: %v\n", kt.Key)
			switch kt.Key {
			case "enter", "space", "return":
				ab.SelectNode()
			case "ctrl-f", "f", "right_arrow":
				ab.OpenNode()
			case "ctrl-b", "b", "left_arrow":
				ab.CollapseNode()
			}
		}
	})
}

var NodeWidgetProps = []map[string]interface{}{
	{
		"border-width":  "0px",
		"border-radius": "0px",
		"padding":       "1px",
		"margin":        "1px",
		// "font-family":         "Arial", // this is crashing
		"font-size":        "24pt", // todo: need to get dpi!
		"text-align":       "left",
		"color":            "black",
		"background-color": "#FFF", // todo: get also from user, type on viewed node
	}, { // selected
		"background-color": "#CFC", // todo: also
	}, { // focused
		"background-color": "#CCF", // todo: also
	},
}

func (g *NodeWidget) Style2D() {
	// we can focus by default
	ki.SetBitFlag64(&g.NodeFlags, int(CanFocus))
	// first do our normal default styles
	g.Style.SetStyle(nil, &StyleDefault, NodeWidgetProps[0])
	// then style with user props
	g.Style2DWidget()
	// now get styles for the different states
	for i := 0; i < int(NodeWidgetStatesN); i++ {
		g.StateStyles[i] = g.Style
		g.StateStyles[i].SetStyle(nil, &StyleDefault, NodeWidgetProps[i])
		g.StateStyles[i].SetUnitContext(&g.Viewport.Render, 0)
	}
	// todo: how to get state-specific user prefs?  need an extra prefix..
}

func (g *NodeWidget) Layout2D(iter int) {
	if iter == 0 {
		g.InitLayout2D()
		st := &g.Style
		pc := &g.Paint
		var w, h float64

		if g.HasCollapsedParent() {
			// g.AllocSize
			return // nothing
		}

		label := g.GetLabel()

		w, h = pc.MeasureString(label)
		if st.Layout.Width.Dots > 0 {
			w = ki.Max64(st.Layout.Width.Dots, w)
		}
		if st.Layout.Height.Dots > 0 {
			h = ki.Max64(st.Layout.Height.Dots, h)
		}
		w += 2.0*st.Padding.Dots + 2.0*st.Layout.Margin.Dots
		h += 2.0*st.Padding.Dots + 2.0*st.Layout.Margin.Dots

		g.WidgetSize = Size2D{w, h}

		if !g.IsCollapsed() {
			// we layout children under us
			for _, kid := range g.Children {
				_, gi := KiToNode2D(kid)
				if gi != nil {
					gi.Layout.AllocPos.Y = h
					gi.Layout.AllocPos.X = 20 // indent children -- todo: make a property
					h += gi.Layout.AllocSize.Y
				}
			}
		}
		g.Layout.AllocSize = Size2D{w, h}
	} else {
		g.GeomFromLayout()
	}

	// todo: test for use of parent-el relative units -- indicates whether multiple loops
	// are required
	g.Style.SetUnitContext(&g.Viewport.Render, 0)
	// now get styles for the different states
	for i := 0; i < int(NodeWidgetStatesN); i++ {
		g.StateStyles[i].SetUnitContext(&g.Viewport.Render, 0)
	}

}

func (g *NodeWidget) Node2DBBox() image.Rectangle {
	// we have unusual situation of bbox != alloc
	tp := g.Paint.TransformPoint(g.Layout.AllocPos.X, g.Layout.AllocPos.Y)
	ts := g.Paint.TransformPoint(g.WidgetSize.X, g.WidgetSize.Y)
	return image.Rect(int(tp.X), int(tp.Y), int(tp.X+ts.X), int(tp.Y+ts.Y))
}

func (g *NodeWidget) Render2D() {
	// g.DefaultGeom() // set win box from layout data

	// reset for next update
	ki.ClearBitFlag64(&g.NodeFlags, int(NodeFlagFullReRender))

	if g.HasCollapsedParent() {
		return // nothing
	}

	if g.IsSelected() {
		g.Style = g.StateStyles[NodeWidgetSelState]
	} else if g.HasFocus() {
		g.Style = g.StateStyles[NodeWidgetFocusState]
	} else {
		g.Style = g.StateStyles[NodeWidgetNormalState]
	}

	pc := &g.Paint
	rs := &g.Viewport.Render
	st := &g.Style
	pc.Font = st.Font
	pc.Text = st.Text
	pc.Stroke.SetColor(&st.Border.Color)
	pc.Stroke.Width = st.Border.Width
	pc.Fill.SetColor(&st.Background.Color)
	// g.DrawStdBox()
	pos := g.Layout.AllocPos.AddVal(st.Layout.Margin.Dots)
	sz := g.WidgetSize.AddVal(-2.0 * st.Layout.Margin.Dots)
	g.DrawBoxImpl(pos, sz, st.Border.Radius.Dots)

	pc.Stroke.SetColor(&st.Color) // ink color

	pos = g.Layout.AllocPos.AddVal(st.Layout.Margin.Dots + st.Padding.Dots)
	// sz := g.Layout.AllocSize.AddVal(-2.0 * (st.Layout.Margin.Dots + st.Padding.Dots))

	label := g.GetLabel()
	fmt.Printf("rendering: %v\n", label)

	pc.DrawStringAnchored(rs, label, pos.X, pos.Y, 0.0, 0.9)
}

func (g *NodeWidget) CanReRender2D() bool {
	if ki.HasBitFlag64(g.NodeFlags, int(NodeFlagFullReRender)) {
		return false
	} else {
		return true
	}
}

func (g *NodeWidget) FocusChanged2D(gotFocus bool) {
	// todo: good to somehow indicate focus
	// Qt does it by changing the color of the little toggle widget!  sheesh!
	g.UpdateStart()
	g.UpdateEnd()
}

// check for interface implementation
var _ Node2D = &NodeWidget{}
