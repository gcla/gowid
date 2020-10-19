// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A port of urwid's palette_test.py example using gowid widgets.
package main

import (
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/holder"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/radio"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

//======================================================================

// t0 t1 t2 t3 t4 t5 t6 t7 t8 t9 t10 t11 t12 t13 t14 t15 t16 t17 t254 t255

var chart256 = `
brown__   dark_red_   dark_magenta_   dark_blue_   dark_cyan_   dark_green_
yellow_   light_red   light_magenta   light_blue   light_cyan   light_green

              #00f#06f#08f#0af#0df#0ff         black_______    dark_gray___  
            #60f#00d#06d#08d#0ad#0dd#0fd         light_gray__    white_______
          #80f#60d#00a#06a#08a#0aa#0da#0fa
        #a0f#80d#60a#008#068#088#0a8#0d8#0f8
      #d0f#a0d#80d#608#006#066#086#0a6#0d6#0f6
    #f0f#d0d#a0a#808#606#000#060#080#0a0#0d0#0f0#0f6#0f8#0fa#0fd#0ff
      #f0d#d0a#a08#806#600#660#680#6a0#6d0#6f0#6f6#6f8#6fa#6fd#6ff#0df
        #f0a#d08#a06#800#860#880#8a0#8d0#8f0#8f6#8f8#8fa#8fd#8ff#6df#0af
          #f08#d06#a00#a60#a80#aa0#ad0#af0#af6#af8#afa#afd#aff#8df#6af#08f
            #f06#d00#d60#d80#da0#dd0#df0#df6#df8#dfa#dfd#dff#adf#8af#68f#06f
              #f00#f60#f80#fa0#fd0#ff0#ff6#ff8#ffa#ffd#fff#ddf#aaf#88f#66f#00f
                                    #fd0#fd6#fd8#fda#fdd#fdf#daf#a8f#86f#60f
      #66d#68d#6ad#6dd                #fa0#fa6#fa8#faa#fad#faf#d8f#a6f#80f
    #86d#66a#68a#6aa#6da                #f80#f86#f88#f8a#f8d#f8f#d6f#a0f
  #a6d#86a#668#688#6a8#6d8                #f60#f66#f68#f6a#f6d#f6f#d0f
#d6d#a6a#868#666#686#6a6#6d6#6d8#6da#6dd    #f00#f06#f08#f0a#f0d#f0f    
  #d6a#a68#866#886#8a6#8d6#8d8#8da#8dd#6ad
    #d68#a66#a86#aa6#ad6#ad8#ada#add#8ad#68d
      #d66#d86#da6#dd6#dd8#dda#ddd#aad#88d#66d       g78_g82_g85_g89_g93_g100_
                    #da6#da8#daa#dad#a8d#86d       g52_g58_g62_g66_g70_g74_
      #88a#8aa        #d86#d88#d8a#d8d#a6d       g27_g31_g35_g38_g42_g46_g50_
    #a8a#888#8a8#8aa    #d66#d68#d6a#d6d       g0__g3__g7__g11_g15_g19_g23_
      #a88#aa8#aaa#88a
            #a88#a8a
`

var chart88 = `
brown__   dark_red_   dark_magenta_   dark_blue_   dark_cyan_   dark_green_
yellow_   light_red   light_magenta   light_blue   light_cyan   light_green

      #00f#08f#0cf#0ff            black_______    dark_gray___
    #80f#00c#08c#0cc#0fc            light_gray__    white_______
  #c0f#80c#008#088#0c8#0f8
#f0f#c0c#808#000#080#0c0#0f0#0f8#0fc#0ff            #88c#8cc
  #f0c#c08#800#880#8c0#8f0#8f8#8fc#8ff#0cf        #c8c#888#8c8#8cc
    #f08#c00#c80#cc0#cf0#cf8#cfc#cff#8cf#08f        #c88#cc8#ccc#88c
      #f00#f80#fc0#ff0#ff8#ffc#fff#ccf#88f#00f            #c88#c8c
                    #fc0#fc8#fcc#fcf#c8f#80f
                      #f80#f88#f8c#f8f#c0f        g62_g74_g82_g89_g100
                        #f00#f08#f0c#f0f        g0__g19_g35_g46_g52
`

var chart16 = `
brown__   dark_red_   dark_magenta_   dark_blue_   dark_cyan_   dark_green_
yellow_   light_red   light_magenta   light_blue   light_cyan   light_green

black_______    dark_gray___    light_gray__    white_______
`

var chart8 = `
black__   red_   green_   yellow_    blue_    magenta_   cyan_   white_
`

var chartMono = `
black__   white_
`

var attrRE = regexp.MustCompile(`(?P<whitespace>[ \n]*)(?P<entry>((#...)|([a-z_]{2,})|(g[0-9]+_+)|(t[0-9]{1,3})))`)

var attrREIndices = makeRegexpNameMap(attrRE)

var fgColorDefault = gowid.NewUrwidColor("light gray")
var bgColorDefault = gowid.MakeTCellColorExt(tcell.ColorBlack)

var chartHolder *holder.Widget

var chart256Content *text.Widget
var chart88Content *text.Widget
var chart16Content *text.Widget
var chart8Content *text.Widget
var chartMonoContent *text.Widget

var foregroundColors = true

//======================================================================

func makeRegexpNameMap(re *regexp.Regexp) map[string]int {
	res := make(map[string]int)
	for i, name := range re.SubexpNames() {
		res[name] = i
	}
	return res
}

//======================================================================

func updateChartHolder(mode gowid.ColorMode, app gowid.IApp) {
	switch mode {
	case gowid.Mode256Colors:
		chart256Content = text.NewFromContent(parseChart(chart256))
		chartHolder.SetSubWidget(chart256Content, app)
	case gowid.Mode88Colors:
		chart88Content = text.NewFromContent(parseChart(chart88))
		chartHolder.SetSubWidget(chart88Content, app)
	case gowid.Mode16Colors:
		chart16Content = text.NewFromContent(parseChart(chart16))
		chartHolder.SetSubWidget(chart16Content, app)
	case gowid.Mode8Colors:
		chart8Content = text.NewFromContent(parseChart(chart8))
		chartHolder.SetSubWidget(chart8Content, app)
	case gowid.ModeMonochrome:
		chartMonoContent = text.NewFromContent(parseChart(chartMono))
		chartHolder.SetSubWidget(chartMonoContent, app)
	default:
		panic(errors.New("Invalid mode, something went wrong!"))
	}
}

//======================================================================

func modeRb(group *[]radio.IWidget, txt string) gowid.IWidget {
	rbt := text.New(" " + txt)
	rb := radio.New(group)
	widp := gowid.RenderFixed{}

	rb.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		if rb.Selected {
			switch txt {
			case "256-Color":
				app.SetColorMode(gowid.Mode256Colors)
			case "88-Color":
				app.SetColorMode(gowid.Mode88Colors)
			case "16-Color":
				app.SetColorMode(gowid.Mode16Colors)
			case "8-Color":
				app.SetColorMode(gowid.Mode8Colors)
			case "Monochrome":
				app.SetColorMode(gowid.ModeMonochrome)
			case "Foreground Colors":
				foregroundColors = true
			case "Background Colors":
				foregroundColors = false
			default:
				panic(errors.New("Invalid mode, something went wrong!"))
			}
			updateChartHolder(app.GetColorMode(), app) // Update the chart displayed
		}
	}})

	c := columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{rb, widp},
		&gowid.ContainerWidget{rbt, widp},
	})

	cm := styled.NewExt(c, gowid.MakePaletteRef("panel"), gowid.MakePaletteRef("focus"))

	return cm
}

