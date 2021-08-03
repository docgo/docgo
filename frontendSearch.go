package main

import (
	"encoding/json"
	"strings"
	"os"
)

func GenerateSearch(index map[string]string) string {
	searchJson, err := json.Marshal(index)
	if err != nil {
		myfmt.Red(err)
		os.Exit(1)
	}
	outHtml := strings.ReplaceAll(PAGE, "{{ CONTENT }}", string(searchJson))
	_, err = CreateDist("godoc_search.html").WriteString(outHtml)
}

const PAGE = `
    <script src="https://cdn.jsdelivr.net/npm/elasticlunr@0.9.5/elasticlunr.min.js"></script>
	<script> window.godoc = {{ CONTENT }}; </script>
	<script> var index = elasticlunr(function () {
    this.addField('body');
    this.setRef('page');
	});
	[...window.godoc.entries()].map( x => index.addDoc({page: x[0], body: x[1]}) );
	const urlSearchParams = new URLSearchParams(window.location.search);
	const params = Object.fromEntries(urlSearchParams.entries());
	document.write(index.search(params[q]));
	</script>
`
