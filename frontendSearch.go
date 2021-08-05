package main

import (
	"encoding/json"
	"os"
	"bytes"
	"html/template"
)

func GenerateSearch(searchTpl *template.Template, index map[string]string) {
	SEARCH_PAGE_HTML := bytes.NewBufferString("")

	/*searchJson, err := json.Marshal(index)
	if err != nil {
		fmt.Red(err)
		os.Exit(1)
	}*/
	_ = json.Marshal
	_ = os.Exit

	searchTpl.Execute(SEARCH_PAGE_HTML, struct {
		JsonContent map[string]string
	}{index})

	CreateDist("godoc_search.html").Write(SEARCH_PAGE_HTML.Bytes())
}
