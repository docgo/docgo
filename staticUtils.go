package main

import (
	"embed"
	"text/template"
	templateHtml "html/template"
)

//go:embed static
var staticFS embed.FS

func LoadMarkdownTemplates(funcMap template.FuncMap) *template.Template {
	t, err := template.New("").Funcs(funcMap).ParseFS(staticFS, "static/*.md")
	if err != nil {
		fmt.Red("Error loading templates", err)
	}
	return t
}

func LoadHTMLTemplates(funcMap templateHtml.FuncMap) *templateHtml.Template {
	tpl, err := templateHtml.New("").Funcs(funcMap).ParseFS(staticFS, "static/*.html")
	if err != nil {
		fmt.Red("Error loading templates", err)
	}
	return tpl
}
