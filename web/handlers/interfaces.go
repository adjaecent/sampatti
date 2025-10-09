package handlers

import "net/http"

type TemplateRenderer interface {
	Render(w http.ResponseWriter, templateName string, data interface{}) error
}