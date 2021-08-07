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
	"reflect"
	"html/template"
	"github.com/docgo/docgo/markdownAnnotate"
)

type Page struct {
	Title string `hcl:"title"`
	Markdown string `hcl:"markdown"`
}
type Document struct {
	Pages []Page `hcl:"page,block"`
	//Templates []Template `hcl:"template,block"`
}
/*
type Template struct {
	Name string `hcl:"name,label"`
	Markdown ExpressionClosureType `hcl:"markdown" cty:""`
}*/

type SerialType struct {
	BaseDef `cty:"BaseDef"`
}
func ex(val interface{}) []*SerialType {
	out := []*SerialType{}
	for i := 0; i < reflect.ValueOf(val).Len(); i++ {
		s := new(SerialType)
		v := reflect.ValueOf(val).Index(i).Elem().FieldByName("BaseDef")
		v.FieldByName("Methods").Set(reflect.ValueOf([]*MethodDef(nil)))
		reflect.ValueOf(s).Elem().FieldByName("BaseDef").Set(v)
		out = append(out, s)
	}
	return out
}
type Pkg struct {
Name         string `cty:"Name"`
Doc          string `cty:"Doc"`
Functions    []*FunctionDef `cty:"Functions"`
	Structs    []*SerialType `cty:"Structs"`
	Interfaces []*SerialType `cty:"Interfaces"`
	Constants  []*SerialType `cty:"Constants"`
	Variables []*SerialType `cty:"Variables"`
}
func Render(doc *ModuleDoc) cty.Value{
	pkgs := []cty.Value{}
	pp := []Pkg{}
	for _, item := range doc.Packages {
		p := Pkg{item.Name, item.Doc, item.Functions, ex(item.Structs), ex(item.Interfaces), ex(item.Constants), ex(item.Variables), }
		pp = append(pp, p)
		obj := map[string]cty.Value{}
		s := cty.StringVal
		obj["Name"] = s(item.Name)
		obj["Doc"] = s(item.Doc)
		fns := []cty.Value{}
		for _, item := range item.Functions {
			o := cty.ObjectVal(map[string]cty.Value{"Name": s(item.Name), "Doc": s(item.Doc),})
			fns = append(fns, o)
		}
		obj["Functions"] = cty.ListVal(fns)
		pkgs = append(pkgs, cty.ObjectVal(obj))
	}
	t, _ := gocty.ImpliedType(pp)
	v, _ := gocty.ToCtyValue(&pp, t)
	//fmt.Debug(cty.CapsuleVal(t, &pp))
	return v
	return cty.ListVal(pkgs)
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
	htmlTemplates := LoadHTMLTemplates(template.FuncMap{"GetPageTitle": func(i int) string{
		return document.Pages[i].Title
	}})
	baseHtmlTemplate := htmlTemplates.Lookup("base.html")
	links := map[int]string{}
	siteInfo := map[string]string{
		"github": "https://github.com/docgo/docgo",
		"gopkg": "https://pkg.go.dev/github.com/docgo/docgo",
		"projectName": "docgo",
	}
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
			SiteInfo  map[string]string
			SearchIndex template.JS
		}{
			item.Title, template.HTML(templateHTML),  links, i, doc, siteInfo, template.JS("3"),
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
			for name, val := range args[0].AsValueMap() {
				ctx2.Variables[name] = val
			}
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