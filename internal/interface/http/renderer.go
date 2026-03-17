package http

import (
	"fmt"
	"html/template"
	"io"
	"strings"

	domain "mltestsuite/internal/domain/testing"
)

const templatesDir = "internal/interface/templates"

// funcMap contiene funciones auxiliares disponibles en todos los templates.
var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"eq":  func(a, b string) bool { return a == b },
	"dict": func(pairs ...any) map[string]any {
		m := make(map[string]any, len(pairs)/2)
		for i := 0; i < len(pairs)-1; i += 2 {
			key, ok := pairs[i].(string)
			if ok {
				m[key] = pairs[i+1]
			}
		}
		return m
	},
	"statusLabel": func(s domain.ExecutionStatus) string {
		return s.Label()
	},
	"statusColor": func(s domain.ExecutionStatus) string {
		return s.Color()
	},
	"allStatuses": func() []domain.ExecutionStatus {
		return domain.AllStatuses()
	},
	"derefBool": func(b *bool) bool {
		if b == nil {
			return false
		}
		return *b
	},
	"not": func(v any) bool {
		if v == nil {
			return true
		}
		return false
	},
}

// Renderer compila un set de templates por pagina.
type Renderer struct {
	templates map[string]*template.Template
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{templates: make(map[string]*template.Template)}

	layout := templatesDir + "/layout.html"
	executionRow := templatesDir + "/partials/execution_row.html"

	// Pages that use the layout
	pages := []string{
		"releases/list.html",
		"releases/new.html",
		"releases/show.html",
		"testcases/list.html",
		"testcases/by_report.html",
		"testcases/new.html",
		"testcases/show.html",
		"testcases/edit.html",
		"testcases/import.html",
		"reports/list.html",
		"reports/new.html",
		"reports/edit.html",
		"teams/list.html",
		"teams/new.html",
		"teams/edit.html",
		"knowledge/show.html",
		"admin/users_list.html",
		"admin/user_form.html",
	}
	for _, page := range pages {
		files := []string{layout, templatesDir + "/" + page}
		// releases/show needs execution_row partial
		if page == "releases/show.html" {
			files = append(files, executionRow)
		}
		t, err := template.New("").Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("error cargando %s: %w", page, err)
		}
		r.templates[page] = t
	}

	// Standalone templates (sin layout)
	standalones := []string{
		"auth/login.html",
		"auth/register.html",
	}
	for _, s := range standalones {
		t, err := template.New("").Funcs(funcMap).ParseFiles(templatesDir + "/" + s)
		if err != nil {
			return nil, fmt.Errorf("error cargando %s: %w", s, err)
		}
		r.templates[s] = t
	}

	// Partials (respuestas HTMX)
	partials := []string{
		"partials/execution_row.html",
	}
	for _, p := range partials {
		t, err := template.New("").Funcs(funcMap).ParseFiles(templatesDir + "/" + p)
		if err != nil {
			return nil, fmt.Errorf("error cargando %s: %w", p, err)
		}
		r.templates[p] = t
	}

	return r, nil
}

// ExecuteTemplate renderiza el template correcto.
func (r *Renderer) ExecuteTemplate(w io.Writer, name string, data any) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template no encontrado: %s", name)
	}
	// Partials: ejecutar el bloque definido por nombre
	if strings.HasPrefix(name, "partials/") {
		return t.ExecuteTemplate(w, name, data)
	}
	// Standalone: ejecutar por nombre base del archivo
	if strings.HasPrefix(name, "auth/") {
		parts := strings.Split(name, "/")
		return t.ExecuteTemplate(w, parts[len(parts)-1], data)
	}
	// Paginas con layout
	return t.ExecuteTemplate(w, "layout.html", data)
}
