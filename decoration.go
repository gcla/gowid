// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package gowid

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gcla/gowid/gwutil"
	"github.com/gdamore/tcell"
	lru "github.com/hashicorp/golang-lru"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/pkg/errors"
)

//======================================================================

// These are used as bitmasks - a style is two AttrMasks. The first bitmask says whether or not the style declares an
// e.g. underline setting; if it's declared, the second bitmask says whether or not underline is affirmatively on or off.
// This allows styles to be layered e.g. the lower style declares underline is on, the upper style does not declare
// an underline preference, so when layered, the cell is rendered with an underline.
const (
	StyleNoneSet tcell.AttrMask = 0 // Just unstyled text.
	StyleAllSet  tcell.AttrMask = tcell.AttrBold | tcell.AttrBlink | tcell.AttrReverse | tcell.AttrUnderline | tcell.AttrDim
)

// StyleAttrs allows the user to represent a set of styles, either affirmatively set (on) or unset (off)
// with the rest of the styles being unspecified, meaning they can be determined by styles layered
// "underneath".
type StyleAttrs struct {
	OnOff tcell.AttrMask // If the specific bit in Set is 1, then the specific bit on OnOff says whether the style is on or off
	Set   tcell.AttrMask // If the specific bit in Set is 0, then no style preference is declared (e.g. for underline)
}

// AllStyleMasks is an array of all the styles that can be applied to a Cell.
var AllStyleMasks = [...]tcell.AttrMask{tcell.AttrBold, tcell.AttrBlink, tcell.AttrDim, tcell.AttrReverse, tcell.AttrUnderline}

// StyleNone expresses no preference for any text styles.
var StyleNone = StyleAttrs{}

// StyleBold specifies the text should be bold, but expresses no preference for other text styles.
var StyleBold = StyleAttrs{tcell.AttrBold, tcell.AttrBold}

// StyleBlink specifies the text should blink, but expresses no preference for other text styles.
var StyleBlink = StyleAttrs{tcell.AttrBlink, tcell.AttrBlink}

// StyleDim specifies the text should be dim, but expresses no preference for other text styles.
var StyleDim = StyleAttrs{tcell.AttrDim, tcell.AttrDim}

// StyleReverse specifies the text should be displayed as reverse-video, but expresses no preference for other text styles.
var StyleReverse = StyleAttrs{tcell.AttrReverse, tcell.AttrReverse}

// StyleUnderline specifies the text should be underlined, but expresses no preference for other text styles.
var StyleUnderline = StyleAttrs{tcell.AttrUnderline, tcell.AttrUnderline}

// StyleBoldOnly specifies the text should be bold, and no other styling should apply.
var StyleBoldOnly = StyleAttrs{tcell.AttrBold, StyleAllSet}

// StyleBlinkOnly specifies the text should blink, and no other styling should apply.
var StyleBlinkOnly = StyleAttrs{tcell.AttrBlink, StyleAllSet}

// StyleDimOnly specifies the text should be dim, and no other styling should apply.
var StyleDimOnly = StyleAttrs{tcell.AttrDim, StyleAllSet}

// StyleReverseOnly specifies the text should be displayed reverse-video, and no other styling should apply.
var StyleReverseOnly = StyleAttrs{tcell.AttrReverse, StyleAllSet}

// StyleUnderlineOnly specifies the text should be underlined, and no other styling should apply.
var StyleUnderlineOnly = StyleAttrs{tcell.AttrUnderline, StyleAllSet}

// MergeUnder merges cell styles. E.g. if a is {underline, underline}, and upper is {!bold, bold}, that
// means a declares that it should be rendered with underline and doesn't care about other styles; and
// upper declares it should NOT be rendered bold, and doesn't declare about other styles. When merged,
// the result is {underline|!bold, underline|bold}.
func (a StyleAttrs) MergeUnder(upper StyleAttrs) StyleAttrs {
	res := a
	for _, am := range AllStyleMasks {
		if (upper.Set & am) != 0 {
			if (upper.OnOff & am) != 0 {
				res.OnOff |= am
			} else {
				res.OnOff &= ^am
			}
			res.Set |= am
		}
	}
	return res
}

//======================================================================

// ColorMode represents the color capability of a terminal.
type ColorMode int

const (
	// Mode256Colors represents a terminal with 256-color support.
	Mode256Colors = ColorMode(iota)

	// Mode88Colors represents a terminal with 88-color support such as rxvt.
	Mode88Colors

	// Mode16Colors represents a terminal with 16-color support.
	Mode16Colors

	// Mode8Colors represents a terminal with 8-color support.
	Mode8Colors

	// Mode8Colors represents a terminal with support for monochrome only.
	ModeMonochrome

	// Mode24BitColors represents a terminal with 24-bit color support like KDE's terminal.
	Mode24BitColors
)

func (c ColorMode) String() string {
	switch c {
	case Mode256Colors:
		return "256 colors"
	case Mode88Colors:
		return "88 colors"
	case Mode16Colors:
		return "16 colors"
	case Mode8Colors:
		return "8 colors"
	case ModeMonochrome:
		return "monochrome"
	case Mode24BitColors:
		return "24-bit truecolor"
	default:
		return fmt.Sprintf("Unknown (%d)", int(c))
	}
}

