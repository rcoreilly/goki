# gogi
Part of the GoKi Go language (golang) full strength tree structure system (ki = tree in Japanese)

`package gogi` -- scenegraph based 2D and 3D GUI / graphics interface (Gi) in Go

GoDoc documentation: https://godoc.org/github.com/rcoreilly/goki/gogi


# Code Map


# Design notes

* `GiNode` base node 
    + Geom / Transform wrt its *Parent* coord system, and then provides its own Geom wrt its children -- one option is to renormalize with proper aspect ratio, height = 1, width = aspect ratio (or maybe the reverse, whichever makes most sense), but it can be ANYTHING, including a pass-through from parent as another supported default option
    + 2D-based children just use x,y but have full x,y,z coords generically for all
    + events (`MouseEvent`, etc) are specifically connected using `Signal` system from a parent `Window` -- not broadcast generically -- simple method to set that up -- automatically finds parent window etc.
	+ maybe want to have a basic 2D and 3D bifurcation KiG2D, KiG3D -- within one RenderPlane, you can only have either 2D or 3D nodes, but not both?  That probably makes sense.

* `RenderPlane` node provides an `Image` that sub-nodes can render into, and it caches the results and composts them up onto any parent RenderPlanes, with geom xform etc -- each KiG node probably caches its immediate parent RenderPlane -- akin to Surface in Qt -- prefer renderplane name?
	+ `RenderPlane2D` -- supports 2D rendering via svg, etc
	+ `RenderPlane3D` -- supports 3D rendering -- by default via OpenGL but good to keep that general
	+ basically a node ONLY really cares about its parent renderplane -- that is the sum total of its context dependency -- renderplanes can be infinitely nested and they only care about their parent renderplanes, supporting basic composting, etc via Image
	+ may need to support some masking stuff etc.  depth order is just order of children?  seems best unless really need something else

* `Window` node is a special RenderPlane that grounds out into actual window -- each node can find their parent window -- this is an OpenGL render target as well and provides all the necessary stuff for that -- can only have one window parent -- see RenderPlane
    + Is the source of events -- has all the Signal objects for each type of event, and nodes connect directly to these to get those events
    + `WindowOffscreen` for offscreen rendering?  want to support that easily as a root of a scenegraph -- could just be the parent RenderPlane but some stuff for events etc will be looking for a window so we want to support that..

* `Layout` nodes (2D): organize layout of sub-items in various ways..  can support iterative layout 

* `Text2D` 2d text renderer using std libs (freetype etc) -- include 2D text in 3D world via RenderPlane2D embedded within a RenderPlane3D

* `Widget` node uses SVG code plus styles in the Props to render 2D GUI interface elements -- draw is all dynamic and exclusively SVG -- need to figure out styling stuff etc

* `Mesh` 3D nodes of various sorts based on standard types in Coin3D and Qt3D
	+ nodes for creating the opengl rendering scripts (forget the name of it) -- see Qt3D -- this stuff is kind of a pain but also powerful..

* `HTML` 2D nodes that basically replicate the DOM!

* `TreeView`, `TreeModel` etc -- see about the model / view stuff from Qt

