# Intro

Welcome to the documentation.

{{ GitHubRepo "ggodoc/ggodoc" }}

{{ range .Packages }}

# {{ .Name }}

{{ .Doc }}

{{ template "snippet.md" .Functions }}
{{ template "snippet.md" .Structs }}
{{ template "snippet.md" .Interfaces }}
{{ template "snippet.md" .Variables }}
{{ template "snippet.md" .Constants }}

{{ end }}