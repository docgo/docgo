package main

import (
	"log"

	oldJson "encoding/json"
	oldFmt "fmt"
	"github.com/docgo/docgo/markdownAnnotate"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	_ "github.com/zclconf/go-cty/cty/gocty"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sort"
	"regexp"
)

type Page struct {
	Title           string   `hcl:"title"`
	Markdown        string   `hcl:"markdown"`
	FullText        string   `hcl:"fulltext"`
	TableOfContents []string `hcl:"table_contents"`
}
type Document struct {
	Pages        []Page       `hcl:"page,block"`
	SiteSettings SiteSettings `hcl:"site_settings,block"`
}
type SiteSettings struct {
	GitHub   string `hcl:"github,attr"`
	GoPkg    string `hcl:"gopkg,attr"`
	SiteName string `hcl:"site_name,attr"`
}

func (p Page) Slug() string {
	return regexp.MustCompile(`[^\p{L}0-9]`).ReplaceAllString(p.Title, "")
}

func ctyValModuleDoc(doc *ModuleDoc) cty.Value {
	t, _ := gocty.ImpliedType(doc.Packages)
	v, _ := gocty.ToCtyValue(&doc.Packages, t)
	return v
}

type Sortable []Page

func (s Sortable) Less(i, j int) bool {
	return strings.Count(s[i].Title, "/") < strings.Count(s[j].Title, "/")
}

func (s Sortable) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Sortable) Len() int {
	return len(s)
}

func ParsePage(doc *ModuleDoc) {
	var document Document
	ctx := hcl.EvalContext{}

	ctx.Functions = hclBaseFunctions()
	ctx.Functions["cleanmarkdown"] = hclCleanMarkdownText()
	ctx.Variables = map[string]cty.Value{"Packages": ctyValModuleDoc(doc)}
	err := decodeHclIntoTarget(Cli.ConfigFile, &ctx, &document)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	settings := document.SiteSettings
	htmlTemplates := LoadHTMLTemplates(cookFuncmap("GetPageTitle", func(i int) string {
		return document.Pages[i].Title
	}))
	baseHtmlTemplate := htmlTemplates.Lookup("base.html")
	links := map[int]string{}
	searchIndex := map[string]string{}

	sort.Sort(Sortable(document.Pages))
	for i := 0; i < len(document.Pages); i++ {
		slug := document.Pages[i].Slug()
		if i == 0 {
			slug = "index"
		}
		links[i] = oldFmt.Sprintf("%s.html", slug)
		searchIndex[slug] = document.Pages[i].FullText
	}
	searchIndexBytes, _ := oldJson.Marshal(searchIndex)
	for i, item := range document.Pages {
		distFile := CreateDist(links[i])
		templateHTML := markdownAnnotate.RenderPage(item.Markdown)
		thisPage := struct {
			Title         string
			Body          template.HTML
			PageLinks     map[int]string
			CurrentPage   int
			ModuleDoc     *ModuleDoc
			SiteInfo      SiteSettings
			SearchIndex   template.JS
			TableContents []string
		}{
			item.Title, template.HTML(templateHTML), links, i, doc, settings, template.JS(string(searchIndexBytes)), item.TableOfContents,
		}
		baseHtmlTemplate.Execute(distFile, thisPage)
	}
}

func decodeHclIntoTarget(filename string, ctx *hcl.EvalContext, target interface{}) error {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Configuration file not found",
					Detail:   oldFmt.Sprintf("The configuration file %s does not exist.", filename),
				},
			}
		}
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read configuration",
				Detail:   oldFmt.Sprintf("Can't read %s: %s.", filename, err),
			},
		}
	}

	return decodeHclFile(filename, src, ctx, target)
}

func decodeHclFile(filename string, src []byte, ctx *hcl.EvalContext, target interface{}) error {
	var file *hcl.File
	var diags hcl.Diagnostics

	switch suffix := strings.ToLower(filepath.Ext(filename)); suffix {
	case ".hcl":
		file, diags = hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	case ".json":
		file, diags = json.Parse(src, filename)
	default:
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unsupported file format",
			Detail:   oldFmt.Sprintf("Cannot read from %s: unrecognized file format suffix %q.", filename, suffix),
		})
		return diags
	}
	if diags.HasErrors() {
		return diags
	}
	userFunctions, parsedBody, _ := userfunc.DecodeUserFunctions(file.Body, "function", func() *hcl.EvalContext {
		return ctx
	})
	for name, val := range userFunctions {
		ctx.Functions[name] = val
	}
	file.Body = parsedBody
	file.Body = dynblock.Expand(file.Body, ctx)
	diags = gohcl.DecodeBody(file.Body, ctx, target)
	if diags.HasErrors() {
		return diags
	}
	return nil
}
