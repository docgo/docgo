package markdownAnnotate

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	gast "github.com/yuin/goldmark/ast"
	"bytes"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
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
				if dg.Attrs["page"] {
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
	k := parser.WithASTTransformers(util.Prioritized(annotationTransformer{}, 0))
	_ = k
	goldmark.New(goldmark.WithExtensions(extension.GFM)).Convert([]byte(markdownString), w)
	return w.String()
}

type annotationTransformer struct{}

func (annotationTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if n.PreviousSibling() == nil || n.PreviousSibling().Kind() != DocGoKind {
			return gast.WalkContinue, nil
		}
		d := n.PreviousSibling().(*DocGoNode)
		_ = d.StringVars
		return gast.WalkContinue, nil
	})
}
