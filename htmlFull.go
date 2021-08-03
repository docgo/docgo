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
	os.RemoveAll("./out")
	buf := bytes.Buffer{}

	ReadTempl("/html/DOCS.tmpl", template.FuncMap{
		"A": func(x int) int{ return 5},
	}).Execute(&buf, doc)

	//htmlBytes := append([]byte{}, htmlBuf.Bytes()...)
	markdownBytes := append([]byte{}, buf.Bytes()...)

	//htmlString := htmlBuf.String()

	type Page struct {
		Title string
		Body template.HTML
		Menu []string
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
					head := ast2.NewString([]byte("[godoc:heading]"))
					n.InsertBefore(n.Parent(), n, head)
					n.InsertAfter(n.Parent(), n, ast2.NewString([]byte("[godoc:heading]")))
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

	const SEP = "[godoc:heading]"
	realIdx := 0
	for i, s := range strings.Split(step2.String(), SEP) {
		if i % 2 == 1 || i < 2  { continue }
		distFile := CreateDist(fmt.Sprintf("%d", realIdx) + ".html")
		//headingBuf := bytes.Buffer{}
		//goldmark.New().Convert([]byte(s), &headingBuf)
		thisPage := Page{
			Title: h1s[realIdx],
			Body:  template.HTML(s),
			Menu:  h1s,
		}
		ReadTempl("/html/base.html", nil).Execute(distFile, thisPage)
		realIdx += 1
	}

}
