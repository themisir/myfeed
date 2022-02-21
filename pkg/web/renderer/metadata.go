package renderer

import (
	"github.com/labstack/echo/v4"
	"io"
)

type MetadataRenderer struct {
	base      echo.Renderer
	values    map[string]interface{}
	dynValues map[string]ResolveFunc
}

type ResolveFunc = func(c echo.Context) interface{}

func Metadata(base echo.Renderer) *MetadataRenderer {
	return &MetadataRenderer{
		base:      base,
		values:    map[string]interface{}{},
		dynValues: map[string]ResolveFunc{},
	}
}

// Set creates new static value on given key
func (m *MetadataRenderer) Set(key string, value interface{}) { m.values[key] = value }

// SetDyn creates new dynamic value on given key
func (m *MetadataRenderer) SetDyn(key string, r ResolveFunc) { m.dynValues[key] = r }

func (m *MetadataRenderer) enrich(data interface{}, c echo.Context) interface{} {
	var dataMap echo.Map
	if data, ok := data.(echo.Map); ok {
		dataMap = data
	}
	if dataMap == nil {
		dataMap = echo.Map{}
	}

	for k, v := range m.dynValues {
		dataMap[k] = v(c)
	}
	for k, v := range m.values {
		dataMap[k] = v
	}
	return dataMap
}

func (m *MetadataRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return m.base.Render(w, name, m.enrich(data, c), c)
}
