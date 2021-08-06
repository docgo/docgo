package main

import (
	"os"
	oldFmt "fmt"
	"errors"
	"bytes"
	"strings"
	"path/filepath"
	"math/rand"
	"github.com/fatih/color"
	"text/template"
	templateHtml "html/template"
	"encoding/json"
	"github.com/docgo/docgo/markdownAnnotate"
)

func transformGodocToMarkdown(godocString string) string {
	const TABWIDTH = "    "
	const MARKDOWN_CODEFENCE = "```"

	// Remove all CR
	godocString = strings.ReplaceAll(godocString, "\r", "")
	// Keep track of indentation
	lastIndentationLevel := -1
	finalOut := ""

	// Increased indentation in a godoc comment always
	// means that the line begins a code/quote block.
	// Example:
	// commentBegin123
	//    code1
	//    code2
	// commentEnd123

	for _, line := range strings.Split(godocString, "\n") {
		line = strings.ReplaceAll(line, "\t", TABWIDTH)
		normalLen := len(line)
		trimmedLen := len(strings.TrimLeft(line, " "))
		indentLevel := normalLen - trimmedLen
		if lastIndentationLevel == -1 {
			lastIndentationLevel = indentLevel
		}
		out := strings.TrimLeft(line, " ")
		if indentLevel > lastIndentationLevel {
			out = "\n" + MARKDOWN_CODEFENCE + "\n" + out
		}
		if indentLevel < lastIndentationLevel {
			out = MARKDOWN_CODEFENCE + "\n" + out
		}
		lastIndentationLevel = indentLevel
		finalOut += out + "\n"
	}
	return finalOut
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
			return transformGodocToMarkdown(source)
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
	pages := []string{}
	pages, headingTitles = markdownAnnotate.SplitPages(markdownOutputBytes)
	var cleanPages = []string{}
	for _, page := range pages {
		cleanPages = append(cleanPages, markdownAnnotate.CleanPage(page))
	}

	type Page struct {
		Title       string
		Body        templateHtml.HTML
		PageLinks   map[int]string
		CurrentPage int
		ModuleDoc   *ModuleDoc
		SiteInfo  map[string]string
	}

	var pageLinks = map[int]string{}
	var pageLinksInverted = map[string]int{}
	var pageNameToSearchableContent = map[string]string {}

	realIndex := 0
	for counter, page := range pages {
		s := markdownAnnotate.RenderPage(page)
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