const (
	colorDefaultName      = "default"
	colorBlackName        = "black"
	colorRedName          = "red"
	colorDarkRedName      = "dark red"
	colorGreenName        = "green"
	colorDarkGreenName    = "dark green"
	colorBrownName        = "brown"
	colorBlueName         = "blue"
	colorDarkBlueName     = "dark blue"
	colorMagentaName      = "magenta"
	colorDarkMagentaName  = "dark magenta"
	colorCyanName         = "cyan"
	colorDarkCyanName     = "dark cyan"
	colorLightGrayName    = "light gray"
	colorDarkGrayName     = "dark gray"
	colorLightRedName     = "light red"
	colorLightGreenName   = "light green"
	colorYellowName       = "yellow"
	colorLightBlueName    = "light blue"
	colorLightMagentaName = "light magenta"
	colorLightCyanName    = "light cyan"
	colorWhiteName        = "white"
)

var (
	basicColors = map[string]int{
		colorDefaultName:      0,
		colorBlackName:        1,
		colorDarkRedName:      2,
		colorDarkGreenName:    3,
		colorBrownName:        4,
		colorDarkBlueName:     5,
		colorDarkMagentaName:  6,
		colorDarkCyanName:     7,
		colorLightGrayName:    8,
		colorDarkGrayName:     9,
		colorLightRedName:     10,
		colorLightGreenName:   11,
		colorYellowName:       12,
		colorLightBlueName:    13,
		colorLightMagentaName: 14,
		colorLightCyanName:    15,
		colorWhiteName:        16,
		colorRedName:          10,
		colorGreenName:        11,
		colorBlueName:         13,
		colorMagentaName:      14,
		colorCyanName:         15,
	}

	tBasicColors = map[string]int{
		colorDefaultName:      0,
		colorBlackName:        1,
		colorDarkRedName:      2,
		colorDarkGreenName:    3,
		colorBrownName:        4,
		colorDarkBlueName:     5,
		colorDarkMagentaName:  6,
		colorDarkCyanName:     7,
		colorLightGrayName:    8,
		colorDarkGrayName:     1,
		colorLightRedName:     2,
		colorLightGreenName:   3,
		colorYellowName:       4,
		colorLightBlueName:    5,
		colorLightMagentaName: 6,
		colorLightCyanName:    7,
		colorWhiteName:        8,
		colorRedName:          2,
		colorGreenName:        3,
		colorBlueName:         5,
		colorMagentaName:      6,
		colorCyanName:         7,
	}

	CubeStart    = 16 // first index of color cube
	CubeSize256  = 6  // one side of the color cube
	graySize256  = 24
	grayStart256 = gwutil.IPow(CubeSize256, 3) + CubeStart
	cubeWhite256 = grayStart256 - 1
	cubeSize88   = 4
	graySize88   = 8
	grayStart88  = gwutil.IPow(cubeSize88, 3) + CubeStart
	cubeWhite88  = grayStart88 - 1
	cubeBlack    = CubeStart

	cubeSteps256 = []int{0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff}
	graySteps256 = []int{
		0x08, 0x12, 0x1c, 0x26, 0x30, 0x3a, 0x44, 0x4e, 0x58, 0x62,
		0x6c, 0x76, 0x80, 0x84, 0x94, 0x9e, 0xa8, 0xb2, 0xbc, 0xc6, 0xd0,
		0xda, 0xe4, 0xee,
	}

	cubeSteps88 = []int{0x00, 0x8b, 0xcd, 0xff}
	graySteps88 = []int{0x2e, 0x5c, 0x73, 0x8b, 0xa2, 0xb9, 0xd0, 0xe7}

	cubeLookup256 = makeColorLookup(cubeSteps256, 256)
	grayLookup256 = makeColorLookup(append([]int{0x00}, append(graySteps256, 0xff)...), 256)

	cubeLookup88 = makeColorLookup(cubeSteps88, 256)
	grayLookup88 = makeColorLookup(append([]int{0x00}, append(graySteps88, 0xff)...), 256)

	cubeLookup256_16  []int
	grayLookup256_101 []int

	cubeLookup88_16  []int
	grayLookup88_101 []int

	// ColorNone means no preference if anything is layered underneath
	ColorNone = MakeTCellNoColor()

	// ColorDefault is an affirmative preference for the default terminal color
	ColorDefault = MakeTCellColorExt(tcell.ColorDefault)

	// Some pre-initialized color objects for use in applications e.g.
	// MakePaletteEntry(ColorBlack, ColorRed)
	ColorBlack      = MakeTCellColorExt(tcell.ColorBlack)
	ColorRed        = MakeTCellColorExt(tcell.ColorRed)
	ColorGreen      = MakeTCellColorExt(tcell.ColorGreen)
	ColorLightGreen = MakeTCellColorExt(tcell.ColorLightGreen)
	ColorYellow     = MakeTCellColorExt(tcell.ColorYellow)
	ColorBlue       = MakeTCellColorExt(tcell.ColorBlue)
	ColorLightBlue  = MakeTCellColorExt(tcell.ColorLightBlue)
	ColorMagenta    = MakeTCellColorExt(tcell.ColorDarkMagenta)
	ColorCyan       = MakeTCellColorExt(tcell.ColorDarkCyan)
	ColorWhite      = MakeTCellColorExt(tcell.ColorWhite)
	ColorDarkRed    = MakeTCellColorExt(tcell.ColorDarkRed)
	ColorDarkGreen  = MakeTCellColorExt(tcell.ColorDarkGreen)
	ColorDarkBlue   = MakeTCellColorExt(tcell.ColorDarkBlue)
	ColorLightGray  = MakeTCellColorExt(tcell.ColorLightGray)
	ColorDarkGray   = MakeTCellColorExt(tcell.ColorDarkGray)
	ColorPurple     = MakeTCellColorExt(tcell.ColorPurple)
	ColorOrange     = MakeTCellColorExt(tcell.ColorOrange)

	longColorRE    = regexp.MustCompile(`^#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$`)
	shortColorRE   = regexp.MustCompile(`^#([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])$`)
	grayHexColorRE = regexp.MustCompile(`^g#([0-9a-fA-F][0-9a-fA-F])$`)
	grayDecColorRE = regexp.MustCompile(`^g(1?[0-9][0-9]?)$`)

	colorfulBlack8   = colorful.Color{R: 0.0, G: 0.0, B: 0.0}
	colorfulWhite8   = colorful.Color{R: 1.0, G: 1.0, B: 1.0}
	colorfulRed8     = colorful.Color{R: 1.0, G: 0.0, B: 0.0}
	colorfulGreen8   = colorful.Color{R: 0.0, G: 1.0, B: 0.0}
	colorfulBlue8    = colorful.Color{R: 0.0, G: 0.0, B: 1.0}
	colorfulYellow8  = colorful.Color{R: 1.0, G: 1.0, B: 1.0}
	colorfulMagenta8 = colorful.Color{R: 1.0, G: 0.0, B: 1.0}
	colorfulCyan8    = colorful.Color{R: 0.0, G: 1.0, B: 1.0}

	colorfulBlack16         = colorful.Color{R: 0.0, G: 0.0, B: 0.0}
	colorfulWhite16         = colorful.Color{R: 0.66, G: 0.66, B: 0.66}
	colorfulRed16           = colorful.Color{R: 0.5, G: 0.0, B: 0.0}
	colorfulGreen16         = colorful.Color{R: 0.0, G: 0.5, B: 0.0}
	colorfulBlue16          = colorful.Color{R: 0.0, G: 0.0, B: 0.5}
	colorfulYellow16        = colorful.Color{R: 0.5, G: 0.5, B: 0.5}
	colorfulMagenta16       = colorful.Color{R: 0.5, G: 0.0, B: 0.5}
	colorfulCyan16          = colorful.Color{R: 0.0, G: 0.5, B: 0.5}
	colorfulBrightBlack16   = colorful.Color{R: 0.33, G: 0.33, B: 0.33}
	colorfulBrightWhite16   = colorful.Color{R: 1.0, G: 1.0, B: 1.0}
	colorfulBrightRed16     = colorful.Color{R: 1.0, G: 0.0, B: 0.0}
	colorfulBrightGreen16   = colorful.Color{R: 0.0, G: 1.0, B: 0.0}
	colorfulBrightBlue16    = colorful.Color{R: 0.0, G: 0.0, B: 1.0}
	colorfulBrightYellow16  = colorful.Color{R: 1.0, G: 1.0, B: 1.0}
	colorfulBrightMagenta16 = colorful.Color{R: 1.0, G: 0.0, B: 1.0}
	colorfulBrightCyan16    = colorful.Color{R: 0.0, G: 1.0, B: 1.0}

	// Used in mapping RGB colors down to 8 terminal colors.
	colorful8 = []colorful.Color{
		colorfulBlack8,
		colorfulWhite8,
		colorfulRed8,
		colorfulGreen8,
		colorfulBlue8,
		colorfulYellow8,
		colorfulMagenta8,
		colorfulCyan8,
	}

	// Used in mapping RGB colors down to 16 terminal colors.
	colorful16 = []colorful.Color{
		colorfulBlack16,
		colorfulWhite16,
		colorfulRed16,
		colorfulGreen16,
		colorfulBlue16,
		colorfulYellow16,
		colorfulMagenta16,
		colorfulCyan16,
		colorfulBrightBlack16,
		colorfulBrightWhite16,
		colorfulBrightRed16,
		colorfulBrightGreen16,
		colorfulBrightBlue16,
		colorfulBrightYellow16,
		colorfulBrightMagenta16,
		colorfulBrightCyan16,
	}

	term8 = []TCellColor{
		ColorBlack,
		ColorWhite,
		ColorRed,
		ColorGreen,
		ColorBlue,
		ColorYellow,
		ColorMagenta,
		ColorCyan,
	}

	term16 = []TCellColor{
		ColorBlack,
		ColorLightGray,
		ColorDarkRed,
		ColorDarkGreen,
		ColorDarkBlue,
		ColorYellow,
		ColorMagenta,
		ColorCyan,
		ColorDarkGray,
		ColorWhite,
		ColorRed,
		ColorGreen,
		ColorBlue,
		ColorYellow,
		ColorMagenta,
		ColorCyan, // TODO - figure out these colors
	}

	term2Cache  *lru.Cache
	term8Cache  *lru.Cache
	term16Cache *lru.Cache
)

