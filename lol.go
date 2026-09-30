package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	esc        = '\x1b'
	reset      = "\x1b[0m"
	hideCursor = "\x1b[?25l"
	showCursor = "\x1b[?25h"
)

// Options controls how text is colored.
type Options struct {
	Freq     float64   // rainbow frequency (higher = tighter bands)
	Spread   float64   // how many columns share one color step
	Seed     int       // starting offset into the rainbow
	Invert   bool      // color the background instead of the foreground
	Mode     ColorMode // truecolor or 256
	Animate  bool      // animate each line
	Duration int       // animation frames per line
	Speed    float64   // animation frames per second
}

// Lolcat colors text line by line and writes it to w.
type Lolcat struct {
	opts Options
	w    *bufio.Writer
	line int // running line counter, kept across files
}

func NewLolcat(w io.Writer, opts Options) *Lolcat {
	opts.Mode = opts.Mode.Resolve()
	if opts.Spread <= 0 {
		opts.Spread = 3.0
	}
	return &Lolcat{opts: opts, w: bufio.NewWriter(w)}
}

// Process reads r until EOF and writes the rainbow version. It flushes after
// every line so it behaves well inside interactive pipes.
func (l *Lolcat) Process(r io.Reader) error {
	if l.opts.Animate {
		l.w.WriteString(hideCursor)
		defer func() {
			l.w.WriteString(showCursor)
			l.w.Flush()
		}()
	}

	br := bufio.NewReader(r)
	for {
		raw, err := br.ReadString('\n')
		if len(raw) > 0 {
			body := strings.TrimRight(raw, "\r\n")
			ending := raw[len(body):]
			l.colorLine(body)
			l.w.WriteString(ending)
			l.line++
			if ferr := l.w.Flush(); ferr != nil {
				return ferr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// colorLine paints one line, animating it if requested.
func (l *Lolcat) colorLine(body string) {
	base := float64(l.opts.Seed + l.line)
	if !l.opts.Animate || body == "" {
		l.paint(body, base)
		return
	}
	delay := time.Duration(float64(time.Second) / l.opts.Speed)
	for i := 0; i < l.opts.Duration; i++ {
		l.w.WriteString("\r")
		l.paint(body, base+float64(i))
		l.w.Flush()
		time.Sleep(delay)
	}
}

// paint writes body with one color per rune. Existing ANSI escape sequences
// are copied through untouched and do not advance the column counter, so
// output from tools like cowsay or ls --color keeps working.
func (l *Lolcat) paint(body string, offset float64) {
	col := 0
	for i := 0; i < len(body); {
		if body[i] == esc {
			n := escLen(body[i:])
			l.w.WriteString(body[i : i+n])
			i += n
			continue
		}
		ch, size := utf8.DecodeRuneInString(body[i:])
		if ch == utf8.RuneError && size == 1 {
			// invalid UTF-8: pass the byte through without coloring
			l.w.WriteByte(body[i])
			i++
			continue
		}
		red, green, blue := rgb(l.opts.Freq, offset+float64(col)/l.opts.Spread)
		fmt.Fprintf(l.w, "\x1b[%sm%c", colorSeq(l.opts.Mode, red, green, blue, l.opts.Invert), ch)
		col++
		i += size
	}
	l.w.WriteString(reset)
}

// escLen returns the byte length of the escape sequence at the start of s.
// s[0] must be ESC.
func escLen(s string) int {
	if len(s) < 2 {
		return len(s)
	}
	switch s[1] {
	case '[': // CSI: parameters, then a final byte in 0x40-0x7E
		for j := 2; j < len(s); j++ {
			if s[j] >= 0x40 && s[j] <= 0x7e {
				return j + 1
			}
		}
		return len(s)
	case ']': // OSC: ends with BEL or ESC \
		for j := 2; j < len(s); j++ {
			if s[j] == '\a' {
				return j + 1
			}
			if s[j] == esc && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2
			}
		}
		return len(s)
	default: // two-byte escape such as ESC 7
		return 2
	}
}
