package resource

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*.html
var templateFS embed.FS

type Renderer interface {
	// Render a template to writer with coresponding template name and data (data is optional)
	Render(w http.ResponseWriter, name string, data any)
}

type RendererImpl struct {
	template *template.Template
}

func (renderer *RendererImpl) Render(w http.ResponseWriter, name string, data any) {
	renderer.template.ExecuteTemplate(w, name, data)
}

// return a new [Renderer]
func NewRenderer() Renderer {
	return &RendererImpl{
		template: template.Must(template.ParseFS(templateFS, "templates/*.html")),
	}
}