//======================================================================

func init() {
	cubeLookup256_16 = make([]int, 16)
	cubeLookup88_16 = make([]int, 16)
	grayLookup256_101 = make([]int, 101)
	grayLookup88_101 = make([]int, 101)

	for i := 0; i < 16; i++ {
		cubeLookup256_16[i] = cubeLookup256[intScale(i, 16, 0x100)]
		cubeLookup88_16[i] = cubeLookup88[intScale(i, 16, 0x100)]
	}
	for i := 0; i < 101; i++ {
		grayLookup256_101[i] = grayLookup256[intScale(i, 101, 0x100)]
		grayLookup88_101[i] = grayLookup88[intScale(i, 101, 0x100)]
	}

	var err error
	for _, cache := range []**lru.Cache{&term2Cache, &term8Cache, &term16Cache} {
		*cache, err = lru.New(100)
		if err != nil {
			panic(err)
		}
	}
}

// makeColorLookup([0, 7, 9], 10)
// [0, 0, 0, 0, 1, 1, 1, 1, 2, 2]
//
func makeColorLookup(vals []int, length int) []int {
	res := make([]int, length)

	vi := 0
	for i := 0; i < len(res); i++ {
		if vi+1 < len(vals) {
			if i <= (vals[vi]+vals[vi+1])/2 {
				res[i] = vi
			} else {
				vi++
				res[i] = vi
			}
		} else if vi < len(vals) {
			// only last vi is valid
			res[i] = vi
		}
	}
	return res
}

