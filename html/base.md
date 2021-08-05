# Intro

Welcome to the documentation.

{{ GitHubRepo "ggodoc/ggodoc" }}

{{ range .Packages }}

# {{ .Name }}

{{ .Doc }}

{{ template "snippet" .Functions }}
{{ template "snippet" .Structs }}
{{ template "snippet" .Interfaces }}

{{ end }}