package main

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	"strings"
	"path/filepath"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
	"io/ioutil"
	"os"
	"github.com/hashicorp/hcl/v2/ext/transform"
	"github.com/davecgh/go-spew/spew"
	oldFmt "fmt"
	_ "github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/gocty"
	"html/template"
	"github.com/docgo/docgo/markdownAnnotate"
	"github.com/hashicorp/hcl/v2/ext/userfunc"
)

type Page struct {
	Title string `hcl:"title"`
	Markdown string `hcl:"markdown"`
}
type Document struct {
	Pages []Page `hcl:"page,block"`
	SiteSettings SiteSettings `hcl:"site_settings,block"`
}
type SiteSettings struct {
	GitHub string `hcl:"github,attr"`
	GoPkg string `hcl:"gopkg,attr"`
	SiteName string `hcl:"site_name,attr"`
}
func Render(doc *ModuleDoc) cty.Value{
	t, _ := gocty.ImpliedType(doc.Packages)
	v, _ := gocty.ToCtyValue(&doc.Packages, t)
	return v
}

func ParsePage(doc *ModuleDoc) {
	var document Document
	ctx := hcl.EvalContext{}

	ctx.Functions = BaseFunctions()
	ctx.Variables = map[string]cty.Value{"Packages": Render(doc)}
	err := DecodeFile(Cli.ConfigFile, &ctx, &document)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	settings := document.SiteSettings
	htmlTemplates := LoadHTMLTemplates(template.FuncMap{"GetPageTitle": func(i int) string{
		return document.Pages[i].Title
	}})
	baseHtmlTemplate := htmlTemplates.Lookup("base.html")
	links := map[int]string{}

	for i := 0; i < len(document.Pages); i++ {
		if i == 0 {
			links[0] = "index.html"
		} else {
			links[i] = oldFmt.Sprintf("%d.html", i)
		}
	}
	for i, item := range document.Pages {
		distFile := CreateDist(links[i])
		templateHTML := markdownAnnotate.RenderPage(item.Markdown)
		thisPage := struct{
			Title       string
			Body        template.HTML
			PageLinks   map[int]string
			CurrentPage int
			ModuleDoc   *ModuleDoc
			SiteInfo  SiteSettings
			SearchIndex template.JS
		}{
			item.Title, template.HTML(templateHTML),  links, i, doc, settings, template.JS("3"),
		}
		baseHtmlTemplate.Execute(distFile, thisPage)
	}
}
func main_() {
	//ParsePage()
}

