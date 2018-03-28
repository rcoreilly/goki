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
	n.FunDownMeFirst(0, nil, func(k Ki, level int, d interface{}) bool {
		mn := KiToMyNode(k)
	    mn.DoSomething()
		...
		return true // return value determines whether tree traversal continues or not
	})
}
```

Two other core features include:

* A `Signal` mechanism that allows nodes to communicate changes and other events to arbitrary lists of other nodes (similar to the signals and slots from Qt).

* `UpdateStart()` and `UpdateEnd()` functions that use an atomic int counter to wrap around code that changes the tree structure or contents -- only when the counter goes back to 0 does the highest affected node in the tree send an `Updated` signal.  This allows arbitrarily nested modifications to proceed independently, each wrapped in their own Start / End blocks, with the optimal minimal update signaling automatically computed.  Garbage collection is performed after these optimized tree updates, minimizing impact and need for unplanned GC interruptions, mitigating one of the major complaints about Go.  In effect, when predominantly using Ki trees, the GC becomes an automatic memory management framework with all the benefits of not having to worry about memory, and very minimal performance impact.

In the `GoGi` graphical interface system built on top of GoKi, the scene graph of Ki elements thus automatically drives minimal refresh updates, and the signaling framework supports gui event delivery and e.g., the "onclick" event signaling from the `Button` widget, etc.  In short, GoGi provides a complete interactive 2D and 3D GUI environment in native Go, in perhaps the fewest lines of code of any such system.  Part of this is the natural elegance of Go, but GoKi enhances that by providing the robust natural primitives needed to express all the GUI functionality.  Because GoGi is based around standard CSS styles, SVG rendering, and supports all the major HTML elements, it provides a lightweight, transparent, good-enough-for-many-apps native web browser.  This provides an potential alternative to chromium / electron for universal web-app deployment. 

GoKi also leverages the type reflection system in Go, to provide automatic JSON and XML-based I/O of tree structures -- there is a type registry that allows the types of the tree nodes to be saved by name, and then used at load time to reconstruct the full tree.  GoGi likewise can display any Ki-based tree using a tree-browser, and selection events there can easily be connected to a property-panel editor for viewing and editing the fields of the node structures.  Thus, GoGi provides a powerful automatic GUI for arbitrary data structures, with no additional GUI coding required!  Tags on the struct fields are used to customize the display and editor interfaces, supporting things like combo-boxes for enumerated (const int) types, spin boxes or sliders for value fields, toggles for bools, etc.  Furthermore, functions can be called using buttons or menus (uses go generate to extract directives from the function comments).  This paradigm was used in the *emergent* neural network simulation system, which implemented reflection in C++ and allowed the main scientific code to focus on the science, with a good automatic GUI coming along for free.

The automatic tree GUI in GoGi enables rapid debugging and coding of tree structures across all domains, and supports with very little additional code the following functionality: file navigator, program editor, type navigator, SVG browser, web inspector and debugger, etc. Furthermore, a GUI designer is trivially available, as is a basic SVG-based drawing system like Inkscape.  A bit of extra work can make those tools competitive with any available.

Another planned framework is a tree-based parsing system that replaces limited parser-generators (yacc / bison) with the power and speed typical of hand-written recursive descent parsers, but with a much easier-to-understand manifest tree-based grammar.  The Grammar tree generates a Parse tree based on input text, and this Parse tree can then be further processed in a highly flexible way by any further steps, creating further trees.  This system should allow virtually anyone to create a parser and see it operate in the GUI, making this typically-onerous task a fun and powerful way to support all manner of parsing needs, including native support for parsing popular languages such as JavaScript, and Go itself.  Maybe it won't be the fastest JavaScript in the world, but it is the most transparent: anyone can browse the entire parsing logic in its Language tree, and then see the resulting Parse tree output from any given input code -- all the sudden, parsing is fun!

In summary, the goal of GoKi is a fun, elegant, simple, efficient framework that handles many common needs by leveraging the power of the most important data structure: trees.  Along with the extremely well-thought-out nature of the Go language, it aims to provide the most esthetically pleasing programming experience possible!  Like a stroll through a Japanese garden, everything is perfectly placed to provide an elegant, beautiful conceptual landscape.  

The contrast with the current state of C++ is striking.  The C++ language is undergoing rapid changes, but its standard library remains one of the most ugly and poorly named available.  The wonderful Qt GUI framework is now directly in conflict with the new standard library in many ways, and over the years Qt has accumulated no less than 3 entirely different widget sets, and is burdened with the binary compatibility requirements of C++, requiring everything to be written *four* different times (once in the public .h header, again in the main .cpp, and then two more times in the private implementation (PIMPL) .h and .cpp files).  It takes nearly an hour to compile, and the project is suffering from major problems with its CI infrastructure due to this.  The new Qt3D system is powerful, but has duplicative front and backend scene graphs and is nearly impossible to understand or debug.  In short, coming from this world, the future was looking increasingly bleak, and programming in this environment was unpleasant, inelegant, slow, and unsatisfying.

The other major existing alternatives also have significant flaws.  Although Python is ubiquitous and powerful, it has major limitations from being an interpreted and weakly typed language.  And its object-oriented syntax seems awkward and inelegant.  Javascript likewise suffers from excessive flexibility, etc.  In short, although necessarily subjective, I happen to agree with all of the major design decisions in the Go language, and in practice it has even exceeded my expectations.  Many other people agree, and it is rapidly increasing in popularity.  Hopefully, having a powerful Go-centric GUI system will help propel it even further!

# Links

GoKi borrows a lot of ideas and experience from *emergent* and the TA (type access) system: https://grey.colorado.edu/emergent -- TA is basically equivalent to the `reflect` reflection system in Go, and it provides the same kind of generic access to classes in C++ for IO, GUI, etc as reflect does in Go.  So far, it seems that the ki package can replicate much of the 67,438 LOC in emergent's `ta_core` directory using a mere 1,896 LOC!

