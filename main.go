//go:generate wget https://github.com/fikisipi/cloudflare-workers-go/releases/download/0.0.1/pkged.go -O pkged.go

package main

import (
	"go/doc"
	"go/token"
	"go/ast"
	"go/parser"
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"io/fs"
	"github.com/alecthomas/kong"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"io"
	"bytes"
	"net/http"
)

var Cli struct {
	ModulePath string `arg help:"Path to module"`
	Open bool
}

func cliParse() {
	kong.Parse(&Cli)
	absModPath, err := filepath.Abs(Cli.ModulePath)
	mInfo, err := os.Stat(absModPath)
	if err != nil {
		fmt.Println("Error loading '", mInfo, "': ", err)
		os.Exit(1)
	}
	mDirPath := absModPath
	if !mInfo.IsDir() {
		mDirPath = filepath.Dir(Cli.ModulePath)
	}

	ModulePath = mDirPath
	MdPackages = ParsePackages(mDirPath)
}

var ModulePath string
var MdPackages map[string]map[string]MarkdownFile

func Generate() (distPath string) {
	fmt.Println("ModulePath =", ModulePath)
	m := token.NewFileSet()
	files := make([]*ast.File, 0)
	paths := make(map[string]bool)
	filepath.WalkDir(ModulePath, func(path string, d fs.DirEntry, err error) error {
		if d.Name() == ".git" { return filepath.SkipDir }
		if d.IsDir() { paths[path] = true; return nil }
		if !strings.HasSuffix(d.Name(), ".go") { return nil }
		//fullpath := filepath.Join(path, d.Name())
		inf, _ := d.Info()
		fmt.Println("Adding file:", path)
		m.AddFile(path, m.Base(), int(inf.Size()))
		return nil
	})

	var myPkgs = make(map[string]*ast.Package)
	for path, _ := range paths {
		pkgMap, _ := parser.ParseDir(m, path, nil, parser.ParseComments)
		for n, pkg := range pkgMap {
			myPkgs[n] = pkg
		}
	}

	for name, pkg := range myPkgs {
		fmt.Println("Parsed pkg:", name)
		for _, f := range pkg.Files {
			files = append(files, f)
		}
	}

	buffer := ""
	write := func(body string, a ...interface{}) {
		S := fmt.Sprintf(body, a...)
		buffer += S
	}

	pkg, _ := doc.NewFromFiles(m, files, "github.com/fikisipi/cloudflare-workers-go/cfgo", doc.AllMethods)
	write(pkg.Doc)

	type DocFile struct {
		astFile *ast.File
		baseName string
		title string
		description string
	}
	docFiles := make([]*DocFile, 0)

	for _, f := range files {
		fullName := m.File(f.Pos()).Name()
		baseName := filepath.Base(fullName)

		mdDoc := ""
		mdFile, ok := MdPackages["cfgo"][baseName]
		if ok {
			mdDoc = mdFile.description
		}

		niceName := strings.TrimSuffix(baseName, ".go")
		if strings.TrimSpace(mdFile.niceName) != "" {
			niceName = mdFile.niceName
		}
		docFiles = append(docFiles, &DocFile{f, baseName, niceName, mdDoc})
	}

	for _, docFile := range docFiles {
		notHere := func(pos token.Pos) bool {
			return m.File(pos) != m.File(docFile.astFile.Pos())
		}
		snippet := func(pos token.Pos, pos2 token.Pos) string {
			srcB, _ := os.ReadFile( m.File(pos).Name())
			srcStr := string(srcB)
			return srcStr[m.Position(pos).Offset : m.Position(pos2).Offset]
		}
		if strings.Contains(docFile.baseName, "_test") {
			continue
		}
		write("# %s \n", docFile.title)
		write(docFile.description + "\n")

		for _, vr := range pkg.Vars {
			if notHere(vr.Decl.Pos()) { continue }
			varDecl := (snippet(vr.Decl.Pos(), vr.Decl.End()))
			write("%s\n```go\n%s\n```\n", vr.Doc, varDecl)
		}

		for _, function := range pkg.Funcs {
			if notHere(function.Decl.Pos()) { continue; }
			write("### %s\n", function.Name)
			write("```go\n%s\n```\n", snippet(function.Decl.Pos(), function.Decl.End()))
			write("%s\n", function.Doc)
		}
		//ch := types.NewChecker(&types.Config{Error: myErr, Importer: importer.Default(), IgnoreFuncBodies: true}, m, types.NewPackage("..", "cfgo"), &info)
		//err := ch.Files(files)
		//fmt.Println(err)

		for _, newF := range pkg.Types {
			if notHere(newF.Decl.Pos()) { continue; }
			for _, s := range newF.Decl.Specs {
				t := s.(*ast.TypeSpec)
				declName := t.Name.Name
				st, ok := t.Type.(*ast.StructType)
				if ok {
					write("### struct " + declName + "\n\n```go\ntype %s struct {\n", declName)
					for _, field := range st.Fields.List {
						write("  %s\n", snippet(field.Pos(), field.End()))
					}
					write("}\n```\n")
				} else {
					it, ok := t.Type.(*ast.InterfaceType)
					if !ok { continue }
					write("### interface %s\n```go\ntype %s interface {\n", declName, declName)
					for _, meth := range it.Methods.List {
						snip := (snippet(meth.Pos(), meth.End()))
						write("  %s\n", snip)
					}
					write("}\n```\n")
				}
				write("%s\n", newF.Doc)
			}
			for _, fff := range newF.Funcs {
				write("### %s\n", fff.Name)
				write("```go\n%s\n```\n", snippet(fff.Decl.Pos(), fff.Decl.End()))
				write(fff.Doc + "\n")
			}
			for _, m := range newF.Methods {
				decl := m.Decl
				snip := (snippet(decl.Pos(), decl.End()))
				write(fmt.Sprintf("```go\n%s\n```\n%s\n", snip, m.Doc))
				for _, e := range m.Examples {
					write("Example:\n```go\n%s\n```\n", snippet(e.Code.Pos(), e.Code.End()))
				}
			}
		}
	}

	w := bytes.Buffer{}
	goldmark.New(goldmark.WithRendererOptions(html.WithXHTML())).Convert([]byte(buffer), &w)

	bufOut, _ := io.ReadAll(&w)

	pkgNameList := make([]string, 0)
	for pkgName, _ := range myPkgs {
		pkgNameList = append(pkgNameList, pkgName)
	}
	metadata := Meta{
		Packages: myPkgs,
		PackageNames: pkgNameList,
	}
	distPath = GenerateHTML(string(bufOut), metadata)
	return
}

func main() {
	cliParse()

	//sf := fmt.Sprintf
	distPath := Generate()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		distPath = Generate()
		http.FileServer(http.Dir(filepath.Dir(distPath))).ServeHTTP(writer, request)
	})
	http.ListenAndServe(":8080", mux)
}
