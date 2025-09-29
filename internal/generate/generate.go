package generate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/lftk/anki-vocab/internal/dyntmpl"
	"github.com/lftk/anki-vocab/internal/notetype"
	"github.com/lftk/anki-vocab/internal/registry"
	"github.com/lftk/anki-vocab/internal/tmplfunc"
	"github.com/lftk/anki-vocab/internal/tmpljson"
)

type Generator struct {
	fields      []*dyntmpl.Template
	queryers    []*dictQueryer
	pronouncers []*dictPronouncer
}

func New(r *registry.Registry, fields []*notetype.Field) (*Generator, error) {
	tmpls := make([]*dyntmpl.Template, 0, len(fields))
	for _, f := range fields {
		t, err := dyntmpl.Parse(f.Name, f.Template)
		if err != nil {
			return nil, err
		}
		tmpls = append(tmpls, t)
	}

	queryers, pronouncers, err := buildDicts(r, tmpls)
	if err != nil {
		return nil, err
	}

	return &Generator{
		fields:      tmpls,
		queryers:    queryers,
		pronouncers: pronouncers,
	}, nil
}

type Writer interface {
	Write(fields []string, media map[string]io.Reader) error
}

func (g *Generator) Generate(ctx context.Context, w Writer, word string) error {
	data, err := g.query(ctx, word)
	if err != nil {
		return err
	}

	type pron struct {
		*dictPronouncer
		Format   string
		Filename string
	}
	var prons []pron

	funcs := tmplfunc.Builtins()
	funcs["highlight_word"] = tmplfunc.Highlight(word)

	for _, p := range g.pronouncers {
		fname := dictPronunciation(p.Name, p.Accent)
		funcs[fname] = func() string {
			format := "mp3"
			if len(p.Caps.Formats) > 0 {
				format = p.Caps.Formats[0]
			}
			filename := fmt.Sprintf(
				"%s_%s_%s.%s", word, p.Name, p.Accent, format,
			)
			pron := pron{
				dictPronouncer: p,
				Format:         format,
				Filename:       filename,
			}
			prons = append(prons, pron)
			return fmt.Sprintf("[sound:%s]", filename)
		}
	}

	fields, err := g.execute(word, funcs, data)
	if err != nil {
		return err
	}

	media := make(map[string]io.Reader)
	for _, p := range prons {
		audio, err := p.Dict.Pronounce(ctx, word, p.Accent, p.Format)
		if err != nil {
			return err
		}
		defer audio.Close()

		media[p.Filename] = audio
	}

	return w.Write(fields, media)
}

func (g *Generator) query(ctx context.Context, word string) (map[string]any, error) {
	data := map[string]any{
		"word": word,
	}
	for _, q := range g.queryers {
		b, err := q.Dict.Query(ctx, word)
		if err != nil {
			return nil, err
		}

		if q.Caps.AI {
			b = unquote(b)
		}
		b, err = tmpljson.Normalize(b)
		if err != nil {
			return nil, err
		}

		var m map[string]any
		err = json.Unmarshal(b, &m)
		if err != nil {
			return nil, err
		}

		data[q.Name] = m
	}
	return data, nil
}

func (g *Generator) execute(word string, funcs dyntmpl.FuncMap, data any) ([]string, error) {
	fields := make([]string, 0, len(g.fields)+1)
	fields = append(fields, word)
	for _, t := range g.fields {
		var buf bytes.Buffer
		err := t.Execute(&buf, funcs, data)
		if err != nil {
			return nil, err
		}
		fields = append(fields, strings.TrimSpace(buf.String()))
	}
	return fields, nil
}

func unquote(b []byte) []byte {
	b = bytes.TrimSpace(b)
	if bytes.HasPrefix(b, []byte("```")) && bytes.HasSuffix(b, []byte("```")) {
		b = b[3 : len(b)-3]
		b = bytes.TrimPrefix(b, []byte("json"))
		return bytes.TrimSpace(b)
	}
	return b
}
