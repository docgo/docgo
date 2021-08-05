package customMarkdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	gast "github.com/yuin/goldmark/ast"
)

func SplitPages(markdownOutputBytes []byte) []string{
	pageIdx := make([]int, 0)
	mdString := string(markdownOutputBytes)
	cleanAST := goldmark.New(goldmark.WithExtensions(extension.GFM, DocgoExtension)).Parser().Parse(text.NewReader(markdownOutputBytes))
	gast.Walk(cleanAST, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			if n.Kind() == DocGoKind {
				dg := n.(*DocGoNode)
				pageIdx = append(pageIdx, dg.LineEnd)
			}
		}
		return gast.WalkContinue, nil
	})
	lastIdx := -1
	documents := make([]string, 0)
	for i, idx := range pageIdx {
		if i == 0 {
			lastIdx = 0
		}
		documents = append(documents, mdString[lastIdx:idx])
		lastIdx = idx
	}
	return documents
}