//go:generate statik -src=data

// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// A simple implementation of a CSV table widget.
package table

import (
	"encoding/csv"
	"io"
	"sort"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/isselected"
	"github.com/gcla/gowid/widgets/radio"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	log "github.com/sirupsen/logrus"
)

//======================================================================

type LayoutOptions struct {
	Widths []gowid.IWidgetDimension
}

type StyleOptions struct {
	VerticalSeparator   gowid.IWidget
	HorizontalSeparator gowid.IWidget
	TableSeparator      gowid.IWidget
	HeaderStyleProvided bool
	HeaderStyleNoFocus  gowid.ICellStyler
	HeaderStyleSelected gowid.ICellStyler
	HeaderStyleFocus    gowid.ICellStyler
	CellStyleProvided   bool
	CellStyleNoFocus    gowid.ICellStyler
	CellStyleSelected   gowid.ICellStyler
	CellStyleFocus      gowid.ICellStyler
}

type SimpleOptions struct {
	NoDefaultSorters bool
	Comparators      []ICompare
	Style            StyleOptions
	Layout           LayoutOptions
}

// SimpleModel implements table.IModel and can be used as a simple model for
// table.IWidget. Fill in the headers, the data; initialize the SortOrder array
// and provide any styling needed. The resulting struct can then be rendered
// as a table.
type SimpleModel struct {
	Headers      []string
	Data         [][]string
	Comparators  []ICompare
	SortOrder    []int // table row order as displayed -> table row identifier (RowId)
	InvSortOrder []int // table row identifier (RowId) -> table row order as displayed
	Style        StyleOptions
	Layout       LayoutOptions
}

var _ IBoundedModel = (*SimpleModel)(nil)

func defaultOptions() SimpleOptions {
	return SimpleOptions{
		Style: StyleOptions{
			HorizontalSeparator: divider.NewAscii(),
			TableSeparator:      divider.NewAscii(),
			VerticalSeparator:   fill.New('|'),
		},
	}
}

// NewCsvModel returns a SimpleTable built from CSV data in the supplied reader. SimpleTable
// implements IModel, and so can be used as a source for table.IWidget.
func NewCsvModel(csvFile io.Reader, firstLineIsHeaders bool, opts ...SimpleOptions) *SimpleModel {
	haveHeaders := false
	reader := csv.NewReader(csvFile)

	res := make([][]string, 0)
	var headers []string
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		if firstLineIsHeaders && !haveHeaders {
			headers = line
			haveHeaders = true
		} else {
			res = append(res, line)
		}
	}

	return NewSimpleModel(headers, res, opts...)
}

// NewSimpleModel returns a SimpleTable built from caller-supplied header data and table data.
// SimpleTable implements IModel, and so can be used as a source for table.IWidget.
func NewSimpleModel(headers []string, res [][]string, opts ...SimpleOptions) *SimpleModel {
	var opt SimpleOptions
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = defaultOptions()
	}

	sortOrder := make([]int, len(res))
	invSortOrder := make([]int, len(res))
	for i := 0; i < len(sortOrder); i++ {
		sortOrder[i] = i
		invSortOrder[i] = i
	}
	tbl := &SimpleModel{
		Headers:      headers,
		Data:         res,
		SortOrder:    sortOrder,
		InvSortOrder: invSortOrder,
	}
	sorters := opt.Comparators
	if sorters == nil {
		var cols int
		if headers != nil {
			cols = len(headers)
		} else if len(res) > 0 {
			cols = len(res[0])
		}
		sorters = make([]ICompare, cols)
	}
	if !opt.NoDefaultSorters {
		for i := 0; i < len(sorters); i++ {
			if sorters[i] == nil {
				sorters[i] = StringCompare{}
			}
		}
	}
	tbl.Comparators = sorters
	tbl.Style = opt.Style
	tbl.Layout = opt.Layout
	return tbl
}

func (c *SimpleModel) Columns() int {
	return len(c.Data[0])
}

func (c *SimpleModel) Rows() int {
	return len(c.Data)
}

var widthOneHeightMax RenderWithUnitsMax = RenderWithUnitsMax{
	RenderWithUnits: gowid.RenderWithUnits{1},
}

func (c *SimpleModel) HeaderWidget(ws []gowid.IWidget, focus int) gowid.IWidget {
	var flowVertDivider *gowid.ContainerWidget
	if c.VerticalSeparator() != nil {
		flowVertDivider = &gowid.ContainerWidget{IWidget: c.VerticalSeparator(), D: widthOneHeightMax}
	}

	cws := make([]gowid.IContainerWidget, 0)
	if flowVertDivider != nil {
		cws = append(cws, flowVertDivider)
	}

	for i, w := range ws {
		var dim gowid.IWidgetDimension = gowid.RenderWithWeight{W: 1}
		if c.Widths() != nil && i < len(c.Widths()) {
			dim = c.Widths()[i]
		}
		cws = append(cws, &gowid.ContainerWidget{IWidget: w, D: dim})
		if flowVertDivider != nil {
			cws = append(cws, flowVertDivider)
		}
	}
	hw := columns.New(cws, columns.Options{
		StartColumn: focus,
	})

	return hw
}

