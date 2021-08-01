package main

import (
	"strings"
	"os"
	"github.com/markbates/pkger"
	"io"
	"fmt"
	"path/filepath"
	"errors"
)

func GenerateHTML(html string) (path string) {
	raw, err := pkger.Open("/html/index.html")
	if err != nil {
		fmt.Println("pkger error: ", err)
		os.Exit(1)
	}
	data, _ := io.ReadAll(raw)
	htmlRaw := string(data)

	parts := strings.Split(htmlRaw, "<!--content-block-->")
	newRaw := parts[0] + html + parts[2]
	ferr := os.Mkdir("out", 0755)
	if ferr != nil {
		if !errors.Is(ferr, os.ErrExist) {
			fmt.Println(ferr)
			os.Exit(1)
		}
	}
	f, _ := os.Create("out/index.html")
	f.WriteString(newRaw)
	outAbs, _ := filepath.Abs("./out/index.html")
	return outAbs
}
