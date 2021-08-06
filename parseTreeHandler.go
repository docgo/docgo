package main

import "go/doc"

func ParseFunc(fn *doc.Func) {
	for _, field := range fn.Decl.Type.Params.List {
		_ = field
	}
	_ = fn
}