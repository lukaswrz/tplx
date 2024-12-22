// Package tplx wraps the standard html/template library to provide a little more
// structure and ease of use.
package tplx

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

var (
	// ErrUnknownTemplate is returned when a template is not found in the renderer.
	ErrUnknownTemplate = errors.New("template not found in renderer")

	// ErrInvalidSpec is returned when the template renderer specification is invalid.
	ErrInvalidSpec = errors.New("template renderer spec is invalid")
)

// Renderer is an interface for rendering templates.
type Renderer interface {
	Render(w io.Writer, name string, data any, funcs template.FuncMap) error
}

type renderer struct {
	m map[string]*template.Template
}

// Spec describes the structure of all templates managed by the renderer.
//
// The keys of the Spec map represent top-level template names. Each key maps
// to a slice of Meta, where each Meta defines the name, path, and functions
// associated with a template fragment.
type Spec map[string][]Meta

// Meta represents metadata for a single template fragment.
//
// Name specifies the name of the template fragment. Path specifies the path to
// the template file in the file system. Funcs provides template-specific
// functions.
type Meta struct {
	Name  string
	Path  string
	Funcs template.FuncMap
}

// NewRenderer creates a new Renderer instance from a file system, specification,
// and global function map.
//
// The fsys parameter specifies the file system from which template files are
// loaded. The spec parameter defines the structure of the templates, mapping
// top-level template names to their fragments. The funcs parameter provides
// global template functions.
//
// Returns a Renderer instance or an error if the templates cannot be initialized
// according to the specification.
func NewRenderer(fsys fs.FS, spec Spec, funcs template.FuncMap) (Renderer, error) {
	r := renderer{
		m: make(map[string]*template.Template, len(spec)),
	}

	for name, metas := range spec {
		inc := false

		t := template.New(name).Funcs(funcs)

		for _, meta := range metas {
			if meta.Name == name {
				inc = true
			}

			text, err := fs.ReadFile(fsys, meta.Path)
			if err != nil {
				return nil, fmt.Errorf("unable to read template file: %w", err)
			}

			t = t.New(meta.Name).Funcs(meta.Funcs)

			t, err = t.Parse(string(text))
			if err != nil {
				return nil, err
			}
		}

		if !inc {
			return nil, ErrInvalidSpec
		}

		r.m[name] = t
	}

	return r, nil
}

// Render writes the rendered output of a named template to the provided writer.
//
// The wr parameter specifies the writer where the rendered template output will
// be written. The name parameter specifies the name of the template to render
// The data parameter provides the context data for rendering, and the funcs
// parameter provides additional template functions.
//
// Returns an error if the template cannot be rendered or does not exist.
func (r renderer) Render(wr io.Writer, name string, data any, funcs template.FuncMap) error {
	t, ok := r.m[name]
	if !ok {
		return ErrUnknownTemplate
	}
	err := t.ExecuteTemplate(wr, name, data)
	if err != nil {
		return fmt.Errorf("cannot render template: %w", err)
	}
	return nil
}
