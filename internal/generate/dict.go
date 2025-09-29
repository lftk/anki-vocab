package generate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/lftk/anki-vocab/internal/dict"
	"github.com/lftk/anki-vocab/internal/dyntmpl"
	"github.com/lftk/anki-vocab/internal/registry"
)

type dictQueryer struct {
	Name string
	Dict dict.Queryer
	Caps *dict.QueryCapabilities
}

type dictPronouncer struct {
	Name   string
	Accent string
	Dict   dict.Pronouncer
	Caps   *dict.PronounceCapabilities
}

func loadOrNewQueryer(r *registry.Registry, name string) (*dictQueryer, error) {
	d, err := r.LoadOrNew(name)
	if err != nil {
		return nil, err
	}

	if d.Queryer == nil || d.Capabilities.Query == nil {
		return nil, fmt.Errorf("dictionary %q does not support query", name)
	}

	return &dictQueryer{
		Name: name,
		Dict: d.Queryer,
		Caps: d.Capabilities.Query,
	}, nil
}

func loadOrNewPronouncer(r *registry.Registry, name, accent string) (*dictPronouncer, error) {
	d, err := r.LoadOrNew(name)
	if err != nil {
		return nil, err
	}

	if d.Pronouncer == nil || d.Capabilities.Pronounce == nil {
		return nil, fmt.Errorf("dictionary %q does not support pronunciation", name)
	}
	if !slices.Contains(d.Capabilities.Pronounce.Accents, accent) {
		return nil, fmt.Errorf("dictionary %q does not support %q accent for pronunciation", name, accent)
	}

	return &dictPronouncer{
		Name:   name,
		Accent: accent,
		Dict:   d.Pronouncer,
		Caps:   d.Capabilities.Pronounce,
	}, nil
}

func buildDicts(r *registry.Registry, tmpls []*dyntmpl.Template) ([]*dictQueryer, []*dictPronouncer, error) {
	type pron struct {
		name, accent string
	}

	var (
		qs linkedSet[string, *dictQueryer]
		ps linkedSet[pron, *dictPronouncer]
	)

	for _, t := range tmpls {
		for _, f := range t.Fields() {
			name, ok := parseDictQueryer(f)
			if !ok {
				continue
			}
			_, err := qs.addFunc(name, func() (*dictQueryer, error) {
				return loadOrNewQueryer(r, name)
			})
			if err != nil {
				return nil, nil, err
			}
		}

		for _, f := range t.Funcs() {
			name, accent, ok := parseDictPronouncer(f)
			if !ok {
				continue
			}
			_, err := ps.addFunc(pron{name, accent}, func() (*dictPronouncer, error) {
				return loadOrNewPronouncer(r, name, accent)
			})
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return qs.values(), ps.values(), nil
}

func parseDictQueryer(field string) (string, bool) {
	dict, _, ok := strings.Cut(field, ".")
	return dict, ok
}

func parseDictPronouncer(fn string) (string, string, bool) {
	return isDictPronunciation(fn)
}

func dictPronunciation(dict, accent string) string {
	return fmt.Sprintf("%s_%s_pronunciation", dict, accent)
}

func isDictPronunciation(fn string) (string, string, bool) {
	if s, ok := strings.CutSuffix(fn, "_pronunciation"); ok {
		return strings.Cut(s, "_")
	}
	return "", "", false
}
