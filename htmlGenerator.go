package main

import (
	"errors"
	"github.com/docgo/docgo/markdownAnnotate"
	"os"
	"path/filepath"
	"strings"
	"github.com/microcosm-cc/bluemonday"
	"html"
)

func markdownToCleanText(markdown string) string {
	markdown = markdownAnnotate.RenderPage(markdown)
	strict := bluemonday.StrictPolicy().SkipElementsContent("code")
	return html.UnescapeString(strict.Sanitize(markdown))
}

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