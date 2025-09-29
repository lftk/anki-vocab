package dyntmpl

import (
	"html/template"
	"io"
	"slices"
	"text/template/parse"

	"github.com/lftk/anki-vocab/internal/tmplinspect"
)

type Template struct {
	tmpl   *template.Template
	fields []string
	funcs  []string
}

func Parse(name, tmpl string) (*Template, error) {
	t := parse.New(name)
	t.Mode = parse.SkipFuncCheck

	t, err := t.Parse(tmpl, "", "", make(map[string]*parse.Tree))
	if err != nil {
		return nil, err
	}

	fields, funcs, err := tmplinspect.InspectTree(t)
	if err != nil {
		return nil, err
	}

	tt, err := template.New(name).AddParseTree(name, t)
	if err != nil {
		return nil, err
	}

	return &Template{
		tmpl:   tt,
		fields: fields,
		funcs:  funcs,
	}, nil
}

func (t *Template) Name() string {
	return t.tmpl.Name()
}

func (t *Template) Fields() []string {
	return slices.Clone(t.fields)
}

func (t *Template) Funcs() []string {
	return slices.Clone(t.funcs)
}

type FuncMap = template.FuncMap

func (t *Template) Execute(w io.Writer, funcs FuncMap, data any) error {
	tt, err := t.tmpl.Clone()
	if err != nil {
		return err
	}
	return tt.Funcs(funcs).Execute(w, data)
}
