package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type GenStruct struct {
	Order  int
	Name   string
	Fields []Field
}

type Field struct {
	Name    string
	RPCName string
	Type    string
	RPCType string
	Comment []string
	Out     string
}

type Generator struct {
	output  *bytes.Buffer
	structs map[string]GenStruct
}

func (g *Generator) P(args ...string) {
	for _, v := range args {
		g.output.WriteString(v)
	}
	g.output.WriteByte('\n')
}

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "rpc/document.pb.go", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Handle struct generation
	generator := Generator{
		output:  bytes.NewBuffer(nil),
		structs: make(map[string]GenStruct),
	}
	generator.collect(fset, file)
	generator.process()
	generator.formatAndSave("doc/document.go")

	// Handle conversion functions
	conGen := Generator{
		output:  bytes.NewBuffer(nil),
		structs: generator.structs,
	}
	conGen.createConverters()
	conGen.formatAndSave("rpc/conversion.go")
}
func (g *Generator) collect(fset *token.FileSet, file *ast.File) {
	var currentStruct string
	var structNum int

	ast.Inspect(file, func(x ast.Node) bool {
		if x == nil {
			return false
		}
		switch n := x.(type) {
		case *ast.TypeSpec:
			currentStruct = n.Name.String()
			var fields []Field

			g.structs[currentStruct] = GenStruct{
				Order:  structNum,
				Name:   currentStruct,
				Fields: fields,
			}

			structNum++

		case *ast.StructType:

			s := g.structs[currentStruct]

			for _, field := range n.Fields.List {
				f := Field{
					RPCName: field.Names[0].Name,
				}

				// Ignore unexported
				if f.RPCName != strings.Title(f.RPCName) {
					continue
				}

				f.Name = mapName(f.RPCName)

				// Handle field comments
				if field.Doc != nil && len(field.Doc.List) > 0 {
					for _, c := range field.Doc.List {
						f.Comment = append(f.Comment, c.Text)
					}
				}

				f.RPCType = getFieldType(field, fset)

				// Replace google timestamp with regular time.Time
				replaceField(field, "timestamppb", "Timestamp", "time", "Time")

				f.Type = getFieldType(field, fset)

				tag := reflect.StructTag(field.Tag.Value)
				jsonTag := tag.Get("json")
				f.Out = fmt.Sprintf("%s %s `json:\"%s\"`", f.Name, f.Type, jsonTag)
				s.Fields = append(s.Fields, f)
			}

			g.structs[currentStruct] = s
		}
		return true
	})
}

func (g *Generator) getStructs() []GenStruct {
	var structs []GenStruct

	for _, s := range g.structs {
		structs = append(structs, s)
	}
	sort.Slice(structs, func(i, j int) bool {
		return structs[i].Order < structs[j].Order
	})

	return structs
}

var translateFields = map[string]string{
	"Id":   "ID",
	"Uri":  "URI",
	"Url":  "URL",
	"Uuid": "UUID",
}

func mapName(rpcName string) string {
	t, ok := translateFields[rpcName]
	if ok {
		return t
	}

	return rpcName
}

func (g *Generator) process() {
	g.P(`// Code generated by doc-generator. DO NOT EDIT.`)
	g.P(`package doc`)
	g.P(`import `, `(`)
	g.P(`"time"`)
	g.P(``)
	g.P(`)`)

	structs := g.getStructs()
	generateSchema(structs)

	// Section were the struct and field is generated
	for _, s := range structs {
		// Ignore unexported
		if s.Name != strings.Title(s.Name) {
			continue
		}

		g.P(`type `, s.Name, ` struct {`)
		for _, f := range s.Fields {
			for _, comment := range f.Comment {
				g.P(comment)
			}
			if f.Type == "[]*Block" || f.Type == "[]*Property" {
				g.P(strings.Replace(f.Out, "*", "", 1))
			} else {
				g.P(f.Out)
			}
		}
		g.P(`}`)
	}
}

