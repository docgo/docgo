package main

import (
	"os"
	"github.com/markbates/pkger"
	"io"
	oldFmt "fmt"
	"errors"
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	mdAst "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"strings"
	"path/filepath"
	"math/rand"
	"github.com/fatih/color"
	"text/template"
	templateHtml "html/template"
)

func CreateDist(file string) *os.File {
	ferr := os.Mkdir(Cli.Out, 0755)
	if ferr != nil {
		if !errors.Is(ferr, os.ErrExist) {
			fmt.Red("creating dist folder error", ferr)
			os.Exit(1)
		}
	}
	f, _ := os.Create(filepath.Join(Cli.Out, file))
	return f
}
func ReadTemplates(funcMap template.FuncMap) *template.Template {
	t := template.New("main")
	if funcMap != nil {
		t.Funcs(funcMap)
	}
	var TEMPLATES = map[string]string{"baseMarkdown": "/html/base.md", "snippet": "/html/snippet.md"}
	for templateName, templatePath := range TEMPLATES {
		file, err := pkger.Open(templatePath)
		if err != nil {
			fmt.Red("Error opening", templatePath, err)
			os.Exit(1)
		}
		templateRawBytes, err := io.ReadAll(file)
		if err != nil {
			fmt.Red("Error reading", templatePath, err)
			os.Exit(1)
		}

		_, err = t.New(templateName).Parse(string(templateRawBytes))
		if err != nil {
			fmt.Red("Error in template", templateName, err)
			os.Exit(1)
		}
	}
	return t
}
func ReadBaseHTMLTemplate(funcMap templateHtml.FuncMap) *templateHtml.Template{
	file, err := pkger.Open("/html/base.html")
	if err != nil {
		fmt.Red("Error opening base.html", err)
		os.Exit(1)
	}
	templateRawBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Red("Error reading base.html", err)
		os.Exit(1)
	}

	t, err := templateHtml.New("baseHtml").Funcs(funcMap).Parse(string(templateRawBytes))
	if err != nil {
		fmt.Red("Error in base.html", err)
		os.Exit(1)
	}

	return t
}

type PkgConfig string

func (c PkgConfig) Group(x ...string) {
	fmt.Debug(x)
}
func GenerateHTML(doc *ModuleDoc) {
	os.RemoveAll(Cli.Out)

	var headingTitles []string
	var subHeadingTitles []string

	markdownOutputBuffer := bytes.Buffer{}
	githubRepo := ""
	templateFunctions := template.FuncMap{
		"GitHubRepo": func(repo string) string {
			githubRepo = repo
			return strings.Repeat(repo, 0)
		},
		"TransformDoc": func(source string) string {
			source = strings.ReplaceAll(source, "\r", "")
			lastLevel := -1
			finalOut := ""
			for _, line := range strings.Split(source, "\n") {
				line = strings.ReplaceAll(line, "\t", "    ")
				normalLen := len(line)
				trimmedLen := len(strings.TrimLeft(line, " "))
				indentLevel := normalLen - trimmedLen
				if lastLevel == -1 {
					lastLevel = indentLevel
				}
				out := strings.TrimLeft(line, " ")
				if indentLevel > lastLevel {
					out = "\n```go\n" + out
				}
				if indentLevel < lastLevel {
					out = "```\n" + out
				}
				lastLevel = indentLevel
				finalOut += out + "\n"
			}
			return finalOut
		},
		"PackageConfig": func(pkgName string) PkgConfig {
			return PkgConfig(pkgName)
		},
	}
	templates := ReadTemplates(templateFunctions)
	htmlTemplate := ReadBaseHTMLTemplate(templateHtml.FuncMap{
		"GetPageTitle": func(idx int) string {
			return headingTitles[idx]
		},
	})

	err := templates.Lookup("baseMarkdown").Execute(&markdownOutputBuffer, doc)
	if err != nil {
		fmt.Red("Error parsing markdown", err)
		os.Exit(1)
	}

	markdownOutputBytes := append([]byte{}, markdownOutputBuffer.Bytes()...)

	type Page struct {
		Title       string
		Body        templateHtml.HTML
		PageLinks   map[int]string
		CurrentPage int
		ModuleDoc   *ModuleDoc
		GitHubRepo  string
	}

	markdownAST := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(markdownOutputBytes))

	mdAst.Walk(markdownAST, func(n mdAst.Node, entering bool) (mdAst.WalkStatus, error) {
		if n.Kind() == mdAst.KindHeading {
			nHeading := n.(*mdAst.Heading)
			if !entering {
				t := oldFmt.Sprintf("%s", n.Text(markdownOutputBytes))

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
		fmt.Red("Error rendering markdown to HTML", err)
		os.Exit(1)
	}

	var pageLinks = map[int]string{}
	var pageLinksInverted = map[string]int{}
	var pageNameToSearchableContent = map[string]string {}

	realIndex := 0
	for counter, s := range strings.Split(htmlBuffer.String(), "<h1></h1>") {
		if counter == 0 {
			continue
		}
		pageName := ""
		if realIndex == 0 {
			pageName = "index"
			pageLinks[0] = "index.html"
			pageLinksInverted[pageName] = 0
		} else {
			dumbLink := strings.Join(strings.Fields(headingTitles[realIndex]), "-")
			if _, exists := pageLinksInverted[dumbLink]; exists {
				dumbLink += oldFmt.Sprintf("%d", rand.Uint32())
			}
			pageName = dumbLink
			pageLinks[realIndex] = dumbLink + ".html"
			pageLinksInverted[dumbLink] = realIndex
		}
		pageNameToSearchableContent[pageName] = s
		defer func(realIndex int, s string, repo string) {
			distFile := CreateDist(pageLinks[realIndex])
			thisPage := Page{
				Title:       headingTitles[realIndex],
				Body:        templateHtml.HTML(s),
				GitHubRepo:  repo,
				PageLinks:   pageLinks,
				CurrentPage: realIndex,
				ModuleDoc:   doc,
			}
			err := htmlTemplate.Execute(distFile, thisPage)
			if err != nil {
				fmt.Red(err)
				return
			}
		}(realIndex, s, githubRepo)
		realIndex += 1
	}
	GenerateSearch(pageNameToSearchableContent)
	color.Green("Generated docs ✔")
}
