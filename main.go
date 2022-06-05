// An implementation of Conway's Game of Life.
package main

// Initial version: https://go.dev/doc/play/life.go

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/kerrigan29a/drawille-go"
	"golang.org/x/exp/slices"
)

// Field represents a two-dimensional field of cells.
type Field struct {
	s    [][]bool
	w, h uint
}

// NewField returns an empty field of the specified width and height.
func NewField(w, h uint) *Field {
	s := make([][]bool, h)
	for i := range s {
		s[i] = make([]bool, w)
	}
	return &Field{s: s, w: w, h: h}
}

// Set sets the state of the specified cell to the given value.
func (f *Field) Set(x, y uint, b bool) {
	f.s[y][x] = b
}

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b            *Field
	w, h            uint
	birth, survival []uint
}

// NewLife returns a new Life game state with a random initial state.
func NewLife(birth, survival []uint, w, h uint, maxDensity float64) *Life {
	a := NewField(w, h)
	for i := uint(0); i < uint(float64(w*h)*maxDensity); i++ {
		a.Set(uint(rand.Intn(int(w))), uint(rand.Intn(int(h))), true)
	}
	return &Life{
		a:        a,
		b:        NewField(w, h),
		w:        w,
		h:        h,
		birth:    birth,
		survival: survival,
	}
}

// Alive reports whether the specified cell is alive.
// If the x or y coordinates are outside the field boundaries they are wrapped
// toroidally. For instance, an x value of -1 is treated as width-1.
func (l *Life) Alive(x, y int) bool {
	return l.a.s[uint(y+int(l.a.h))%l.a.h][uint(x+int(l.a.w))%l.a.w]
}

func contains(x uint, xs []uint) bool {
	_, ok := slices.BinarySearch(xs, x)
	return ok
}

// Next returns the state of the specified cell at the next time step.
func (l *Life) Next(x, y uint) bool {
	// Count the adjacent cells that are alive.
	neighbors := uint(0)
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if (j != 0 || i != 0) && l.Alive(int(x)+i, int(y)+j) {
				neighbors++
			}
		}
	}
	// Return next state according to the game rules:
	//   neighbors in BIRTH: on,
	//   neighbors in SURVIVAL: maintain current state,
	//   otherwise: off.
	return contains(neighbors, l.birth) || contains(neighbors, l.survival) && l.Alive(int(x), int(y))
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).
	for y := uint(0); y < l.h; y++ {
		for x := uint(0); x < l.w; x++ {
			l.b.Set(x, y, l.Next(x, y))
		}
	}
	// Swap fields a and b.
	l.a, l.b = l.b, l.a
}

// String returns the game board as a string.
func (l *Life) String() string {
	g := drawille.NewCanvas()
	for y := 0; y < int(l.h); y++ {
		for x := 0; x < int(l.w); x++ {
			if l.Alive(x, y) {
				g.Set(x, y)
			}
		}
	}
	return g.String()
}

func draw(screen tcell.Screen, l *Life) {
	for y, line := range strings.Split(l.String(), "\n") {
		pos := 0
		for _, r := range line { // iterates over runes, not positions
			screen.SetCell(pos, y, tcell.StyleDefault, r)
			pos++
		}
	}
	screen.Show()
}

func next(l *Life, screen tcell.Screen, epoch uint) uint {
	l.Step()
	draw(screen, l)
	return epoch + 1
}

func parseDigits(name, s string) []uint {
	var result []uint
	for _, r := range s {
		if !unicode.IsDigit(r) || (r < '0' || r > '8') {
			panic(fmt.Errorf("invalid %s rule, use only [0-8] digits: %s", name, s))
		}
		result = append(result, uint(r-'0'))
	}
	slices.Sort(result)

	return result
}

func parseBS(s string) ([]uint, []uint) {
	re := regexp.MustCompile(`(?i)B([0-8]+)/S([0-8]*)`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		panic(fmt.Errorf("invalid B/S rule: %s", s))
	}
	return parseDigits("birth", m[1]), parseDigits("survival", m[2])
}

func parseSB(s string) ([]uint, []uint) {
	re := regexp.MustCompile(`([0-8]*)/([0-8]+)`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		panic(fmt.Errorf("invalid S/B rule: %s", s))
	}
	return parseDigits("survival", m[1]), parseDigits("birth", m[2])
}

