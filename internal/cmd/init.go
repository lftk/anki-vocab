package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

// newInitCmd creates the unified 'init' command.
func newInitCmd(dictsExample []byte, defaultNotetype fs.FS) *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize configuration files (e.g., --dicts, --notetype)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "dicts",
				Usage: "Create a dictionaries configuration file at the specified path.",
			},
			&cli.StringFlag{
				Name:  "notetype",
				Usage: "Create a notetype template directory at the specified path.",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force overwrite if the file or directory already exists.",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			force := cmd.Bool("force")

			if cmd.IsSet("dicts") {
				dictsPath := cmd.String("dicts")
				if dictsPath == "" {
					dictsPath = "./dicts.yaml"
				}
				err := initDicts(dictsPath, dictsExample, force)
				if err != nil {
					return err
				}
			}

			if cmd.IsSet("notetype") {
				notetypePath := cmd.String("notetype")
				if notetypePath == "" {
					notetypePath = "./notetype"
				}
				err := exportNotetype(notetypePath, defaultNotetype, force)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func initDicts(path string, content []byte, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file %q already exists, use --force to overwrite", path)
		}
	}
	fmt.Printf("Generating new config file: %s\n", path)
	return os.WriteFile(path, content, 0644)
}

func exportNotetype(dest string, src fs.FS, force bool) error {
	if _, err := os.Stat(dest); err == nil {
		entries, err := os.ReadDir(dest)
		if err != nil {
			return fmt.Errorf("failed to read destination directory: %w", err)
		}
		if len(entries) > 0 && !force {
			return fmt.Errorf("directory %q is not empty, use --force to overwrite", dest)
		}
	}

	fmt.Printf("Exporting built-in notetype to %s/ ...\n", dest)

	// Use the standard library fs.WalkDir for a safe, merge-overwrite copy.
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		outpath := filepath.Join(dest, path)
		if d.IsDir() {
			return os.MkdirAll(outpath, 0755)
		}
		data, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		return os.WriteFile(outpath, data, 0644)
	})
}