func BaseFunctions() map[string]function.Function{
	var fnn = map[string]function.Function{
		"absolute": stdlib.AbsoluteFunc, "add": stdlib.AddFunc, "and": stdlib.AndFunc, "byteslen": stdlib.BytesLenFunc, "bytesslice": stdlib.BytesSliceFunc, "csvdecode": stdlib.CSVDecodeFunc, "ceil": stdlib.CeilFunc, "chomp": stdlib.ChompFunc, "chunklist": stdlib.ChunklistFunc, "coalesce": stdlib.CoalesceFunc, "coalescelist": stdlib.CoalesceListFunc, "compact": stdlib.CompactFunc, "concat": stdlib.ConcatFunc, "contains": stdlib.ContainsFunc, "distinct": stdlib.DistinctFunc, "divide": stdlib.DivideFunc, "element": stdlib.ElementFunc, "equal": stdlib.EqualFunc, "flatten": stdlib.FlattenFunc, "floor": stdlib.FloorFunc, "formatdate": stdlib.FormatDateFunc, "format": stdlib.FormatFunc, "formatlist": stdlib.FormatListFunc, "greaterthan": stdlib.GreaterThanFunc, "greaterthanorequalto": stdlib.GreaterThanOrEqualToFunc, "hasindex": stdlib.HasIndexFunc, "indent": stdlib.IndentFunc, "index": stdlib.IndexFunc, "int": stdlib.IntFunc, "jsondecode": stdlib.JSONDecodeFunc, "jsonencode": stdlib.JSONEncodeFunc, "join": stdlib.JoinFunc, "keys": stdlib.KeysFunc, "length": stdlib.LengthFunc, "lessthan": stdlib.LessThanFunc, "lessthanorequalto": stdlib.LessThanOrEqualToFunc, "log": stdlib.LogFunc, "lookup": stdlib.LookupFunc, "lower": stdlib.LowerFunc, "max": stdlib.MaxFunc, "merge": stdlib.MergeFunc, "min": stdlib.MinFunc, "modulo": stdlib.ModuloFunc, "multiply": stdlib.MultiplyFunc, "negate": stdlib.NegateFunc, "notequal": stdlib.NotEqualFunc, "not": stdlib.NotFunc, "or": stdlib.OrFunc, "parseint": stdlib.ParseIntFunc, "pow": stdlib.PowFunc, "range": stdlib.RangeFunc, "regexall": stdlib.RegexAllFunc, "regex": stdlib.RegexFunc, "regexreplace": stdlib.RegexReplaceFunc, "replace": stdlib.ReplaceFunc, "reverse": stdlib.ReverseFunc, "reverselist": stdlib.ReverseListFunc, "sethaselement": stdlib.SetHasElementFunc, "setintersection": stdlib.SetIntersectionFunc, "setproduct": stdlib.SetProductFunc, "setsubtract": stdlib.SetSubtractFunc, "setsymmetricdifference": stdlib.SetSymmetricDifferenceFunc, "setunion": stdlib.SetUnionFunc, "signum": stdlib.SignumFunc, "slice": stdlib.SliceFunc, "sort": stdlib.SortFunc, "split": stdlib.SplitFunc, "strlen": stdlib.StrlenFunc, "substr": stdlib.SubstrFunc, "subtract": stdlib.SubtractFunc, "timeadd": stdlib.TimeAddFunc, "title": stdlib.TitleFunc, "trim": stdlib.TrimFunc, "trimprefix": stdlib.TrimPrefixFunc, "trimspace": stdlib.TrimSpaceFunc, "trimsuffix": stdlib.TrimSuffixFunc, "upper": stdlib.UpperFunc, "values": stdlib.ValuesFunc, "zipmap": stdlib.ZipmapFunc,
	}
	return fnn
}


func Decode(filename string, src []byte, ctx *hcl.EvalContext, target interface{}) error {
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
	x := transform.TransformerFunc(func(body hcl.Body) hcl.Body {
		sch := hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: "template",
					LabelNames: []string{"templateName",},
				},
			},
		}

		//sch.Attributes = []hcl.AttributeSchema{{"markdown", true}}
		b, remain, diags := body.PartialContent(&sch)
		_ = spew.Dump

		for _, item := range b.Blocks.OfType("template") {
			attr, _ := item.Body.JustAttributes()
			ctx.Functions[item.Labels[0]] = TemplateFn(attr["markdown"].Expr, ctx)
		}
		/*
		for _, item := range c.Blocks.OfType("template") {
			item.Body.
		}
		fmt.Println(c.MissingItemRange.String())
*/
		return transform.BodyWithDiagnostics(remain, diags)
	})

	fn, b, _ := userfunc.DecodeUserFunctions(file.Body, "function", func() *hcl.EvalContext {
		return ctx
	})
	for x, y := range fn {
		ctx.Functions[x] = y
	}
	file.Body = b
	file.Body = x(dynblock.Expand(file.Body, ctx))
	diags = gohcl.DecodeBody(file.Body, ctx, target)
	if diags.HasErrors() {
		return diags
	}
	return nil
}
func TemplateFn(e hcl.Expression, ctx *hcl.EvalContext) function.Function{
	x := function.Spec{
		Params:   []function.Parameter{{
			Name:             "data",
			Type:             cty.DynamicPseudoType,
		}},
		VarParam: nil,
		Type: func(args []cty.Value) (cty.Type, error) {
			return cty.String, nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ctx2 := ctx.NewChild()
			ctx2.Variables = map[string]cty.Value{}
			ctx2.Variables["args"] = args[0]
			v, _ := e.Value(ctx2)
			return v, nil
		},
	}
	return function.New(&x)
}


func DecodeFile(filename string, ctx *hcl.EvalContext, target interface{}) error {
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

	return Decode(filename, src, ctx, target)
}