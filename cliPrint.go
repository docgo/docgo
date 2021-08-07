package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"log"
	"os"
	"reflect"
)

type myCliFormatterFn func(args ...interface{})

type myCliFormatter struct {
	dbgLogger *log.Logger
	Red       myCliFormatterFn
	Green     myCliFormatterFn
	Yellow    myCliFormatterFn
}

var fmt = myCliFormatter{nil, myCliColor(color.FgRed), myCliColor(color.FgGreen), myCliColor(color.FgHiYellow)}

// Pretty-prints a data structure for debugging purposes
// At the moment, ran only inside JetBrains
func (m myCliFormatter) Debug(args ...interface{}) {
	if os.Getenv("TERMINAL_EMULATOR") != "JetBrains-JediTerm" || os.Getenv("APPDEBUG") != "1" {
		return
	}
	spew.Config.DisableCapacities = true
	for i, arg := range args {
		if arg != nil && reflect.TypeOf(arg) != nil && reflect.TypeOf(arg).Kind() == reflect.String {
			os.Stdout.WriteString(" " + arg.(string) + " ")
		} else {
			spew.Dump(args[i])
		}
	}
}

func myCliColor(attribute color.Attribute) myCliFormatterFn {
	c := color.New(attribute)
	return func(args ...interface{}) {
		c.Println(args...)
	}
}
