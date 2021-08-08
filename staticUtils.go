package main

import (
	"embed"
	templateHtml "html/template"
	"io"
	"io/fs"
	"os"
	"text/template"
	"strings"
	"encoding/base64"
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
	defer f.Close()
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

// Given a filename, GetStaticCSS will fetch the corresponding file
// from staticFS memory and cast it to a template.CSS type so that
// it can be embedded into a Golang template in a safe way.
func GetStaticCSS(filename string) templateHtml.CSS {
	setTemplateFs()
	style := string(ReadStaticFile(filename))
	style = strings.ReplaceAll(style, "\n", "")
	style = strings.ReplaceAll(style, "\r", "")
	return templateHtml.CSS(style)
}

// Given a filename, GetStaticSVG will fetch the correponding file
// and give you a template.URL representing the SVG using Data URI
// scheme with base64 encoding.
func GetStaticSVG(filename string) templateHtml.URL {
	setTemplateFs()
	base64Content := base64.StdEncoding.EncodeToString(ReadStaticFile(filename))
	return templateHtml.URL("data:image/svg+xml;base64," + base64Content)
}

// Returns a FuncMap containing UsualFuncMap + some
// extra ones supplied as arguments. The arguments
// should look like an unzipped list:
//     name1, fn1, name2, fn2, ...
func cookFuncmap(fnNamePair ...interface{}) templateHtml.FuncMap{
	var funcMap = templateHtml.FuncMap{}

	// Add the usual ingredients
	for x, y := range UsualFuncMap {
		funcMap[x] = y
	}

	// Mix new ones
	for i, item := range fnNamePair {
		if i % 2 == 0 { continue }
		name, ok := fnNamePair[i - 1].(string)
		if !ok {
			return funcMap
		}
		funcMap[name] = item
	}

	// Stir and bake
	return funcMap
}

var UsualFuncMap = templateHtml.FuncMap{
	"GetStaticCSS": GetStaticCSS,
	"GetStaticSVG": GetStaticSVG,
}