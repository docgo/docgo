package main

import (
	"encoding/json"
	"strings"
	"os"
	"github.com/markbates/pkger"
	"io"
)
var SEARCH_PAGE_HTML string

func GenerateSearch(index map[string]string) {
	searchFile, _ := pkger.Open("/html/search.html")
	pageBytes, _ := io.ReadAll(searchFile)
	SEARCH_PAGE_HTML = string(pageBytes)

	searchJson, err := json.Marshal(index)
	if err != nil {
		myfmt.Red(err)
		os.Exit(1)
	}
	outHtml := strings.ReplaceAll(SEARCH_PAGE_HTML, "{{ CONTENT }}", string(searchJson))
	_, err = CreateDist("godoc_search.html").WriteString(outHtml)
}
