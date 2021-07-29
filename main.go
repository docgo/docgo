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
	"github.com/markbates/pkger"
	"io/ioutil"
	"os/exec"
	"runtime"
	"io/fs"
	"github.com/alecthomas/kong"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark"
)

var Cli struct {
	ConfFile string `arg help:"Path to conf.md"`
	Open bool
}

func cli() {
	kong.Parse(&Cli)
	mdData, err := os.ReadFile(Cli.ConfFile)
	if err != nil {
		fmt.Println("Error reading config: ", err)
		return
	}
	mdReader := text.NewReader(mdData)
	md := goldmark.DefaultParser().Parse(mdReader)
	MdPackages = parseJson(func(segment text.Segment) string {
		return string(mdData[segment.Start : segment.Stop])
	}, md)
}

var MdPackages map[string]map[string]MarkdownFile

func main() {
	cli()

	if inf, err := os.Stat("src"); err != nil || !inf.IsDir() {
		err = os.Mkdir("src", os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}

	sf := fmt.Sprintf

	summary, _ := os.Create("src/SUMMARY.md")
	summary.WriteString("# Summary\n\n")


	for pkgName, mdFiles := range MdPackages {
		summary.WriteString(sf("- [%s](%s.md)", pkgName, pkgName))
		for mdFile, _ := range mdFiles {
			summary.WriteString(sf("- [%s](%s.md)", mdFile, mdFile))
		}
	}

	cnt := make(map[string]int)
	write := func(file *os.File, body string, a ...interface{}) {
		S := fmt.Sprintf(body, a...)
		file.WriteString(S)
		cnt[file.Name()] += len(strings.TrimSpace(S))
	}

	entries, _ := os.ReadDir(".")

	m := token.NewFileSet()
	files := make([]*ast.File, 0)
	for _, f := range entries {
		inf, _ := f.Info()
		m.AddFile("./" + f.Name(), m.Base(), int(inf.Size()))
	}
	myPkg, _ := parser.ParseDir(m, ".", nil, parser.ParseComments)
	fmt.Println(myPkg)
	for _, f := range myPkg["cfgo"].Files {
		files = append(files, f)
	}

	pkg, _ := doc.NewFromFiles(m, files, "github.com/fikisipi/cloudflare-workers-go/cfgo", doc.AllMethods)
	f, _ := os.Create("src/cfgo.md")
	write(f, pkg.Doc)

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
		mdFile, _ := os.Create("src/" + docFile.baseName + ".md")
		mdFile.WriteString(fmt.Sprintf("# %s \n", docFile.title))
		mdFile.WriteString(docFile.description + "\n")

		for _, vr := range pkg.Vars {
			if notHere(vr.Decl.Pos()) { continue }
			varDecl := (snippet(vr.Decl.Pos(), vr.Decl.End()))
			write(mdFile, "%s\n```go\n%s\n```\n", vr.Doc, varDecl)
		}

		for _, function := range pkg.Funcs {
			if notHere(function.Decl.Pos()) { continue; }
			write(mdFile, "### %s\n", function.Name)
			write(mdFile, "```go\n%s\n```\n", snippet(function.Decl.Pos(), function.Decl.End()))
			write(mdFile, "%s\n", function.Doc)
		}
		mdFile.WriteString("\n")
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
					write(mdFile, "### struct " + declName + "\n\n```go\ntype %s struct {\n", declName)
					for _, field := range st.Fields.List {
						write(mdFile, "  %s\n", snippet(field.Pos(), field.End()))
					}
					write(mdFile, "}\n```\n")
				} else {
					it, ok := t.Type.(*ast.InterfaceType)
					if !ok { continue }
					write(mdFile, "### interface %s\n```go\ntype %s interface {\n", declName, declName)
					for _, meth := range it.Methods.List {
						snip := (snippet(meth.Pos(), meth.End()))
						write(mdFile, "  %s\n", snip)
					}
					write(mdFile, "}\n```\n")
				}
				write(mdFile, "%s\n", newF.Doc)
			}
			for _, fff := range newF.Funcs {
				write(mdFile, "### %s\n", fff.Name)
				write(mdFile, "```go\n%s\n```\n", snippet(fff.Decl.Pos(), fff.Decl.End()))
				write(mdFile, fff.Doc + "\n")
			}
			for _, m := range newF.Methods {
				decl := m.Decl
				snip := (snippet(decl.Pos(), decl.End()))
				write(mdFile, fmt.Sprintf("```go\n%s\n```\n%s\n", snip, m.Doc))
				for _, e := range m.Examples {
					write(mdFile, "Example:\n```go\n%s\n```\n", snippet(e.Code.Pos(), e.Code.End()))
				}
			}
		}

		if cnt[mdFile.Name()] > 1 {
			write(summary, "   - [%s](%s.md)\n", docFile.title, docFile.baseName)
		}
		mdFile.Close()
	}
	win, _ := pkger.Open("/pkger/mdbook.exe")
	linux, _ := pkger.Open("/pkger/mdbook.bin")
	plat := win
	fmt.Println("Detected GOOS =", runtime.GOOS, "\n")
	if runtime.GOOS == "windows" {
		plat = win
	} else {
		plat = linux
	}
	platFile := filepath.Base(plat.Name())
	fileBytes, _ := ioutil.ReadAll(plat)
	ioutil.WriteFile(platFile, fileBytes, 0777)
	//extra := make([]string, 0)
	cmd := []string{"build", "-d", "src"}
	if Cli.Open {
		cmd = append(cmd, "-o")
	}
	err := exec.Command("./" + platFile, cmd...).Run()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Wrote docs to \"src/\".")
		fmt.Println("Directory contents:")
		filepath.WalkDir("./src", func(path string, d fs.DirEntry, err error) error {
			if strings.HasSuffix(path, "html") {
				fmt.Print(path, " ")
			}
			return nil
		})
		fmt.Println("")
	}
	os.Remove(platFile)
}
