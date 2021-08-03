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
	"math/rand"
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

	var headingTitles []string
	var subHeadingTitles []string

	markdownOutputBuffer := bytes.Buffer{}

	templateFunctions := template.FuncMap{
		"GetPageTitle": func(idx int) string {
			return headingTitles[idx]
		},
	}
	templates := ReadTemplates(templateFunctions)

	err := templates.Lookup("baseMarkdown").Execute(&markdownOutputBuffer, doc)
	if err != nil {
		fmt.Println("Error parsing markdown", err)
		os.Exit(1)
	}

	markdownOutputBytes := append([]byte{}, markdownOutputBuffer.Bytes()...)

	type Page struct {
		Title string
		Body template.HTML
		PageLinks map[int]string
		CurrentPage int
		ModuleDoc *ModuleDoc
	}

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
					subHeadingTitles = append(subHeadingTitles, t)
				}
			}
		}
		return mdAst.WalkContinue, nil
	})
	htmlBuffer := bytes.Buffer{}
	err = goldmark.New(goldmark.WithExtensions(extension.GFM)).Renderer().Render(&htmlBuffer, markdownOutputBytes, markdownAST)
	if err != nil {
		fmt.Println("Error rendering markdown to HTML", err)
		os.Exit(1)
	}

	var pageLinks = map[int]string{}
	var pageLinksInverted = map[string]int{}

	realIndex := 0
	for counter, s := range strings.Split(htmlBuffer.String(), "<h1></h1>") {
		if counter == 0 {
			continue
		}
		if realIndex == 0 {
			pageLinks[0] = "index.html"
			pageLinksInverted["index"] = 0
		} else {
			dumbLink := strings.Join(strings.Fields(headingTitles[realIndex]), "-")
			if _, exists := pageLinksInverted[dumbLink]; exists {
				dumbLink += fmt.Sprintf("%d", rand.Uint32())
			}
			pageLinks[realIndex] = dumbLink + ".html"
			pageLinksInverted[dumbLink] = realIndex
		}
		defer func(realIndex int, s string) {
			distFile := CreateDist(pageLinks[realIndex])
			thisPage := Page{
				Title:     headingTitles[realIndex],
				Body:      template.HTML(s),
				PageLinks: pageLinks,
				CurrentPage: realIndex,
				ModuleDoc: doc,
			}
			templates.Lookup("baseHTML").Execute(distFile, thisPage)
		}(realIndex, s)
		realIndex += 1
	}

}
