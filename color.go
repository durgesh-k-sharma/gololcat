package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// ColorMode selects which ANSI color encoding is emitted.
type ColorMode int

const (
	ModeAuto      ColorMode = iota // detect from COLORTERM
	ModeTrueColor                  // 24-bit: ESC[38;2;R;G;Bm
	Mode256                        // xterm 256-color cube: ESC[38;5;Nm
)

// ParseMode converts a flag value into a ColorMode.
func ParseMode(s string) (ColorMode, error) {
	switch strings.ToLower(s) {
	case "auto", "":
		return ModeAuto, nil
	case "truecolor", "24bit", "true":
		return ModeTrueColor, nil
	case "256":
		return Mode256, nil
	}
	return ModeAuto, fmt.Errorf("unknown color mode %q (use auto, truecolor or 256)", s)
}

// Resolve turns ModeAuto into a concrete mode using the environment.
func (m ColorMode) Resolve() ColorMode {
	if m != ModeAuto {
		return m
	}
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	if strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit") {
		return ModeTrueColor
	}
	return Mode256
}

// rgb maps a position on the rainbow to a color. It is the same idea as the
// original lolcat: three sine waves shifted by 120 degrees.
func rgb(freq, i float64) (r, g, b int) {
	r = int(math.Sin(freq*i+0)*127 + 128)
	g = int(math.Sin(freq*i+2*math.Pi/3)*127 + 128)
	b = int(math.Sin(freq*i+4*math.Pi/3)*127 + 128)
	return
}

// colorSeq returns the SGR parameter string (without ESC[ and m) for a color.
// When bg is true the color is applied to the background and a black
// foreground is added so the text stays readable.
func colorSeq(mode ColorMode, r, g, b int, bg bool) string {
	kind := 38
	prefix := ""
	if bg {
		kind = 48
		prefix = "30;"
	}
	if mode == Mode256 {
		return fmt.Sprintf("%s%d;5;%d", prefix, kind, to256(r, g, b))
	}
	return fmt.Sprintf("%s%d;2;%d;%d;%d", prefix, kind, r, g, b)
}

// to256 maps an RGB triple onto the 6x6x6 color cube (indices 16-231).
func to256(r, g, b int) int {
	c := func(v int) int { return (v*5 + 127) / 255 }
	return 16 + 36*c(r) + 6*c(g) + c(b)
}
