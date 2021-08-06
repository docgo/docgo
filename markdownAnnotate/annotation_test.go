package markdownAnnotate

import (
	"testing"
	"github.com/yuin/goldmark"
	"bytes"
	"strings"
	"github.com/yuin/goldmark/text"
)

func TestLeak(t *testing.T) {
	var sourceCode = []byte(`
# Heading 1
This is a text block.
* and a list
[docgo: a="string1"]
`)
	htmlBuffer := bytes.NewBufferString("")
	goldmark.New(goldmark.WithExtensions(DocgoExtension)).Convert(sourceCode, htmlBuffer)
	htmlStr := htmlBuffer.String()
	if strings.Contains(htmlStr, "docgo") {
		t.Error("Annotations shouldn't leak to HTML source code.")
	}
}

func TestAST(t *testing.T) {
	var sourceCode = []byte(`# H1
[docgo: key1="str1" key2="str2"]
`)
	ast := goldmark.New(goldmark.WithExtensions(DocgoExtension)).Parser().Parse(text.NewReader(sourceCode))
	firstChild := ast.FirstChild()
	if firstChild.Kind().String() != "Heading" {
		t.Error("First child not properly parsed")
	}
	para := firstChild.NextSibling()
	expectedAnnotate := para.FirstChild()
	if expectedAnnotate.Kind() != DocGoKind {
		t.Error("Failed to parse annotation")
	}
	ann := expectedAnnotate.(*DocGoNode)
	if y := len(ann.StringVars); y != 2 {
		t.Error("Incorrect annotation data length: ", y)
	}
	if len(ann.BoolVars) > 0 || len(ann.IntVars) > 0 {
		t.Error("Found incorrect type of annotations")
	}
	val1, val2 := ann.StringVars["key1"], ann.StringVars["key2"]
	if val1 != "str1" {
		t.Error("Invalid key1 value", val1)
	}
	if val2 != "str2" {
		t.Error("Invalid key2 value", val2)
	}
}

