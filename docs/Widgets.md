# Gowid Widgets

Gowid supplies a number of widgets out-of-the-box. 

## asciigraph

**Purpose:** The `asciigraph` widget renders line graphs. It uses the Go package `github.com/guptarohit/asciigraph`.

![desc](https://drive.google.com/uc?export=view&id=19bDRxGYjtL00c1y6StrSZG63AM86Zm7C)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-asciigraph` 
 - `github.com/gcla/gowid/examples/gowid-overlay2` 

## bargraph

**Purpose**: renders bar graphs. Based heavily on urwid's `graph.py`. 

![desc](https://drive.google.com/uc?export=view&id=1mgG-4TnefC6xEwkQM2GlwHctcIB3bH-g)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-graph` 

## boxadapter

**Purpose**: allow a box widget to be rendered in a flow context.

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-dir` 

## button

**Purpose**: a clickable widget. The app can register callbacks to handle click events.

![desc](https://drive.google.com/uc?export=view&id=19kVB4t4c0dLwLusRaGj6oNrTWhNr02u2)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-dir` 
 - `github.com/gcla/gowid/examples/gowid-menu` 
 - `github.com/gcla/gowid/examples/gowid-palette` 
 - `github.com/gcla/gowid/examples/gowid-graph` 
 - `github.com/gcla/gowid/examples/gowid-tree1` 
 - 
## cellmod

**Purpose**: modify the canvas of a child widget by applying a user-supplied function to each `Cell` .

**Examples:**

 - `github.com/gcla/gowid/widgets/dialog` 

## checkbox

**Purpose**: a clickable widget with two states - selected and unselected. 

![desc](https://drive.google.com/uc?export=view&id=1a7OBwLMzithJDwtylG1aLct0dLsyK2kn)

**Examples:**

 - `github.com/gcla/gowid/gowid-asciigraph` 

## clicktracker

**Purpose**: to highlight a widget that has been clicked with the mouse, but which has not yet been activated because the mouse button has not been released. The idea is to highlight which widget will be activated when the mouse is released, if focus remains over that widget.

**Examples**:
- `github.com/gcla/gowid/examples/gowid-graph`
- `github.com/gcla/gowid/examples/gowid-widgets3`

## columns

**Purpose**: arrange child widgets into vertical columns, with configurable column widths.
 
![desc](https://drive.google.com/uc?export=view&id=1kZI6n7wvO16PFu_-nTJ8t24WNsyasdU7)

**Examples:**
 - `github.com/gcla/gowid/examples/gowid-widgets2` 
 - `github.com/gcla/gowid/examples/gowid-widgets3` 
 - `github.com/gcla/gowid/examples/gowid-editor` 
 - `github.com/gcla/gowid/examples/gowid-graph` 
 - 
## dialog

**Purpose**: a modal dialog box that can be opened on top of another widget and will process the user input preferentially.

![desc](https://drive.google.com/uc?export=view&id=1gq_HJLXdnr0KJ2gPK4nalJ56WyPuVkJh)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-editor` 

## divider

**Purpose**: a configurable horizontal line that can be used to separate widgets arranged vertically. Can render using ascii or unicode.

![desc](https://drive.google.com/uc?export=view&id=1YRHeQvckXIVPwPf-8sLAWd3ZuT_qTfuL)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-graph` 
 - `github.com/gcla/gowid/examples/gowid-helloworld` 
 - `github.com/gcla/gowid/examples/gowid-palette` 

## edit

**Purpose**: a text area that will display text typed in by the user, with an optional caption/prefix.

![desc](https://drive.google.com/uc?export=view&id=1Wr_nFrRawGN_FyGv8pqBEegSrwboe3Fm)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-editor` 
 - `github.com/gcla/gowid/examples/gowid-widgets4` 
 - `github.com/gcla/gowid/examples/gowid-widgets6` 

## fill

**Purpose**: a widget that when rendered returns a canvas full of the same user-supplied `Cell`.

![desc](https://drive.google.com/uc?export=view&id=1_O4P4YlikeP6j0vK8A7VPCXwxcnOh9a4)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-overlay1` 
 - `github.com/gcla/gowid/examples/gowid-terminal` 
 - `github.com/gcla/gowid/examples/gowid-widgets2` 

## fixedadapter

**Purpose**: a simple way to allow a fixed widget to be used in a box or flow context.

**Examples:**

 - `github.com/gcla/gowid/widgets/list/list_test.go` 

## framed

**Purpose**: surround a child widget with a configurable "frame", using unicode or ascii characters.

![desc](https://drive.google.com/uc?export=view&id=1EUU4DZxPb6B4-u0XPN9rTgDxmHB9Z0lw)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-widgets1` 
 - `github.com/gcla/gowid/examples/gowid-tree1` 
 - `github.com/gcla/gowid/examples/gowid-graph` 
 - `github.com/gcla/gowid/examples/gowid-terminal` 

## grid

**Purpose**: a way to arrange widgets in a grid, with configurable horizontal alignment.

![desc](https://drive.google.com/uc?export=view&id=1ngHp3pzFzw7qSM8uQ-UmKiijIE4b_ULq)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-widgets3` 

## holder

**Purpose**: wraps a child widget and defers all behavior to it. Allows the child to be swapped out for another.

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-editor` 
 - `github.com/gcla/gowid/examples/gowid-palette` 
 - `github.com/gcla/gowid/examples/gowid-terminal` 

## hpadding

**Purpose**: a widget to render and align a child widget horizontally in a wider space.

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-fib` 
 - `github.com/gcla/gowid/examples/gowid-graph` 
 - `github.com/gcla/gowid/examples/gowid-menu` 

## list

**Purpose**: a flexible widget to navigate a vertical list of widgets rendered in flow mode.

![desc](https://drive.google.com/uc?export=view&id=1uJ3Muv5zEu8HHK5DlBU9NEB_v9aHyemi)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-fib` 
 - `github.com/gcla/gowid/examples/gowid-menu` 
 - `github.com/gcla/gowid/examples/gowid-widgets4` 
 - `github.com/gcla/gowid/examples/gowid-widgets7` 

## menu

**Purpose**: a drop-down menu supporting arbitrarily many sub-menus.

![desc](https://drive.google.com/uc?export=view&id=1kLrAyPAbRi37VjxoVikSfjvrfxKiDVtQ)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-menu` 

## overlay

**Purpose**: a widget to render one widget over another, only passing user input to the occluded widget if the input coordinates are outside the boundaries of the widget on top.

![desc](https://drive.google.com/uc?export=view&id=1q8LcHhl-ZTA9AEIEWmgRGeRTnwjGyX4u)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-overlay1` 
 - `github.com/gcla/gowid/examples/gowid-overlay2` 

## palettemap

**Purpose**: a widget that will render an inner widget with style X if it would otherwise be rendered with style Y, if configured to map Y -> X, and if the widget is styled using references to palette entries.

**Examples:**

 - `github.com/gcla/gowid/gowid-widgets1` 
 - `github.com/gcla/gowid/gowid-widgets4` 
 - `github.com/gcla/gowid/gowid-fib`
 - `github.com/gcla/gowid/gowid-tree1` 

The `palettemap` widget is best used in conjunction with an app that relies on a global palette for its styling and colors. Let's say somewhere in your app's widget hierarchy you have a widget like this:

```go
w := styled.New(x, gowid.MakePaletteRef("red"))
```
The widget `w` will obviously be rendered in `red` according to the app's palette. But if you wrap `w` in something like

```go
z := palettemap.New(w, palettemap.Map{"red": "green"}, palettemap.Map{})
```
Then when `w` is rendered and is the focus widget, the app's palette will be looked up with the name "green" instead. This provides a convenient way of changing the color of widgets, especially when they are in focus.


## pile

**Purpose**: arrange child widgets into horizontal bands, with configurable heights.

![desc](https://drive.google.com/uc?export=view&id=1Bnzhu-hHsr0Ok3hFP5Q0LlaQMTOPEepD)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-helloworld` 
 - `github.com/gcla/gowid/examples/gowid-palette` 
 - `github.com/gcla/gowid/examples/gowid-widgets5` 
 - `github.com/gcla/gowid/examples/gowid-widgets6` 

## progress

**Purpose**: a simple progress monitor.

![desc](https://drive.google.com/uc?export=view&id=15GK6PIlh_CswM6OQ0WNy2O6LC-EwMvhh)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-graph` 
 - `github.com/gcla/gowid/examples/gowid-widgets1` 

 Here is an example initialization of a progress bar:
 ```go
pb := progress.New(progress.Options{
	Normal:   gowid.MakePaletteRef("pg normal"),
	Complete: gowid.MakePaletteRef("pg complete"),
})
```
The struct `progress.Options` is used to pass arguments to the progress bar. You can also set the target number of units for completion, and the current number of units completed. If target is not set, it will default to 100; if current is not set it will default to 0.

Any type implementing `progress.IWidget` can be rendered as a progress bar. Here is an example of how to customize the widget, from `gowid-widgets1`:
```go
type PBWidget struct {
	*progress.Widget
}

func (w *PBWidget) Text() string {
	cur, done := w.Progress(), w.Target()
	percent := gwutil.Min(100, gwutil.Max(0, cur*100/done))
	return fmt.Sprintf("At %d %% (%d/%d)", percent, cur, done)
}

func (w *PBWidget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return progress.Render(w, size, focus, app)
}
```
The `Text()` method provides an alternative label inside the rendered progress bar. Note that we also provided a `Render()` function. The `Text()` function is called when rendering the progress bar, and we need our implementation to take effect. Without a dedicated `Render()` method for `PBWidget`, when the enclosing widget - presumably holding an `IWidget` type - calls `Render()`, the implementation will be provided by the embedded `*progress.Widget`, meaning that will be the receiver type when `Render()` is called. It will defer to the free function `progress.Render()`, but the `progress.IWidget` will hold a `*progress.Widget` not a `*PBWidget`, and so `progress.Render()` will not use our new `Text()` implementation. For a fuller explanation, see the [FAQ](FAQ).
 
## radio

**Purpose**: a widget that, as part of a group, can be in a selected state (if no others in the group are selected) or is otherwise unselected. 

![desc](https://drive.google.com/uc?export=view&id=1cJa4m-DaDpbGwUWChASh-m9AE0sI6tk3)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-graph` 
 - `github.com/gcla/gowid/examples/gowid-overlay2` 
 - `github.com/gcla/gowid/examples/gowid-palette` 
 - `github.com/gcla/gowid/examples/gowid-widgets3` 

## selectable

**Purpose**: make a widget always be selectable, even if it rejects user input.

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-widgets2` 
 - `github.com/gcla/gowid/examples/gowid-dir` 
 - `github.com/gcla/gowid/examples/gowid-tree1` 

## shadow

**Purpose**: adds a drop-shadow effect to a widget.

![desc](https://drive.google.com/uc?export=view&id=1BtI1f0nbyxDUhgsgYaHhOqV3sQJwZnoy)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-graph` 
 - `github.com/gcla/gowid/widgets/dialog/dialog.go` 

## styled

**Purpose**: apply foreground and background coloring and text styling to a widget.

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-dir` 
 - `github.com/gcla/gowid/examples/gowid-fib` 
 - `github.com/gcla/gowid/examples/gowid-helloworld` 
 - `github.com/gcla/gowid/examples/gowid-menu` 

## table

**Purpose**: a widget to display tabular data in columns and rows.

![desc](https://drive.google.com/uc?export=view&id=1TZXfT_VVf5g2sNYi9hia2h36krcGUBI8)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-table` 

Any type that implements the following interface can be used as the source for a table widget:

```go
type IModel interface {
	Columns() int
	RowIdentifier(row int) (RowId, bool)  // return a unique ID for row
	RowWidgets(row RowId) []gowid.IWidget // nil means EOD
	HeaderWidgets() []gowid.IWidget       // nil means no headers
	VerticalSeparator() gowid.IWidget
	HorizontalSeparator() gowid.IWidget
	HeaderSeparator() gowid.IWidget
	Widths() []gowid.IWidgetDimension
}
```
The interface distinguishes a row number (int) from a row identifier (RowId). In order to render the nth row, the widget asks the IModel for the identifier of the nth row. With that in hand, the widget then asks the IModel for the row widgets corresponding to the provided RowId - and with these it can render the row. For simple tables, the RowId value and row number will be the same - n for the nth row. For tables that support sorting (e.g. on a specific column), the underlying implementation of IModel can track the new ordering and ensure the correct row widgets are returned for the row to be displayed at the nth position - perhaps using a simple map. 

Any implementation of IModel should consider caching the widgets returned via calls to `RowWidgets()`. If their state has changed from the default, then returning a cached widget when `RowWidgets()` is called will result in the display reflecting that changed state - color change, clicked checkboxes, etc. This can be seen in the `gowid-table` example. If the widgets are "read-only" and do not change state, then they can be safely generated anew each time `RowWidgets()` is called.

`HeaderWidgets()` should return an array of widgets used as column headers. It can also return `nil`, which means the rendered widget will have no column headers. `VerticalSeparator()`, `HorizontalSeparator()` and `HeaderSeparator()` can also return nil, if the cells don't need to be explicitly boxed or separated from the column headers.

`Widths()` determines how the row widgets are laid out. 

- To use equal space for each column, return an array of `gowid.IRenderWithWeight` with value 1.
- To have a table column use a fixed number of display columns and overflow into subsequent display rows, return a `gowid.IRenderFlow` in that column's position.
- If a column widget is fixed (determines its own render size), return a `gowid.IRenderFixed` at that column's index.

Each table row is rendered using `columns.Widget`, so values suitable for column widths are also suitable for table widths.

An implementation of IModel that returns data from a CSV file is available as `table.NewCsvTable()`. Here is an example of its use:

```go
model := table.NewCsvTable(csvFile, table.SimpleOptions{
	FirstLineIsHeaders: true,
	Style: table.StyleOptions{
		HorizontalSeparator: divider.NewAscii(),
		TableSeparator:      divider.NewUnicode(),
		VerticalSeparator:   fill.New('|'),
	},
})
```

The `csvFile` argument should implement `io.Reader` and be suitable for processing by the standard library's `csv.NewReader()`. The table widget itself is then easily created:

```go
table := table.New(model)
```


## terminal

**Purpose**: a VT-220 capable terminal emulator widget, heavily plagiarized from urwid's `vterm.py`. 

![desc](https://drive.google.com/uc?export=view&id=1PNcMPwiybGBBu48Oot7_hm-uV55TJPTw)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-terminal` 

This widget lets you embed a featureful terminal - or collection of terminals - in your application. The `gowid-terminal` example included demonstrates a very simple copy of tmux - three terminals open running different programs. You can resize the terminal panes using the default hotkey `ctrl-b` and then hitting any of `>`, `<`, `+` or `-`. 

You can have code execute on specific terminal events by registering callbacks:
- when the terminal's process exits
- when the terminal's bell rings
- when the terminal's title is set

You can create a terminal widget simply like this:
```go
tw, err := terminal.New("/bin/bash")
```
There is also a `NewExt()` if you need more control, which takes a `terminal.Params` struct:
```go
type Params struct {
	Command           []string
	Env               []string
	HotKey            IHotKeyProvider
	HotKeyPersistence IHotKeyPersistence
}
```
With that you can provide the environment for the terminal's running process. When a terminal widget has focus, it makes sense for the terminal to be able to process all of the user's keypresses. The `HotKey` field lets you choose a specific keypress (a `tcell.Key`) that will temporarily cause the terminal widget to reject keyboard input. If you have a terminal embedded in your app, this gives the user an opportunity to switch focus to another widget using the keyboard, just like the default `ctrl-b` key in tmux. You can configure how long the hotkey keypress will remain in effect with the `HotKeyPersistence` field. 

Terminal widgets expect to be rendered in box-mode. If your application reorganizes its widget layout, or perhaps if the user simply resizes the terminal window in which your app is running, the terminal widget(s) may be rendered with a different size than was used in the last call to `Render()`. `Gowid` will detect this and send `syscall.TIOCSWINSZ` to the underlying PTY.

When the process underlying the terminal starts running (at least before the first render), the widget will start a goroutine to read from the terminal's master file descriptor. During normal operation, the data read will be terminal-specific control codes. For some examples, see http://www.termsys.demon.co.uk/vtansi.htm. Many of these codes will represent characters that are to be emitted at the current cursor position on the terminal screen, advancing the cursor. Other codes will have a special meaning, like "move the cursor" or "erase part of the screen". The widget implements some simple state machines to track multi-byte sequences, such as the ANSI CSI codes - `ESC[3;4H` - "move the cursor to row 3 column 4". The full-set of `terminal.Widget`'s emulation amounts approximately to VT-220 support. The widget has been tested by running within it the standard `vttest` program and checking the output. All of the credit for this terminal code parsing and state tracking belong's to `urwid`s `vterm.py` implementation.

As with all other `gowid` widgets, `terminal.Widget` supports use of the mouse, where possible. The `gowid-terminal` example demonstrates this - try clicking inside `vim`, or change focus to `emacs` and do the same (you might need to run `M-x xterm-mouse-mode` first). There are several standards for encoding mouse events in the terminal. An SGR encoding of a left-mouse click, for example, might be `ESC[0;3;4M` - a click at row 4, column 3. An older style encoding might be `ESC[M $%` - where the click position is translated to a printable character. The terminal library underlying an application will typically send CSI codes to advertise the various modes the terminal supports and expects - `terminal.Widget` tracks these, and in particular which mouse mode is enabled. `Gowid`'s user input is provided via `tcell` APIs, meaning key-presses and mouse-clicks appear to `gowid` as one of `tcell.EventKey` or `tcell.EventMouse` - that is, because `gowid` runs on top of `tcell`, it does not see the exact byte sequences that the terminal containing the `app` generates. Instead it sees `tcell`'s representation. `terminal.Widget` will convert these `tcell` structs back into byte sequences to send to the widget's underlying terminals file descriptor, and will use its knowledge of the terminal's current mode to choose the correct conversion. To illustrate, let's say you have written a `gowid` application which embeds a `terminal.Widget`. When your app has started, `tcell` will have a PTY to talk to the terminal in which you started the application; and `terminal.Widget` will have a PTY to talk to the terminal running the command embedded in your widget. Your `gowid` application's `TERM` environment variable will determine which `terminfo` database is used to encode and decode terminal sequences to  and from `tcell`. And the environment of the process running in your widget will determine the same for `gowid.Widget`. Let's say the user clicks a mouse button inside your widget.
1. Under your `gowid` app, `tcell` talks to the terminal. The app's `TERM` environment variable will determine the byte sequence `tcell` receives for the mouse event from the app's terminal.
2. `tcell` will translate that sequence into a `tcell.MouseEvent`
3. The `gowid` application will pass that event down through the widget hierarchy until it is accepted by the `terminal.Widget`
4. The `terminal.Widget` will understand from `tcell` that a mouse button has been clicked, and the coordinates of the click. `Gowid` itself will have translated the coordinates of the click as the event was pushed down through the widget hierarchy. The `terminal.Widget` will know the mouse-mode of its underlying terminal because it has tracked the CSI codes sent by its underlying terminal, determined by its process's `TERM` variable.
5. The `terminal.Widget` will convert the `tcell.EventMouse` back to a sequence of bytes according to the correct mouse mode, and send it to the underlying terminal's file descriptor.

The terminal widget defers most of its state tracking to a specialized implementation of `gowid.ICanvas`. The terminal canvas embeds a `gowid.Canvas`, which it renders as normal, but also contains the state-machines and logic to decode and encode terminal byte sequences. The terminal's canvas, when rendered, will always represent the latest state of the terminal underlying the widget. The code is in `github.com/gcla/gowid/widgets/terminal/term_canvas.go`. The terminal canvas implements `io.Writer` allowing a client to write ANSI codes using this standard Golang interface.


## text

**Purpose**: a widget to render text, optionally styled and aligned.

![desc](https://drive.google.com/uc?export=view&id=1XJaSfqljC5ullPj5M9O2ld4OGtP9Nvji)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-widgets1` 
 - `github.com/gcla/gowid/examples/gowid-widgets2` 
 - `github.com/gcla/gowid/examples/gowid-widgets3` 
 - `github.com/gcla/gowid/examples/gowid-widgets4` 

This widget provides a way to render styled and unstyled text. It represents text using this interface:

```go
type IContent interface {
	Length() int
	ChrAt(idx int) rune
	RangeOver(start, end int, attrs gowid.IRenderContext, proc gowid.ICellProcessor)
	AddAt(idx int, markup ContentSegment)
	DeleteAt(idx, length int)
	fmt.Stringer
}
```
The user constructs the widget by providing an array of `ContentSegment` each of which is a string with an associated `gowid.ICellStyler`. When rendering, the widget will built an array of `Cell` using the `RangeOver()` function - that will use the supplied `IRenderContext` (implemented by `gowid.App`) to turn each rune of the markup, along with its `ICellStyler` into a `Cell` - respecting the current color mode of the terminal.

Here is an example of how to build a simple text widget:
```go
w := text.New("Do you want to quit?")
```
Here is an example of how to build a styled text widget:
```go
w := text.NewFromContent(
	text.NewContent([]text.ContentSegment{
		text.StyledContent("hello", gowid.MakePaletteRef("red")),
		text.StringContent(" "),
		text.StyledContent("world", gowid.MakePaletteRef("green")),
	}))
```
There is also a `NewExt()` function that you can supply with these arguments:
```go
type Options struct {
	Wrap  WrapType
	Align gowid.IHAlignment
}
```
- Wrap supports `WrapAny` meaning text will be wrapped to the next line, and `WrapClip` which means the text will be clipped at the end of the current line (and so will render to one canvas line only).
- Align supports any of `HAlignLeft`, `HAlignRight` and `HAlignMiddle`. This option can be used to e.g. center each rendered line of text by sharing the white-space at either edge.

## tree

**Purpose**: a generalization of the `list` widget to render a tree structure.

![desc](https://drive.google.com/uc?export=view&id=1GDirTv-CeKH8CNBjSNrHG7WtE7ccU93P)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-dir` 
 - `github.com/gcla/gowid/examples/gowid-tree1` 

## vpadding

**Purpose**: a widget to render and align a child widget vertically in a wider space.

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-helloworld` 
 - `github.com/gcla/gowid/examples/gowid-overlay2` 
 - `github.com/gcla/gowid/examples/gowid-widgets1` 

## vscroll

**Purpose**: a vertical scroll bar with clickable arrows on either end.

![desc](https://drive.google.com/uc?export=view&id=1dUArLQ1KuzwQmthpTs-HXwxkHvTOVCuT)

**Examples:**

 - `github.com/gcla/gowid/examples/gowid-editor` 

