package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/signal"

	"github.com/lftk/anki-vocab/internal/cmd"
)

const version = "v0.1.0"

//go:embed dicts.yaml.example
var dictsExample []byte

//go:embed notetype
var notetypeFS embed.FS

// defaultNotetype represents the content of the 'notetype' directory.
var defaultNotetype fs.FS

func init() {
	fsys, err := fs.Sub(notetypeFS, "notetype")
	if err != nil {
		panic(fmt.Sprintf("Failed to load embedded notetype: %v", err))
	}
	defaultNotetype = fsys
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.Run(ctx, version, os.Args, dictsExample, defaultNotetype); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
