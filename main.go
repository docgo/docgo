package main

// This comment is under.

import (
	"go/ast"
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"github.com/alecthomas/kong"
	"net/http"
	"golang.org/x/tools/godoc"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/godoc/vfs"
	"time"
	"go/token"
	"io/fs"
)

var Cli struct {
	Out    string `default:"dist" short:"o" help"Where to put documentation/assets."`
	Module string `arg help:"Path to module/package for documentation generation."`
	Open   bool
}

func cliParse() {
	kong.Parse(&Cli)
	cliOutputAbs, err := filepath.Abs(Cli.Out)
	if err != nil {
		myfmt.Red("Couldn't parse directory for output", err)
		os.Exit(1)
	}
	Cli.Out = cliOutputAbs
	if cliStat, err := os.Stat(Cli.Out); err == nil {
		if !cliStat.IsDir() {
			myfmt.Red("Output is not a directory, but a file.")
			os.Exit(1)
		}
		isFine := true
		filepath.WalkDir(Cli.Out, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() && filepath.Ext(path) != ".html" {
				myfmt.Red("Out path not empty (contains non-assets):", Cli.Out)
				os.Exit(1)
				isFine = false
				return filepath.SkipDir
			}
			return nil
		})
		if !isFine {
			return
		}
	}

	myfmt.Yellow("Using \"" + Cli.Out + "\" as an output directory...")

	absModPath, err := filepath.Abs(Cli.Module)
	mInfo, err := os.Stat(absModPath)
	if err != nil {
		myfmt.Red("Error loading '", mInfo, "': ", err)
		os.Exit(1)
	}
	mDirPath := absModPath
	if !mInfo.IsDir() {
		mDirPath = filepath.Dir(Cli.Module)
	}

	_modDoc = ModuleParse(mDirPath)
}

var _modDoc *ModuleDoc = nil

func ModuleParse(modFilePath string) (parsedModuleDoc *ModuleDoc) {
	parsedModuleDoc = new(ModuleDoc)
	parsedModuleDoc.Packages = []*PackageDoc{}
	parsedModuleDoc.SimpleExports = SimpleExportsByType{}

	myfmt.Debug("modFilePath", modFilePath)
	c := godoc.NewCorpus(vfs.OS(modFilePath))

	err := c.Init()
	if err != nil {
		fmt.Println(err)
	}
	go func() {
		c.RunIndexer()
	}()
	<-time.NewTicker(time.Millisecond * 200).C

	idx, _ := c.CurrentIndex()

	goModBuffer, err := os.ReadFile(filepath.Join(modFilePath, "go.mod"))
	modImportPath := modfile.ModulePath(goModBuffer)

	parsedModuleDoc.AbsolutePath = modFilePath
	parsedModuleDoc.ImportPath = modImportPath

	pkgList := map[string]string{}
	for kind, symbols := range idx.Idents() {
		if kind.Name() == "Packages" {
			for _, sym := range symbols {
				pkgList[sym[0].Path] = sym[0].Name
			}
		} else {
			for name, symTable := range symbols {
				for _, symbol := range symTable {
					scopedId := ScopedIdentifier{
						PackagePath: symbol.Path,
						Name:        name,
						IsFunction:  kind == godoc.FuncDecl,
						IsMethod:    kind == godoc.MethodDecl,
						isType:      kind == godoc.TypeDecl,
					}
					parsedModuleDoc.SimpleExports[name] = append(parsedModuleDoc.SimpleExports[name], scopedId)
				}
			}
		}
	}
	parsedModuleDoc.DebugPrint()
	myfmt.Green("Loaded packages:", pkgList)

	godocPresentation := godoc.NewPresentation(c)
	for path, pkgName := range pkgList {
		parsedPackage := new(PackageDoc)
		info := godocPresentation.GetPkgPageInfo(path, pkgName, godoc.NoFiltering)
		if info == nil {
			continue
		}

		parsedPackage.ParentModule = parsedModuleDoc
		parsedPackage.AbsolutePath = filepath.Join(modFilePath, strings.TrimPrefix(path, "/"))
		parsedPackage.FileSet = info.FSet
		parsedPackage.RelativePath = path
		parsedPackage.Name = pkgName
		parsedPackage.Doc = info.PDoc.Doc

		parsedModuleDoc.Packages = append(parsedModuleDoc.Packages, parsedPackage)

		info.FSet.Iterate(func(file *token.File) bool {
			if file == nil {
				return false
			}
			baseName := filepath.Base(file.Name())
			q, _ := os.ReadFile(filepath.Join(parsedPackage.AbsolutePath, baseName))
			_ = q
			//fmt.Println(string(q))
			return true
		})

		for _, tp := range info.PDoc.Types {
			for _, spec := range tp.Decl.Specs {
				ParseTypeDecl(spec, parsedPackage)
			}
		}

		for _, fn := range info.PDoc.Funcs {
			parsedFn := FunctionDef{}

			parsedFn.FoundInFile = GetDeclFile(fn.Decl, parsedPackage)
			parsedFn.Snippet = CreateSnippet(fn.Decl, parsedPackage)
			parsedFn.Name = fn.Name
			parsedFn.Doc = fn.Doc
			parsedPackage.Functions = append(parsedPackage.Functions, parsedFn)
		}

		for _, varVal := range info.PDoc.Vars {
			_ = varVal
		}

		for _, constVal := range info.PDoc.Consts {
			_ = constVal
		}

		//fmt.Println(info.CallGraphIndex)
	}
	return
}

func ParseTypeDecl(s ast.Spec, docPackage *PackageDoc) {
	t := s.(*ast.TypeSpec)
	declName := t.Name.Name
	st, ok := t.Type.(*ast.StructType)
	if ok {
		sDef := StructDef{}
		sDef.Snippet = CreateSnippet(st, docPackage, "type ", declName, " ")
		sDef.FoundInFile = GetDeclFile(st, docPackage)
		sDef.Name = declName
		sDef.Type = st

		for _, field := range st.Fields.List {
			_ = field
		}
		docPackage.Structs = append(docPackage.Structs, sDef)
	} else {
		it, ok := t.Type.(*ast.InterfaceType)
		if !ok {
			return
		}
		interDef := InterfaceDef{}
		interDef.FoundInFile = GetDeclFile(it, docPackage)
		interDef.Name = declName
		interDef.Type = it
		interDef.Snippet = CreateSnippet(it, docPackage, "type ", declName, " ")
		docPackage.Interfaces = append(docPackage.Interfaces, interDef)

		for _, meth := range it.Methods.List {
			_ = meth
		}
	}
}

var ModulePath string

func main() {
	cliParse()

	GenerateHTML(_modDoc)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		GenerateHTML(_modDoc)

		http.FileServer(http.Dir(Cli.Out)).ServeHTTP(writer, request)
	})
	http.ListenAndServe(":8080", mux)
}
