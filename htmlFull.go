package main

import (
	"strings"
	"os"
	"github.com/markbates/pkger"
	"io"
	"fmt"
)

func GenerateHTML(html string) {
	raw, err := pkger.Open("goSuperDoc/html/index.html")
	if err != nil {
		fmt.Println("pkger error: ", err)
		os.Exit(1)
	}
	data, _ := io.ReadAll(raw)
	htmlRaw := string(data)

	newRaw := strings.Replace(htmlRaw, "{empty}", html, 1)
	f, _ := os.Create("out.html")
	f.WriteString(newRaw)
}
