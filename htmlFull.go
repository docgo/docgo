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
	mdAst "github.com/yuin/goldmark/ast"
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
	t.New("snip").Parse(`
{{ range $idx, $p := . }}
### {{ .Name }} [{{.FoundInFile}}]
{{ .Doc }}
` + "```" + `go
{{ .Snippet }}
` + "```" + `
{{ end }}
`)
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
	markdownOutputBuffer := bytes.Buffer{}

	ReadTempl("/html/DOCS.tmpl", template.FuncMap{
		"A": func(x int) int{ return 5},
	}).Execute(&markdownOutputBuffer, doc)

	//htmlBytes := append([]byte{}, htmlBuf.Bytes()...)
	markdownOutputBytes := append([]byte{}, markdownOutputBuffer.Bytes()...)
	//fmt.Printf("%s\n", markdownOutputBytes)

	//htmlString := htmlBuf.String()

	type Page struct {
		Title string
		Body template.HTML
		Menu []string
	}
	//entries := []Entry{}

	var h1s []string
	var h2s []string

	markdownAST := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(markdownOutputBytes))

	mdAst.Walk(markdownAST, func(n mdAst.Node, entering bool) (mdAst.WalkStatus, error) {
		if n.Kind() == mdAst.KindHeading {
			nHeading := n.(*mdAst.Heading)
			if !entering {
				t := fmt.Sprintf("%s", n.Text(markdownOutputBytes))

				if nHeading.Level == 1 {
					h1s = append(h1s, t)
					//head := mdAst.NewString([]byte("[godoc:heading]"))
					//n.InsertBefore(n.Parent(), n, head)
					n.RemoveChildren(n)
					//n.InsertAfter(n.Parent(), n, mdAst.NewString([]byte("[godoc:heading]")))
				}
				if nHeading.Level == 2 {
					h2s = append(h2s, t)
					//eentries = append(entries, currentEntry)
				}
			}
		}
		return mdAst.WalkContinue, nil
	})
	step2 := bytes.Buffer{}
	goldmark.New(goldmark.WithExtensions(extension.GFM)).Renderer().Render(&step2, markdownOutputBytes, markdownAST)

	realIndex := 0
	for counter, s := range strings.Split(step2.String(), "<h1></h1>") {
		if counter == 0 {
			continue
		}

		distFile := CreateDist(fmt.Sprintf("%d", realIndex) + ".html")
		thisPage := Page{
			Title: h1s[realIndex],
			Body:  template.HTML(s),
			Menu:  h1s,
		}
		ReadTempl("/html/base.html", nil).Execute(distFile, thisPage)
		realIndex += 1
	}

}
