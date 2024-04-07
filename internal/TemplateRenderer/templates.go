package TemplateRenderer

import (
	"html/template"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/relepega/doujinstyle-downloader/internal/taskQueue"
)

type UIRenderAction struct {
	Action   string `json:"action"`
	Target   string `json:"target"`
	Receiver string `json:"receiver"`
	Node     string `json:"node"`
}

func NewUIRenderAction(action, target, receiver string, task *taskQueue.Task) UIRenderAction {
	node := NewTemplates().RenderToString("task", task)

	return UIRenderAction{
		Action:   action,
		Target:   target,
		Receiver: receiver,
		Node:     node,
	}
}

type Templates struct {
	templates *template.Template
}

func (t *Templates) RenderToString(name string, data interface{}) string {
	buf := new(strings.Builder)
	t.templates.ExecuteTemplate(buf, name, data)

	return buf.String()
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("./views/*.html")),
	}
}
