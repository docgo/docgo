package main

import (
	"go/ast"
	"os"
	"strings"
	"path/filepath"
	"github.com/alecthomas/kong"
	"net/http"
	"golang.org/x/tools/godoc"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/godoc/vfs"
	"time"
	"io/fs"
	"github.com/pkg/browser"
	"context"
	"errors"
	"go/doc"
)

var Cli struct {
	Out        string `default:"dist/" short:"o" help:"Where to put documentation/assets."`
	ModulePath string `arg help:"Path to module/package for documentation generation."`
	ServerPort int    `default:8080 short:"p" help:"Port for hot-reload server. 0 to disable server."`
}

func cliParse() {
	kong.Parse(&Cli)
	cliOutputAbs, err := filepath.Abs(Cli.Out)
	if err != nil {
		fmt.Red("Couldn't parse directory for output", err)
		os.Exit(1)
	}
	Cli.Out = cliOutputAbs
	if cliStat, err := os.Stat(Cli.Out); err == nil {
		if !cliStat.IsDir() {
			fmt.Red("Output is not a directory, but a file.")
			os.Exit(1)
		}
		isFine := true
		filepath.WalkDir(Cli.Out, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() && filepath.Ext(path) != ".html" {
				fmt.Red("Out path not empty (contains non-assets):", Cli.Out)
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

	fmt.Yellow("Using \"" + Cli.Out + "\" as an output directory...")

	absModPath, err := filepath.Abs(Cli.ModulePath)
	mInfo, err := os.Stat(absModPath)
	if err != nil {
		fmt.Red("Error loading '", mInfo, "': ", err)
		os.Exit(1)
	}
	mDirPath := absModPath
	if !mInfo.IsDir() {
		mDirPath = filepath.Dir(Cli.ModulePath)
	}

	_modDoc = ModuleParse(mDirPath)
}

var _modDoc *ModuleDoc = nil

func ModuleParse(modFilePath string) (parsedModuleDoc *ModuleDoc) {
	parsedModuleDoc = new(ModuleDoc)
	parsedModuleDoc.Packages = []*PackageDoc{}
	parsedModuleDoc.SimpleExports = SimpleExportsByType{}

	fmt.Debug("modFilePath", modFilePath)
	c := godoc.NewCorpus(vfs.OS(modFilePath))

	err := c.Init()
	if err != nil {
		fmt.Red(err)
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

	for pkgName, exportMap := range idx.Exports() {
		_ = pkgName
		for symbolName, val := range exportMap {
			fmt.Debug(symbolName, val)
		}
	}

	/*
	for i := 0; idx.Snippet(i) != nil; i++{
		fmt.Debug(idx.Snippet(i))
	}*/

	pkgList := map[string]string{}
	for pkgPath, symMap := range idx.Exports() {
		hasMain := false
		for sym, _ := range symMap {
			if sym == "main" {
				hasMain = true
				break
			}
		}

		name := filepath.Base(pkgPath)
		if hasMain {
			name = "main"
		}
		pkgList[pkgPath] = name
	}

	for kind, symbols := range idx.Idents() {
		if kind.Name() == "Packages" {
			for _, sym := range symbols {
				// pkgList[sym[0].Path] = sym[0].Name
				_ = sym
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
	fmt.Green("Loaded packages:", pkgList)

	godocPresentation := godoc.NewPresentation(c)
	for path, pkgName := range pkgList {
		parsedPackage := new(PackageDoc)
		info := godocPresentation.GetPkgPageInfo(path, pkgName, godoc.NoFiltering)
		if info == nil {
			continue
		}

		parsedPackage.FileDecls = make(map[string][]BaseDef)
		parsedPackage.ParentModule = parsedModuleDoc
		parsedPackage.AbsolutePath = filepath.Join(modFilePath, strings.TrimPrefix(path, "/"))
		parsedPackage.FileSet = info.FSet
		parsedPackage.RelativePath = path
		parsedPackage.Name = pkgName
		parsedPackage.Doc = info.PDoc.Doc

		parsedModuleDoc.Packages = append(parsedModuleDoc.Packages, parsedPackage)

		for _, tp := range info.PDoc.Types {
			for _, spec := range tp.Decl.Specs {
				ParseTypeDecl(spec, parsedPackage, tp.Methods)
			}
		}

		for _, fn := range info.PDoc.Funcs {
			parsedFn := FunctionDef{}

			parsedFn.Snippet = CreateSnippet(fn.Decl, parsedPackage)
			parsedFn.Name = fn.Name
			parsedFn.Doc = fn.Doc
			parsedPackage.Functions = append(parsedPackage.Functions, &parsedFn)
			parsedFn.FoundInFile = GetDeclFile(fn.Decl, parsedFn.BaseDef, parsedPackage)
		}

		for _, varVal := range info.PDoc.Vars {
			for _, varName := range varVal.Names {
				fmt.Debug(varName)
			}
			fmt.Debug("specs", varVal.Decl.Specs)
			_ = varVal
		}

		for _, constVal := range info.PDoc.Consts {
			for _, constName := range constVal.Names {
				fmt.Debug(constName)
			}
			constDef := ConstDef{}
			constDef.Name = strings.Join(constVal.Names, ", ")
			constDef.Doc = constVal.Doc
			parsedPackage.Constants = append(parsedPackage.Constants, &constDef)
			constDef.Snippet = CreateSnippet(constVal.Decl, parsedPackage, "")
		}

		for _, constVal := range info.PDoc.Vars {
			for _, constName := range constVal.Names {
				fmt.Debug(constName)
			}
			constDef := VarDef{}
			constDef.Name = strings.Join(constVal.Names, ", ")
			constDef.Doc = constVal.Doc
			parsedPackage.Variables = append(parsedPackage.Variables, &constDef)
			constDef.Snippet = CreateSnippet(constVal.Decl, parsedPackage, "")
		}

		//fmt.Println(info.CallGraphIndex)
		for file, decls := range parsedPackage.FileDecls {
			fmt.Debug(file, decls)
		}
	}

	return
}

func ParseTypeDecl(s ast.Spec, docPackage *PackageDoc, methods []*doc.Func) {
	methodDefs := make([]*MethodDef, 0)
	for _, method := range methods {
		methodDef := MethodDef{}
		methodDef.Name = method.Name
		methodDef.Doc = method.Doc
		methodDef.Snippet = CreateSnippet(method.Decl, docPackage)
		methodDefs = append(methodDefs, &methodDef)
	}

	t := s.(*ast.TypeSpec)
	declName := t.Name.Name
	st, ok := t.Type.(*ast.StructType)
	if ok {
		sDef := StructDef{}
		sDef.Snippet = CreateSnippet(st, docPackage, "type ", declName, " ")
		sDef.Name = declName
		sDef.Type = st
		sDef.FoundInFile = GetDeclFile(st, sDef.BaseDef, docPackage)
		sDef.Methods = methodDefs

		for _, field := range st.Fields.List {
			_ = field
		}
		docPackage.Structs = append(docPackage.Structs, &sDef)
	} else {
		it, ok := t.Type.(*ast.InterfaceType)
		if !ok {
			return
		}
		interDef := InterfaceDef{}
		interDef.FoundInFile = GetDeclFile(it, interDef.BaseDef, docPackage)
		interDef.Name = declName
		interDef.Type = it
		interDef.Snippet = CreateSnippet(it, docPackage, "type ", declName, " ")
		docPackage.Interfaces = append(docPackage.Interfaces, &interDef)

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
		http.FileServer(http.Dir(Cli.Out)).ServeHTTP(writer, request)
	})

	if Cli.ServerPort != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 200 * time.Millisecond)
		go func() {
			<- ctx.Done()
			if !errors.Is(ctx.Err(), context.Canceled) {
				fmt.Green("Listening on :8080")
				browser.OpenURL("http://localhost:8080")
			}
		}()
		err := http.ListenAndServe(":8080", mux)
		if err != nil {
			cancel()
			fmt.Red("Cannot listen on :8080\n", err)
			os.Exit(1)
		}
	}
}
