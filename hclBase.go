package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/transform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func hclBaseFunctions() map[string]function.Function {
	var fnn = map[string]function.Function{
		"readfile": hclReadFile(),
		"absolute": stdlib.AbsoluteFunc, "add": stdlib.AddFunc, "and": stdlib.AndFunc, "byteslen": stdlib.BytesLenFunc, "bytesslice": stdlib.BytesSliceFunc, "csvdecode": stdlib.CSVDecodeFunc, "ceil": stdlib.CeilFunc, "chomp": stdlib.ChompFunc, "chunklist": stdlib.ChunklistFunc, "coalesce": stdlib.CoalesceFunc, "coalescelist": stdlib.CoalesceListFunc, "compact": stdlib.CompactFunc, "concat": stdlib.ConcatFunc, "contains": stdlib.ContainsFunc, "distinct": stdlib.DistinctFunc, "divide": stdlib.DivideFunc, "element": stdlib.ElementFunc, "equal": stdlib.EqualFunc, "flatten": stdlib.FlattenFunc, "floor": stdlib.FloorFunc, "formatdate": stdlib.FormatDateFunc, "format": stdlib.FormatFunc, "formatlist": stdlib.FormatListFunc, "greaterthan": stdlib.GreaterThanFunc, "greaterthanorequalto": stdlib.GreaterThanOrEqualToFunc, "hasindex": stdlib.HasIndexFunc, "indent": stdlib.IndentFunc, "index": stdlib.IndexFunc, "int": stdlib.IntFunc, "jsondecode": stdlib.JSONDecodeFunc, "jsonencode": stdlib.JSONEncodeFunc, "join": stdlib.JoinFunc, "keys": stdlib.KeysFunc, "length": stdlib.LengthFunc, "lessthan": stdlib.LessThanFunc, "lessthanorequalto": stdlib.LessThanOrEqualToFunc, "log": stdlib.LogFunc, "lookup": stdlib.LookupFunc, "lower": stdlib.LowerFunc, "max": stdlib.MaxFunc, "merge": stdlib.MergeFunc, "min": stdlib.MinFunc, "modulo": stdlib.ModuloFunc, "multiply": stdlib.MultiplyFunc, "negate": stdlib.NegateFunc, "notequal": stdlib.NotEqualFunc, "not": stdlib.NotFunc, "or": stdlib.OrFunc, "parseint": stdlib.ParseIntFunc, "pow": stdlib.PowFunc, "range": stdlib.RangeFunc, "regexall": stdlib.RegexAllFunc, "regex": stdlib.RegexFunc, "regexreplace": stdlib.RegexReplaceFunc, "replace": stdlib.ReplaceFunc, "reverse": stdlib.ReverseFunc, "reverselist": stdlib.ReverseListFunc, "sethaselement": stdlib.SetHasElementFunc, "setintersection": stdlib.SetIntersectionFunc, "setproduct": stdlib.SetProductFunc, "setsubtract": stdlib.SetSubtractFunc, "setsymmetricdifference": stdlib.SetSymmetricDifferenceFunc, "setunion": stdlib.SetUnionFunc, "signum": stdlib.SignumFunc, "slice": stdlib.SliceFunc, "sort": stdlib.SortFunc, "split": stdlib.SplitFunc, "strlen": stdlib.StrlenFunc, "substr": stdlib.SubstrFunc, "subtract": stdlib.SubtractFunc, "timeadd": stdlib.TimeAddFunc, "title": stdlib.TitleFunc, "trim": stdlib.TrimFunc, "trimprefix": stdlib.TrimPrefixFunc, "trimspace": stdlib.TrimSpaceFunc, "trimsuffix": stdlib.TrimSuffixFunc, "upper": stdlib.UpperFunc, "values": stdlib.ValuesFunc, "zipmap": stdlib.ZipmapFunc,
	}
	return fnn
}

func hclReadFile() function.Function {
	x := function.Spec{
		Params: []function.Parameter{{
			Name: "filename",
			Type: cty.String,
		}},
		VarParam: nil,
		Type: func(args []cty.Value) (cty.Type, error) {
			return cty.String, nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.StringVal(string(ReadStaticFile(args[0].AsString()))), nil
		},
	}
	return function.New(&x)
}

func hclTemplateFn(e hcl.Expression, ctx *hcl.EvalContext) function.Function {
	x := function.Spec{
		Params: []function.Parameter{{
			Name: "data",
			Type: cty.DynamicPseudoType,
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

func hclTemplateTransformer(ctx *hcl.EvalContext, body hcl.Body) hcl.Body {
	sch := hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "template",
				LabelNames: []string{"templateName"},
			},
		},
	}

	//sch.Attributes = []hcl.AttributeSchema{{"markdown", true}}
	b, remain, diags := body.PartialContent(&sch)
	_ = spew.Dump

	for _, item := range b.Blocks.OfType("template") {
		attr, _ := item.Body.JustAttributes()
		ctx.Functions[item.Labels[0]] = hclTemplateFn(attr["markdown"].Expr, ctx)
	}
	/*
		for _, item := range c.Blocks.OfType("template") {
			item.Body.
		}
		fmt.Println(c.MissingItemRange.String())
	*/
	return transform.BodyWithDiagnostics(remain, diags)
}