// Scale val in the range [0, val_range-1] to an integer in the range
// [0, out_range-1].  This implementation uses the "round-half-up" rounding
// method.
//
func intScale(val int, val_range int, out_range int) int {
	num := val*(out_range-1)*2 + (val_range - 1)
	dem := (val_range - 1) * 2
	return num / dem
}

//======================================================================

type ColorModeMismatch struct {
	Color IColor
	Mode  ColorMode
}

var _ error = ColorModeMismatch{}

func (e ColorModeMismatch) Error() string {
	return fmt.Sprintf("Color %v of type %T not supported in mode %v", e.Color, e.Color, e.Mode)
}

type InvalidColor struct {
	Color interface{}
}

var _ error = InvalidColor{}

func (e InvalidColor) Error() string {
	return fmt.Sprintf("Color %v of type %T is invalid", e.Color, e.Color)
}

//======================================================================

// ICellStyler is an analog to urwid's AttrSpec (http://urwid.org/reference/attrspec.html). When provided
// a RenderContext (specifically the color mode in which to be rendered), the GetStyle() function will
// return foreground, background and style values with which a cell should be rendered. The IRenderContext
// argument provides access to the global palette, so an ICellStyle implementation can look up palette
// entries by name.
type ICellStyler interface {
	GetStyle(IRenderContext) (IColor, IColor, StyleAttrs)
}

// IColor is implemented by any object that can turn itself into a TCellColor, meaning a color with
// which a cell can be rendered. The display mode (e.g. 256 colors) is provided. If no TCellColor is
// available, the second argument should be set to false e.g. no color can be found given a particular
// string name.
type IColor interface {
	ToTCellColor(mode ColorMode) (TCellColor, bool)
}

// MakeCellStyle constructs a tcell.Style from gowid colors and styles. The return value can be provided
// to tcell in order to style a particular region of the screen.
func MakeCellStyle(fg TCellColor, bg TCellColor, attr StyleAttrs) tcell.Style {
	var fgt, bgt tcell.Color
	if fg == ColorNone {
		fgt = tcell.ColorDefault
	} else {
		fgt = fg.ToTCell()
	}
	if bg == ColorNone {
		bgt = tcell.ColorDefault
	} else {
		bgt = bg.ToTCell()
	}
	st := StyleNone.MergeUnder(attr)
	return tcell.Style(st.OnOff).Foreground(fgt).Background(bgt)
}

//======================================================================

// Color satisfies IColor, embeds an IColor, and allows a gowid Color to be
// constructed from a string alone. Each of the more specific color types is
// tried in turn with the string until one succeeds.
type Color struct {
	IColor
	Id string
}

func (c Color) String() string {
	return fmt.Sprintf("%v", c.IColor)
}

// MakeColorSafe returns a Color struct specified by the string argument, in a
// do-what-I-mean fashion - it tries the Color struct maker functions in
// a pre-determined order until one successfully initialized a Color, or
// until all fail - in which case an error is returned. The order tried is
// TCellColor, RGBColor, GrayColor, UrwidColor.
func MakeColorSafe(s string) (Color, error) {
	var col IColor
	var err error
	col, err = MakeTCellColor(s)
	if err == nil {
		return Color{col, s}, nil
	}
	col, err = MakeRGBColorSafe(s)
	if err == nil {
		return Color{col, s}, nil
	}
	col, err = MakeGrayColorSafe(s)
	if err == nil {
		return Color{col, s}, nil
	}
	col, err = NewUrwidColorSafe(s)
	if err == nil {
		return Color{col, s}, nil
	}

	return Color{}, errors.WithStack(InvalidColor{Color: s})
}

func MakeColor(s string) Color {
	res, err := MakeColorSafe(s)
	if err != nil {
		panic(err)
	}
	return res
}

//======================================================================

type ColorByMode struct {
	Colors map[ColorMode]IColor // Indexed by ColorMode
}

var _ IColor = (*ColorByMode)(nil)

func MakeColorByMode(cols map[ColorMode]IColor) ColorByMode {
	res, err := MakeColorByModeSafe(cols)
	if err != nil {
		panic(err)
	}
	return res
}

func MakeColorByModeSafe(cols map[ColorMode]IColor) (ColorByMode, error) {
	return ColorByMode{Colors: cols}, nil
}

func (c ColorByMode) ToTCellColor(mode ColorMode) (TCellColor, bool) {
	if col, ok := c.Colors[mode]; ok {
		col2, ok := col.ToTCellColor(mode)
		return col2, ok
	}
	panic(ColorModeMismatch{Color: c, Mode: mode})
}

