site_settings {
  github = "https://github.com/docgo/docgo"
  gopkg = "https://pkg.go.dev/"
  site_name = "docgo"
}


page {
  title = "Intro page"
  markdown = readfile("static/intro.md")
  fulltext = "This is the intro page containing awesome markdown"
  table_contents = ["docgo", "HCL syntax"]
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
    fulltext = join(" ", [for item in getSections(it.value.CodeDef) : getSectionText(item)])
    table_contents = flatten([for section in getSections(it.value.CodeDef) : [for item in section : item.BaseDef.Name]])
  }
}

function "getSections" {
  params = [cdef]
  result = [cdef.Constants, cdef.Functions, cdef.Structs, cdef.Variables, cdef.Interfaces]
}
function "getSectionText" {
  params = [section]
  result = join(" ", [for item in section : "${item.BaseDef.Name} ${item.BaseDef.Doc}"])
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
  ${ join("\n", [for item in structs : "${snippet(item.BaseDef)}${renderMethods(item.BaseDef.Name, item.MethodList)}" if length(item.MethodList) != 0 ])}
  ${ join("\n", [for item in structs : snippet(item.BaseDef) if length(item.MethodList) == 0 ])}
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

function "renderMethods" {
  params = [name, methods]
  result = <<EOT
  ### ⤷ *Methods on ${name}*
  ${join("\n", [for method in methods : snippet(method.BaseDef, "❯ ") ])}
  EOT
}