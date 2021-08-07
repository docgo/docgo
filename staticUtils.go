package main

import (
	"embed"
	templateHtml "html/template"
	"io"
	"io/fs"
	"os"
	"text/template"
)

//go:embed static
var staticFS embed.FS
var templateFs fs.FS

func setTemplateFs() {
	if os.Getenv("TERMINAL_EMULATOR") == "JetBrains-JediTerm" || os.Getenv("APPDEBUG") == "1" {
		templateFs = os.DirFS("./")
	} else {
		templateFs = staticFS
	}
}
func ReadStaticFile(name string) []byte {
	setTemplateFs()
	f, err := templateFs.Open(name)
	if err != nil {
		panic(err)
	}
	out, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return out
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
