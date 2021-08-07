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

${snippet("Functions", it.value.Functions) }
${snippet("Structs", it.value.Structs) }
${snippet("Interfaces", it.value.Interfaces) }
    EOF
  }
}
function "snippet" {
  params = [snippetType, obj]
  result = length(obj) == 0 ? "" : <<EOF

### ${snippetType}
${ join("\n", [for item in obj : baseDef(item.BaseDef) ])}

  EOF
}

function "baseDef" {
  params = [def]
  result = <<EOF
*${def.Name}*

${def.Doc}

```go
${def.Snippet}
```
  EOF
}
