page {
  title = "A"
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
    markdown = <<-MD
${it.value.Doc}
${snippet({type="Functions", obj=it.value.Functions }) }
${snippet({type="Structs", obj=it.value.Structs }) }
    MD
  }
}
template "snippet" {
  markdown = <<-MD
### ${type}
${ join("\n", [for item in obj : "${item.BaseDef.Name}\n```go\n${item.BaseDef.Snippet}\n```" ])}
  MD
}