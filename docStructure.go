package main

import (
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
	Name        string
	IsFunction  bool
	IsMethod    bool
	isType      bool
}

type ExportType string
type SimpleExportsByType map[string][]ScopedIdentifier

func CreateSnippet(node ast.Node, pkg *PackageDoc, prefix ...string) string {
	snipFile := pkg.FileSet.File(node.Pos())
	baseName := filepath.Base(snipFile.Name())
	q, _ := os.ReadFile(filepath.Join(pkg.AbsolutePath, baseName))
	if len(q) == 0 {
		fmt.Red(pkg.AbsolutePath, baseName)
		os.Exit(1)
	}
	snipStr := string(q)[snipFile.Offset(node.Pos()):snipFile.Offset(node.End())]
	return strings.Join(prefix, "") + snipStr
}

func GetDeclFile(node ast.Node, ourDecl BaseDef, pkg *PackageDoc) string {
	fileName := pkg.FileSet.File(node.Pos()).Name()
	if pkg.FileDecls[fileName] == nil {
		pkg.FileDecls[fileName] = make([]BaseDef, 0)
	}
	pkg.FileDecls[fileName] = append(pkg.FileDecls[fileName], ourDecl)
	return fileName
}

type BaseDef struct {
	Snippet   string`cty:"Snippet"`
	Name        string `cty:"Name"`
	FoundInFile string
	Doc string
	Methods    []*MethodDef
}

type ConstDef struct {
	BaseDef `cty:"BaseDef"`
}

type VarDef struct {
	BaseDef `cty:"BaseDef"`
}

type FunctionDef struct {
	BaseDef `cty:"BaseDef"`
}

type StructDef struct {
	BaseDef `cty:"BaseDef"`
	Type    *ast.StructType
}

type InterfaceDef struct {
	BaseDef `cty:"BaseDef"`
	Type    *ast.InterfaceType
}

type MethodDef struct {
	BaseDef `cty:"BaseDef"`
}
type Typedef struct {
	BaseDef `cty:"BaseDef"`
}

type PackageDoc struct {
	CodeDef
	Name         string
	Doc          string
	AbsolutePath string
	RelativePath string
	FileSet      *token.FileSet
	ParentModule *ModuleDoc
	FileDecls    map[string][]BaseDef
}

type CodeDef struct {
	Functions  []*FunctionDef
	Typedefs   []*Typedef
	Structs    []*StructDef
	Interfaces []*InterfaceDef
	Constants  []*ConstDef
	Variables []*VarDef
}

type PackageFileDoc struct {
	CodeDef
	Name string
}
