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
func ReadTemplates(funcMap template.FuncMap) *template.Template{
	t := template.New("main")
	if funcMap != nil {
		t.Funcs(funcMap)
	}
	var TEMPLATES = map[string]string{"baseHTML": "/html/base.html", "baseMarkdown": "/html/base.md", "snippet": "/html/snippet.md"}
	for templateName, templatePath := range TEMPLATES {
		file, err := pkger.Open(templatePath)
		if err != nil {
			fmt.Println("Error opening", templatePath, err)
			os.Exit(1)
		}
		templateRawBytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Println("Error reading", templatePath, err)
			os.Exit(1)
		}

		_, err = t.New(templateName).Parse(string(templateRawBytes))
		if err != nil {
			fmt.Println("Error in template", templateName, err)
		}
	}
	return t
}


func GenerateHTML2(doc *ModuleDoc)  {
	os.RemoveAll("./out")
	markdownOutputBuffer := bytes.Buffer{}

	templates := ReadTemplates(template.FuncMap{

	})

	templates.Lookup("baseMarkdown").Execute(&markdownOutputBuffer, doc)

	markdownOutputBytes := append([]byte{}, markdownOutputBuffer.Bytes()...)

	type Page struct {
		Title string
		Body template.HTML
		Menu []string
	}

	var headingTitles []string
	var h2s []string

	markdownAST := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(markdownOutputBytes))

	mdAst.Walk(markdownAST, func(n mdAst.Node, entering bool) (mdAst.WalkStatus, error) {
		if n.Kind() == mdAst.KindHeading {
			nHeading := n.(*mdAst.Heading)
			if !entering {
				t := fmt.Sprintf("%s", n.Text(markdownOutputBytes))

				if nHeading.Level == 1 {
					headingTitles = append(headingTitles, t)
					n.RemoveChildren(n)
				}
				if nHeading.Level == 2 {
					h2s = append(h2s, t)
				}
			}
		}
		return mdAst.WalkContinue, nil
	})
	htmlBuffer := bytes.Buffer{}
	goldmark.New(goldmark.WithExtensions(extension.GFM)).Renderer().Render(&htmlBuffer, markdownOutputBytes, markdownAST)

	realIndex := 0
	for counter, s := range strings.Split(htmlBuffer.String(), "<h1></h1>") {
		if counter == 0 {
			continue
		}

		distFile := CreateDist(fmt.Sprintf("%d", realIndex) + ".html")
		thisPage := Page{
			Title: headingTitles[realIndex],
			Body:  template.HTML(s),
			Menu:  headingTitles,
		}
		templates.Lookup("baseHTML").Execute(distFile, thisPage)
		realIndex += 1
	}

}
