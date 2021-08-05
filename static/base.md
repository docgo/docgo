[docgo: title = "intro" ]

Welcome to the documentation.

{{ SetSiteInfo
 "github" "https://github.com/docgo/docgo"
 "gopkg" "https://pkg.go.dev/github.com/docgo/docgo"
 "projectName" "docgo"
}}

{{ range .Packages }}
[docgo: title = "{{ .Name }}" ]
{{ .Doc }}

{{ template "snippet.md" .Functions }}
{{ template "snippet.md" .Structs }}
{{ template "snippet.md" .Interfaces }}
{{ template "snippet.md" .Variables }}
{{ template "snippet.md" .Constants }}

{{ end }}