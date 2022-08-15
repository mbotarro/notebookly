package main

import (
	"github.com/alecthomas/kong"
)

type Options struct {
	Debug bool
}

var CLI struct {
	Debug bool     `help:"Enable debug mode."`
	Clone CloneCmd `cmd:"" help:"Clone a notebook."`
}

func main() {
	ctx := kong.Parse(&CLI)

	err := ctx.Run(&Options{Debug: CLI.Debug})
	ctx.FatalIfErrorf(err)
}
