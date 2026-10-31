package web

import (
	"embed"
	"net/http"
)

//go:embed template/*
var TemplateFS embed.FS

// FileServer returns an http.Handler for the embedded web assets.
func FileServer() http.Handler {
	return http.FileServer(http.FS(TemplateFS))
}
