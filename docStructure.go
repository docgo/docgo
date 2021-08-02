package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
)

type ModuleDoc struct {
	AbsolutePath  string
	ImportPath    string
	Packages      []PackageDoc
	SimpleExports SimpleExportsByType
}

type ExportType string
type SimpleExportsByType map[ExportType][]string

func (m ModuleDoc) Print() {
	fmt.Println("ModuleDoc{ ", "AbsolutePath =", m.AbsolutePath, " ImportPath =", m.ImportPath)
	for exportType, exports := range m.SimpleExports {
		fmt.Println(" - ", exportType, ":", exports)
	}
}

type Snippet struct {
	SnippetText string
}

func CreateSnippet(node ast.Node, pkg *PackageDoc) Snippet {
	snipFile := pkg.FileSet.File(node.Pos())
	baseName := filepath.Base(snipFile.Name())
	q, _ := os.ReadFile(filepath.Join(pkg.AbsolutePath, baseName))
	if len(q) == 0 {
		fmt.Println(pkg.AbsolutePath, baseName)
		os.Exit(1)
	}
	snipStr := string(q)[snipFile.Offset(node.Pos())  : snipFile.Offset(node.End()) ]
	return Snippet{snipStr}
}

type FunctionDef struct {
	Snippet
	Name string
	Doc string
}

type StructDef struct {
	Snippet
	Name string
	Type *ast.StructType
}

type InterfaceDef struct {
	Snippet
	Name string
	Type *ast.InterfaceType
}

type Method struct {
	Snippet
}
type Typedef struct{
	Snippet
}

type CodeDef struct {
	Functions []FunctionDef
	Methods []Method
	Typedefs []Typedef
}

type PackageFile struct {}

type PackageDoc struct {
	Name            string
	Doc             string
	AbsolutePath    string
	RelativePath    string
	CodeDefinitions CodeDef
	Files           []PackageFile
	FileSet         *token.FileSet
}
