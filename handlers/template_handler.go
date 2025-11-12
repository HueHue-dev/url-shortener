package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

type Metrics struct {
	Count int64
}

type TemplateData struct {
	Title          string
	Error          string
	ShortURL       string
	MetricsURL     string
	ExpirationDate string
	Metrics        Metrics
	Content        interface{}
}

type TemplateHandler struct {
	templates map[string]*template.Template
}

func NewTemplateHandler(templateDir string) (*TemplateHandler, error) {
	templates := make(map[string]*template.Template)

	basePath := filepath.Join(templateDir, "base.html")

	pageTemplates := []string{"home.html", "shorten.html", "metrics.html"}

	for _, page := range pageTemplates {
		pagePath := filepath.Join(templateDir, page)
		tmpl, err := template.ParseFiles(basePath, pagePath)
		if err != nil {
			return nil, err
		}
		templates[page] = tmpl
	}

	return &TemplateHandler{
		templates: templates,
	}, nil
}

func (th *TemplateHandler) Render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := th.templates[name]
	if !ok {
		log.Printf("Template %s not found", name)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (th *TemplateHandler) RenderWithStatus(w http.ResponseWriter, name string, status int, data interface{}) {
	w.WriteHeader(status)
	th.Render(w, name, data)
}
