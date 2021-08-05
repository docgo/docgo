# Intro

Welcome to the documentation.

{{ SetSiteInfo
 "github" "gGodoc/gGodoc"
 "projectName" "gGodoc" }}

{{ range .Packages }}

# {{ .Name }}

{{ .Doc }}

{{ template "snippet.md" .Functions }}
{{ template "snippet.md" .Structs }}
{{ template "snippet.md" .Interfaces }}
{{ template "snippet.md" .Variables }}
{{ template "snippet.md" .Constants }}

{{ end }}