func generateField(f *Field, direction, from, to string) []string {
	fromName := f.Name
	toName := f.RPCName

	if direction == "to" {
		fromName, toName = toName, fromName
	}

	res := []string{
		to, ".", toName, " = ", from, ".", fromName,
	}

	switch f.RPCType {
	case "*timestamppb.Timestamp":
		res = []string{
			to, ".", fromName, " = ", direction, "DocTime(", from, ".", toName, ")",
		}
	case "[]*Block":
		res = []string{
			to, ".", fromName, " = ", direction, "DocBlocks(", from, ".", toName, ")",
		}
	case "[]*Property":
		res = []string{
			to, ".", fromName, " = ", direction, "DocProperties(", from, ".", toName, ")",
		}
	}
	return res
}

func (g *Generator) createConverters() {
	g.P(`// Code generated by doc-generator. DO NOT EDIT.`)
	g.P(`package rpc`)
	g.P(`import `, `(`)
	g.P(`"time"`)
	g.P(``)
	g.P(`doc "bitbucket.org/infomaker/doc-format/v2/doc"`)
	g.P(`"google.golang.org/protobuf/types/known/timestamppb"`)
	g.P(`)`)

	// Generate methods to convert structs

	structs := g.getStructs()

	for _, s := range structs {
		// Ignore unexported
		if s.Name != strings.Title(s.Name) {
			continue
		}

		g.P(`func (d *`, s.Name, ` ) FromDoc`, s.Name, `(doc *doc.`, s.Name, `) error  {`)
		for _, f := range s.Fields {
			fld := f
			g.P(generateField(&fld, "from", "doc", "d")...)
		}
		g.P(`return nil`)
		g.P(`}`)
	}

	for _, s := range structs {
		// Ignore unexported
		if s.Name != strings.Title(s.Name) {
			continue
		}

		g.P(`func (d *`, s.Name, `) ToDoc`, s.Name, `(doc *doc.`, s.Name, `) error  {`)
		for _, f := range s.Fields {
			fld := f
			g.P(generateField(&fld, "to", "d", "doc")...)
		}
		g.P(`return nil`)
		g.P(`}`)
	}
	g.P(`
    func fromDocBlocks(blocks []doc.Block) []*Block {
            var rpcBlocks []*Block
        for _, b := range blocks {
                rpcBlock := Block{}
                rpcBlock.FromDocBlock(&b)
                rpcBlocks = append(rpcBlocks, &rpcBlock)
        }
        return rpcBlocks
    } 
    `)

	g.P(`
    func toDocBlocks(blocks []*Block) []doc.Block {
            var docBlocks []doc.Block
            for _, b := range blocks {
                    docBlock := doc.Block{}
                    b.ToDocBlock(&docBlock)
                    docBlocks = append(docBlocks, docBlock)
            }
            return docBlocks
    }
    `)

	g.P(`
    func fromDocProperties(properties []doc.Property) []*Property {
        var rpcProperties []*Property
        for _, p := range properties {
                rpcProp := Property{}
                rpcProp.FromDocProperty(&p)
                rpcProperties = append(rpcProperties, &rpcProp)
        }
        return rpcProperties
    }
    `)

	g.P(`
    func toDocProperties(properties []*Property) []doc.Property {
            var docProperties []doc.Property
            for _, p := range properties {
                    docProp := doc.Property{}
                    p.ToDocProperty(&docProp)
                    docProperties = append(docProperties, docProp)
            }
            return docProperties
    }
    `)

	g.P(`
    func fromDocTime(ts *time.Time) *timestamppb.Timestamp {
            if ts != nil {
                    time := timestamppb.New(*ts)
                    return time
            }
            return nil
    }
    `)

	g.P(`
    func toDocTime(ts *timestamppb.Timestamp) *time.Time {
            if ts != nil {
                    time := ts.AsTime()
                    return &time
            }
            return nil
    }
    `)
}
func (g *Generator) formatAndSave(out string) {
	result, err := format.Source(g.output.Bytes())
	if err != nil {
		fmt.Println("Failed format: %w", err)
	}

	err = ioutil.WriteFile(out, result, 0600)
	if err != nil {
		fmt.Println("Failed to write file %w", err)
	}
}

