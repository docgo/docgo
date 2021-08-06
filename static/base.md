@docgo [ title = "intro", 
 page = true ]

Welcome to the documentation. It's written
in ***Markdown*** (i.e. \*markdown\*). The page
was created with a md annotation:
```
@docgo[ page = true, title = "intro" ]
```

{{ SetSiteInfo
 "github" "https://github.com/docgo/docgo"
 "gopkg" "https://pkg.go.dev/github.com/docgo/docgo"
 "projectName" "docgo"
}}

{{ range .Packages }}
@docgo[ title = "{{ .Name }}", page = true ]
{{ .Doc }}
{{ template "snippet.md" .Functions }}
{{ template "snippet.md" .Structs }}
{{ template "snippet.md" .Interfaces }}
{{ template "snippet.md" .Variables }}
{{ template "snippet.md" .Constants }}

{{ end }}