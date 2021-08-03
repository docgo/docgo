{{ range $idx, $p := . }}
### {{ .Name }}
{{ TransformDoc .Doc }}
```go
{{ .Snippet }}
```
{{ end }}