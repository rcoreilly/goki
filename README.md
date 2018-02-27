# goki
Go language (golang) full strength tree structures (ki = tree in Japanese)

# Motivation

The **Tree** is the most powerful data structure in programming, and it underlies all the best tech, such as the WWW (the DOM is a tree structure), 3D scene graphs (now used for all 2D and 3D gui description in Qt), JSON, XML, SVG, filesystems, programs themselves, etc.  This is a prototype of Go with a powerful native Tree container type, that can support all of these things just by defining derived Node types.

Instead of LISP (a programming language built on the list data type), the key idea here is "Go with Trees" (GoKi) -- an awesome, simple programming language with an awesome builtin infrastructure for doing everything you typically need to do with Trees.

This will enable a purely Go-native 2D / 3D scenegraph based gui (on top of OpenGL, etc), which itself must have native support for rendering said Tree objects, thus enabling automatic visualization of this most powerful of data structures. This then enables rapid debugging and coding of Tree structures across all domains. The Tree, augmented with property flags and various other interface methods, can provide an entirely suitable GUI framework for many apps, all by itself. That is the core of a file navigator, program editor, type navigator, SVG browser, web debugger, etc. With a powerful Tree system, an optimized DOM system should be relatively straightforward, with built-in dirty-bit optimized update logic, etc. A lightweight, transparent, good-enough-for-many-apps native web browser should be well within scope, and a welcome alternative to the massive chromium project. IO is automatic: Tree <-> JSON, XML, etc would of course be built in.

The entire language itself should be representable as a Tree structure with various parsing and semantics attributes, and function hooks to handle all the tricky bits -- providing that infrastructure, which takes a Language Tree and turns it into a Parse Tree, for general use would facilitate all manner of other parsers, including all the JSON, XML, etc, and enable things like a native JavaScript interpreter, which any program can then use to support an interactive coding component on top of the compiled Go code. Maybe it isn't the fastest JavaScript in the world, but it is the most transparent: anyone can browse the entire parsing logic in its LanguageTree, and then see the resulting ParseTree output from any given input code -- all the sudden, parsing is fun again. Subsequent optimization and bytecode generation passes are all coded as Tree transformation functions that walk one tree and generate another, each of which can be visualized dynamically in a gui.

This is a vision that will take Go to the next level of fun and productivity. It sorely lacks a native GUI. Building one on top of native trees, with full reflection into the GUI scenegraph itself, will put all the power of the Qt framework into the much friendlier Go language. It is horrific seeing how much redundant PIMPL-hidden, glacially-compiling C++ code is in the current Qt framework: all to produce a very elegant QML declarative GUI that could be reproduced with an elegant, powerful Tree system with vastly less code. Qt is dead-ending with integer-based pixel coordinates, in a hiDPI world, while also abandoning native widgets in favor of the power of a full OpenGL scenegraph renderer. We can fast-forward to this same kind of solution without dragging along all the cruft, and without doing too much work at all with a native Tree that supports all the required functionality of a scenegraph, which again overlaps with everything needed for a DOM, etc.

In sort, write it once, in the core of the language, and provide the kind of modern infrastructure that enables people to build amazing things. This is the Go philosophy, and this seems like the obvious next step.

## Other benefits:

* The Tree can naturally support an ownership logic that could automate memory management and greatly minimize the need for garbage collection (GC). GC remains one of the oft-cited issues that prevents C / C++ folks from making the jump to Go. Thus, for many cases, GC could function as a kind of backup safety-net, with the Tree handling most of the heavy lifting.

* Another major issue with Go is support for generics. Generics are most important for containers, and Go's robust collection of native containers has minimized their need, and enabled the language to thrive despite their absence. Following this logic, adding a much more powerful container with appropriate native functionality will go even further toward mitigating the need for generics. If much of the program logic is transforming one tree into another, and defining new Node types etc, then you don't really need generics.

## Links

### GUI efforts

* Shiny (not much progress recently, only works on android?):  https://github.com/golang/go/issues/11818 https://github.com/golang/exp/tree/master/shiny

* Current plans for GUI based on OpenGL: https://docs.google.com/document/d/1mXev7TyEnvM4t33lnqoji-x7EqGByzh4RpE4OqEZck4

* Window events: https://github.com/skelterjohn/go.wde, Material gui https://github.com/skelterjohn/go.wde

* scenegraphs / 3D game engines: 
	+ https://github.com/g3n/engine
	+ https://github.com/oakmound/oak
	+ https://github.com/walesey/go-engine
