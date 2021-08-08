package main

import (
	"context"
	"errors"
	"github.com/alecthomas/kong"
	"github.com/pkg/browser"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var Cli struct {
	Out        string `default:"docgo-dist/" type:"path" short:"o" help:"Where to put HTML assets"`
	ConfigFile string `default:"DOCGO.hcl" name:"conf" type:"path" short:"c" help:"A config file for extending docs"`
	ModulePath string `arg:"" type:"path" name:"path" help:"Path to module/package for documentation generation"`
	ServerPort int    `default:"8080" name:"port" short:"p" help:"Port for launching a server, set 0 for no server"`
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
		if !EmptyOutDirectory(struct{Out string}{Cli.Out}) {
			os.Exit(1)
		}
	}
	configAbsolute, _ := filepath.Abs(Cli.ConfigFile)
	configBytes := ReadStaticFile("static/DOCGO.hcl")
	os.WriteFile(configAbsolute, configBytes, fs.ModePerm)

	fmt.Yellow("Generating docs into:\n'" + Cli.Out + "' [as HTML assets]")

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

var ModulePath string

func main() {
	cliParse()
	ParsePage(_modDoc)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		ParsePage(_modDoc)
		http.FileServer(http.Dir(Cli.Out)).ServeHTTP(writer, request)
	})

	if Cli.ServerPort != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		go func() {
			<-ctx.Done()
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

func EmptyOutDirectory(conf struct{Out string}) bool{
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
	return isFine
}
