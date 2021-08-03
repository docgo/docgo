package main

import (
	"os"
	"github.com/markbates/pkger"
	"io"
	"fmt"
	"path/filepath"
	"errors"
	"html/template"
	"bytes"
	"go/ast"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Meta struct {
	Packages map[string]*ast.Package
	PackageNames []string
	Content template.HTML
}
func CreateDist() *os.File{
	ferr := os.Mkdir("out", 0755)
	if ferr != nil {
		if !errors.Is(ferr, os.ErrExist) {
			fmt.Println("creating dist folder error", ferr)
			os.Exit(1)
		}
	}
	f, _ := os.Create("out/index.html")
	return f
}
func ReadTempl(name string, funcMap template.FuncMap) *template.Template{
	t := template.New("main")
	raw, err := pkger.Open(name)
	if err != nil {
		fmt.Println("pkger error: ", err)
		os.Exit(1)
	}
	if funcMap != nil {
		t.Funcs(funcMap)
	}
	data, _ := io.ReadAll(raw)
	templ, err := t.Parse(string(data))
	if err != nil {
		fmt.Println("template loading error", err)
		os.Exit(1)
	}
	return templ
}
func GenerateHTML2(doc *ModuleDoc) string {
	distFile := CreateDist()
	buf := bytes.Buffer{}
	md := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(html.WithHardWraps()))

	errT := ReadTempl("/html/DOCS.tmpl", template.FuncMap{
		"A": func(x int) int{ return 5},
	}).Execute(&buf, doc)
	buf2 := bytes.Buffer{}
	md.Convert(buf.Bytes(), &buf2)

	markdownHTML := buf2.String()
	ReadTempl("/html/base.html", nil).Execute(distFile, markdownHTML)

	if errT != nil {
		fmt.Println(errT)
	}
	if y, err := filepath.Abs("./out/index.html"); err == nil {
		return y
	}
	return ""
}

func GenerateHTML(html string, metadata Meta) (path string) {
	t := template.New("main")
	raw, err := pkger.Open("/html/index.html")
	if err != nil {
		fmt.Println("pkger error: ", err)
		os.Exit(1)
	}
	data, _ := io.ReadAll(raw)
	htmlRaw := string(data)
	t.Funcs(template.FuncMap{
		"PackageFiles": func(p *ast.Package) []string {
			out := make([]string, 0)
			for f, _ := range p.Files {
				out = append(out, filepath.Base(f))
			}
			return out
		},
	})
	t, _ = t.Parse(htmlRaw)
	byteBuffer := bytes.Buffer{}

	//t.Lookup("q").Execute(&byteBuffer, metadata)
	//fmt.Println(byteBuffer.String())
	byteBuffer = bytes.Buffer{}
	metadata.Content = template.HTML(html)
	err = t.Execute(&byteBuffer, metadata)
	if err != nil {
		fmt.Println("template err", err)
		os.Exit(1)
	}

	ferr := os.Mkdir("out", 0755)
	if ferr != nil {
		if !errors.Is(ferr, os.ErrExist) {
			fmt.Println(ferr)
			os.Exit(1)
		}
	}
	f, _ := os.Create("out/index.html")
	f.Write(byteBuffer.Bytes())
	outAbs, _ := filepath.Abs("./out/index.html")
	return outAbs
}
