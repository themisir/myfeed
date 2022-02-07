package renderer

import (
	"bytes"
	"io"

	"github.com/labstack/echo/v4"
)

func Layout(layout string, inner echo.Renderer) echo.Renderer {
	return &layoutedRenderer{
		base:   inner,
		layout: layout,
	}
}

// Custom renderer for encapsulating inner page on outer layout templates
type layoutedRenderer struct {
	base   echo.Renderer
	layout string
}

// Data struct passed to layout tempaltes
type layoutData struct {
	Inner string
	Data  interface{}
}

// Render renders a template document
func (t *layoutedRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	inner := new(bytes.Buffer)

	// Render inner template
	if err := t.base.Render(inner, name, data, c); err != nil {
		return err
	}

	// Render outer template
	return t.base.Render(w, t.layout, layoutData{Inner: inner.String(), Data: data}, c)
}
