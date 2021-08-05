package customMarkdown

import (
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"bytes"
	"strings"
	"fmt"
	"github.com/yuin/goldmark/extension"
	"io"
	"encoding/json"
	"errors"
)

type gcParser struct {
}

var gcParserInstance = &gcParser{}

func NewGCParser() parser.InlineParser {
	return gcParserInstance
}

func (s *gcParser) Trigger() []byte {
	return []byte{'['}
}

func (s *gcParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	// Given AST structure must be like
	// - List
	//   - ListItem         : parent.Parent
	//     - TextBlock      : parent
	//       (current line)
	_, seg := block.Position()
	line, _ := block.PeekLine()
	sline := string(line)
	vars := make(map[string]string)
	if !strings.HasPrefix(sline, "[docgo:") {
		return nil
	}
	block.Advance(len("[docgo:"))
	sline = sline[len("[docgo:"):]

	dec := strings.NewReader(sline)
	lenBefore := dec.Len()
	for {
		name := ""
		finish := false
		for {
			b, err := dec.ReadByte()
			if b == ']' {
				finish = true
				break
			}
			if b == '=' {
				break
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					finish = true
					break
				}
				panic("FAILED" +  err.Error())
			}
			name += string(b)
		}
		if finish {
			break
		}
		jString := ""
		err := json.NewDecoder(dec).Decode(&jString)
		if err != nil {
			panic("FAILED" + err.Error())
		}
		vars[strings.TrimSpace(name)] = jString
	}
	block.Advance(lenBefore - dec.Len())

	out := &DocGoNode{}
	out.Vars = vars
	out.LineStart = seg.Start
	out.LineEnd = seg.Stop
	return out
}
type DocGoNode struct{
	gast.BaseInline
	Code string
	Vars map[string]string
	LineStart int
	LineEnd int
}

// Dump implements Node.Dump.
func (n *DocGoNode) Dump(source []byte, level int) {
	m := map[string]string{
		"Checked": fmt.Sprintf("%v", n.Code),
	}
	gast.DumpHelper(n, source, level, m, nil)
}

// KindTaskCheckBox is a NodeKind of the TaskCheckBox node.
var DocGoKind = gast.NewNodeKind("gc")

// Kind implements Node.Kind.
func (n *DocGoNode) Kind() gast.NodeKind {
	return DocGoKind
}

func (s *gcParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// TaskCheckBoxHTMLRenderer is a renderer.NodeRenderer implementation that
// renders checkboxes in list items.
type GcRenderer struct {
	html.Config
}

// NewTaskCheckBoxHTMLRenderer returns a new TaskCheckBoxHTMLRenderer.
func NewGCRenderer(opts ...html.Option) renderer.NodeRenderer {
	return &GcRenderer{}
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *GcRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(DocGoKind, r.renderGC)
}

type gcWriter interface {
	io.Writer
}

func (r *GcRenderer) renderGC(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}
	n := node.(*DocGoNode)
	_ = n
	return gast.WalkContinue, nil
}

type gcExtender struct {
}

// TaskList is an extension that allow you to use GFM task lists.
var DocgoExtension = &gcExtender{}

func (e *gcExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewGCParser(), 0),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewGCRenderer(), 500),
	))
}

func CleanPage(page string) string{
	cleanAST := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader([]byte(page)))
	gast.Walk(cleanAST, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			if n.Kind() == gast.KindCodeBlock || n.Kind() == gast.KindFencedCodeBlock || n.Kind() == gast.KindCodeSpan {
				//n.RemoveChildren(n)
				n.PreviousSibling().SetNextSibling(n.NextSibling())
			}
		}
		return gast.WalkContinue, nil
	})
	cleanBuf := bytes.NewBufferString("")
	goldmark.New(goldmark.WithExtensions(extension.GFM)).Renderer().Render(cleanBuf, []byte(page), cleanAST)
	return cleanBuf.String()
}