site_settings {
  github = "https://github.com/docgo/docgo"
  gopkg = "https://pkg.go.dev/"
  site_name = "docgo"
}

page {
  title = "Intro page"
  markdown = <<EOF
  # This is the intro page
  Containing awesome **Markdown**
  EOF
}

dynamic "page" {
  for_each = Packages
  iterator = it
  content {
    title = it.value.Name
    markdown = <<EOF
${it.value.Doc}
${typeSection("Constants", it.value.CodeDef.Constants) }
${typeSection("Functions", it.value.CodeDef.Functions) }
${typeSection("Structs", it.value.CodeDef.Structs) }
${typeSection("Variables", it.value.CodeDef.Variables) }
${typeSection("Interfaces", it.value.CodeDef.Interfaces) }
    EOF
  }
}
function "typeSection" {
  params = [snippetType, obj]
  result = length(obj) == 0 ? "" : <<EOF

### ${snippetType}
${ join("\n", [for item in obj : snippet(item.BaseDef) ])}

  EOF
}

function "snippet" {
  params = [def]
  result = <<EOF
*${def.Name}*

${def.Doc}

```go
${def.Snippet}
```
  EOF
}
