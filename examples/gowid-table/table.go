//go:generate statik -src=data

// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// A demonstration of gowid's table widget.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gcla/gowid/examples"
	_ "github.com/gcla/gowid/examples/gowid-table/statik"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/gdamore/tcell"
	"github.com/rakyll/statik/fs"
	log "github.com/sirupsen/logrus"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/table"
)

//======================================================================

type handler struct{}

func (h handler) UnhandledInput(app gowid.IApp, ev interface{}) bool {
	if evk, ok := ev.(*tcell.EventKey); ok {
		if evk.Key() == tcell.KeyCtrlC || evk.Rune() == 'q' || evk.Rune() == 'Q' {
			app.Quit()
			return true
		}
	}
	return false
}

//======================================================================

var (
	file     = kingpin.Arg("file", "CSV file to read.").String()
	colTypes = kingpin.Flag("column-types", "Column data types (for sorting e.g. 0:int,2:string,5:float).").Short('t').String()
)

//======================================================================

func main() {

	//f := examples.RedirectLogger("table.log")
	//defer f.Close()

	kingpin.Parse()

	palette := gowid.Palette{
		"green": gowid.MakePaletteEntry(gowid.ColorDarkGreen, gowid.ColorDefault),
		"red":   gowid.MakePaletteEntry(gowid.ColorRed, gowid.ColorDefault),
	}

	var csvFile io.Reader

	if *file == "" {
		statikFS, err := fs.New()
		if err != nil {
			log.Fatal(err)
		}

		stFile, err := statikFS.Open("/worldcitiespop1k.csv")
		if err != nil {
			log.Fatal(err)
		}
		defer stFile.Close()

		csvFile = stFile
	} else {
		fsFile, err := os.Open(*file)
		if err != nil {
			log.Fatal(err)
		}
		defer fsFile.Close()

		csvFile = fsFile
	}

	model := table.NewCsvModel(csvFile, true, table.SimpleOptions{
		Style: table.StyleOptions{
			HorizontalSeparator: divider.NewAscii(),
			TableSeparator:      divider.NewUnicode(),
			VerticalSeparator:   fill.New('|'),
		},
	})

	if *file == "" {
		// Requires knowledge of the CSV file loaded, of course...
		model.Comparators[3] = table.IntCompare{}
		model.Comparators[4] = table.IntCompare{}
		model.Comparators[5] = table.FloatCompare{}
		model.Comparators[6] = table.FloatCompare{}
	} else {
		if *colTypes != "" {
			types := strings.Split(*colTypes, ",")
			for _, typ := range types {
				colPlusType := strings.Split(typ, ":")
				if len(colPlusType) == 2 {
					if colNum, err := strconv.Atoi(colPlusType[0]); err == nil {
						switch colPlusType[1] {
						case "int":
							model.Comparators[colNum] = table.IntCompare{}
						case "string":
							model.Comparators[colNum] = table.StringCompare{}
						case "float":
							model.Comparators[colNum] = table.FloatCompare{}
						default:
							panic(fmt.Errorf("Did not recognize column type %v", typ))
						}
					}
				}
			}
		}
	}

	table := table.New(model)

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    table,
		Palette: &palette,
		Log:     log.StandardLogger(),
	})
	examples.ExitOnErr(err)

	app.MainLoop(handler{})
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
