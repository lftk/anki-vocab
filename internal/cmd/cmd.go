package cmd

import (
	"context"
	"io/fs"

	"github.com/urfave/cli/v3"
)

// Run builds the app and runs it, returning any error.
func Run(ctx context.Context, version string, args []string, dictsExample []byte, defaultNotetype fs.FS) error {
	app := &cli.Command{
		Name:    "anki-vocab",
		Version: version,
		Usage:   "Generate Anki decks for vocabulary.",
		Commands: []*cli.Command{
			newGenerateCmd(defaultNotetype),
			newInitCmd(dictsExample, defaultNotetype),
		},
	}
	return app.Run(ctx, args)
}
