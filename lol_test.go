package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

func strip(s string) string { return ansi.ReplaceAllString(s, "") }

func newOpts() Options {
	return Options{Freq: 0.1, Spread: 3, Seed: 1, Mode: ModeTrueColor, Duration: 1, Speed: 1000}
}

func render(t *testing.T, opts Options, in string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := NewLolcat(&buf, opts).Process(strings.NewReader(in)); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestTextPreserved(t *testing.T) {
	in := "hello\nwörld ☃ 日本語\n\nno trailing newline"
	if got := strip(render(t, newOpts(), in)); got != in {
		t.Fatalf("text changed:\n got %q\nwant %q", got, in)
	}
}

func TestCRLFPreserved(t *testing.T) {
	in := "a\r\nb\r\n"
	if got := strip(render(t, newOpts(), in)); got != in {
		t.Fatalf("got %q want %q", got, in)
	}
}

func TestTrueColorSequence(t *testing.T) {
	out := render(t, newOpts(), "x")
	if !strings.Contains(out, "\x1b[38;2;") {
		t.Fatalf("missing truecolor sequence: %q", out)
	}
}

func Test256Mode(t *testing.T) {
	o := newOpts()
	o.Mode = Mode256
	out := render(t, o, "x")
	if !strings.Contains(out, "\x1b[38;5;") || strings.Contains(out, "38;2;") {
		t.Fatalf("expected 256-color output: %q", out)
	}
}

func TestInvertUsesBackground(t *testing.T) {
	o := newOpts()
	o.Invert = true
	if out := render(t, o, "x"); !strings.Contains(out, "48;2;") {
		t.Fatalf("expected background color: %q", out)
	}
}

func TestExistingEscapesPassThrough(t *testing.T) {
	in := "\x1b[1mbold\x1b[0m"
	out := render(t, newOpts(), in)
	if !strings.Contains(out, "\x1b[1m") {
		t.Fatalf("user escape lost: %q", out)
	}
	if strip(out) != "bold" {
		t.Fatalf("escape leaked into text: %q", strip(out))
	}
}

func TestEscapeDoesNotShiftColors(t *testing.T) {
	plain := render(t, newOpts(), "ab")
	withEsc := render(t, newOpts(), "a\x1b[1mb")
	pick := func(s string) string { return regexp.MustCompile(`38;2;[0-9;]+`).FindAllString(s, -1)[1] }
	if pick(plain) != pick(withEsc) {
		t.Fatalf("escape shifted the gradient")
	}
}

func TestLinesShiftGradient(t *testing.T) {
	out := render(t, newOpts(), "a\na\n")
	c := regexp.MustCompile(`38;2;[0-9;]+`).FindAllString(out, -1)
	if len(c) != 2 || c[0] == c[1] {
		t.Fatalf("expected different colors per line, got %v", c)
	}
}

func TestInvalidUTF8DoesNotPanic(t *testing.T) {
	_ = render(t, newOpts(), "ok\xff\xfeok\n")
}

func TestEscLen(t *testing.T) {
	cases := map[string]int{
		"\x1b[31mX":        5,
		"\x1b[38;2;1;2;3m": 13,
		"\x1b]0;title\aX":  10,
		"\x1b7X":           2,
		"\x1b":             1,
		"\x1b[31":          4,
	}
	for in, want := range cases {
		if got := escLen(in); got != want {
			t.Errorf("escLen(%q)=%d want %d", in, got, want)
		}
	}
}

func TestRGBRange(t *testing.T) {
	for i := 0; i < 2000; i++ {
		r, g, b := rgb(0.1, float64(i)/3)
		for _, v := range []int{r, g, b} {
			if v < 0 || v > 255 {
				t.Fatalf("out of range at %d: %d", i, v)
			}
		}
	}
}

func TestTo256(t *testing.T) {
	if to256(0, 0, 0) != 16 || to256(255, 255, 255) != 231 || to256(255, 0, 0) != 196 {
		t.Fatal("bad 256-color mapping")
	}
}

func TestParseMode(t *testing.T) {
	if m, _ := ParseMode("256"); m != Mode256 {
		t.Fatal("256")
	}
	if _, err := ParseMode("nope"); err == nil {
		t.Fatal("expected error")
	}
}
