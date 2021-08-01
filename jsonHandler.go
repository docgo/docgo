package main

import (
	mdAst "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"os"
	"fmt"
	"github.com/yuin/goldmark"
	"path/filepath"
)

type PkgList map[string]map[string]MarkdownFile
func ParsePackages(dirPath string) (list PkgList) {
	mdFile := filepath.Join(dirPath, "DOCS.md")
	mdData, err := os.ReadFile(mdFile)
	if err != nil {
		fmt.Println("[note] no DOCS.md found, will use default.")
		return
	}
	mdReader := text.NewReader(mdData)
	md := goldmark.DefaultParser().Parse(mdReader)
	list = parseJson(func(segment text.Segment) string {
		return string(mdData[segment.Start : segment.Stop])
	}, md)
	return list
}

type MarkdownFile struct {
	pkg string
	name string
	niceName string
	description string
}

func parseJson(getSegment func(segment text.Segment) string, node mdAst.Node) map[string]map[string]MarkdownFile {
	currentText := ""
	var currentFile = new(MarkdownFile)
	modules := make([]*MarkdownFile, 0)
	mdAst.Walk(node, func(n mdAst.Node, entering bool) (mdAst.WalkStatus, error) {
		switch orig := n.(type) {
		case *mdAst.Heading:
			if !entering {
				if orig.Level == 1 {
					currentFile.pkg = currentText
				}
				if orig.Level == 2 {
					currentFile.niceName = currentText
				}
				currentText = ""
				break
			} else {
				if currentFile.name != "" {
					currentFile.description = currentText
					samePkg := currentFile.pkg
					currentText = ""
					modules = append(modules, currentFile)
					currentFile = new(MarkdownFile)
					currentFile.pkg = samePkg
				}
			}
		case *mdAst.Document:
			if !entering {
				currentFile.description = currentText
				modules = append(modules, currentFile)
			}
		case *mdAst.CodeSpan:
			if !entering {
				currentFile.name = currentText
				currentText = ""
			}
		case *mdAst.Emphasis:
			if orig.Level == 1 {
				currentText += "*"
			} else {
				currentText += "**"
			}
		case *mdAst.Paragraph:
			currentText += "\n"
		case *mdAst.ListItem:
			if entering {
				currentText += "\n* "
			} else {
			}
		case *mdAst.Text:
			if !entering {
				/*
					currentFile.description += currentText
				currentLevel = -1 */
			} else {
				currentText += getSegment(orig.Segment)
			}
		}
		//fmt.Println(n.Type(), n.Kind())
		return mdAst.WalkContinue, nil
	})
	var out = make(map[string]map[string]MarkdownFile)
	for _, mmod := range modules {
		_, ok := out[mmod.pkg]
		if !ok {
			out[mmod.pkg] = make(map[string]MarkdownFile)
		}
		out[mmod.pkg][mmod.name] = *mmod
	}
	return out
}
