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

var prefix = []byte("@docgo")

const unicodeIdentifier = `[\p{L}][\p{L}_0-9]*`

type docGoAnnotationParser struct{}

func NewDocgoParser() parser.InlineParser {
	return &docGoAnnotationParser{}
}

func (s *docGoAnnotationParser) Trigger() []byte {
	return []byte{'@'}
}

func (s *docGoAnnotationParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	// Parses:
	// [prefix id=val id2=val2 ...]
	// into: {intVars: (id, val), boolVars: (id2, val), ...}

	_, seg := block.Position()
	line, _ := block.PeekLine()
	stringVars := make(map[string]string)
	boolVars, intVars := make(map[string]bool), make(map[string]int)
	attributes := make(map[string]bool)
	lineStart := seg.Start
	lineEnd := seg.Stop

	if !bytes.HasPrefix(line, prefix) {
		return nil
	}

	failedDecode := ""

	decoder := bytes.NewReader(line)
	decoderAdvance := func(i int) {
		for x := 0; x < i; x++ {
			_, err := decoder.ReadByte()
			if err != nil {
				break
			}
		}
		line = line[i:]
		block.Advance(i)
	}
	decoderAdvance(len(prefix))

	rIdentifier := regexp.MustCompile(`^\s*(` + unicodeIdentifier + `)`)
	rComma := regexp.MustCompile(`^\s*,`)
	rBegin := regexp.MustCompile(`^\s*\[`)
	rEquals := regexp.MustCompile(`^\s*=`)
	parsedCounter := 0
	for {
		if parsedCounter == 0 {
			start := rBegin.Find(line)
			if start == nil {
				failedDecode = "no [ after annotation"
				break
			}
			decoderAdvance(len(start))
		}
		if parsedCounter > 0 {
			comma := rComma.Find(line)
			if comma == nil {
				break
			}
			decoderAdvance(len(comma))
			if len(strings.TrimSpace(string(line))) == 0 {
				block.AdvanceLine()
				line, _ = block.PeekLine()
				decoder = bytes.NewReader(line)
				_, seg := block.Position()
				lineEnd = seg.Stop
			}
		}
		matches := rIdentifier.FindSubmatch(line)
		if len(matches) != 2 {
			break
		}
		varName := string(matches[1])
		decoderAdvance(len(matches[0]))

		equals := rEquals.Find(line)
		if equals == nil {
			attributes[varName] = true
			parsedCounter++
			continue
		}
		decoderAdvance(len(equals))

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
			if strings.Contains(parsedVal.String(), ".") {
				continue
			}
			pInt, _ := parsedVal.Int64()
			intVars[varName] = int(pInt)
		case bool:
			boolVars[varName] = parsedVal
		default:
			failedDecode = "not a string"
		}
		decoderAdvance(int(jDecoder.InputOffset()))
		parsedCounter++
	}
	end := regexp.MustCompile(`^\s*\]`).Find(line)
	if end != nil {
		decoderAdvance(len(end))
	}
	out := &DocGoNode{}
	out.Attrs = attributes
	out.StringVars = stringVars
	out.BoolVars = boolVars
	out.IntVars = intVars
	out.DecodeStatus = failedDecode
	out.LineStart = lineStart
	out.LineEnd = lineEnd
	return out
}

// DocGoNode represents an AST node in the Markdown hierarchy.
// It contains annotations (that have attributes and begin/end marks) used later for rendering purposes.
type DocGoNode struct {
	gast.BaseInline
	Code         string
	Attrs        map[string]bool
	StringVars   map[string]string
	IntVars      map[string]int
	BoolVars     map[string]bool
	LineStart    int
	LineEnd      int
	DecodeStatus string
}

// Dump makes the DocGoNode conform to AST interface
func (n *DocGoNode) Dump(source []byte, level int) {
	m := map[string]string{
		"AnnotateData": fmt.Sprintln(n.StringVars, n.IntVars),
	}
	gast.DumpHelper(n, source, level, m, nil)
}

func (n *DocGoNode) String() string {
	return fmt.Sprint("DocGoNode data =", n.StringVars, n.IntVars, n.BoolVars)
}

// DocGoKind is a `NodeKind` that needs to be registered to the goldmark parser, so that
// it can be properly marked and each AST walker can know the type of this node.
var DocGoKind = gast.NewNodeKind("gc")

// Kind makes sure DocGoNode reports its correct ID/Kind to other AST processors
func (n *DocGoNode) Kind() gast.NodeKind {
	return DocGoKind
}

func (s *docGoAnnotationParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// `DocgoRenderer` is an inline renderer that fetches the AST nodes containing docgo annotation
// and converts them to the appropriate HTML.
type DocgoRenderer struct {
	html.Config
}

// NewDocgoRenderer returns a `NodeRenderer` instance capable of rendering Markdown
// for code documentation pages, containing annotations.
func NewDocgoRenderer(opts ...html.Option) renderer.NodeRenderer {
	return &DocgoRenderer{}
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *DocgoRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(DocGoKind, r.renderGC)
}

type gcWriter interface {
	io.Writer
}

func (r *DocgoRenderer) renderGC(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	if !entering {
		return gast.WalkContinue, nil
	}
	n := node.(*DocGoNode)
	fmt.Println(n)
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
		util.Prioritized(NewDocgoRenderer(), 500),
	))
}

// CleanPage takes a Markdown source as an input and deletes AST nodes that wouldn't be useful for searching/metadata
// purposes.
func CleanPage(page string) string {
	cleanAST := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader([]byte(page)))
	pageBytes := []byte(page)
	out := []byte("")
	gast.Walk(cleanAST, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			if n.Kind() == gast.KindCodeBlock || n.Kind() == gast.KindFencedCodeBlock || n.Kind() == gast.KindCodeSpan {
				//n.RemoveChildren(n)
				//n.PreviousSibling().SetNextSibling(n.NextSibling())
			}
			if n.Kind() == gast.KindText {
				out = append(out, n.Text(pageBytes)...)
				out = append(out, ' ')
			}
		}
		return gast.WalkContinue, nil
	})
	return string(out)
	cleanBuf := bytes.NewBufferString("")
	goldmark.New(goldmark.WithExtensions(extension.GFM)).Renderer().Render(cleanBuf, []byte(page), cleanAST)
	return cleanBuf.String()
}
