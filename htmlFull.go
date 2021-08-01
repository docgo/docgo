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
)

type Meta struct {
	Packages map[string]*ast.Package
	PackageNames []string
	Content template.HTML
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
