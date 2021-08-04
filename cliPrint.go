package main

import (
	"github.com/fatih/color"
	"log"
	"os"
	"github.com/markbates/pkger"
)

// Needed for the `pkger` tool to autoload the required files.
func extraPkgerOpens() {
	pkger.Open("/html/base.md")
	pkger.Open("/html/base.html")
	pkger.Open("/html/snippet.md")
}

type myCliFormatterFn func(args ...interface{})

type myCliFormatter struct {
	dbgLogger *log.Logger
	Red       myCliFormatterFn
	Green     myCliFormatterFn
	Yellow    myCliFormatterFn
}

var fmt = myCliFormatter{nil, myCliColor(color.FgRed), myCliColor(color.FgGreen), myCliColor(color.FgHiYellow)}

func (m myCliFormatter) Debug(args ...interface{}) {
	if os.Getenv("NODEBUG") != "" {
		return
	}
	if m.dbgLogger == nil {
		m.dbgLogger = log.New(os.Stdout, "", (0))
	}
	m.dbgLogger.Println(args...)
}

func myCliColor(attribute color.Attribute) myCliFormatterFn {
	c := color.New(attribute)
	return func(args ...interface{}) {
		c.Println(args...)
	}
}
