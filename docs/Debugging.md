# Debugging Techniques

In no particular order, here is a list of tricks for debugging gowid applications.

## Run on a different tty

- Make a local clone of the `tcell` repo:

```bash
git clone https://github.com/gdamore/tcell
cd tcell
```

- Apply the patch from this gist: https://gist.github.com/gcla/29628006828e57ece336554f26e0bde9

```bash
patch -p1 < gowidtty.patch
```

- Make your gowid application compile against your local clone of `tcell`. Adjust your application's `go.mod` like this, replacing `<me>` with your username (or adjust to where you cloned `tcell`)

```bash
replace github.com/gdamore/tcell => /home/<me>/tcell
```

- Run your application in tmux - make a split screen. 
  - On one side, determine the tty, then block input
  - On the other side, set the environment variable `GOWID_TTY`

![Screenshot-20190616154511-1085x724](https://user-images.githubusercontent.com/45680/59569057-33a9e100-9052-11e9-8d51-4171a870a872.png)

- Run your gowid application e.g. using [tm.sh](https://gist.github.com/gcla/e52ea391c4001cedcfa2cf22d124a750)

```bash
tm.sh 1 go run examples/gowid-fib/fib.go
```

![image](https://user-images.githubusercontent.com/45680/59569085-bfbc0880-9052-11e9-8d17-eaebcca25b6b.png)

Then you can add `fmt.Printf(...)` calls to quickly debug and not have them interfere with your application's tty.

## Watch Flow of User Input Events

Gowid widgets are arranged in a hierarchy, with outer widgets passing events through to inner widgets for processing, 
possibly altering them or handling them themselves on the way. Outer widgets could call

```go
child.UserInput(ev, ...)
```

to determine whether or not the child is handling the event. But instead, all gowid widgets currently call

```go
gowid.UserInput(child, ev, ...)
```

instead. This has the same effect, but means that `gowid.UserInput()` can be used to inspect events flowing through
the application. For example, you can modify the function in `support.go`:

```go
func UserInput(w IWidget, ev interface{}, size IRenderSize, focus Selector, app IApp) bool {
  if evm, ok := ev.(*tcell.EventMouse); ok {
    // Do something
    fmt.Printf("Sending event %v of type %T to widget %v", ev, ev, w)
  }
  return w.UserInput(ev, size, focus, app)
}