//======================================================================

// RGBColor allows for use of colors specified as three components, each with values from 0x0 to 0xf.
// Note that an RGBColor should render as close to the components specify regardless of the color mode
// of the terminal - 24-bit color, 256-color, 88-color. Gowid constructs a color cube, just like urwid,
// and for each color mode, has a lookup table that maps the rgb values to a color cube value which is
// closest to the intended color. Note that RGBColor is not supported in 16-color, 8-color or
// monochrome.
type RGBColor struct {
	Red, Green, Blue int
}

var _ IColor = (*RGBColor)(nil)

// MakeRGBColor constructs an RGBColor from a string e.g. "#f00" is red. Note that
// MakeRGBColorSafe should be used unless you are sure the string provided is valid
// (otherwise there will be a panic).
func MakeRGBColor(s string) RGBColor {
	res, err := MakeRGBColorSafe(s)
	if err != nil {
		panic(err)
	}
	return res
}

func (r RGBColor) String() string {
	return fmt.Sprintf("RGBColor(#%02x,#%02x,#%02x)", r.Red, r.Green, r.Blue)
}

// MakeRGBColorSafe does the same as MakeRGBColor except will return an
// error if provided with invalid input.
func MakeRGBColorSafe(s string) (RGBColor, error) {
	var mult int64 = 1
	match := longColorRE.FindAllStringSubmatch(s, -1)
	if len(match) == 0 {
		match = shortColorRE.FindAllStringSubmatch(s, -1)
		if len(match) == 0 {
			return RGBColor{}, errors.WithStack(InvalidColor{Color: s})
		}
		mult = 16
	}

	d1, _ := strconv.ParseInt(match[0][1], 16, 16)
	d2, _ := strconv.ParseInt(match[0][2], 16, 16)
	d3, _ := strconv.ParseInt(match[0][3], 16, 16)

	d1 *= mult
	d2 *= mult
	d3 *= mult

	x := MakeRGBColorExt(int(d1), int(d2), int(d3))
	return x, nil
}

// MakeRGBColorExtSafe builds an RGBColor from the red, green and blue components
// provided as integers. If the values are out of range, an error is returned.
func MakeRGBColorExtSafe(r, g, b int) (RGBColor, error) {
	col := RGBColor{r, g, b}
	if r > 0xff || g > 0xff || b > 0xff {
		return RGBColor{}, errors.WithStack(errors.WithMessage(InvalidColor{Color: col}, "RGBColor parameters must be between 0x00 and 0xfff"))
	}
	return col, nil
}

// MakeRGBColorExt builds an RGBColor from the red, green and blue components
// provided as integers. If the values are out of range, the function will panic.
func MakeRGBColorExt(r, g, b int) RGBColor {
	res, err := MakeRGBColorExtSafe(r, g, b)
	if err != nil {
		panic(err)
	}

	return res
}

// Implements golang standard library's color.Color
func (rgb RGBColor) RGBA() (r, g, b, a uint32) {
	r = uint32(rgb.Red << 16)
	g = uint32(rgb.Green << 16)
	b = uint32(rgb.Blue << 16)
	a = 0xffff
	return
}

func (r RGBColor) findClosest(from []colorful.Color, corresponding []TCellColor, cache *lru.Cache) TCellColor {
	var best float64 = 100.0
	var j int

	if res, ok := cache.Get(r); ok {
		return res.(TCellColor)
	}

	ccol, _ := colorful.MakeColor(r)

	for i, c := range from {
		x := c.DistanceLab(ccol)
		if x < best {
			best = x
			j = i
		}
	}

	cache.Add(r, corresponding[j])

	return corresponding[j]
}

// ToTCellColor converts an RGBColor to a TCellColor, suitable for rendering to the screen
// with tcell. It lets RGBColor conform to IColor.
func (r RGBColor) ToTCellColor(mode ColorMode) (TCellColor, bool) {
	switch mode {
	case Mode24BitColors:
		c := tcell.Color((r.Red << 16) | (r.Green << 8) | (r.Blue << 0) | int(tcell.ColorIsRGB))
		return MakeTCellColorExt(c), true
	case Mode256Colors:
		rd := cubeLookup256_16[r.Red>>4]
		g := cubeLookup256_16[r.Green>>4]
		b := cubeLookup256_16[r.Blue>>4]
		c := tcell.Color((CubeStart + (((rd * CubeSize256) + g) * CubeSize256) + b) + 0)
		return MakeTCellColorExt(c), true
	case Mode88Colors:
		rd := cubeLookup88_16[r.Red>>4]
		g := cubeLookup88_16[r.Green>>4]
		b := cubeLookup88_16[r.Blue>>4]
		c := tcell.Color((CubeStart + (((rd * cubeSize88) + g) * cubeSize88) + b) + 0)
		return MakeTCellColorExt(c), true
	case Mode16Colors:
		return r.findClosest(colorful16, term16, term16Cache), true
	case Mode8Colors:
		return r.findClosest(colorful8, term8, term8Cache), true
	case ModeMonochrome:
		return r.findClosest(colorful8[0:1], term8[0:1], term2Cache), true
	default:
		return TCellColor{}, false
	}
}

//======================================================================

// UrwidColor is a gowid Color implementing IColor and which allows urwid color names to be used
// (http://urwid.org/manual/displayattributes.html#foreground-and-background-settings) e.g.
// "dark blue", "light gray".
type UrwidColor struct {
	Id     string
	cached bool
	cache  [2]TCellColor
}

