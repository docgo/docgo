package markdownAnnotate

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	gast "github.com/yuin/goldmark/ast"
	"bytes"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
	"fmt"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"strings"
	"html/template"
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
	goldmark.WithRendererOptions()
	w := bytes.NewBufferString("")

	cfRenderer := codeFenceRenderer{html.DefaultWriter, nil}
	renderOpt := renderer.WithNodeRenderers(util.Prioritized(&cfRenderer, 0))

	k := parser.WithASTTransformers(util.Prioritized(&annotationTransformer{[]byte(markdownString)}, 1000))
	_ = k
	goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(renderOpt)).Convert([]byte(markdownString), w)
	return w.String()
}

type codeFenceRenderer struct {
	html.Writer
	entries []string
}

func (r *codeFenceRenderer) writeLines(w util.BufWriter, source []byte, n gast.Node) {
	l := n.Lines().Len()
	buffer := ""
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		buffer += string(line.Value(source))
	}
	parts := strings.Split(buffer, "@[docgo-info-begin]")
	sourceCode := template.HTMLEscapeString(parts[0])
	if len(parts) == 2 {
		r.entries = make([]string, 0)
		for _, entry := range strings.Split(parts[1], "@[docgo-entry-end]") {
			entryParts := strings.Split(entry, "=")
			if len(entryParts) == 1 {
				break
			}
			entryName := entryParts[0]
			entryVal := entryParts[1]
			fmt.Println(entryName, entryVal)

			sourceCode = strings.ReplaceAll(sourceCode, entryName, fmt.Sprintf("<a href='#%s'>%s</a>", entryVal, entryName))
			r.entries = append(r.entries, entry) // process sth here?
		}
	}
	w.WriteString(sourceCode)
	//r.Writer.RawWrite(w, []byte(sourceCode))
}

func (r *codeFenceRenderer) RegisterFuncs(registerer renderer.NodeRendererFuncRegisterer) {
	registerer.Register(gast.KindFencedCodeBlock, func(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
		n := node.(*gast.FencedCodeBlock)
		if entering {
			_, _ = w.WriteString("<pre><code")
			language := n.Language(source)
			if language != nil {
				_, _ = w.WriteString(" class=\"language-")
				r.Writer.Write(w, language)
				_, _ = w.WriteString("\"")
			}
			_ = w.WriteByte('>')
			r.writeLines(w, source, n)
		} else {
			_, _ = w.WriteString("</code></pre>\n")
		}
		return gast.WalkContinue, nil
	})
}

type annotationTransformer struct{
	source []byte
}

func (t *annotationTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if n.Kind() == gast.KindFencedCodeBlock {
			fb := n.(*gast.FencedCodeBlock)
			x, y := fb.Info.Segment.Start, fb.Info.Segment.Stop
			fmt.Println(t.source[x:y])
			fmt.Println(string(reader.Value(fb.Info.Segment)))
		}
		/*
		if n.PreviousSibling() == nil || n.PreviousSibling().Kind() != DocGoKind {
			return gast.WalkContinue, nil
		}
		d := n.PreviousSibling().(*DocGoNode)
		_ = d.StringVars */
		return gast.WalkContinue, nil
	})
}
