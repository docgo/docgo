package main

import (
	"github.com/fatih/color"
	"log"
	"os"
	"github.com/markbates/pkger"
)

func _extraOpens() {
	pkger.Open("/html/base.md")
	pkger.Open("/html/base.html")
	pkger.Open("/html/snippet.md")
}
type _mPrintlnType func(...interface{})

func _mWrapColor(c color.Attribute) _mPrintlnType {
	return func(x ...interface{}) { color.New(c).Println(x...) }
}

func _mDoDebug() _mPrintlnType {
	if os.Getenv("NODEBUG") != "" {
		return func(i ...interface{}) {
			// noop
		}
	}
	return log.New(os.Stdout, "DBG ", log.Flags()).Println
}

var myfmt = struct {
	Red   _mPrintlnType
	Green _mPrintlnType
	Yellow _mPrintlnType
	Debug _mPrintlnType
}{
	_mWrapColor(color.FgRed), _mWrapColor(color.FgGreen), _mWrapColor(color.FgHiYellow), _mDoDebug(),
}
