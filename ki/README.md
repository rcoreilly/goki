# ki
Part of the GoKi Go language (golang) full strength tree structure system (ki = tree in Japanese)

`package ki` -- core `Ki` interface (`ki.go`) and `Node` struct (`node.go`), plus other supporting players.

**THIS IS ARCHIVED -- active version at: https://github.com/goki**

A `Ki` tree is recursively composed of Ki `Node` structs, in a one-Parent / multiple-Child structure.  The typical use is to embed Node in other structs that then implement specific tree-based functionality.  See other packages in GoKi for examples, and top-level README in GoKi for overall motivation and design.

# Code Map

* `kit` package: `kit.Types` `TypeRegistry` provides name-to-type map for looking up types by name, and types can have default properties. `kit.Enums` `EnumRegistry` provides enum (const int) <-> string conversion, including `bitflag` enums.  Also has robust generic `ki.ToInt` `ki.ToFloat` etc converters from `interface{}` to specific type, for processing properties, and several utilties in `embeds.go` for managing embedded structure types (e.g., ``TypeEmbeds` checks if one type embeds another, and `EmbeddedStruct` returns the embedded struct from a given struct, providing flexible access to elements of an embedded type hierarchy -- there are also methods for navigating the flattened list of all embedded fields within a struct).  Also has a `kit.Type` struct that supports saving / loading of type information using type names.

* `bitflag` package: simple bit flag setting, checking, and clearing methods that take bit position args as ints (from const int eunum iota's) and do the bit shifting from there

* `ki.go` = `Ki` interface for all major tree node functionality.

* `slice.go` = `ki.Slice []Ki` supports saving / loading of Ki objects in a slice, by recording the size and types of elements in the slice -- requires `ki.Types` type registry to lookup types by name.

* `props.go` = `ki.Props map[string]interface{}` supports saving / loading of property values using actual `struct` types and named const int enums, using the `kit` type registries.  Used for CSS styling in `GoGi`.

* `signal.go` = `Signal` that calls function on a receiver Ki objects that have been previously `Connect`ed to the signal -- also supports signal type so the same signal sender can send different types of signals over the same connection -- used for signaling changes in tree structure, and more general tree updating signals.

* `ptr.go` = `ki.Ptr` struct that supports saving / loading of pointers using paths.


# Go Language (golang) Notes (esp for people coming from C++)

general summary: https://github.com/golang/go/wiki/GoForCPPProgrammers

## Naming Conventions

https://golang.org/doc/effective_go.html#names

* Don't use Get* -- just use the name itself, when a "getter" is necessary

* Do use Set* -- setter + field name -- this is convention in Qt too

* Uppercase all Fields in general -- Strongly prefer to just deal with the fields directly without having to go through a getter method, and we want embedded objects and anyone else to be able to access those fields, and they also need to be saved / loaded through JSON, etc

* Above means that there can be conflicts for any interfaces that need to provide access to those fields, if you also want to export the fields in the struct implementation.  You have to be creative in coming up with different but equally sensible names in both the interface and the implementer.

* `Delete` instead of `Remove`  `Iface` for `interface{}`

* Unless it is a one-liner converter to a given type or value like Stringer, it can be challenging to name an interface and a base type for that interface differently.
	+ The Interface should generally be given priority, and have the cleaner name.  Base types are only typed relatively rarely at start of structs that embed them, so they are less important.
	+ One not-so-good idea: Add a capital I at the end of an interface when it is designed for derived types of a given base, e.g., `EventI` for structs that embed type `Event` -- I couldn't find anything about this in searching but somehow it doesn't seem like the "Go" way..
	
* It *IS* ok to have types and fields / members of the same name!  So EventType() EventType is perfectly valid and that's a relief :)

* It is hard to remember, but important, that everything will be prefixed by the package name for users of the package, so *don't put a redundant prefix on anything*

* Use `AsType()` for methods that give you that give you that concrete type from a struct (where it isn't a conversion, just selecting that view of it)

### Enums (const int)

* Use plural for enum type, instead of using a "Type" suffix -- e.g., `NodeSignals` instead of `NodeSignalType`, and in general use a consistent prefix for all enum values: NodeSignalAdded etc 

* my version of gp generate stringer utility generates conversions back from string to given type: https://github.com/rcoreilly/stringer

* ki.EnumRegister (kit.AddEnum) -- see kit/types.go -- adds a lot of important functionality to enums

## Struct structure

* Ki Nodes can be used as fields in a struct -- they function much like pre-defined Children elements, and all the standard FuncDown* iterators traverse the fields automatically.  The Ki Init function automatically names these structs with their field names, and sets the parent to the parent struct.

## Interfaces, Embedded types

* In C++ terms, an interface in go creates a virtual table -- anytime you need virtual functions, you must create an interface.
	+ I was confused about the nature of inheritance in Go for a long time, in part because the Go folks like to emphasize the idea that *there is no inheritance in Go*, and in general like to emphasize the differences from C++.  In fact, through anonymous embedding, you can inherit base-class definitions of interface methods, and *selectively override* (redefine) the ones you want to change for the sub-class, no problem!  The only thing that is actually missing is an automatic conversion of a derived class to a base class, which is probably actually a good thing -- you always need to know what type of thing you have.  But, the interface pointer actually does always know what type of thing it is, so if we keep a "This" interface pointer around on the struct, we can always call the proper derived versions of any interface methods.  Basically, you can really get all the functionality of C++ in Go that you want, no problem, plus all the flexibility of a more polymorphic system!
	+ An interface is the *only* way to create an equivalence class of types -- otherwise Go is strict about dealing with each struct in terms of the actual type it is, *even if it embeds a common base type*.
	
* Anonymous embedding a type at the start of a struct gives transparent access to the embedded types (and so on for the embedded types of that type), so it *looks* like inheritance in C++, but critically those inherited methods are *not* virtual in any way, and you must explicitly convert a given "derived" type into its base type -- you cannot do something like: `bt := derived.(*BaseType)` to get the base type from a derived object -- it will just complain that derived is its actual type, not a base type.  Although this seems like a really easy thing to fix in Go that would support derived types, it would require a (little bit expensive) dynamic (runtime) traversal of the reflect type info.  Those methods are avail in package `kit` as `TypeEmbeds` to check if one type embeds another, and `EmbededStruct` to get an embedded struct out of a struct that embeds it (at any level of depth).  In general it is much better to provide an explicit interface method to provide access to embedded structs that serve as base-types for a given level of functionality.  Or provide access to everything you might need via the interface, but just giving the base struct is typically much easier.

* Although the derived types have direct access to members and methods of embedded types, at any level of depth of embedding, the `reflect` system presents each embedded type as a *single field* as it is declared in the actual code.  Thus, you have to recursively traverse these embedded structs -- methods to flatten these field lists out by calling a given function on each field in turn are provided in `kit embed.go`: `FlatFieldsTypeFun` and `FlatFieldsValueFun`

* The Ki / Node framework may be seen as antithetical to the Go way of doing things, and it is at first glance more similar to C++ paradigms where a common base type is used to provide a comprehensive set of common functionality (e.g., the QObject in Qt).  However, Go also makes extensive use of special base types with considerable special, powerful functionality, in the form of slices and maps.  A Tree is likewise a kind of universal container object, and the only way to provide an appropriate general interface for it is by embedding a base type with the relevant functionality.
	+ You still have all the critical Go interface mix-and-match power: Any number of polymorphic interfaces can be implemented on top of the basic Ki tree structure, and multiple different anonymous embedded types can be used, so there is still none of the C++ rigidity and awkwardness of dealing with multiple inheritance, etc.
	+ The Ki interface itself is NOT designed to be implemented more than once -- there should be no reason to do so -- and all of the functionality that it does implement is fully interdependent and would not be sensibly separated into smaller interfaces.  Every attempt was made to separate out dissociable elements such as the `Ptr`, `Slice`, and `Signal`, and all the separate more general-purpose type management functionality in `kit`
	+ The `Node` object maintains its own Ki interface pointer in the `This` field -- another seeming C++ throwback, which has turned out to be absolutely essential to getting the whole enterprise to work.  First, an interface type is automatically a pointer, so even though it is declared as a `Ki` type, it is automatically a pointer to the struct.  Most importantly, this Ki pointer retains the true underlying identity of the struct, in whatever context it is being used -- by contrast, when you call any method with a specific receiver type, the struct becomes only that type, and loses all memory of its "true" underlying type.  For example, if you try to call another interface-based function within a given receiver-defined function, you'll *only* get the interface function for that specific type, not the one for the "true" underlying type.  Therefore, you need to call `ob.This.Function()` to get the one defined for the full type of the object.  This is probably the functionality most people would expect, especially coming from the world of C++ virtual functions.

## interface{} type: universal Variant

Go makes extensive use of the `interface{}` type which is effectively a built-in universal Variant type in Qt http://doc.qt.io/qt-5/qvariant.html -- you use type switches or conversions or the reflect system to figure out what kind of thing it actually is -- extremely powerful yet not very dangerous it seems. 

## Closures & anonymous functions

It is very convenient to use anonymous functions directly in the `FuncDown` (etc) and `Signal Connect` cases, but for performance reasons, it is important to be careful about capturing local variables from the parent function, thereby creating a *closure*, which creates a local stack to represent those variables.  In the case of FunDown / FunUp etc, the impact is minimized because the function is ONLY used during the lifetime of the outer function.  However, for `Signal Connect`, the function is itself saved and used later, so using a closure there creates extra memory overhead for each time the connection is created.  Thus, it is generally better to avoid capturing local variables in such functions -- typically all the relevant info can be made available in the recv, send, sig, and data args for the connection function.

## Notes on things that would perhaps be nice to change about Go..

As the docs state, you really have to use it for a while to appreciate all of the design decisions in the language, so these are very preliminary and subject to later reconsideration.. 

* Reference arg type -- the receiver arg (the "this" pointer in C++) has
  automagical ability to either be a pointer or a value depending on the
  function declaration, but other args do NOT have this flexibility.  It means
  that the caller has to do a little bit of extra work adding a & or not where
  necessary.  The counter-argument is presumably that this means that the
  caller KNOWS that by passing a pointer, the value of the arg will be
  changed, and that is probably a good thing.  The same point could be made
  for the receiver arg, but there I suppose the syntax would have been a bit
  clunky and confusing.

# TODO

* what about Kind == reflect.Interface fields in structs -- should they be set to zero?  probably..
* XML IO -- first pass done, but more should be in attr instead of full elements
* FindChildRecursive functions?
* port to better logging for buried errors, with debug mode: https://github.com/sirupsen/logrus


