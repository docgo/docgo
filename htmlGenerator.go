package main

import (
	"os"
	oldFmt "fmt"
	"errors"
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	mdAst "github.com/yuin/goldmark/ast"
	"strings"
	"path/filepath"
	"math/rand"
	"github.com/fatih/color"
	"text/template"
	templateHtml "html/template"
	"encoding/json"
	"github.com/docgo/docgo/customMarkdown"
	"github.com/yuin/goldmark/text"
)

type IDGen struct{
}

func (I IDGen) Generate(value []byte, kind mdAst.NodeKind) []byte {
	fmt.Green(value)
	return []byte("TEST")
}

func (I IDGen) Put(value []byte) {
}

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
	_ = githubRepo
	siteInfo := make(map[string]string )
	templateFunctions := template.FuncMap{
		"GitHubRepo": func(repo string) string {
			githubRepo = repo
			return strings.Repeat(repo, 0)
		},
		"SetSiteInfo": func(keyValues ...string) string {
			if len(keyValues) == 0 || (len(keyValues) % 2) == 1 {
				fmt.Red("Invalid SetSiteInfo arguments. Must be even 'arg1' 'key1' ...")
				return ""
			}
			for i := 0; i < len(keyValues); i++ {
				if i % 2 == 1 { continue }
				siteInfo[keyValues[i]] = keyValues[i + 1]
			}
			return ""
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
					out = "\n```\n" + out
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
	//templates := ReadTemplates(templateFunctions)
	templates := LoadMarkdownTemplates(templateFunctions)
	htmlTemplates := LoadHTMLTemplates(templateHtml.FuncMap{
		"GetPageTitle": func(idx int) string {
			return headingTitles[idx]
		},
	})
	baseHtmlTemplate := htmlTemplates.Lookup("base.html")

	err := templates.Lookup("base.md").Execute(&markdownOutputBuffer, doc)
	if err != nil {
		fmt.Red("Error parsing markdown", err)
		os.Exit(1)
	}

	markdownOutputBytes := append([]byte{}, markdownOutputBuffer.Bytes()...)
	//pages := customMarkdown.SplitPages(markdownOutputBytes)
	var cleanPages = customMarkdown.CleanPage(markdownOutputBytes)

	type Page struct {
		Title       string
		Body        templateHtml.HTML
		PageLinks   map[int]string
		CurrentPage int
		ModuleDoc   *ModuleDoc
		SiteInfo  map[string]string
	}

	markdownAST := goldmark.New(goldmark.WithExtensions(extension.GFM, customMarkdown.DocgoExtension)).Parser().Parse(text.NewReader(markdownOutputBytes))
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
	err = goldmark.New(goldmark.WithExtensions(extension.GFM, customMarkdown.DocgoExtension)).Renderer().Render(&htmlBuffer, markdownOutputBytes, markdownAST)
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
		pageNameToSearchableContent[pageName] = cleanPages[counter]
		defer func(realIndex int, s string, siteInfo map[string]string) {
			distFile := CreateDist(pageLinks[realIndex])
			jsonIndex, _ := json.Marshal(pageNameToSearchableContent)
			thisPage := struct{
				Title       string
				Body        templateHtml.HTML
				PageLinks   map[int]string
				CurrentPage int
				ModuleDoc   *ModuleDoc
				SiteInfo  map[string]string
				SearchIndex templateHtml.JS
			}{
				headingTitles[realIndex], templateHtml.HTML(s),  pageLinks, realIndex, doc, siteInfo, templateHtml.JS(jsonIndex),
			}
			err := baseHtmlTemplate.Execute(distFile, thisPage)
			if err != nil {
				fmt.Red(err)
				return
			}
		}(realIndex, s, siteInfo)
		realIndex += 1
	}
	if FirstRun {
		color.Green("Generated docs ✔")
		FirstRun = false
	}
}

var FirstRun = true
