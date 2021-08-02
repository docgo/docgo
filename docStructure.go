package main

type ModuleDoc struct {
	AbsolutePath string
	ImportPath string
	Packages []Package
}

type Snippet interface {
	Snip() string
}

type Function struct {}
type Method struct {}
type Typedef struct{}

type CodeDefinition struct {
	Functions []Function
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
}