var _ IColor = (*UrwidColor)(nil)

// NewUrwidColorSafe returns a pointer to an UrwidColor struct and builds the UrwidColor from
// a string argument e.g. "yellow". Note that in urwid proper (python), a color can also specify
// a style, like "yellow, underline". UrwidColor does not support specifying styles in that manner.
func NewUrwidColorSafe(val string) (*UrwidColor, error) {
	return &UrwidColor{
		Id: val,
	}, nil
}

// NewUrwidColorSafe returns a pointer to an UrwidColor struct and builds the UrwidColor from
// a string argument e.g. "yellow"; this function will panic if the there is an error during
// initialization.
func NewUrwidColor(val string) *UrwidColor {
	res, err := NewUrwidColorSafe(val)
	if err != nil {
		panic(err)
	}

	return res
}

func (r UrwidColor) String() string {
	return fmt.Sprintf("UrwidColor(%s)", r.Id)
}

// ToTCellColor converts the receiver UrwidColor to a TCellColor, ready for rendering to a
// tcell screen. This lets UrwidColor conform to IColor.
func (s *UrwidColor) ToTCellColor(mode ColorMode) (TCellColor, bool) {
	if s.cached {
		switch mode {
		case Mode24BitColors, Mode256Colors, Mode88Colors, Mode16Colors:
			return s.cache[0], true
		case Mode8Colors, ModeMonochrome:
			return s.cache[1], true
		default:
			panic(errors.WithStack(ColorModeMismatch{Color: s, Mode: mode}))
		}
	}

	idx := -1
	switch mode {
	case Mode24BitColors, Mode256Colors, Mode88Colors, Mode16Colors:
		idx = posInMap(s.Id, basicColors)
	case Mode8Colors, ModeMonochrome:
		idx = posInMap(s.Id, tBasicColors)
	default:
		panic(errors.WithStack(ColorModeMismatch{Color: s, Mode: mode}))
	}

	if idx == -1 {
		panic(errors.WithStack(InvalidColor{Color: s}))
	}

	idx = idx - 1 // offset for tcell, which stores default at -1

	c := MakeTCellColorExt(tcell.Color(idx))

	switch mode {
	case Mode24BitColors, Mode256Colors, Mode88Colors, Mode16Colors:
		s.cache[0] = c
	case Mode8Colors, ModeMonochrome:
		s.cache[1] = c
	}
	s.cached = true

	return c, true
}

//======================================================================

// GrayColor is an IColor that represents a greyscale specified by the
// same syntax as urwid - http://urwid.org/manual/displayattributes.html
// and search for "gray scale entries". Strings may be of the form "g3",
// "g100" or "g#a1", "g#ff" if hexadecimal is preferred. These index the
// grayscale color cube.
type GrayColor struct {
	Val int
}

func (g GrayColor) String() string {
	return fmt.Sprintf("GrayColor(%d)", g.Val)
}

// MakeGrayColorSafe returns an initialized GrayColor provided with a string
// input like "g50" or "g#ab". If the input is invalid, an error is returned.
func MakeGrayColorSafe(val string) (GrayColor, error) {
	var d uint64
	match := grayDecColorRE.FindAllStringSubmatch(val, -1)
	if len(match) == 0 || len(match[0]) != 2 {
		match := grayHexColorRE.FindAllStringSubmatch(val, -1)
		if len(match) == 0 || len(match[0]) != 2 {
			return GrayColor{}, errors.WithStack(InvalidColor{Color: val})
		}
		d, _ = strconv.ParseUint(match[0][1], 16, 8)
	} else {
		d, _ = strconv.ParseUint(match[0][1], 10, 8)
		if d > 100 {
			return GrayColor{}, errors.WithStack(InvalidColor{Color: val})
		}
	}

	return GrayColor{int(d)}, nil
}

// MakeGrayColor returns an initialized GrayColor provided with a string
// input like "g50" or "g#ab". If the input is invalid, the function panics.
func MakeGrayColor(val string) GrayColor {
	res, err := MakeGrayColorSafe(val)
	if err != nil {
		panic(err)
	}

	return res
}

func grayAdjustment88(val int) int {
	if val == 0 {
		return cubeBlack
	}
	val -= 1
	if val == graySize88 {
		return cubeWhite88
	}
	y := grayStart88 + val
	return y
}

func grayAdjustment256(val int) int {
	if val == 0 {
		return cubeBlack
	}
	val -= 1
	if val == graySize256 {
		return cubeWhite256
	}
	y := grayStart256 + val
	return y
}

// ToTCellColor converts the receiver GrayColor to a TCellColor, ready for rendering to a
// tcell screen. This lets GrayColor conform to IColor.
func (s GrayColor) ToTCellColor(mode ColorMode) (TCellColor, bool) {
	switch mode {
	case Mode24BitColors:
		adj := intScale(s.Val, 101, 0x100)
		c := tcell.Color((adj << 16) | (adj << 8) | (adj << 0) | int(tcell.ColorIsRGB))
		return MakeTCellColorExt(c), true
	case Mode256Colors:
		x := tcell.Color(grayAdjustment256(grayLookup256_101[s.Val]) + 1)
		return MakeTCellColorExt(x), true
	case Mode88Colors:
		x := tcell.Color(grayAdjustment88(grayLookup88_101[s.Val]) + 1)
		return MakeTCellColorExt(x), true
	default:
		panic(errors.WithStack(ColorModeMismatch{Color: s, Mode: mode}))
	}
}

