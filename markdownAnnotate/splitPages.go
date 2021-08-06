package markdownAnnotate

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	gast "github.com/yuin/goldmark/ast"
	"bytes"
)
type pair struct { from, to int }

// Splits a single Markdown source string into
// N strings that represent Markdown sections.
// The sections are defined using Markdown annotations.
func SplitPages(markdownOutputBytes []byte) ([]string, []string){
	pageIdx := []pair{}
	titles := make([]string, 0)
	mdString := string(markdownOutputBytes)
	cleanAST := goldmark.New(goldmark.WithExtensions(extension.GFM, DocgoExtension)).Parser().Parse(text.NewReader(markdownOutputBytes))
	gast.Walk(cleanAST, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			if n.Kind() == DocGoKind {
				dg := n.(*DocGoNode)
				if dg.BoolVars["page"] {
					pageIdx = append(pageIdx, pair{dg.LineStart, dg.LineEnd})
					titles = append(titles, dg.StringVars["title"])
				}
			}
		}
		return gast.WalkContinue, nil
	})
	documents := make([]string, 0)
	for i, idx := range pageIdx {
		if i + 1 >= len(pageIdx) {
			documents = append(documents, mdString[idx.to:])
		} else {
			documents = append(documents, mdString[idx.to:pageIdx[i + 1].from])
		}
	}
	return documents, titles
}

// Renders a markdown source with GitHub flavor
// into HTML
func RenderPage(markdownString string) string {
	w := bytes.NewBufferString("")
	goldmark.New(goldmark.WithExtensions(extension.GFM, DocgoExtension)).Convert([]byte(markdownString), w)
	return w.String()
}