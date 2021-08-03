package main

import (
	"os"
	"github.com/markbates/pkger"
	"io"
	"fmt"
	"errors"
	"html/template"
	"bytes"
	"go/ast"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	ast2 "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"strings"
	"path/filepath"
)

type Meta struct {
	Packages map[string]*ast.Package
	PackageNames []string
	Content template.HTML
}
func CreateDist(file string) *os.File{
	ferr := os.Mkdir("out", 0755)
	if ferr != nil {
		if !errors.Is(ferr, os.ErrExist) {
			fmt.Println("creating dist folder error", ferr)
			os.Exit(1)
		}
	}
	f, _ := os.Create(filepath.Join(".", "out", file))
	return f
}
func ReadTempl(name string, funcMap template.FuncMap) *template.Template{
	t := template.New("main")
	raw, err := pkger.Open(name)
	if err != nil {
		fmt.Println("pkger error: ", err)
		os.Exit(1)
	}
	if funcMap != nil {
		t.Funcs(funcMap)
	}
	data, _ := io.ReadAll(raw)
	templ, err := t.Parse(string(data))
	if err != nil {
		fmt.Println("template loading error", err)
		os.Exit(1)
	}
	return templ
}


func GenerateHTML2(doc *ModuleDoc)  {
	buf := bytes.Buffer{}

	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	ReadTempl("/html/DOCS.tmpl", template.FuncMap{
		"A": func(x int) int{ return 5},
	}).Execute(&buf, doc)
	htmlBuf := bytes.Buffer{}
	markdownBytes := append([]byte{}, buf.Bytes()...)
	md.Convert(markdownBytes, &htmlBuf)
	//htmlString := htmlBuf.String()

	type Entry struct {
		h1 string
		h2 string
		markdownRest string
	}
	//entries := []Entry{}

	var h1s []string
	var h2s []string
	parsedMarkdown := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(markdownBytes))

	ast2.Walk(parsedMarkdown, func(n ast2.Node, entering bool) (ast2.WalkStatus, error) {
		if n.Kind() == ast2.KindHeading {
			nHeading := n.(*ast2.Heading)
			if !entering {
				t := fmt.Sprintf("%s", n.Text(markdownBytes))

				if nHeading.Level == 1 {
					h1s = append(h1s, t)
					n.Parent().ReplaceChild(n.Parent(), n, ast2.NewString([]byte("```godoc\nheading_1\n```")))
				}
				if nHeading.Level == 2 {
					h2s = append(h2s, t)
					//eentries = append(entries, currentEntry)
				}
			}
		}
		return ast2.WalkContinue, nil
	})
	step2 := bytes.Buffer{}
	goldmark.DefaultRenderer().Render(&step2, markdownBytes, parsedMarkdown)

	const SEP = "```godoc\nheading_1\n```"
	for i, s := range strings.Split(step2.String(), SEP) {
		distFile := CreateDist(h1s[i] + ".html")
		fmt.Println(s)
		headingBuf := bytes.Buffer{}
		goldmark.New().Convert([]byte(s), &headingBuf)
		ReadTempl("/html/base.html", nil).Execute(distFile, template.HTML(headingBuf.String()))
	}

}
