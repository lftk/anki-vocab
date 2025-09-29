package notetype

import (
	"io/fs"
	"path"
	"strings"

	"github.com/lftk/anki"
)

type Field struct {
	Name     string
	Template string
}

type Template struct {
	Name  string
	Front string
	Back  string
}

type Notetype struct {
	name      string
	fields    []*Field
	templates []*Template
	style     string
}

func Load(name string, fsys fs.FS) (*Notetype, error) {
	fields, err := loadFields(fsys)
	if err != nil {
		return nil, err
	}

	templates, err := loadTemplates(fsys)
	if err != nil {
		return nil, err
	}

	style, err := loadStyle(fsys)
	if err != nil {
		return nil, err
	}

	return &Notetype{
		name:      name,
		fields:    fields,
		templates: templates,
		style:     style,
	}, nil
}

func (nt *Notetype) Name() string {
	return nt.name
}

func (nt *Notetype) Fields() []*Field {
	return nt.fields
}

func (nt *Notetype) Templates() []*Template {
	return nt.templates
}

func (nt *Notetype) Style() string {
	return nt.style
}

func (nt *Notetype) ToAnki() *anki.Notetype {
	fields := make([]*anki.Field, 0, len(nt.fields)+1)
	fields = append(fields, anki.NewField("word"))
	for _, f := range nt.fields {
		fields = append(fields, anki.NewField(f.Name))
	}

	templates := make([]*anki.Template, 0, len(nt.templates))
	for _, t := range nt.templates {
		templates = append(templates, anki.NewTemplate(t.Name, t.Front, t.Back))
	}

	return &anki.Notetype{
		Name:      nt.name,
		Fields:    fields,
		Templates: templates,
		Config:    anki.NewNotetypeConfig(nt.style, false),
	}
}

func loadFields(fsys fs.FS) ([]*Field, error) {
	entries, err := fs.ReadDir(fsys, "fields")
	if err != nil {
		return nil, err
	}

	var fields []*Field
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".tmpl") {
			continue
		}

		name := strings.TrimSuffix(filename, ".tmpl")
		tmpl, err := fs.ReadFile(fsys, path.Join("fields", filename))
		if err != nil {
			return nil, err
		}

		fields = append(fields, &Field{
			Name:     name,
			Template: string(tmpl),
		})
	}
	return fields, nil
}

func loadTemplates(fsys fs.FS) ([]*Template, error) {
	entries, err := fs.ReadDir(fsys, "templates")
	if err != nil {
		return nil, err
	}

	var templates []*Template
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		dir := path.Join("templates", name)

		front, err := fs.ReadFile(fsys, path.Join(dir, "front.html"))
		if err != nil {
			return nil, err
		}

		back, err := fs.ReadFile(fsys, path.Join(dir, "back.html"))
		if err != nil {
			return nil, err
		}

		templates = append(templates, &Template{
			Name:  name,
			Front: string(front),
			Back:  string(back),
		})
	}

	return templates, nil
}

func loadStyle(fsys fs.FS) (string, error) {
	b, err := fs.ReadFile(fsys, "style.css")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
