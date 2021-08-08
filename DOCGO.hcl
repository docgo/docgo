site_settings {
  github = "https://github.com/docgo/docgo"
  gopkg = "https://pkg.go.dev/"
  site_name = "docgo"
}


page {
  title = "Intro page"
  markdown = readfile("static/intro.md")
  fulltext = cleanmarkdown(readfile("static/intro.md"))
  table_contents = ["What's docgo"]
}

dynamic "page" {
  for_each = Packages
  iterator = it
  content {
    title = it.value.Name
    markdown = join("\n", [
      it.value.Doc,
      "\n",
      typeSection("Constants", it.value.CodeDef.Constants),
      typeSection("Functions", it.value.CodeDef.Functions),
      structSection("Structs", it.value.CodeDef.Structs),
      typeSection("Variables", it.value.CodeDef.Variables),
      typeSection("Interfaces", it.value.CodeDef.Interfaces),
    ])
    fulltext = join(" ", [ for section in getSections(it.value.CodeDef) : cleanmarkdown(typeSection("", section)) ] )
    table_contents = flatten([for section in getSections(it.value.CodeDef) : [for item in section : item.BaseDef.Name]])
  }
}

page {
  title = "HCL Syntax"
  markdown = readfile("static/conf_syntax.md")
  fulltext = cleanmarkdown(readfile("static/conf_syntax.md"))
  table_contents = ["HCL syntax"]
}

function "getSections" {
  params = [cdef]
  result = [cdef.Constants, cdef.Functions, cdef.Structs, cdef.Variables, cdef.Interfaces]
}

function "typeSection" {
  params = [sectionTitle, obj]
  result = length(obj) == 0 ? "" : <<-MULTILINE
  ## ${sectionTitle}
  ----
  ${ join("\n", [for item in obj : snippet(item.BaseDef) ])}
  MULTILINE
}

function "structSection" {
  params = [title, structs]
  result = <<-MULTILINE
  ## Structs
  ----
  %{ for item in structs }
  ${snippet(item.BaseDef)}
  %{ if length(item.MethodList) != 0 }
  ### ⤷ *Methods on ${item.BaseDef.Name}*
  ${join("\n", [for method in item.MethodList : snippet(method.BaseDef, "❯ ") ])}
  %{ endif }
  %{ endfor }
  MULTILINE
}

function "snippet" {
  params = [def]
  variadic_param = extra
  result = <<-MULTILINE
  ### ${join(" ", extra)}${def.Name}

  ${def.Doc}

  ```go
  ${def.Snippet}
  ```
  MULTILINE
}