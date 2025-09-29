package cmd

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/lftk/anki"
	"github.com/urfave/cli/v3"

	"github.com/lftk/anki-vocab/internal/generate"
	"github.com/lftk/anki-vocab/internal/notetype"
	"github.com/lftk/anki-vocab/internal/registry"
	"github.com/lftk/anki-vocab/internal/wordlist"
)

// newGenerateCmd creates the generate command, injecting the default notetype filesystem.
func newGenerateCmd(defaultNotetype fs.FS) *cli.Command {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		// If we can't even get the user's cache dir, something is wrong with the environment.
		// It's better to fail fast than to silently fall back to a local directory.
		panic(fmt.Errorf("failed to determine user cache directory: %w", err))
	}
	defaultCacheDir := filepath.Join(userCacheDir, "anki-vocab")

	return &cli.Command{
		Name:      "generate",
		Usage:     "Generate Anki package from a wordlist file",
		ArgsUsage: "<wordlist_file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Aliases:  []string{"n"},
				Usage:    "The base name of the vocabulary deck to generate.",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Path for the output .apkg file. Defaults to <name>.apkg.",
			},
			&cli.StringFlag{
				Name:  "notetype",
				Usage: "Path to the custom notetype directory.",
			},
			&cli.StringFlag{
				Name:  "dicts",
				Value: "./dicts.yaml",
				Usage: "Path to the dictionaries configuration file.",
			},
			&cli.StringFlag{
				Name:  "cache-dir",
				Value: defaultCacheDir,
				Usage: "Path to the cache directory.",
			},
			&cli.BoolFlag{
				Name:  "no-cache",
				Usage: "Disable caching.",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output.",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			wordlistPath := cmd.Args().First()
			if wordlistPath == "" {
				return fmt.Errorf("missing required argument: wordlist_file")
			}

			name := cmd.String("name")
			notetypeDir := cmd.String("notetype")
			dictsPath := cmd.String("dicts")
			verbose := cmd.Bool("verbose")

			output := cmd.String("output")
			if output == "" {
				output = name + ".apkg"
			}

			cacheDir := cmd.String("cache-dir")
			if cmd.Bool("no-cache") {
				cacheDir = ""
			}

			return runGenerate(ctx, defaultNotetype, name, output, dictsPath, notetypeDir, wordlistPath, cacheDir, verbose)
		},
	}
}

func runGenerate(ctx context.Context, defaultNotetype fs.FS, name, apkgPath, dictsPath, notetypeDir, wordlistPath, cacheDir string, verbose bool) error {
	nt, err := loadNotetype(defaultNotetype, notetypeDir)
	if err != nil {
		return err
	}

	g, err := newGenerator(nt.Fields(), dictsPath, cacheDir)
	if err != nil {
		return err
	}

	col, err := anki.Create()
	if err != nil {
		return err
	}
	defer col.Close()

	ntid, err := addAnkiNotetype(col, nt)
	if err != nil {
		return err
	}

	fmt.Printf("Generating deck '%s' from '%s'...\n", name, wordlistPath)

	var count int
	for deck, err := range wordlist.Load(wordlistPath) {
		if err != nil {
			return err
		}

		deckName := []string{name}
		if deck.Name != "" {
			deckName = append(deckName, deck.Name)
		}
		did, err := addAnkiDeck(col, deckName...)
		if err != nil {
			return err
		}

		dw := &deckWriter{
			col:  col,
			did:  did,
			ntid: ntid,
		}
		for _, word := range deck.Words {
			if count++; verbose {
				fmt.Printf("[%04d] Processing: %s\n", count, word.Text)
			}
			err = g.Generate(ctx, dw, word.Text)
			if err != nil {
				return fmt.Errorf("failed to generate for word %q: %w", word.Text, err)
			}
		}
	}

	fmt.Printf("Successfully generated %d words. Saving to %s...\n", count, apkgPath)

	return col.SaveAs(apkgPath)
}

type deckWriter struct {
	col  *anki.Collection
	did  int64
	ntid int64
}

func (dw *deckWriter) Write(fields []string, media map[string]io.Reader) error {
	n := &anki.Note{
		Fields:     fields,
		NotetypeID: dw.ntid,
	}
	err := dw.col.AddNote(dw.did, n)
	if err != nil {
		return err
	}

	for name, r := range media {
		w, err := dw.col.CreateMedia(name)
		if err != nil {
			return err
		}
		defer w.Close()

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func newGenerator(fields []*notetype.Field, dictsPath, cacheDir string) (*generate.Generator, error) {
	r, err := registry.New(dictsPath, cacheDir)
	if err != nil {
		return nil, err
	}
	return generate.New(r, fields)
}

func loadNotetype(defaultNotetype fs.FS, dir string) (*notetype.Notetype, error) {
	name := "Vocab"
	fsys := defaultNotetype
	if dir != "" {
		name = filepath.Base(dir)
		fsys = os.DirFS(dir)
	}
	return notetype.Load(name, fsys)
}

func addAnkiNotetype(col *anki.Collection, nt *notetype.Notetype) (int64, error) {
	ant := nt.ToAnki()
	err := col.AddNotetype(ant)
	if err != nil {
		return 0, err
	}
	return ant.ID, nil
}

func addAnkiDeck(col *anki.Collection, name ...string) (int64, error) {
	d := &anki.Deck{
		Name: anki.JoinDeckName(name...),
	}
	err := col.AddDeck(d)
	if err != nil {
		return 0, err
	}
	return d.ID, nil
}
