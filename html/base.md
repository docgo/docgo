# Intro

Welcome to the documentation.

{{ range .Packages }}

# {{ .Name }}

{{ .Doc }}

{{ template "snippet" .Functions }}
{{ template "snippet" .Structs }}
{{ template "snippet" .Interfaces }}

{{ end }}

{{ $x := PackageConfig "myPkg" }}
{{ $x.Group "a.go" "b.go" "Name" }}