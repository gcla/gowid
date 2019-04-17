# Gowid Tutorial

This tutorial closely follows the structure of the [urwid tutorial](http://urwid.org/tutorial/index.html). When a type is named without an explicit Go package specifier for brevity, then the package is `gowid`.

## Minimal Application

![enter image description here](https://drive.google.com/uc?export=view&id=12c4uZCWCynsusX6ELW9q08HB--YgKvFb)

Here is the traditional Hello World program, written for gowid. It displays "hello world" in the top left-hand corner of the terminal and will run until terminated with one of a few keypresses - Escape, Ctrl-c, q or Q. You can find this example at `github.com/gcla/gowid/examples/gowid-tutorial1` and run it via `gowid-tutorial1`. 

```go
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/text"
)

func main() {
	txt := text.New("hello world")
	app, _ := gowid.NewApp(gowid.AppArgs{View: txt})
	app.SimpleMainLoop()
}
```

 - txt is a `text.Widget` and renders strings to its canvas. This widget also supports rendering collections of text with style and color attributes attached, called markup. A `text.Widget` can render in urwid's "flow-mode", meaning its `Render()` function is provided with a number of columns, but with no specified number of rows. The widget will create a canvas with as many rows as it needs to render suitably. A widget that renders in urwid's "box-mode" will be given both a number of columns *and* a number of rows, and must create a canvas of that size.
 - - The second value returned from `NewApp` is an error, which you should check - though there's not much to do except exit gracefully.
 - The `app`'s `SimpleMainLoop()` function will hand control over to gowid. Terminal events will be handled by gowid, and in particular, user input will be processed by the hierarchy of widgets that constitute the user interface. Input is handed to the root widget provided as the `View` parameter to `NewApp()`. It may handle the event and it may also hand the event to its children. In this case, `text.Widget` is the root of the hierarchy, and it does not accept user input. Gowid will then hand the input to an `IUnhandledInput` which is provided in this case by `SimpleMainLoop()`. It checks for Escape, Ctrl-c, q or Q - if any are detected, the `app`'s `Quit()` function is called. After processing input, `SimpleMainLoop()` will then terminate.

## Global Input

![desc](https://drive.google.com/uc?export=view&id=1SgDht4cup0hhgMrwnQaE3e3lqvQvuKlC)

The second example features a function that processes user input. If the user does not press a quit key, the "hello world" message is updated to show what key was pressed. You can find this example at `github.com/gcla/gowid/examples/gowid-tutorial2` and run it via `gowid-tutorial2`. 

```go
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
)

var txt *text.Widget

func unhandled(app gowid.IApp, ev tcell.Event) bool {
	if evk, ok := ev.(*tcell.EventKey); ok {
		switch evk.Rune() {
		case 'q', 'Q':
			app.Quit()
		default:
			txt.SetText(fmt.Sprintf("hello world - %c", evk.Rune()), app)
		}
	}
	return true
}

func main() {
	txt = text.New("hello world")
	app, _ := gowid.NewApp(gowid.AppArgs{View: txt})
	app.MainLoop(gowid.UnhandledInputFunc(unhandled))
}
```
- The main loop is now provided an explicit function to process input that is not handled by any widget in the hierarchy. The `app`'s`MainLoop()` function expects a type that implements `IUnhandledInput`. The gowid type `UnhandledInputFunc` is a simple function adapter that allows use of a regular Go function.
- The function `unhandled()` is given the `app` and the user input in the form of a `tcell.Event`.  Gowid relies throughout on the Go package `tcell` and its representation of terminal input, both from the keyboard and the mouse. If the input provided is from the keyboard and is not one of the quit keys, the root `text.Widget` is updated to display the key that was pressed.

## Display Attributes

![desc](https://drive.google.com/uc?export=view&id=1D2rT70O_NPRGVFyFuHKZORt6jgVUn0Im)
![desc](https://drive.google.com/uc?export=view&id=1iSVmGLqUVbF4amSGWSGfHeJhQ4HU2-Rq)

The third example demonstrates the use of color. You can find this example at `github.com/gcla/gowid/examples/gowid-tutorial3` and run it via `gowid-tutorial3`. 

```go
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

func main() {
	palette := gowid.Palette{
		"banner": gowid.MakePaletteEntry(gowid.ColorBlack, gowid.NewUrwidColor("light gray")),
		"streak": gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorRed),
		"bg":     gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorDarkBlue),
	}

	txt := text.NewFromContentExt(
		text.NewContent([]text.TextContentSegment{
			text.StyledContent("hello world", gowid.MakePaletteRef("banner")),
		}), text.Options{
			Align: gowid.HAlignMiddle{},
		})

	map1 := styled.New(txt, gowid.MakePaletteRef("streak"))
	vert := vpadding.New(map1, gowid.VAlignMiddle{}, gowid.RenderFlow{})
	map2 := styled.New(vert, gowid.MakePaletteRef("bg"))
	app, _ := gowid.NewApp(gowid.AppArgs{
		View:    map2,
		Palette: palette,
	})
	app.SimpleMainLoop()
}
```
- Display attributes are defined and named in a `Palette`. The first argument to `MakePaletteEntry()` represents a foreground color and the second a background color. A similar gowid API allows for a third argument which represents text "styles" like underline and bold. 
- Gowid allows colors to be defined in a number of ways. Each color type must implement `IColor`, an interface which provides for a conversion to `tcell` color primitives (depending on the color mode of the terminal), ready for rendering on the terminal screen. 
	- `ColorBlack` is one of a set of predefined `TCellColor`s you can use. It trivially implements `IColor`.
	- `NewUrwidColor()` allows you to provide the name of a color that would be accepted by urwid and returns a `*UrwidColor`. You can read about urwid's color options [here](http://urwid.org/manual/displayattributes.html).
- You can pass the palette when initializing an `App`. Certain gowid widgets that use colors and styles can then refer to palette entries by name when rendering by using the `app`'s `GetCellStyler()` function and providing the name of the palette entry. For example, "hello world" appears in a called to `text.StyledContent()` which binds the display string together with a "cell styler" that comes from a reference to the palette. When this text widget is rendered, the string hello world is displayed in black text with a light gray background.
- You can also give `text.Widget` an alignment parameter. When rendering, the widget will then shift the text left or right depending on how many columns are required. But note that only "hello world" is styled, so the extra space on the left and right is blank.
- The text widget is enclosed in a `styled.Widget` and then inside a `vpadding.Widget` that is also styled. The `styled.Widget` will apply the supplied style "underneath" any styling currently in use for the given widget. This has the effect of applying "streak" in the unstyled areas to the left and right of "hello world". Similarly, `map2` will apply "bg" in the unstyled areas above and below "hello world". 
- `vpadding.New()` has a third argument, `RenderFlow{}`. This determines how the inner widget, `map1`, is rendered. In this case, it says that whatever size argument is provided when rendering `vert`, use flow-mode to render `map`.

The screenshots above show how the app reacts to being resized. You can see here that gowid's text widget is less sophisticated than urwid's. When made too narrow to fit on one line, the widget should really break "hello world" on the space in the middle. At the moment it doesn't do that. Room for improvement!

## High-Color Modes
![desc](https://drive.google.com/uc?export=view&id=1PQbFW5O-_qE0C-tAdMQVi54GZbbXZzZS)

This program is a glitzier "hello world". This example is at `github.com/gcla/gowid/examples/helloworld` and you can run it via `gowid-helloworld`. 

```go
import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

func main() {

	palette := gowid.Palette{
		"banner":  gowid.MakePaletteEntry(gowid.ColorWhite, gowid.MakeRGBColor("#60d")),
		"streak":  gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#60a")),
		"inside":  gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#808")),
		"outside": gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#a06")),
		"bg":      gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#d06")),
	}

	div := divider.NewBlank()
	outside := styled.New(div, gowid.MakePaletteRef("outside"))
	inside := styled.New(div, gowid.MakePaletteRef("inside"))

	helloworld := styled.New(
		text.NewFromContentExt(
			text.NewContent([]text.TextContentSegment{
				text.StyledContent("Hello World", gowid.MakePaletteRef("banner")),
			}),
			text.Options{
				Align: gowid.HAlignMiddle{},
			},
		),
		gowid.MakePaletteRef("streak"),
	)

	f := gowid.RenderFlow{}

	view := styled.New(
		vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: outside, D: f},
				&gowid.ContainerWidget{IWidget: inside, D: f},
				&gowid.ContainerWidget{IWidget: helloworld, D: f},
				&gowid.ContainerWidget{IWidget: inside, D: f},
				&gowid.ContainerWidget{IWidget: outside, D: f},
			}),
			gowid.VAlignMiddle{},
			f),
		gowid.MakePaletteRef("bg"),
	)

	app, _ := gowid.NewApp(gowid.AppArgs{
		View:    view,
		Palette: &palette,
	})

	app.SimpleMainLoop()
}
```
- To create the vertical effect, a `pile.Widget` is used. The blank lines are made with a `divider.Widget`, where `outside`  and `inside` are styled with different colors. The widget pile is centered with a `vpadding.Widget` and `VAlignMiddle{}`, and the rest of the blank space is styled with "bg".
- This example uses a new `IColor`-creating function, `MakeRGBColor()`. You can provide hex values for red, green and blue, where each value should range from 0x0 to 0xF. If the terminal is in a mode with fewer color combinations, such as 256-color mode, the chosen RGB value is interpolated into an 8x8x8 color cube to find the closest match - in exactly the same fashion as urwid.

## Question and Answer

![desc](https://drive.google.com/uc?export=view&id=1gFKp4b48Jx3t2TUVoydNFyehR3wHfObO)
![desc](https://drive.google.com/uc?export=view&id=1wy67y8sfa6Pkjs22EIC-Epx27cbx95Mv)

The next example asks for the user's name. When the user presses enter, it displays a friendly personalized message. The q or Q key will terminate the app. You can find this example at `github.com/gcla/gowid/examples/gowid-tutorial4` and run it via `gowid-tutorial4`. 

```go
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
)

//======================================================================

type QuestionBox struct {
	gowid.IWidget
}

func (w *QuestionBox) UserInput(ev tcell.Event, size gowid.IRenderSize, focus bool, app gowid.IApp) bool {
	res := true
	if evk, ok := ev.(*tcell.EventKey); ok {
		switch evk.Key() {
		case tcell.KeyEnter:
			w.IWidget = text.New(fmt.Sprintf("Nice to meet you, %s.\n\nPress Q to exit.", w.IWidget.(*edit.Widget).Text()))
		default:
			res = gowid.UserInput(w.IWidget, ev, size, focus, app)
		}
	}
	return res
}

func main() {
	edit := edit.New(edit.Options{Caption: "What is your name?\n"})
	qb := &QuestionBox{edit}
	app, _ := gowid.NewApp(gowid.AppArgs{View: qb})
	app.MainLoop(gowid.UnhandledInputFunc(gowid.HandleQuitKeys))
}
```
- This example shows how you can extend a widget. `QuestionBox` embeds an `IWidget` meaning that it itself implements `IWidget`. The `main()` function sets up a `QuestionBox` widget that extends an `edit.Widget`. That means `QuestionBox` will render like `edit.Widget`. But `QuestionBox` provides a new implementation of `UserInput()`, one of the requirements of `IWidget`. If the key pressed is not "enter" then it defers to its embedded `IWidget`'s implementation of `UserInput()`. That means the embedded `edit.Widget` will process it, and it will accumulate the user's typed input and display that when rendered. But if the user presses "enter", `QuestionBox` replaces its embedded widget with a new `text.Widget` that displays a message to the name the user has typed in.
- When constructing the "Nice to meet you" message, the embedded `IWidget` is cast to an `*edit.Widget`. That's safe because we control the embedded widget, so we know its type. Note that the concrete type is a pointer - gowid widgets have pointer-receiver functions, for the most part, including all methods used to implement `IWidget`. 
- There are pitfalls if your mindset is "object-oriented" like Java or older-style C++. My first instinct was to view `UserInput()` as "overriding" the embedded widget's `UserInput()`. And it's true that our new implementation will be called from an `IWidget` if the interface's type is a `QuestionBox` pointer. But let's say you also provide a specialized implementation for `RenderSize()` another `IWidget` requirement. And let's say `UserInput()` calls a method which is not "overridden" in `edit.Widget`, and that in turn calls `RenderSize()`; then your new version will not be called. The receiver will be the `edit.Widget` pointer. Go does not support dynamic dispatch except for calls through an interface. I certainly misunderstood that when getting going. More details here: https://golang.org/doc/faq#How_do_I_get_dynamic_dispatch_of_methods.

## Widget Callbacks
![desc](https://drive.google.com/uc?export=view&id=11MhLJtGfjTnOtWehJvkdnQv1zHm3-9a_)
![desc](https://drive.google.com/uc?export=view&id=1EL8E6GPvitgPznUZ3B7XOaiNBmBH9syq)

This example shows how you can respond to widget actions, like a button click. See this example at `github.com/gcla/gowid/examples/gowid-tutorial5` and run it via `gowid-tutorial5`. 

```go
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
)

//======================================================================

func main() {

	ask := edit.New(edit.Options{Caption: "What is your name?\n"})
	reply := text.New("")
	btn := button.New(text.New("Exit"))
	sbtn := styled.New(btn, gowid.MakeStyledAs(gowid.StyleReverse))
	div := divider.NewBlank()

	btn.OnClick(gowid.WidgetChangedCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		app.Quit()
	}})

	ask.OnTextSet(gowid.WidgetChangedCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		if ask.Text() == "" {
			reply.SetText("", app)
		} else {
			reply.SetText(fmt.Sprintf("Nice to meet you, %s", ask.Text()), app)
		}
	}})

	f := gowid.RenderFlow{}

	view := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: ask, D: f},
		&gowid.ContainerWidget{IWidget: div, D: f},
		&gowid.ContainerWidget{IWidget: reply, D: f},
		&gowid.ContainerWidget{IWidget: div, D: f},
		&gowid.ContainerWidget{IWidget: sbtn, D: f},
	})

	app, _ := gowid.NewApp(gowid.AppArgs{View: view})

	app.SimpleMainLoop()
}
```
- The bottom-most widget in the pile is a `button.Widget`. It itself wraps an inner widget, and when rendered will add characters on the left and right of the inner widget to create a button effect.
- `button.Widget` can call an interface method when it's clicked. `OnClick()` expects an `IWidgetChangedCallback`. You can use the `WidgetChangedCallback()` adapter to pass a simple function. 
- The first parameter of `WidgetChangedCallback` is an `interface{}`. It's meant to uniquely identify this callback instance so that if you later need to remove the callback, you can by passing the same `interface{}`. Here I've used a simple string, "cb". The callbacks are scoped to the widget, so you can use the same callback identifier when registering callbacks for other widgets. 
- `edit.Widget` can call an interface method when its text changes. In this example, every time the user enters a character, `ask` will update the `reply` widget so that it displays a message.
- The callback will be called with two arguments - the application `app` and the widget issuing the callback. But if it's more convenient, you can rely on Go's scope rules to capture the widgets that you need to modify in the callback. `ask`'s callback refers to `reply` and not the callback parameter `w`. 
- The `<exit>` button is styled using `MakeStyleAs()`, which applies a text style like underline, bold or reverse-video. No colors are given, so the button will use the terminal's default colors.

## Multiple Questions
![desc](https://drive.google.com/uc?export=view&id=18F4F_34YzK9aFMHAx9gsimt_GBk8WZY2)

The final example asks the same question over and over, and collects the results. You can go back and edit previous answers and the program will update its response. It demonstrates the use of a gowid listbox. This example is available at `github.com/gcla/gowid/examples/gowid-tutorial6` and you can run it via `gowid-tutorial6`. 

```go
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
)

//======================================================================

func question() *pile.Widget {
	return pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{
			IWidget: edit.New(edit.Options{Caption: "What is your name?\n"}),
			D:       gowid.RenderFlow{},
		},
	})
}

func answer(name string) *gowid.ContainerWidget {
	return &gowid.ContainerWidget{
		IWidget: text.New(fmt.Sprintf("Nice to meet you, %s", name)),
		D:       gowid.RenderFlow{},
	}
}

type ConversationWidget struct {
	*list.Widget
}

func NewConversationWidget() *ConversationWidget {
	widgets := make([]gowid.IWidget, 1)
	widgets[0] = question()
	lb := list.New(list.NewSimpleListWalker(widgets))
	return &ConversationWidget{lb}
}

func (w *ConversationWidget) UserInput(ev tcell.Event, size gowid.IRenderSize, focus bool, app gowid.IApp) bool {
	res := false
	if evk, ok := ev.(*tcell.EventKey); ok && evk.Key() == tcell.KeyEnter {
		res = true
		focus := w.Walker().Focus()
		focusPile := focus.Widget.(*pile.Widget)
		pileChildren := focusPile.SubWidgets()
		ed := pileChildren[0].(*gowid.ContainerWidget).SubWidget().(*edit.Widget)
		focusPile.SetSubWidgets(append(pileChildren[0:1], answer(ed.Text())), app)
		walker := w.Widget.Walker().(*list.SimpleListWalker)
		walker.Widgets = append(walker.Widgets, question())
		nextPos := walker.Next(focus.Pos).Pos
		walker.SetFocus(nextPos)
		w.Widget.GoToBottom(app)
	} else {
		res = gowid.UserInput(w.Widget, ev, size, focus, app)
	}
	return res
}

func main() {
	app, _ := gowid.NewApp(gowid.AppArgs{View: NewConversationWidget()})
	app.SimpleMainLoop()
}
```
- In this example I've created a new widget called `ConversationWidget`. It embeds a `*list.Widget` and renders like one, but its input is handled specially. A `list.Widget` is a more general form of `pile.Widget`. You provide a `list.Widget` with a `list.IListWalker`which is like a widget iterator. It can return the current "focus" widget, move to the next widget and move to the previous widget. This allows it, potentially, to be unbounded. For an example of that in action, see `github.com/gcla/gowid/examples/gowid-fib` which is plagiarized heavily from urwid's `fib.py` example. 
- The list walker in this example is a wrapper around a Go array of widgets. Each widget in the list is a `pile.Widget` containing either
	- A single `edit.Widget` asking for a name, or
	- An `edit.Widget` asking for a name and the user's response as a `text.Widget`.
- When the user presses "enter" in an `edit.Widget`, the current focus `pile.Widget` is manipulated. Any previous answer is eliminated, and a new answer is appended. The walker is advanced one position, and finally, the `list.Widget` is told to render so the focus widget is at the bottom of the canvas. There is a good deal of type-casting here, but again it's safe because we control the concrete types involved in the construction of this widget hierarchy. 
- When the user presses the up and down cursor keys in the context of a `list.Widget`, the widget's walker adjusts its focus widget. In this example, focus will move from one `pile.Widget` to another. Within that `pile.Widget` there are at most two widgets - one `edit.Widget` and one `text.Widget`. An `edit.Widget` is "selectable", which means it is useful for it to be given the focus. A `text.Widget` is not selectable, which means there's no point in it being given the focus. That means that as the user moves up and down, focus will always be given to an `edit.Widget`. Do note though that just like in urwid, a non-selectable widget can still be given focus e.g. if there is no other selectable widget in the current widget scope. 
- If you're following along with the urwid tutorial, you'll noticed that this example is a little longer than the corresponding urwid program. I attribute that to Go having fewer short-cuts than python, and forcing the programmer to be more explicit. I like that, personally.