//======================================================================

// TCellColor is an IColor using tcell's color primitives. If you are not porting from urwid or translating
// from urwid, this is the simplest approach to using color. Note that the underlying tcell.Color value is
// stored offset by 2 from the value tcell would use to actually render a colored cell. Tcell represents
// e.g. black by 0, maroon by 1, and so on - that means the default/empty/zero value for a tcell.Color object
// is the color black. Gowid's layering approach means that the empty value for a color should mean "no color
// preference" - so we want the zero value to mean that. A tcell.Color of -1 means "default color". So gowid
// coopts -2 to mean "no color preference". We store the tcell.Color offset by 2 so the empty value for a
// TCellColor means "no color preference". When we convert to a tcell.Color, we subtract 2 (but since a value
// of -2 is meaningless to tcell, the caller should check and not pass on a value of -2 to tcell APIs - see
// gowid.ColorNone)
type TCellColor struct {
	tc tcell.Color
}

var (
	_ IColor       = (*TCellColor)(nil)
	_ fmt.Stringer = (*TCellColor)(nil)
)

// MakeTCellColor returns an initialized TCellColor given a string input like "yellow". The names that can be
// used are provided here: https://github.com/gdamore/tcell/blob/master/color.go#L821.
func MakeTCellColor(val string) (TCellColor, error) {
	if col, ok := tcell.ColorNames[val]; !ok {
		return TCellColor{}, errors.WithStack(InvalidColor{Color: val})
	} else {
		return MakeTCellColorExt(col), nil
	}
}

// MakeTCellColor returns an initialized TCellColor given a tcell.Color input. The values that can be
// used are provided here: https://github.com/gdamore/tcell/blob/master/color.go#L41.
func MakeTCellColorSafe(val tcell.Color) (TCellColor, error) {
	return TCellColor{val + 2}, nil
}

// MakeTCellColor returns an initialized TCellColor given a tcell.Color input. The values that can be
// used are provided here: https://github.com/gdamore/tcell/blob/master/color.go#L41.
func MakeTCellColorExt(val tcell.Color) TCellColor {
	res, _ := MakeTCellColorSafe(val)
	return res
}

// MakeTCellNoColor returns an initialized TCellColor that represents "no color" - meaning if another
// color is rendered "under" this one, then the color underneath will be displayed.
func MakeTCellNoColor() TCellColor {
	res := MakeTCellColorExt(-2)
	return res
}

// String implements Stringer for '%v' support.
func (r TCellColor) String() string {
	c := r.tc - 2
	if c == -2 {
		return "[no-color]"
	} else {
		return fmt.Sprintf("TCellColor(%v)", tcell.Color(c))
	}
}

// ToTCell converts a TCellColor back to a tcell.Color for passing to tcell APIs.
func (r TCellColor) ToTCell() tcell.Color {
	return r.tc - 2
}

// ToTCellColor is a no-op, and exists so that TCellColor conforms to the IColor interface.
func (r TCellColor) ToTCellColor(mode ColorMode) (TCellColor, bool) {
	return r, true
}

//======================================================================

// NoColor implements IColor, and represents "no color preference", distinct from the default terminal color,
// white, black, etc. This means that if a NoColor is rendered over another color, the color underneath will
// be displayed.
type NoColor struct{}

// ToTCellColor converts NoColor to TCellColor. This lets NoColor conform to the IColor interface.
func (r NoColor) ToTCellColor(mode ColorMode) (TCellColor, bool) {
	return ColorNone, true
}

func (r NoColor) String() string {
	return "NoColor"
}

//======================================================================

// DefaultColor implements IColor and means use whatever the default terminal color is. This is
// different to NoColor, which expresses no preference.
type DefaultColor struct{}

// ToTCellColor converts DefaultColor to TCellColor. This lets DefaultColor conform to the IColor interface.
func (r DefaultColor) ToTCellColor(mode ColorMode) (TCellColor, bool) {
	return MakeTCellColorExt(tcell.ColorDefault), true
}

func (r DefaultColor) String() string {
	return "DefaultColor"
}

//======================================================================

// ColorInverter implements ICellStyler, and simply swaps foreground and background colors.
type ColorInverter struct {
	ICellStyler
}

func (c ColorInverter) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	y, x, z = c.ICellStyler.GetStyle(prov)
	return
}

//======================================================================

// PaletteEntry is typically used by a gowid application to represent a set of color and style
// preferences for use by different application widgets e.g. black text on a white background
// with text underlined. PaletteEntry implements the ICellStyler interface meaning it can
// provide a triple of foreground and background IColor, and a StyleAttrs struct.
type PaletteEntry struct {
	FG    IColor
	BG    IColor
	Style StyleAttrs
}

var _ ICellStyler = (*PaletteEntry)(nil)

// MakeStyledPaletteEntry simply stores the three parameters provided - a foreground and
// background IColor, and a StyleAttrs struct.
func MakeStyledPaletteEntry(fg, bg IColor, style StyleAttrs) PaletteEntry {
	return PaletteEntry{fg, bg, style}
}

// MakePaletteEntry stores the two IColor parameters provided, and has no style preference.
func MakePaletteEntry(fg, bg IColor) PaletteEntry {
	return PaletteEntry{fg, bg, StyleNone}
}

