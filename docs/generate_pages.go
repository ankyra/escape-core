package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

type Page struct {
	Name       string
	Slug       string
	SrcFile    string
	StructName string
}

var Pages = map[string]Page{
	"consumer":  Page{"Consumers", "consumers", "consumer.go", "ConsumerConfig"},
	"depends":   Page{"Dependencies", "dependencies", "dependency_config.go", "DependencyConfig"},
	"downloads": Page{"Downloads", "downloads", "download_config.go", "DownloadConfig"},
	"errands":   Page{"Errands", "errands", "errand.go", "Errand"},
	"templates": Page{"Templates", "templates", "templates/templates.go", "Template"},
	"variables": Page{"Input and Output Variables", "input-and-output-variables", "variables/variable.go", "Variable"},
}

const PageHeader = `---
date: 2017-11-11 00:00:00
title: "%s"
slug: %s
type: "docs"
toc: true
---

%s

Field | Type | Description
------|------|-------------
%s
`

func GetJsonFieldFromTag(tag string) string {
	for _, s := range strings.Split(tag, " ") {
		s = strings.Trim(s, "`")
		if strings.HasPrefix(s, "json:\"") {
			s = s[6 : len(s)-1]
			return s
		}
	}
	return ""
}

func StructTable(page Page, topLevelDoc string, s *ast.TypeSpec) string {
	structType := s.Type.(*ast.StructType)
	result := ""
	for _, field := range structType.Fields.List {
		doc := strings.TrimSpace(field.Doc.Text())
		tag := GetJsonFieldFromTag(field.Tag.Value)
		typ := ""
		result += "|" + tag + "|" + typ + "|" + doc + "|\n"
	}
	return fmt.Sprintf(PageHeader, page.Name, page.Slug, topLevelDoc, result)
	return s.Name.String() + ": " + topLevelDoc + result
}

func GenerateStructDocs(f *ast.File, page Page) string {
	for _, decl := range f.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
			for _, spec := range gen.Specs {
				if s, ok := spec.(*ast.TypeSpec); ok {
					switch s.Type.(type) {
					case *ast.StructType:
						if s.Name.String() == page.StructName {
							return StructTable(page, gen.Doc.Text(), s)
						}
					}
				}
			}
		}
	}
	return ""
}

func main() {
	os.Mkdir("docs/generated/", 0755)
	for _, page := range Pages {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, page.SrcFile, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		str := GenerateStructDocs(f, page)
		filename := "docs/generated/" + page.Slug + ".md"
		fmt.Println("Writing ", filename)
		ioutil.WriteFile(filename, []byte(str), 0644)
	}
}
