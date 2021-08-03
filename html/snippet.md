{{ range $idx, $p := . }}
### {{ .Name }}
{{ .Doc }}
```go
{{ .Snippet }}
```
{{ end }}