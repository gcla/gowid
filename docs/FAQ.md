# FAQ
Gowid is a new Go package, so this is my best guess at some useful tips and tricks.

## What is the difference between `RenderBox`, `RenderFlowWith` and `RenderFixed`?

This concept comes directly from [urwid](http://urwid.org/manual/widgets.html#box-flow-and-fixed-widgets). Each widget supports being rendered in one or more of these three modes:

- Box
- Flow
- Fixed

A box widget is a widget that can render itself given a width and a height i.e. #columns and #rows. It should render a canvas of the correct size. You can render a box widget by calling its `Render()` function with a `RenderBox` struct for the `size` argument e.g. `RenderBox{C: 20, R:16}`. Gowid will always render the root of the widget hierarchy with a `RenderBox` `size` argument, so the root widget should be a box widget.

A flow widget is a widget that can render itself given a width only. The idea is that the widget itself should determine how many rows it requires. A good example of this is a `text.Widget`. If its given fewer columns in which to render, it might need to build a canvas with more rows. A `listbox.Widget` renders its children with a `RenderFlowWith{C:...}` argument and lets each child determine how many lines it needs. You can see this in action by running `gowid-fib` and paging down a few times until the numbers are so long that each scrolls onto the next line.

A fixed widget is a widget that will render itself without any guidance about the width and height. For example, a `checkbox.Widget` can render itself this way - it will make a 3x1 canvas which contains the text `[x]`. 

When a container widget, like a `pile.Widget` or `columns.Widget` renders its children, it will use one of these types of size arguments for each child. Sometimes the child widget may not support being rendered with a particular size type. For example, a fixed widget won't automatically expand to accommodate the size given with `RenderBox`. Gowid provides adapter widgets to let you choose how your application should handle this. `boxadadapter.Widget` is initialized with a child widget and an integer that means number-of-rows. The child widget should be a box widget. `boxadapter` allows it to be rendered in flow mode e.g. to be used in a `listbox.Widget`. When `boxadapter.Widget` renders its child, it turns its flow size into a box size by setting the number of rows to render from its initialization parameter. Another option is `vpadding.Widget`. It is initialized with a child widget, an alignment, and a "subsize" that tells the widget how to transform its size argument when rendering its child. A `vpadding.Widget` can turn a box size into a flow size, render its child in flow mode, and then align the rendered child within a canvas of the right size determined by the box, potentially chopping lines from the top and bottom if the child is too large. 

## How does Gowid use goroutines? How can I stay thread-safe?

A gowid app is typically launched with a line of code like this:

```go
	app.SimpleMainLoop()
```
That function does the following:

 1. Starts a goroutine to collect events from tcell. These are pushed into a gowid channel for tcell events.
 2. Enters a loop that runs a `select` on three channels:
 - tcell events channel
 - run-after-render channel
 - quit channel
 3. The loop is terminated by an event on the quit channel. Then gowid will tell tcell to stop sending events and wait for the tcell event-collecting goroutine to stop.

Example events that appear on the tcell event channel are key-presses, mouse-clicks and terminal changes like a resize. Gowid responds to user input by calling the root widget's `UserInput()` function. 

The quit channel receives an event when the gowid application calls `app.Quit()`. 

The run-after-render channel receives functions to be executed. A gowid application can send such a function by calling `app.Run()` and passing the function to call. On receipt of such a function, gowid will call the function, then redraw the widgets. Note that the function is executed by the main application goroutine - the one executing `app.SimpleMainLoop()` - so the `Run()` function will not race with any widget rendering or user-input processing. 

If your application starts other goroutines that might update the widgets' state or hierarchy, it is best to make those state changes in a function that is issued via `app.Run()`. For an example of this, see `github.com/gcla/gowid/examples/gowid-editor` - in particular code that runs on a timer and updates the editor's status bar.

## How do I write code to respond to a button click?

Here is an example of a callback issued in response to a button click:

```go
	rb.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		if rb.Selected {
			switch txt {
			case "256-Color":
				app.SetColorMode(gowid.Mode256Colors)
				[...elided for brevity...]
			case "Monochrome":
				app.SetColorMode(gowid.ModeMonochrome)
			}
			updateChartHolder(app.GetColorMode(), app)
		}
	}})
```

This is based on the example `github.com/gcla/gowid/examples/gowid-palette`. 

These callbacks are run in the main gowid goroutine, within the call stack that starts with the root widget's `UserInput()`. 

The `OnClick()` function takes an `IWidgetChangedCallback`. A simple implementation of that is `WidgetCallback`. To satisfy the interface, you need an ID ("cb" here) and a function that is called with the app and the widget issuing the callback. The ID is present so you can easily remove the callback later if necessary - you just supply the ID, so it must be comparable. Note that if it's more convenient, you can just exploit Go's scoping rules to refer to and capture the callback-issuing widget by its name in the outer scope i.e. "rb". 

## Why do all your interfaces start with "I" and not all end in "er"?

When I started writing gowid, I didn't appreciate or understand the convention. For official discussion, you can read this - https://golang.org/doc/effective_go.html#interface-names. When programming I found I wanted a visible way to distinguish arguments to functions - interfaces or values. Using the old I-prefix made that simple for me. As time passed I could see the elegance of small simple interfaces that "do things", hence `Doer()`, of limiting functions with receivers and instead using free functions. But I haven't gone back to try to retro-fit to that new appreciation. So as gowid stands, interfaces start with an I. One of the most important gowid interfaces is `IWidget`. In light of a better understanding of this recommended naming convention, could `IWidget` be broken down? For example

```go
type InputHandler interface {
	UserInput(ev tcell.Event, size IRenderSize, focus bool, app IApp) bool
}
```
So perhaps each composite widget could store one or more `InputHandler`s, and defer input to the right child. But in figuring out which is the right child to accept the input (e.g. mouse click), a composite widget such as `columns.Widget` may need to figure out the rendered-size of each child and do some arithmetic on the input event coordinates. Now the children need to implement `RenderedSizeProvider` too. Some widgets fall back to calling `Render()` in order to compute the rendered canvas size (slower but simpler) so then they would need to also implement a `Renderer` interface. This gets the requirements close to the current `IWidget` interface. So my view has been that `IWidget` represents a reasonable interface needed to satisfy all widget processing, and it's still pretty small - only four methods.


## How do I write a new widget?
<a name="how-do-write-a-new-widget"></a>

The quick answer is you need to implement `IWidget`:

```go
type IWidget interface {
	Render(size IRenderSize, focus Selector, app IApp) ICanvas
	RenderSize(size IRenderSize, focus Selector, app IApp) IRenderBox
	UserInput(ev tcell.Event, size IRenderSize, focus Selector, app IApp) bool
	Selectable() bool
}
```

The `focus.Focus` argument will be true if your widget has the application focus; otherwise false (`focus.Selected` is intended for letting widgets like columns and pile highlight the subwidget that *would* be in focus if the outer widget was in focus).

Some helper types are provided for the common case. If your widget is always selectable, you can do this:

```go
type MyWidget struct {
   ...
   gowid.IsSelectable
}
```

Or if the opposite, use `NotSelectable`. If your widget will always reject user input, you can embed `RejectUserInput` which will provide a default implementation returning `false`.

Sometimes it's simpler to extend an existing widget. There are some examples of this e.g. `github.com/gcla/gowid/examples/gowid-tutorial4` - see `QuestionBox`. It chooses to embed an interface, `IWidget`, so that it can replace the implementation at runtime. It starts out as an `*edit.Widget` and then is replaced with a `*text.Widget`. `QuestionBox` provides its own `UserInput()` function but the embedded `IWidget` provides the other functions needed to satisfy the widget interface. But be careful and remember that Go does not have dynamic dispatch for structs. If you embed another widget, and that embedded widget's method is called, the receiver will be the embedded widget, not the containing widget. You can't "escape" back to the containing widget. I misunderstood this fundamental design feature when I started programming with Go.

Most gowid widgets are structured into two groups of functions. The essence of the widget is distilled into an interface that rests on `IWidget` - for example, here is a checkbox (in the `github.com/gcla/gowid/widgets/checkbox` package):

```go
// IWidget scoped to "checkbox" here.
type IWidget interface {
	gowid.IWidget
	IChecked
}
```

So checkbox is a widget that satisfies `checkbox.IChecked`. The `checkbox` package provides a free function that implements some of the expected widget functionality:

```go
func Render(w IChecked, size gowid.IRenderSize, focus Selector, app gowid.IApp) gowid.ICanvas
```

The checkbox widget's `Render()` method looks like this:

```go
func (w *Widget) Render(size gowid.IRenderSize, focus Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}
```
The goal is to make it easier to override parts of a widget, and use the default implementations for the rest. The rendering algorithm is contained in a free function that needs only an `IChecked`, so a new implementation that can be rendered similarly can call the same free function. If instead your new widget's `Render()` function did this:

```go
func (w *Widget) Render(size gowid.IRenderSize, focus Selector, app gowid.IApp) gowid.ICanvas {
	return w.IWidget.Render(size, focus, app)
}
```
then the embedded `IWidget` - presumably a `*gowid.Checkbox` - would call `Render()` with a `*gowid.Checkbox` as the receiver, so calling the free function `Render()` with the `IChecked` argument being a `*gowid.Checkbox` instead of your new type. The effect would be your new widget would render like the original checkbox.

## What is the difference between being selectable and handling user input?

A widget that is selectable is intended to be able to take the focus. For example, if a `listbox` is displaying a range of widgets, hitting the down arrow will make the `listbox` look for the next selectable widget to take the focus. If it can, it will skip any widget that is not selectable, like `text.Widget`s. But note that if *no* candidate widgets are selectable, then one will be chosen anyway. So your widget may still be rendered and provided user input (which you can then just reject).

A widget that returns `false` to a call to its `UserInput()` function is indicating that it has not handled the input. Gowid will then try to give the input to another widget. For example, if you hit down arrow in a `listbox`, it will first see if the currently focused listbox widget will accept the keypress. If that widget is an `edit.Widget`, it might move the cursor down a line inside its editing area. The `edit.Widget` will return `true`, and the `listbox` will not process the keypress further. But if, say, the focus `edit.Widget`is on the last line in its editing area, it can't move down a line, and will return `false` to the invocation of its `UserInput()`. The parent `listbox` will then accept the keypress and try to change its own focus widget. You can see this in action in `github.com/gcla/gowid/examples/gowid-widgets2` with left and right arrow keypresses.

## How do I color my widget?

You can use `styled.Widget` and pass it your own widget e.g.

```go
styled.New(myWidget, gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorCyan))
```
The second argument must be an `ICellStyler`:

```go
type ICellStyler interface {
	GetStyle(IRenderContext) (IColor, IColor, StyleAttrs)
}
```
You can use `MakePaletteEntry()` to construct one on the fly. The first argument is foreground color, the second background. If you register a palette when you initialize your `app`, you can refer to entries in that palette:
```go
styled.New(myWidget, gowid.MakePaletteRef("eyegrabbing"))
```
If you would like a different style to be used when your widget is in focus, then you can do this:

```go
styled.NewWithFocus(myWidget,
                    gowid.MakePaletteRef("boring")),
                    gowid.MakePaletteRef("eyegrabbing"))
)
```
You can easily just invert the colors on focus by using `styled.NewWithSimpleFocus()`. It simply defers to `NewWithFocus()` and uses `ColorInverter{s}` as its third argument where `s` is the second argument.

## How do I apply text styles like underline?

The `StyledAs` struct implements `ICellStyler`, providing no color preferences and the requested "style". So something like this:

```go
styled.New(myTextWidget, gowid.MakeStyledAs(gowid.StyleUnderline))
```

will do the job.

## Why do all the Set...() functions take an IApp Argument?

I decided that it could be useful for widgets to support issuing callbacks when properties change - so that you could tie together the behavior of groups of widgets. Those callbacks might also wish to interact with the app e.g. to run the `Quit()` function, or to inspect the state of the mouse buttons. So that decision necessitates having access to the `App`. To make access possible, there are a couple of other options:
- having a single, global app
- having every widget store a pointer to its app when initialized

Both aren't ideal, and put arbitrary restrictions on the applications using the widgets (though in practice, surely each application will only have one `App`?) Having a magic global `App` also seems to go against Go best practices such as those described in https://peter.bourgon.org/blog/2017/06/09/theory-of-modern-go.html. So I added an explicit `IApp` parameter to each function that might be connected to a subsequent use of the `App`, like calling `Quit()`. 

