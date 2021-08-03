# Intro

Welcome to the documentation.

{{ range .Packages }}

# {{ .Name }}
## Documentation

{{ .Doc }}

{{ template "snippet" .Functions }}
{{ template "snippet" .Structs }}
{{ template "snippet" .Interfaces }}

{{ end }}