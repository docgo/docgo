package markdownAnnotate

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
	"regexp"
	"errors"
)

var prefix = []byte("[docgo:")
const unicodeIdentifier = `[\p{L}][\p{L}_0-9]*`

type docGoAnnotationParser struct{}

func NewDocgoParser() parser.InlineParser {
	return &docGoAnnotationParser{}
}

func (s *docGoAnnotationParser) Trigger() []byte {
	return []byte{'['}
}

func (s *docGoAnnotationParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	// Parses:
	// [prefix id=val id2=val2 ...]
	// into: {intVars: (id, val), boolVars: (id2, val), ...}

	_, seg := block.Position()
	line, _ := block.PeekLine()
	stringVars := make(map[string]string)
	boolVars, intVars := make(map[string]bool), make(map[string]int)

	if !bytes.HasPrefix(line, prefix) {
		return nil
	}

	failedDecode := ""

	decoder := bytes.NewReader(line)
	oldLen := decoder.Len()
	decoderAdvance := func (i int) {
		for x := 0; x < i; x++ {
			_, err := decoder.ReadByte()
			if err != nil {
				break
			}
		}
		line = line[i:]
	}
	decoderAdvance(len(prefix))

	r := regexp.MustCompile(`\s*(` + unicodeIdentifier + `)\s*=\s*`)
	for {
		matches := r.FindSubmatch(line)
		if len(matches) != 2 {
			break
		}
		varName := string(matches[1])
		decoderAdvance(len(matches[0]))

		jDecoder := json.NewDecoder(bytes.NewReader(line))
		jDecoder.UseNumber()
		t, err := jDecoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				failedDecode = "reached EOF without parsing a value."
				break
			}
			failedDecode = "malformed data: " + err.Error()
		}
		switch parsedVal := t.(type) {
		case string:
			stringVars[varName] = parsedVal
		case json.Number:
			if strings.Contains(parsedVal.String(), ".") { continue }
			pInt, _ := parsedVal.Int64()
			intVars[varName] = int(pInt)
		case bool:
			boolVars[varName] = parsedVal
		default:
			failedDecode = "not a string"
		}
		decoderAdvance(int(jDecoder.InputOffset()))
	}
	block.Advance(oldLen - decoder.Len())
	block.Advance(1)

	out := &DocGoNode{}
	out.StringVars = stringVars
	out.BoolVars = boolVars
	out.IntVars = intVars
	out.DecodeStatus = failedDecode
	out.LineStart = seg.Start
	out.LineEnd = seg.Stop
	return out
}

type DocGoNode struct{
	gast.BaseInline
	Code         string
	StringVars   map[string]string
	IntVars      map[string]int
	BoolVars     map[string]bool
	LineStart    int
	LineEnd      int
	DecodeStatus string
}

// Dump implements Node.Dump.
func (n *DocGoNode) Dump(source []byte, level int) {
	m := map[string]string{
		"AnnotateData": fmt.Sprintln(n.StringVars, n.IntVars),
	}
	gast.DumpHelper(n, source, level, m, nil)
}

func (n *DocGoNode) String() string {
	return fmt.Sprint("DocGoNode data =", n.StringVars, n.IntVars, n.BoolVars)
}

// KindTaskCheckBox is a NodeKind of the TaskCheckBox node.
var DocGoKind = gast.NewNodeKind("gc")

// Kind implements Node.Kind.
func (n *DocGoNode) Kind() gast.NodeKind {
	return DocGoKind
}

func (s *docGoAnnotationParser) CloseBlock(parent gast.Node, pc parser.Context) {
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
	if n.DecodeStatus != "" {
		//fmt.Println("Note: Malformed [docgo:] block")
	}
	return gast.WalkContinue, nil
}

type gcExtender struct {
}

// TaskList is an extension that allow you to use GFM task lists.
var DocgoExtension = &gcExtender{}

func (e *gcExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewDocgoParser(), 0),
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