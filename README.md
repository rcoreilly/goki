# goki
Go language (golang) full strength tree structures (ki = tree in Japanese)

GoDoc documentation: https://godoc.org/github.com/rcoreilly/goki/ki

# Code Map

* `package ki` -- core `Ki` interface (`ki.go`) and `Node` struct (`node.go`), plus other supporting players
* `package gi` -- Graphical Interface based on 2D and 3D scenegraph using Ki trees

# Motivation

The **Tree** is the most powerful data structure in programming, and it underlies all the best tech, such as the WWW (the DOM is a tree structure), 3D scene graphs (now used for 2D and 3D GUI description in Qt), JSON, XML, SVG, filesystems, programs themselves, etc.  GoKi provides a powerful tree container type, that can support all of these things just by embedding and extending the `Node` struct type that implements the `Ki` interface.

Instead of LISP (a programming language built on the list data type), the key idea here is "Go with Trees" (GoKi) -- an awesome, simple programming language with an awesome builtin infrastructure for doing everything you typically need to do with Trees.

This will enable a purely Go-native 2D / 3D scenegraph based GUI (on top of OpenGL, etc), which itself must have native support for rendering said trees, thus enabling automatic visualization of this most powerful of data structures. This then enables rapid debugging and coding of tree structures across all domains. The tree, augmented with inheritable property values and various other interface methods, can provide an entirely suitable GUI framework for many apps, all by itself. That is the core of a file navigator, program editor, type navigator, SVG browser, web debugger, etc. With a powerful tree system, an optimized DOM should be relatively straightforward, with built-in  optimized update logic, etc.  A lightweight, transparent, good-enough-for-many-apps native web browser should be well within scope, and a welcome alternative to the massive chromium project. IO is automatic: tree saving / loading in JSON is already implemented, and XML is coming..

Another planned framework is a tree-based parsing system that replaces limited parser-generators (yacc / bison) with the power and speed typical of hand-written recursive descent parsers, but with a much easier-to-understand manifest tree-based grammar.  The Grammar tree generates a Parse tree based on input text, and this Parse tree can then be further processed in a highly flexible way by any further steps, creating further trees.  This system should allow virtually anyone to create a parser and see it operate in the GUI, making this typically-onerous task a fun and powerful way to support all manner of parsing needs, including native support for parsing popular languages such as JavaScript, and Go itself.  Maybe it won't be the fastest JavaScript in the world, but it is the most transparent: anyone can browse the entire parsing logic in its Language tree, and then see the resulting Parse tree output from any given input code -- all the sudden, parsing is fun!

This is a vision that will take Go to the next level of fun and productivity. It sorely lacks a native GUI. Building one on top of powerful trees, with full reflection into the GUI scenegraph itself, will put the power of the Qt framework into the much friendlier Go language, in a fully native manner. It is horrific seeing how much redundant PIMPL-hidden, glacially-compiling C++ code is in the current Qt framework: all to produce a very elegant QML declarative GUI that could be reproduced with an elegant, powerful tree system with vastly less code. Qt is dead-ending with integer-based pixel coordinates, in a hiDPI world, while also abandoning native widgets in favor of the power of a full OpenGL scenegraph renderer.  We can fast-forward to this same kind of solution without dragging along all the cruft, and without doing too much work at all with a native tree that supports all the required functionality of a scenegraph, which again overlaps with everything needed for a DOM, etc.

In short, write it once, and provide the kind of modern infrastructure that enables people to build amazing things. This is the Go philosophy, and this seems like the obvious next step.

## Going Native?

If this project is successful, it might be interesting to consider a possible future version of the Go language that builds in native support for a tree, given the ubiquity and power of this structure.  It is unclear at this point if this would even be useful -- right now the current implementation is leveraging slices and maps and reflection and many other key features of Go -- it seems like a pretty efficient and minimal implementation, and it is unclear what real advantage could be obtained with a native object.  Here's some possibilities, however:

* The Tree can naturally support an ownership logic that could automate memory management and greatly minimize the need for garbage collection (GC). GC remains one of the oft-cited issues that prevents C / C++ folks from making the jump to Go. Thus, for many cases, GC could function as a kind of backup safety-net, with the Tree handling most of the heavy lifting.

* Another major issue with Go is support for generics. Generics are most important for containers, and Go's robust collection of native containers has minimized their need, and enabled the language to thrive despite their absence. Following this logic, adding a much more powerful container with appropriate native functionality will go even further toward mitigating the need for generics. If much of the program logic is transforming one tree into another, and defining new Node types etc, then you don't really need generics.

# Links

GoKi borrows a lot of ideas and experience from *emergent* and the TA (type access) system: https://grey.colorado.edu/emergent -- TA is basically equivalent to the `reflect` reflection system in Go, and it provides the same kind of generic access to classes in C++ for IO, GUI, etc as reflect does in Go.  So far, it seems that the ki package can replicate much of the 67,438 LOC in emergent's `ta_core` directory using a mere 1,896 LOC!

