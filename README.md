# gololcat

Rainbow-colorize your terminal output. A fast, dependency-free `lolcat` written in Go.

```sh
fortune | gololcat
```

## Features

- Smooth per-character rainbow with a gradient that shifts on every line
- Reads from stdin or files, streaming line by line (works with `tail -f`)
- Truecolor (24-bit) output with automatic 256-color fallback
- Preserves existing ANSI styling such as bold text, `ls --color` and cowsay
- Plays nicely with pipes: no color codes are emitted when output is redirected or `NO_COLOR` is set
- Background mode and line animation
- Full Unicode support
- Standard library only

## Installation

With Go 1.21 or newer:

```sh
go install github.com/durgesh-k-sharma/gololcat@latest
```

Or build from source:

```sh
git clone https://github.com/durgesh-k-sharma/gololcat.git
cd gololcat
make build
```

## Usage

```sh
# Pipe anything into it
ls -la | gololcat
echo "hello world" | gololcat -animate

# Or pass files ("-" means stdin)
gololcat README.md
gololcat header.txt - footer.txt
```

### Flags

| Flag        | Default | Description                                        |
|-------------|---------|----------------------------------------------------|
| `-freq`     | `0.1`   | Rainbow frequency (higher = tighter bands)         |
| `-spread`   | `3.0`   | Columns per color step (higher = wider bands)      |
| `-seed`     | random  | Starting point in the rainbow (`0` = random)       |
| `-invert`   | off     | Color the background instead of the text           |
| `-animate`  | off     | Animate each line                                  |
| `-duration` | `12`    | Animation frames per line                          |
| `-speed`    | `20`    | Animation frames per second                        |
| `-mode`     | `auto`  | Color mode: `auto`, `truecolor` or `256`           |
| `-force`    | off     | Color output even when stdout is not a terminal    |
| `-version`  |         | Print the version and exit                         |

Flags must come before file names.

Use `-seed` with a fixed value for reproducible colors, for example `gololcat -seed 42`.

### Color detection

- **Color depth:** in `auto` mode, truecolor is used when `COLORTERM` is `truecolor` or `24bit`. Otherwise gololcat falls back to the 256-color palette. Override this with `-mode`.
- **Redirected output:** when stdout is not a terminal, or `NO_COLOR` is set, input is passed through unchanged, so `gololcat file > out.txt` never writes escape codes. `-force` overrides this.

### Exit codes

| Code | Meaning                                               |
|------|-------------------------------------------------------|
| `0`  | Success                                               |
| `1`  | One or more files could not be read                   |
| `2`  | Invalid usage, or no input on an interactive terminal |
| `130`| Interrupted during animation (Ctrl-C)                 |

## Development

```sh
make build    # build ./gololcat
make test     # go vet + tests with the race detector
make install  # install into $GOBIN
```

## Limitations

- Animation repaints each line with a carriage return, so lines wider than the terminal don't animate cleanly.
- Wide characters (CJK, emoji) count as one column when computing the gradient.
