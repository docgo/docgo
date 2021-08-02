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
	Packages      []Package
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

func CreateSnippet(node ast.Node, pkg *Package) Snippet {
	snipFile := pkg.FileSet.File(node.Pos())
	q, _ := os.ReadFile(filepath.Join(pkg.Path, snipFile.Name()))
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
}

type InterfaceDef struct {
	Snippet
	
}
typepe Method struct {
	Snippet
}
type Typedef struct{
	Snippet
}

type CodeDefinition struct {
	Functions []FunctionDef
	Methods []Method
	Typedefs []Typedef
}

type PackageFile struct {
	Pkg *Package
	Code CodeDefinition
}

type Package struct {
	Name string
	Path string
	ExportedCode CodeDefinition
	Files []PackageFile
	FileSet *token.FileSet
}
