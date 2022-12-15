package gomarkdoc

import (
	"strings"
	"text/template"

	"github.com/princjef/gomarkdoc/format"
	"github.com/princjef/gomarkdoc/lang"
)

type (
	// Renderer provides capabilities for rendering various types of
	// documentation with the configured format and templates.
	Renderer struct {
		templateOverrides map[string]string
		tmpl              *template.Template
		format            format.Format
	}

	// RendererOption configures the renderer's behavior.
	RendererOption func(renderer *Renderer) error
)

//go:generate ./gentmpl.sh templates templates

// NewRenderer initializes a Renderer configured using the provided options. If
// nothing special is provided, the created renderer will use the default set of
// templates and the GitHubFlavoredMarkdown.
func NewRenderer(opts ...RendererOption) (*Renderer, error) {
	renderer := &Renderer{
		templateOverrides: make(map[string]string),
		format:            &format.GitHubFlavoredMarkdown{},
	}

	for _, opt := range opts {
		if err := opt(renderer); err != nil {
			return nil, err
		}
	}
	for override, val := range renderer.templateOverrides {
		for template := range templates {
			if template == override {
				continue
			}
		}
		// Add the new template if it isn't in the list of predefined templates
		templates[override] = val
	}

	for name, tmplStr := range templates {
		// Use the override if present
		if val, ok := renderer.templateOverrides[name]; ok {
			tmplStr = val
		}

		if renderer.tmpl == nil {
			tmpl := template.New(name)
			tmpl.Funcs(map[string]interface{}{
				"add": func(n1, n2 int) int {
					return n1 + n2
				},
				"spacer": func() string {
					return "\n\n"
				},

				"bold":                renderer.format.Bold,
				"header":              renderer.format.Header,
				"rawHeader":           renderer.format.RawHeader,
				"codeBlock":           renderer.format.CodeBlock,
				"link":                renderer.format.Link,
				"docLink":             renderer.format.DocLink,
				"listEntry":           renderer.format.ListEntry,
				"accordion":           renderer.format.Accordion,
				"accordionHeader":     renderer.format.AccordionHeader,
				"accordionTerminator": renderer.format.AccordionTerminator,
				"localHref":           renderer.format.LocalHref,
				"codeHref":            renderer.format.CodeHref,
				"paragraph":           renderer.format.Paragraph,
				"escape":              renderer.format.Escape,
				"exec":                execTemplate(tmpl),
			})

			if _, err := tmpl.Parse(tmplStr); err != nil {
				return nil, err
			}

			renderer.tmpl = tmpl
		} else if _, err := renderer.tmpl.New(name).Parse(tmplStr); err != nil {
			return nil, err
		}
	}

	return renderer, nil
}

func execTemplate(t *template.Template) func(string, interface{}) (string, error) {
	return func(name string, v interface{}) (string, error) {
		var buf strings.Builder
		err := t.ExecuteTemplate(&buf, name, v)
		return buf.String(), err
	}
}

// WithTemplateOverride adds a template that overrides the template with the
// provided name using the value provided in the tmpl parameter.
func WithTemplateOverride(name, tmpl string) RendererOption {
	return func(renderer *Renderer) error {
		renderer.templateOverrides[name] = tmpl

		return nil
	}
}

// WithFormat changes the renderer to use the format provided instead of the
// default format.
func WithFormat(format format.Format) RendererOption {
	return func(renderer *Renderer) error {
		renderer.format = format
		return nil
	}
}

// File renders a file containing one or more packages to document to a string.
// You can change the rendering of the file by overriding the "file" template
// or one of the templates it references.
func (out *Renderer) File(file *lang.File) (string, error) {
	return out.writeTemplate("file", file)
}

// Package renders a package's documentation to a string. You can change the
// rendering of the package by overriding the "package" template or one of the
// templates it references.
func (out *Renderer) Package(pkg *lang.Package) (string, error) {
	return out.writeTemplate("package", pkg)
}

// Func renders a function's documentation to a string. You can change the
// rendering of the package by overriding the "func" template or one of the
// templates it references.
func (out *Renderer) Func(fn *lang.Func) (string, error) {
	return out.writeTemplate("func", fn)
}

// Type renders a type's documentation to a string. You can change the
// rendering of the type by overriding the "type" template or one of the
// templates it references.
func (out *Renderer) Type(typ *lang.Type) (string, error) {
	return out.writeTemplate("type", typ)
}

// Example renders an example's documentation to a string. You can change the
// rendering of the example by overriding the "example" template or one of the
// templates it references.
func (out *Renderer) Example(ex *lang.Example) (string, error) {
	return out.writeTemplate("example", ex)
}

// writeTemplate renders the template of the provided name using the provided
// data object to a string. It uses the set of templates provided to the
// renderer as a template library.
func (out *Renderer) writeTemplate(name string, data interface{}) (string, error) {
	var result strings.Builder
	if err := out.tmpl.ExecuteTemplate(&result, name, data); err != nil {
		return "", err
	}

	return result.String(), nil
}
