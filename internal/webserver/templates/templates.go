package templates

import (
	"bytes"
	"html/template"
	"io"
)

// templateFile defines the contents of a template to be stored in a file, for testing.
type Templates struct {
	templates *template.Template
	functions template.FuncMap
}

func (t *Templates) AddFunction(name string, fn interface{}) {
	// t.templates.Funcs(template.FuncMap{name: fn})
	t.functions[name] = fn
}

func (t *Templates) ParseFiles(files []string) error {
	tmpl, err := template.New("").Funcs(t.functions).ParseFiles(files...)
	if err != nil {
		return err
	}

	t.templates = tmpl

	return nil
}

func (t *Templates) ParseGlob(pattern string) error {
	tmpl, err := template.New("").Funcs(t.functions).ParseGlob(pattern)
	if err != nil {
		return err
	}

	t.templates = tmpl

	return nil
}

func (t *Templates) Execute(name string, data any) (string, error) {
	var buf bytes.Buffer

	err := t.templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (t *Templates) ExecuteWithWriter(w io.Writer, name string, data any) error {
	err := t.templates.ExecuteTemplate(w, name, data)

	if err != nil {
		w.Write([]byte(err.Error()))
		return err
	}

	return nil
}

/*
Returns an enpty Templates struct

# Usage

	t, err := templates.NewTemplates()

	if err != nil {
		...
	}

	t.AddFunction("add", func(a, b int) int {
		return a + b
	})

	t.ParseFiles([]string{"templates/index.html", "templates/footer.html"})

	t.Execute("index", nil)
*/
func NewTemplates() (*Templates, error) {
	r := &Templates{
		templates: nil,
		functions: template.FuncMap{},
	}

	return r, nil
}
