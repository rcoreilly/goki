# goki
Go language (golang) full strength tree structures (ki = tree in Japanese)

[![Go Report Card](https://goreportcard.com/badge/github.com/rcoreilly/goki)](https://goreportcard.com/report/github.com/rcoreilly/goki)
[![GoDoc](https://godoc.org/github.com/rcoreilly/goki?status.svg)](http://godoc.org/github.com/rcoreilly/goki)

# Code Map

* `package ki` -- core `Ki` interface (`ki.go`) and `Node` struct (`node.go`), plus other supporting players
* `package gi` -- GoGi Graphical Interface based on 2D and 3D scenegraph using Ki trees
* `package ip` -- TBD: Interactive Parsing system
* `package di` -- TBD: Drawing / Design Interface based on gi -- Inkscape / GUI designer all in one
* `package wi` -- TBD: Web Interface based on gi

# Motivation

*Note: the following is an attempt at rationalizing why I'm ~wasting~ investing so much of my time reinventing every wheel.  Basically, esthetics matter.  Programming is an emotionally-driven creative act, and things like elegance and beauty matter.*

The **Tree** is the most powerful data structure in programming, and it underlies all the best tech, such as the WWW (the DOM is a tree structure), scene graphs for 3D and 2D graphics systems, JSON, XML, SVG, filesystems, programs themselves, etc.  GoKi provides a powerful tree container type, that can support all of these things just by embedding and extending the `Node` struct type that implements the `Ki` (Ki = Tree in Japanese) interface.

Much like LISP (a programming language built around the list data type), the key idea here is to create a comprehensive ecosystem for Go built around Trees (GoKi) -- an awesome, simple programming language with an awesome infrastructure for doing everything you typically need to do with Trees.

The goal of GoKi is to create a minimalist, elegant, and powerful environment (like Go itself) where the tree-based primitives are used to simplify otherwise complex operations.  Similar to MATLAB and matricies, you can perform major computational functions using just a few lines of GoKi code.  As is always the case in programming, using the right data structure that captures the underlying structure of the problem is essential, and in many cases, that structure is a tree.  Of necessity, much existing code incorporates tree structures, but the goal of GoKi is to provide a set of carefully thought-out primitive operations that effectively form a new mental basis set for programming.

For example, GoKi provides functions that traverse the tree in all the relevant ways ("natural" me-first, breadth-first, depth-first, etc) and take a `func` function argument, so you can do things like this:

``` go
func (n *MyNode) DoSomethingOnMyTree() {
	n.FuncDownMeFirst(0, nil, func(k Ki, level int, d interface{}) bool {
		mn := KiToMyNode(k)
	    mn.DoSomething()
		...
		return true // return value determines whether tree traversal continues or not
	})
}
```

Three other core features include:

* A `Signal` mechanism that allows nodes to communicate changes and other events to arbitrary lists of other nodes (similar to the signals and slots from Qt).

* `UpdateStart()` and `UpdateEnd()` functions that wrap around code that changes the tree structure or contents -- these automatically and efficiently determine the highest level node that was affected by changes, and only that highest node sends an `Updated` signal.  This allows arbitrarily nested modifications to proceed independently, each wrapped in their own Start / End blocks, with the optimal minimal update signaling automatically computed.

* `ConfigChildren` uses a list of types and names and performs a minimal, efficient update of the children of a node to configure them to match (including no changes if already configured accordingly).  This is used during loading from JSON, and extensively in the `GoGi` GUI system to efficiently re-use existing tree elements.  There is often complex logic to determine what elements need to be present in a Widget, so separating that out from then configuring the elements that actually are present is very efficient and simplifies the code.

In addition, Ki nodes support a general-purpose `Props` property `map`, and the `kit` package provides a `TypeRegistry` and an `EnumRegistry`, along with various `reflect` utilities, to enable fully-automatic saving / loading of Ki trees from JSON or XML, including converting const int (enum) values to / from strings so those numeric values can change in the code without invalidating existing files.

## GoGi

The first and most important application of GoKi is the `GoGi` graphical interface system, in the `gi` package.  The scene graph of Ki elements automatically drives minimal refresh updates, and the signaling framework supports gui event delivery and e.g., the "onclick" event signaling from the `Button` widget, etc.

In short, GoGi provides a complete interactive 2D and 3D GUI environment in native Go, in perhaps the fewest lines of code of any such system.  Part of this is the natural elegance of Go, but GoKi enhances that by providing the robust natural primitives needed to express all the GUI functionality.  Because GoGi is based around standard CSS styles, SVG rendering, and supports all the major HTML elements, it (will) provide a lightweight, transparent, good-enough-for-many-apps native web browser.  This (will) provide an potential alternative to chromium / electron for universal web-app deployment. 

GoGi also leverages the type reflection system in Go, to provide an automatic GUI representation of all the Go native data structures (`struct` `slice` `map`, and basic data types), in addition to a tree browser of `Ki` elements.  This means you can use GoGi to inspect and edit Go data structures automatically with no additional coding -- a minimal first-order GUI that can be enhanced with relevant custom GUI elements as needed.  Tags on the struct fields are used to customize the display and editor interfaces, and methods can be called using buttons or menus.  This paradigm was used in the *emergent* neural network simulation system, which implements reflection in C++ and allows the main scientific code to focus on the science, with a good automatic GUI coming along for free.

Right now you can use the GoGi self-reflection to design GUI layouts, and save / load them from JSON formats, capturing some of the basic functionality from Qt creator and the QtQuick declarative framework.  More sophisticated designer functionality should be straightforward to add.  Plans include extending this to a lightweight SVG-based drawing program like Inkscape -- this functionality would then be available for e.g., scientific graphing apps etc to customize the graphs.

## Parsing made fun?

Another planned framework is a GUI and tree-based parsing system for generating and manipulating abstract syntax trees (AST), to make it (relatively) fun and easy (really!?) to write arbitrary parsing systems. This can provide an interface to the native Go AST system, and will be used for parsing HTML, CSS, JavaScript, etc. for the planned browser based on GoKi.

## Summary

In summary, the goal of GoKi is a fun, elegant, simple, efficient framework that handles many common needs by leveraging the power of the most important data structure: trees.  Along with the extremely well-thought-out nature of the Go language, it aims to provide the most esthetically pleasing programming experience possible!  Like a stroll through a Japanese garden, everything is perfectly placed to provide an elegant, beautiful conceptual landscape.  

The contrast with the current state of C++ is striking.  The C++ language is undergoing rapid changes, but its standard library remains one of the most ugly and poorly named available.  The wonderful Qt GUI framework is now directly in conflict with the new standard library in many ways, and over the years Qt has accumulated no less than 3 entirely different widget sets, and is burdened with the binary compatibility requirements of C++, requiring everything to be written *four* different times (once in the public .h header, again in the main .cpp, and then two more times for the private implementation (PIMPL)).  It takes nearly an hour to compile, compared to the few seconds for GoGi.   In short, coming from this world, the future was looking increasingly bleak, and programming in this environment was unpleasant, inelegant, slow, and unsatisfying.

The other major existing alternatives also have significant flaws.  Although Python is ubiquitous and powerful, it has major limitations from being an interpreted and weakly typed language.  And its object-oriented syntax seems awkward and inelegant (IMO).  Javascript likewise suffers from excessive flexibility, etc.  In short, although necessarily subjective, I happen to agree with all of the major design decisions in the Go language, and in practice it has even exceeded my expectations.  Many other people agree, and it is rapidly increasing in popularity.  Hopefully, having a powerful Go-centric GUI system will help propel it even further!

# Links

GoKi borrows a lot of ideas and experience from *emergent* and the TA (type access) system: https://grey.colorado.edu/emergent -- TA is basically equivalent to the `reflect` reflection system in Go, and it provides the same kind of generic access to classes in C++ for IO, GUI, etc as reflect does in Go.  So far, it seems that the ki package can replicate much of the 67,438 LOC in emergent's `ta_core` directory using a mere 1,896 LOC!

