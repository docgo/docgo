{{ range $idx, $p := . }}
### {{ .Name }}

{{ TransformDoc .Doc }}

```go
{{ .Snippet }}
```

{{ if .Methods }}

***Methods for {{.Name}}***

{{ range $idx, $p := .Methods }}

{{ TransformDoc .Doc }}

```go
{{ .Snippet }}
```

{{ end }}

----

{{ end }}

{{ end }}