func (c *SimpleModel) HeaderWidgets() []gowid.IWidget {
	var res []gowid.IWidget
	if c.Headers != nil {
		rbgroup := make([]radio.IWidget, 0, len(c.Headers)*2)

		res = make([]gowid.IWidget, 0, len(c.Headers))
		for i, s := range c.Headers {
			i2 := i
			var all, label gowid.IWidget
			label = text.New(s + " ")
			label = button.NewBare(label)

			sorters := c.Comparators
			if sorters != nil {
				sorteri := sorters[i2]
				if sorteri != nil {
					rb1 := radio.New(&rbgroup)
					rb1.Decoration.Right = "/"

					rb1.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, widget gowid.IWidget) {
						sorter := &SimpleTableByColumn{
							SimpleModel: c,
							Column:      i2,
						}
						sort.Sort(sorter)
					}})

					rb2 := radio.New(&rbgroup)
					rb2.Decoration.Left = ""

					rb2.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, widget gowid.IWidget) {
						sorter := &SimpleTableByColumn{
							SimpleModel: c,
							Column:      i2,
						}
						sort.Sort(sort.Reverse(sorter))
					}})

					all = columns.NewFixed(label, rb1, rb2)
				}
			}
			var w gowid.IWidget
			if c.Style.HeaderStyleProvided {
				w = isselected.New(
					styled.New(
						all,
						c.GetStyle().HeaderStyleNoFocus,
					),
					styled.New(
						all,
						c.GetStyle().HeaderStyleSelected,
					),
					styled.New(
						all,
						c.GetStyle().HeaderStyleFocus,
					),
				)
			} else {
				w = styled.NewExt(
					all,
					nil,
					gowid.MakeStyledAs(gowid.StyleReverse),
				)
			}
			res = append(res, w)
		}
	}
	return res
}

type ISimpleRowProvider interface {
	GetStyle() StyleOptions
}

func (c *SimpleModel) GetStyle() StyleOptions {
	return c.Style
}

// Provides a "cell" which is stitched together with columns to provide a "row"
func SimpleCellWidget(c ISimpleRowProvider, i int, s string) gowid.IWidget {
	var w gowid.IWidget
	if c.GetStyle().CellStyleProvided {
		b := button.NewBare(text.New(s))
		w = isselected.New(b, styled.New(b, c.GetStyle().CellStyleSelected), styled.New(b, c.GetStyle().CellStyleFocus))
	} else {
		w = styled.NewExt(button.NewBare(text.New(s)), nil, gowid.MakeStyledAs(gowid.StyleReverse))
	}
	return w
}

func (c *SimpleModel) CellWidget(i int, s string) gowid.IWidget {
	return SimpleCellWidget(c, i, s)
}

type ISimpleDataProvider interface {
	GetData() [][]string
	CellWidget(i int, s string) gowid.IWidget
}

func (c *SimpleModel) GetData() [][]string {
	return c.Data
}

func SimpleCellWidgets(c ISimpleDataProvider, row2 RowId) []gowid.IWidget {
	if int(row2) < len(c.GetData()) {
		row := int(row2)
		res := make([]gowid.IWidget, len(c.GetData()[row]))
		for i, s := range c.GetData()[row] {
			res[i] = c.CellWidget(i, s)
		}
		return res
	}
	return nil
}

func (c *SimpleModel) CellWidgets(rowid RowId) []gowid.IWidget {
	return SimpleCellWidgets(c, rowid)
}

func (c *SimpleModel) RowIdentifier(row int) (RowId, bool) {
	if row < 0 || row >= len(c.SortOrder) {
		return RowId(-1), false
	} else {
		return RowId(c.SortOrder[row]), true
	}
}

func (c *SimpleModel) IdentifierToRow(rowid RowId) (int, bool) {
	if rowid < 0 || int(rowid) >= len(c.InvSortOrder) {
		return -1, false
	} else {
		return c.InvSortOrder[rowid], true
	}
}

func (c *SimpleModel) VerticalSeparator() gowid.IWidget {
	return c.Style.VerticalSeparator
}

func (c *SimpleModel) HorizontalSeparator() gowid.IWidget {
	return c.Style.HorizontalSeparator
}

func (c *SimpleModel) HeaderSeparator() gowid.IWidget {
	return c.Style.TableSeparator
}

func (c *SimpleModel) Widths() []gowid.IWidgetDimension {
	return c.Layout.Widths
}

//======================================================================

// SimpleTableByColumn is a SimpleTable with a selected column; it's intended
// to be sortable, with the values in the selected column being those compared.
type SimpleTableByColumn struct {
	*SimpleModel
	Column int
}

func (m *SimpleTableByColumn) Len() int {
	return len(m.Data)
}

func (m *SimpleTableByColumn) Less(i, j int) bool {
	return m.SimpleModel.Comparators[m.Column].Less(m.Data[m.SortOrder[i]][m.Column], m.Data[m.SortOrder[j]][m.Column])
}

func (m *SimpleTableByColumn) Swap(i, j int) {
	invi, invj := m.SortOrder[i], m.SortOrder[j]
	m.SortOrder[i], m.SortOrder[j] = m.SortOrder[j], m.SortOrder[i]
	m.InvSortOrder[invi], m.InvSortOrder[invj] = m.InvSortOrder[invj], m.InvSortOrder[invi]
}

var _ sort.Interface = (*SimpleTableByColumn)(nil)

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
