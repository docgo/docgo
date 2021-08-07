package main

import (
	"context"
	"go/ast"
	"go/doc"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/vfs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var _modDoc *ModuleDoc = nil

func runIndexer(ctx context.Context, corpus *godoc.Corpus) {
	corpus.IndexFullText = false
	corpus.IndexThrottle = 0.5
	go func() {
		corpus.RunIndexer()
	}()
	for {
		<-time.NewTimer(time.Millisecond * 500).C
		corpus.UpdateIndex()
		if i, _ := corpus.CurrentIndex(); i != nil {
			break
		}
		if ctx.Err() != nil {
			break
		}
	}
}

func ModuleParse(modFilePath string) (parsedModuleDoc *ModuleDoc) {
	parsedModuleDoc = new(ModuleDoc)
	parsedModuleDoc.Packages = []*PackageDoc{}
	parsedModuleDoc.SimpleExports = SimpleExportsByType{}

	fmt.Debug("modFilePath", modFilePath)
	c := godoc.NewCorpus(vfs.OS(modFilePath))
	c.IndexFullText = false
	c.IndexThrottle = 0.5
	err := c.Init()
	if err != nil {
		fmt.Red(err)
		os.Exit(1)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	runIndexer(ctx, c)

	idx, _ := c.CurrentIndex()

	goModBuffer, err := os.ReadFile(filepath.Join(modFilePath, "go.mod"))

	modImportPath := modfile.ModulePath(goModBuffer)

	parsedModuleDoc.AbsolutePath = modFilePath
	parsedModuleDoc.ImportPath = modImportPath

	for pkgName, exportMap := range idx.Exports() {
		_ = pkgName
		for symbolName, val := range exportMap {
			_, _ = symbolName, val //fmt.Debug(symbolName, val)
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
		name := strings.ReplaceAll(filepath.Join(strings.TrimLeft(pkgPath, "/")), string(os.PathSeparator), "/")
		if hasMain {
			continue
			name += "/main"
		}
		if name == "" {
			entries, _ := os.ReadDir(modFilePath)
			for _, item := range entries {
				if !item.IsDir() && filepath.Ext(item.Name()) == ".go" {
					full := filepath.Join(modFilePath, item.Name())
					content, _ := os.ReadFile(full)
					matches := regexp.MustCompile(`(?s)package.*?([\p{L}_][\p{L}_0-9]*)`).FindStringSubmatch(string(content))
					if len(matches) > 1 {
						name = matches[1]
					}
				}
			}
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
		info := godocPresentation.GetPkgPageInfo(path, pkgName, godoc.PageInfoMode(0))
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

		parsedPackage.Structs = make([]*StructDef, 0)
		parsedPackage.Interfaces = make([]*InterfaceDef, 0)
		parsedPackage.Functions = make([]*FunctionDef, 0)
		parsedPackage.Variables = make([]*VarDef, 0)
		parsedPackage.Constants = make([]*ConstDef, 0)

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
				_ = varName
			}
			_ = varVal
		}

		for _, constVal := range info.PDoc.Consts {
			constDef := ConstDef{}
			constDef.Name = strings.Join(constVal.Names, ", ")
			constDef.Doc = constVal.Doc
			parsedPackage.Constants = append(parsedPackage.Constants, &constDef)
			constDef.Snippet = CreateSnippet(constVal.Decl, parsedPackage, "")
		}

		for _, constVal := range info.PDoc.Vars {
			constDef := VarDef{}
			constDef.Name = strings.Join(constVal.Names, ", ")
			constDef.Doc = constVal.Doc
			parsedPackage.Variables = append(parsedPackage.Variables, &constDef)
			constDef.Snippet = CreateSnippet(constVal.Decl, parsedPackage, "")
		}

		//fmt.Println(info.CallGraphIndex)
		for file, decls := range parsedPackage.FileDecls {
			fmt.Debug(file)
			_ = decls
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