func parseArgs() (birth, survival []uint, density float64) {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "")
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\n", Version)
		fmt.Fprintln(flag.CommandLine.Output(), "")
	}

	var bs string
	bsDefault := "B3/S23"
	bsHelp := "Birth/Survival (or Golly) `rule`"
	flag.StringVar(&bs, "bs", bsDefault, fmt.Sprintf("%-35s %-20s", bsHelp, "(alias -golly)"))
	flag.StringVar(&bs, "golly", bsDefault, fmt.Sprintf("%-35s %-20s", bsHelp, "(alias -bs)"))

	var sb string
	sbDefault := "23/3"
	sbHelp := "Survival/Birth (or MCell) `rule`"
	flag.StringVar(&sb, "sb", sbDefault, fmt.Sprintf("%-35s %-20s", sbHelp, "(alias -mcell)"))
	flag.StringVar(&sb, "mcell", sbDefault, fmt.Sprintf("%-35s %-20s", sbHelp, "(alias -sb)"))

	densityDefault := 0.5
	densityHelp := "Initial `density`"
	flag.Float64Var(&density, "density", densityDefault, fmt.Sprintf("%-35s %-20s", densityHelp, "(alias -d)"))
	flag.Float64Var(&density, "d", densityDefault, fmt.Sprintf("%-35s %-20s", densityHelp, "(alias -density)"))

	flag.Parse()

	if bs != bsDefault {
		birth, survival = parseBS(bs)
	} else {
		survival, birth = parseSB(sb)
	}
	if birth == nil {
		panic("unknown parsing state")
	}
	return birth, survival, density
}

func handleErrors() {
	// This code allows us to propagate internal errors without having to add error checks everywhere throughout the
	// code. This is only possible because the code does not update shared state and does not manipulate locks.
	if r := recover(); r != nil {
		var rerr runtime.Error
		if err, ok := r.(error); ok && !errors.As(err, &rerr) {
			log.Fatalf("%+v", err)
		} else {
			panic(r)
		}
	}
	os.Exit(0)
}

func main() {
	// Idea from: https://www.youtube.com/watch?v=c78U0MZ4b_c
	// This is also used in:
	//     - <GOLANG_CODEBASE>/src/encoding/json/encode.go
	//     - https://github.com/golang/go/blob/865911424d509184d95d3f9fc6a8301927117fdc/src/encoding/json/encode.go#L322
	defer handleErrors()

	birth, survival, density := parseArgs()

	// Initialize screen
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	defer screen.Fini()
	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
	screen.EnableMouse()
	screen.DisablePaste()
	screen.HideCursor()
	screen.Clear()

	rand.Seed(time.Now().UnixNano())
	w, h := screen.Size()
	if err != nil {
		panic(err)
	}
	l := NewLife(birth, survival, uint(w*2), uint(h*4), density)

	tick := time.NewTicker(time.Second / 10)

	events := make(chan tcell.Event)
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

	epoch := uint(0)
	paused := false
loop:
	for {
		select {
		case event := <-events:
			switch event := event.(type) {
			case *tcell.EventResize:
				screen.Sync()
			case *tcell.EventKey:
				if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyCtrlC {
					break loop
				}
				if unicode.ToLower(event.Rune()) == 'q' {
					break loop
				}
				if unicode.ToLower(event.Rune()) == 'p' {
					paused = !paused
				} else if unicode.ToLower(event.Rune()) == 'c' {
					screen.Sync()
				} else if unicode.ToLower(event.Rune()) == 'n' && paused {
					epoch = next(l, screen, epoch)
				}

			case *tcell.EventMouse:
				button := event.Buttons()
				// Only process button events, not wheel events
				button &= tcell.ButtonMask(0xff)
				if button != tcell.ButtonNone {
					x, y := event.Position()
					l.a.Set(uint(x*2)+0, uint(y*4)+0, button == tcell.Button1)
					l.a.Set(uint(x*2)+0, uint(y*4)+1, button == tcell.Button1)
					l.a.Set(uint(x*2)+0, uint(y*4)+2, button == tcell.Button1)
					l.a.Set(uint(x*2)+0, uint(y*4)+3, button == tcell.Button1)
					l.a.Set(uint(x*2)+1, uint(y*4)+0, button == tcell.Button1)
					l.a.Set(uint(x*2)+1, uint(y*4)+1, button == tcell.Button1)
					l.a.Set(uint(x*2)+1, uint(y*4)+2, button == tcell.Button1)
					l.a.Set(uint(x*2)+1, uint(y*4)+3, button == tcell.Button1)
					draw(screen, l)
				}
			}
		case <-tick.C:
			if paused {
				continue
			}
			epoch = next(l, screen, epoch)
		}
	}
}
