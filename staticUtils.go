package main

import (
	"embed"
	"text/template"
	templateHtml "html/template"
	"io/fs"
	"os"
)

//go:embed static
var staticFS embed.FS
var templateFs fs.FS

func setTemplateFs() {
	if os.Getenv("TERMINAL_EMULATOR") == "JetBrains-JediTerm" {
		templateFs = os.DirFS("./")
	} else {
		templateFs = staticFS
	}
}
func LoadMarkdownTemplates(funcMap template.FuncMap) *template.Template {
	setTemplateFs()
	t, err := template.New("").Funcs(funcMap).ParseFS(templateFs, "static/*.md")
	if err != nil {
		fmt.Red("Error loading templates", err)
	}
	return t
}

func LoadHTMLTemplates(funcMap templateHtml.FuncMap) *templateHtml.Template {
	setTemplateFs()
	tpl, err := templateHtml.New("").Funcs(funcMap).ParseFS(templateFs, "static/*.html")
	if err != nil {
		fmt.Red("Error loading templates", err)
	}
	return tpl
}