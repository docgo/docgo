@[ page, title = "intro" ]

# Hi
Welcome to the documentation. It's written
in ***Markdown*** (i.e. \*markdown\*). The page
was created with a md annotation:
```
@docgo[ page = true, title = "intro" ]
```
---
# @docgo[name]

{{ range .Packages }}
@docgo[ title = "{{ .Name }}", type = "page" ]
{{ .Doc }}
{{ template "snippet.md" .Functions }}
{{ template "snippet.md" .Structs }}
{{ template "snippet.md" .Interfaces }}
{{ template "snippet.md" .Variables }}
{{ template "snippet.md" .Constants }}

{{ end }}