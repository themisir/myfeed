package renderer

import (
	"io"
	"io/fs"
	"text/template"

	"github.com/labstack/echo/v4"
)

func Template(fsys fs.FS, root string) echo.Renderer {
	var tmpl *template.Template = nil

	if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			name := path[len(root):]

			// Skip trailing slash
			if name[0] == '/' {
				name = name[1:]
			}

			// Read file contents
			bytes, err := fs.ReadFile(fsys, path)
			if err != nil {
				panic(err)
			}

			if tmpl == nil {
				tmpl = template.Must(template.New(name).Parse(string(bytes)))
			} else {
				tmpl = template.Must(tmpl.New(name).Parse(string(bytes)))
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}

	return &templateRenderer{tmpl}
}

// templateRenderer is a custom html/template renderer for Echo framework
type templateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	if err := t.templates.ExecuteTemplate(w, name, data); err != nil {
		c.Logger().Errorf("Rendering error: %s", err)
		return err
	}

	return nil
}
