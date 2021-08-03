{{ range $idx, $p := . }}
### {{ .Name }} [{{.FoundInFile}}]
{{ .Doc }}
```go
{{ .Snippet }}
```
{{ end }}