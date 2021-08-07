site_settings {
  github = "https://github.com/docgo/docgo"
  gopkg = "https://pkg.go.dev/"
  site_name = "docgo"
}


page {
  title = "Intro page"
  markdown = readfile("static/intro.md")
  fulltext = "This is the intro page containing awesome markdown"
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
      typeSection("Structs", it.value.CodeDef.Structs),
      typeSection("Variables", it.value.CodeDef.Variables),
      typeSection("Interfaces", it.value.CodeDef.Interfaces),
    ])
    fulltext = join(" ", [for item in getSections(it.value.CodeDef) : getSectionText(item)])
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
  ### ${sectionTitle}
  ${ join("\n", [for item in obj : snippet(item.BaseDef) ])}
  MULTILINE
}

function "snippet" {
  params = [def]
  result = <<-MULTILINE
  *${def.Name}*

  ${def.Doc}

  ```go
  ${def.Snippet}
  ```
  MULTILINE
}