//======================================================================

func parseChart(chart string) *text.Content {
	content := make([]text.ContentSegment, 0)
	replacer := strings.NewReplacer("_", " ")
	for _, match := range attrRE.FindAllStringSubmatch(chart, -1) {
		ws := match[attrREIndices["whitespace"]]
		if ws != "" {
			content = append(content, text.StringContent(ws))
		}
		entry := match[attrREIndices["entry"]]
		entry = replacer.Replace(entry)

		entry2 := strings.Trim(entry, " ")
		scol, err := gowid.MakeColorSafe(entry2)
		if err == nil {
			var attrPair gowid.PaletteEntry
			if foregroundColors {
				attrPair = gowid.MakePaletteEntry(scol, bgColorDefault)
			} else {
				attrPair = gowid.MakePaletteEntry(fgColorDefault, scol)
			}
			content = append(content, text.StyledContent(entry, attrPair))
		} else {
			content = append(content, text.StyledContent(entry, gowid.MakePaletteRef("redinv")))
		}
	}
	return text.NewContent(content)
}

//======================================================================

func main() {

	// If this is set to truecolor when a gowid screen is setup, 24-bit truecolor
	// support is enabled. Then the program won't output the 256-color/16-color
	// terminal codes that this program is supposed to exhibit. So unset the variable
	// right away.
	os.Unsetenv("COLORTERM")

	f := examples.RedirectLogger("palette.log")
	defer f.Close()

	palette := gowid.Palette{
		"header":  gowid.MakeStyledPaletteEntry(gowid.ColorBlack, gowid.ColorWhite, gowid.StyleUnderline),
		"panel":   gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorDarkBlue),
		"focus":   gowid.MakePaletteEntry(gowid.ColorYellow, gowid.ColorRed),
		"red":     gowid.MakePaletteEntry(gowid.ColorRed, gowid.ColorBlack),
		"redinv":  gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorRed),
		"default": gowid.MakePaletteEntry(gowid.NewUrwidColor("light gray"), gowid.ColorBlack),
	}

	header := gowid.MakePaletteRef("header")

	// make this just text
	headerContent := []text.ContentSegment{
		text.StyledContent("Gowid Palette Test", header),
	}

	headerText := styled.New(text.NewFromContent(text.NewContent(headerContent)), header)

	rbgroup := make([]radio.IWidget, 0)
	bggroup := make([]radio.IWidget, 0)

	btn := button.New(text.New("Exit"))

	btn.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		app.Quit()
	}})

	subFlow := gowid.RenderFlow{}

	// Put this in the group first so it is the one selected
	cols256 := modeRb(&rbgroup, "256-Color")

	pw1 := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{modeRb(&rbgroup, "Monochrome"), subFlow},
		&gowid.ContainerWidget{modeRb(&rbgroup, "8-Color"), subFlow},
		&gowid.ContainerWidget{modeRb(&rbgroup, "16-Color"), subFlow},
		&gowid.ContainerWidget{modeRb(&rbgroup, "88-Color"), subFlow},
		&gowid.ContainerWidget{cols256, subFlow},
	})

	pw2 := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{modeRb(&bggroup, "Foreground Colors"), subFlow},
		&gowid.ContainerWidget{modeRb(&bggroup, "Background Colors"), subFlow},
		&gowid.ContainerWidget{divider.NewBlank(), subFlow},
		&gowid.ContainerWidget{divider.NewBlank(), subFlow},
		&gowid.ContainerWidget{
			styled.NewExt(
				btn,
				gowid.MakePaletteRef("panel"),
				gowid.MakePaletteRef("focus"),
			),
			subFlow},
	})

	cs := columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{pw1, gowid.RenderWithWeight{10}},
		&gowid.ContainerWidget{pw2, gowid.RenderWithWeight{10}},
	})

	cs2 := styled.New(cs, gowid.MakePaletteRef("panel"))

	chart256Content = text.NewFromContent(parseChart(chart256))
	chartHolder = holder.New(chart256Content)

	view := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{headerText, subFlow},
		&gowid.ContainerWidget{cs2, subFlow},
		&gowid.ContainerWidget{chartHolder, gowid.RenderWithWeight{1}},
	})

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    view,
		Palette: &palette,
		Log:     log.StandardLogger(),
	})
	examples.ExitOnErr(err)
	app.SetColorMode(gowid.Mode256Colors)

	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