func getFieldType(field *ast.Field, fset *token.FileSet) string {
	var typeNameBuf bytes.Buffer
	err := printer.Fprint(&typeNameBuf, fset, field.Type)
	if err != nil {
		log.Fatalf("failed printing %s", err)
	}
	return typeNameBuf.String()
}

func replaceField(f *ast.Field, oldPackage, oldType, newPackage, newType string) {
	if se, ok := (f.Type).(*ast.StarExpr); ok {
		if selectorExpr, ok := se.X.(*ast.SelectorExpr); ok {
			if id, ok := selectorExpr.X.(*ast.Ident); ok {
				if id.Name == oldPackage && selectorExpr.Sel.Name == oldType {
					id.Name = newPackage
					selectorExpr.Sel.Name = newType
				}
			}
		}
	}
}

type JSONSchema map[string]interface{}

func generateSchema(structs []GenStruct) {
	schema := JSONSchema{}
	// FIXME What to use?
	schema["$id"] = "http://navigalobal.com/navigadoc/schema"
	schema["title"] = "Navigadoc"
	schema["description"] = "Navigadoc Schema"
	schema["type"] = "object"
	docProperties := make(map[string]interface{})
	schema["properties"] = docProperties

	var doc GenStruct
	var block GenStruct
	var property GenStruct
	for _, s := range structs {
		switch s.Name {
		case "Document":
			doc = s
		case "Block":
			block = s
		case "Property":
			property = s
		}
	}

	definitions := make(map[string]interface{})
	schema["definitions"] = definitions
	blockDef := make(map[string]interface{})
	definitions["block"] = blockDef
	blockDef["type"] = "object"
	blockProperties := make(map[string]interface{})
	blockDef["properties"] = blockProperties

	propDef := make(map[string]interface{})
	definitions["property"] = propDef
	propDef["type"] = "object"
	propProperties := make(map[string]interface{})
	propDef["properties"] = propProperties

	schemaFields(block.Fields, blockProperties)
	schemaFields(property.Fields, propProperties)
	schemaFields(doc.Fields, docProperties)

	// FIXME Which are required fields?
	// blockDef["required"] = []string{"id", "uuid", "type"}
	propDef["required"] = []string{"name"}
	schema["required"] = []string{"uuid", "type", "created"}

	js, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return
	}

	err = ioutil.WriteFile("schema/navigadoc-schema.json", js, 0600)
	if err != nil {
		fmt.Println("Failed to write JSON Schema file %w", err)
	}
}

func schemaFields(fields []Field, out map[string]interface{}) {
	rgx := regexp.MustCompile("^.*`json:\"?(.*?),.*\"`")
	for _, f := range fields {
		var name string
		m := rgx.FindStringSubmatch(f.Out)
		if len(m) >= 1 {
			name = m[1]
		} else {
			name = strings.ToLower(f.Name)
		}
		switch name {
		case "uri", "url":
			out[name] = map[string]interface{}{"type": "string", "format": "uri"}
			continue
		case "uuid":
			out[name] = map[string]interface{}{"type": "string", "pattern": "[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}"}
			continue
		}
		switch f.Type {
		case "string":
			out[name] = map[string]interface{}{"type": "string"}
		case "*time.Time":
			out[name] = map[string]interface{}{"type": "string", "format": "date-time"}
		case "[]*Block":
			out[name] = map[string]interface{}{"type": "array", "items": map[string]interface{}{"$ref": "#/definitions/block"}}
		case "[]*Property":
			out[name] = map[string]interface{}{"type": "array", "items": map[string]interface{}{"$ref": "#/definitions/property"}}
		case "map[string]string":
			out[name] = map[string]interface{}{"type": "object"}
		}
	}
}
