package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type ModuleDoc struct {
	AbsolutePath  string
	ImportPath    string
	Packages      []*PackageDoc
	SimpleExports SimpleExportsByType
}

type ScopedIdentifier struct {
	PackagePath string
	Name string
	IsFunction bool
	IsMethod bool
	isType bool
}

type ExportType string
type SimpleExportsByType map[string][]ScopedIdentifier

func (m ModuleDoc) DebugPrint() {
	debugLog("ModuleDoc{ ", "AbsolutePath =", m.AbsolutePath, " ImportPath =", m.ImportPath)
	for exportType, exports := range m.SimpleExports {
		fmt.Println(" - ", exportType, ":", exports)
	}
}

type Snippet struct {
	SnippetText string
}

func (s Snippet) String() string {
	return s.SnippetText
}

func CreateSnippet(node ast.Node, pkg *PackageDoc, prefix ...string) Snippet {
	snipFile := pkg.FileSet.File(node.Pos())
	baseName := filepath.Base(snipFile.Name())
	q, _ := os.ReadFile(filepath.Join(pkg.AbsolutePath, baseName))
	if len(q) == 0 {
		fmt.Println(pkg.AbsolutePath, baseName)
		os.Exit(1)
	}
	snipStr := string(q)[snipFile.Offset(node.Pos())  : snipFile.Offset(node.End()) ]
	return Snippet{strings.Join(prefix, "") + snipStr}
}

func GetDeclFile(node ast.Node, pkg *PackageDoc) string {
	return pkg.FileSet.File(node.Pos()).Name()
}

type BaseDef struct {
	Snippet
	FoundInFile string
}

type FunctionDef struct {
	BaseDef
	Name string
	Doc string
}

type StructDef struct {
	BaseDef
	Name string
	Doc             string
	Type *ast.StructType
}

type InterfaceDef struct {
	BaseDef
	Name string
	Doc         string
	Type *ast.InterfaceType
}

type Method struct {
	BaseDef
}
type Typedef struct{
	BaseDef
}

type CodeDef struct {
	Functions []FunctionDef
	Methods []Method
	Typedefs []Typedef
	Structs []StructDef
	Interfaces []InterfaceDef
}

type PackageDoc struct {
	CodeDef
	Name            string
	Doc             string
	AbsolutePath    string
	RelativePath    string
	FileSet         *token.FileSet
	ParentModule    *ModuleDoc
}

