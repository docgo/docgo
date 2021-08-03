{{ range $idx, $p := . }}
### {{ .Name }}
asdf
{{ TransformDoc .Doc }}

```go
{{ .Snippet }}
```
{{ end }}