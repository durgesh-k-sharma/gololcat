// gololcat - paint your terminal output with rainbow colors.
package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

// version is overridden at build time: go build -ldflags "-X main.version=1.0.0"
var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok &&
		info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}

func main() { os.Exit(run()) }

func run() int {
	var (
		opts    Options
		modeStr string
		force   bool
		showVer bool
	)
	flag.Float64Var(&opts.Freq, "freq", 0.1, "rainbow frequency (higher = tighter bands)")
	flag.Float64Var(&opts.Spread, "spread", 3.0, "rainbow spread (higher = wider bands)")
	flag.IntVar(&opts.Seed, "seed", 0, "rainbow seed, 0 = random")
	flag.BoolVar(&opts.Invert, "invert", false, "color the background instead of the text")
	flag.BoolVar(&opts.Animate, "animate", false, "animate each line")
	flag.IntVar(&opts.Duration, "duration", 12, "animation frames per line")
	flag.Float64Var(&opts.Speed, "speed", 20, "animation frames per second")
	flag.StringVar(&modeStr, "mode", "auto", "color mode: auto, truecolor or 256")
	flag.BoolVar(&force, "force", false, "force color even when output is not a terminal")
	flag.BoolVar(&showVer, "version", false, "print version and exit")
	flag.Usage = usage
	flag.Parse()

	if showVer {
		fmt.Println("gololcat", version)
		return 0
	}

	mode, err := ParseMode(modeStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gololcat:", err)
		return 2
	}
	opts.Mode = mode
	if opts.Speed <= 0 || opts.Duration < 1 {
		fmt.Fprintln(os.Stderr, "gololcat: -speed must be > 0 and -duration must be >= 1")
		return 2
	}
	if opts.Seed == 0 {
		opts.Seed = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(256)
	}

	files := flag.Args()
	if len(files) == 0 {
		if isTerminal(os.Stdin) {
			usage()
			return 2
		}
		files = []string{"-"}
	}

	// Plain passthrough when color would only add noise (redirect, NO_COLOR).
	colorize := force || (isTerminal(os.Stdout) && os.Getenv("NO_COLOR") == "")
	if !colorize {
		opts.Animate = false
	}

	// Put the cursor back if the user hits Ctrl-C mid-animation.
	if opts.Animate {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sig
			fmt.Fprint(os.Stdout, reset+showCursor)
			os.Exit(130)
		}()
	}

	lol := NewLolcat(os.Stdout, opts)
	status := 0
	for _, name := range files {
		if err := processFile(lol, name, colorize); err != nil {
			fmt.Fprintln(os.Stderr, "gololcat:", err)
			status = 1
		}
	}
	return status
}

func processFile(lol *Lolcat, name string, colorize bool) error {
	var r io.Reader = os.Stdin
	if name != "-" {
		f, err := os.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	if !colorize {
		_, err := io.Copy(os.Stdout, r)
		return err
	}
	return lol.Process(r)
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func usage() {
	fmt.Fprintf(os.Stderr, `gololcat %s - rainbow-colorize text

Usage:
  fortune | gololcat [flags]
  gololcat [flags] [file ...]      ("-" reads stdin)

Flags:
`, version)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, `
Flags must come before file names.
Color is disabled when stdout is not a terminal or NO_COLOR is set (use -force to override).`)
}