// GetStyle returns the individual colors and style attributes.
func (a PaletteEntry) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	x, y, z = a.FG, a.BG, a.Style
	return
}

//======================================================================

// PaletteRef is intended to represent a PaletteEntry, looked up by name. The ICellStyler
// API GetStyle() provides an IRenderContext and should return two colors and style attributes.
// PaletteRef provides these by looking up the IRenderContext with the name (string) provided
// to it at initialization.
type PaletteRef struct {
	Name string
}

var _ ICellStyler = (*PaletteRef)(nil)

// MakePaletteRef returns a PaletteRef struct storing the (string) name of the PaletteEntry
// which will be looked up in the IRenderContext.
func MakePaletteRef(name string) PaletteRef {
	return PaletteRef{name}
}

// GetStyle returns the two colors and a style, looked up in the IRenderContext by name.
func (a PaletteRef) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	spec, ok := prov.CellStyler(a.Name)
	if ok {
		x, y, z = spec.GetStyle(prov)
	} else {
		x, y, z = NoColor{}, NoColor{}, StyleAttrs{}
	}
	return
}

//======================================================================

// EmptyPalette implements ICellStyler and returns no preference for any colors or styling.
type EmptyPalette struct{}

var _ ICellStyler = (*EmptyPalette)(nil)

func MakeEmptyPalette() EmptyPalette {
	return EmptyPalette{}
}

// GetStyle implements ICellStyler.
func (a EmptyPalette) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	x, y, z = NoColor{}, NoColor{}, StyleAttrs{}
	return
}

//======================================================================

// StyleMod implements ICellStyler. It returns colors and styles from its Cur field unless they are
// overridden by settings in its Mod field. This provides a way for a layering of ICellStylers.
type StyleMod struct {
	Cur ICellStyler
	Mod ICellStyler
}

var _ ICellStyler = (*StyleMod)(nil)

// MakeStyleMod implements ICellStyler and stores two ICellStylers, one to layer on top of the
// other.
func MakeStyleMod(cur, mod ICellStyler) StyleMod {
	return StyleMod{cur, mod}
}

// GetStyle returns the IColors and StyleAttrs from the Mod ICellStyler if they express an
// affirmative preference, otherwise defers to the values from the Cur ICellStyler.
func (a StyleMod) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	fcur, bcur, scur := a.Cur.GetStyle(prov)
	fmod, bmod, smod := a.Mod.GetStyle(prov)
	var ok bool
	_, ok = fmod.ToTCellColor(prov.GetColorMode())
	if ok {
		x = fmod
	} else {
		x = fcur
	}
	_, ok = bmod.ToTCellColor(prov.GetColorMode())
	if ok {
		y = bmod
	} else {
		y = bcur
	}
	z = scur.MergeUnder(smod)
	return
}

//======================================================================

// ForegroundColor is an ICellStyler that expresses a specific foreground color and no preference for
// background color or style.
type ForegroundColor struct {
	IColor
}

var _ ICellStyler = (*ForegroundColor)(nil)

func MakeForeground(c IColor) ForegroundColor {
	return ForegroundColor{c}
}

// GetStyle implements ICellStyler.
func (a ForegroundColor) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	x = a.IColor
	y = NoColor{}
	z = StyleNone
	return
}

//======================================================================

// BackgroundColor is an ICellStyler that expresses a specific background color and no preference for
// foreground color or style.
type BackgroundColor struct {
	IColor
}

var _ ICellStyler = (*BackgroundColor)(nil)

func MakeBackground(c IColor) BackgroundColor {
	return BackgroundColor{c}
}

// GetStyle implements ICellStyler.
func (a BackgroundColor) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	x = NoColor{}
	y = a.IColor
	z = StyleNone
	return
}

//======================================================================

// StyledAs is an ICellStyler that expresses a specific text style and no preference for
// foreground and background color.
type StyledAs struct {
	StyleAttrs
}

var _ ICellStyler = (*StyledAs)(nil)

func MakeStyledAs(s StyleAttrs) StyledAs {
	return StyledAs{s}
}

// GetStyle implements ICellStyler.
func (a StyledAs) GetStyle(prov IRenderContext) (x IColor, y IColor, z StyleAttrs) {
	x = NoColor{}
	y = NoColor{}
	z = a.StyleAttrs
	return
}

//======================================================================

// Palette implements IPalette and is a trivial implementation of a type that can store
// cell stylers and provide access to them via iteration.
type Palette map[string]ICellStyler

var _ IPalette = (*Palette)(nil)

// CellStyler will return an ICellStyler by name, if it exists.
func (m Palette) CellStyler(name string) (ICellStyler, bool) {
	i, ok := m[name]
	return i, ok
}

// RangeOverPalette applies the supplied function to each member of the
// palette. If the function returns false, the loop terminates early.
func (m Palette) RangeOverPalette(f func(k string, v ICellStyler) bool) {
	for k, v := range m {
		if !f(k, v) {
			break
		}
	}
}

//======================================================================

// IColorToTCell is a utility function that will convert an IColor to a TCellColor
// in preparation for passing to tcell to render; if the conversion fails, a default
// TCellColor is returned (provided to the function via a parameter)
func IColorToTCell(color IColor, def TCellColor, mode ColorMode) TCellColor {
	res := def
	colTC, ok := color.ToTCellColor(mode) // Is there a color specified affirmatively? (i.e. not NoColor)
	if ok && colTC != ColorNone {         // Yes a color specified
		res = colTC
	}
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
