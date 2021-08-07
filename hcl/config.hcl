projectTitle = getName()
page {
  title = "title1"
  content = "My content"
}
page {
  title = "title2"
  content = <<-PageContent
  some content
  ${bestPage}
  PageContent
  dynamic "definition" {
    for_each = ["dyn1", "dyn2"]
    content {
      name = definition.value
    }
  }
}

age=absolute(-4)
someList = [1, 2, 3]
someMap = {x = 3, y = 2}
someConf = [for s in someList : absolute(s)]