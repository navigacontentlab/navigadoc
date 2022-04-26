package newsml

import (
	"bytes"
	"fmt"
	"strings"

	"bitbucket.org/infomaker/doc-format/v2/doc"
	"github.com/Infomaker/etree"
)

// ContentSet represents <newsItem><contentSet>
type ContentSet struct {
	InlineXML InlineXML `xml:"inlineXML"`
}

// InlineXML represents <newsItem><contentSet><inlineXML>
type InlineXML struct {
	Idf Idf `xml:"idf"`
}

// Idf represents <newsItem><contentSet><inlineXML><idf>
type Idf struct {
	Lang  string   `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Dir   string   `xml:"dir,attr,omitempty"`
	Xmlns string   `xml:"xmlns,attr,omitempty"`
	Group []*Group `xml:"group,omitempty"`
}

// Group represents <newsItem><contentSet><inlineXML><grouo>
type Group struct {
	Type  string           `xml:"type,attr,omitempty"`
	ID    string           `xml:"id,attr,omitempty"`
	Child []*ElementObject `xml:",any"`
}

// ElementObject ...
type ElementObject struct {
	Element *IDFElement
	Object  *IDFObject
}

type Property struct {
	Name  string `xml:"name,attr,omitempty"`
	Value string `xml:"value,attr,omitempty"`
}

func NewContentSet(lang string) *ContentSet {
	cs := ContentSet{
		InlineXML: InlineXML{
			Idf: Idf{
				Xmlns: "http://www.infomaker.se/idf/1.0",
				Dir:   "ltr",
				Lang:  lang,
				Group: []*Group{
					{
						Type: "body",
					},
				},
			},
		},
	}
	return &cs
}

func (cs *ContentSet) fromBlocks(blocks []doc.Block, opts *Options) error {
	for _, block := range blocks {
		if isElementType(block.Type, opts) {
			if err := cs.addElement(block, opts); err != nil {
				return err
			}
		} else {
			if err := cs.addObject(block, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *ContentSet) children() []*ElementObject {
	return cs.InlineXML.Idf.Group[0].Child
}

// NPS-346
// func (cs *ContentSet) empty() bool {
// 	return len(cs.InlineXML.Idf.Group[0].Child) == 0
// }

func (cs *ContentSet) addContent(content *ElementObject) {
	cs.InlineXML.Idf.Group[0].Child = append(cs.InlineXML.Idf.Group[0].Child, content)
}

func (cs *ContentSet) addElement(block doc.Block, opts *Options) error {
	element := &IDFElement{}
	if err := element.fromBlock(block, opts); err != nil {
		return err
	}
	cs.addContent(&ElementObject{Element: element})
	return nil
}

func (cs *ContentSet) addObject(block doc.Block, opts *Options) error {
	var object IDFObject
	if err := object.fromBlock(block, opts, ContentContext); err != nil {
		return err
	}

	cs.addContent(&ElementObject{Object: &object})

	return nil
}

// This will take a single child from group or
func handleGroupChild(child ElementObject, opts *Options, context ContextType) (doc.Block, error) {
	block := doc.Block{}

	if child.Element != nil {
		el := child.Element
		err := el.toBlock(&block, opts)
		if err != nil {
			return block, err
		}
	}

	if child.Object != nil {
		ob := child.Object
		err := ob.toBlock(&block, opts, context)
		if err != nil {
			return block, err
		}
	}

	return block, nil
}

func buildList(text string, opts *Options) []doc.Block {
	if text == "" {
		return nil
	}

	var content []doc.Block

	xmlData := fmt.Sprintf("<data>%s</data>", strings.TrimSpace(text))

	listDoc := etree.NewDocument()
	err := listDoc.ReadFromString(xmlData)
	if err != nil {
		return nil
	}

	listItems := listDoc.FindElements("data/list-item")
	if len(listItems) > 0 {
		var innerXML bytes.Buffer
		content = make([]doc.Block, 0)

		for _, listItem := range listItems {
			innerXML.Reset()

			for _, child := range listItem.Child {
				child.WriteTo(&innerXML, &listDoc.WriteSettings)
			}

			html, err := SanitizeHTML(innerXML.String(), opts)
			if err != nil {
				return nil
			}
			newItem := doc.Block{
				Type: "x-im/paragraph",
				Data: map[string]string{
					"text":   html,
					"format": "html",
				},
			}

			content = append(content, newItem)
		}
	}

	return content
}

func (cs *ContentSet) toDoc(document *doc.Document, opts *Options) error {
	cs.addLanguageToDoc(document, &cs.InlineXML.Idf)
	for _, group := range cs.InlineXML.Idf.Group {
		for _, child := range group.Child {
			block, err := handleGroupChild(*child, opts, ContentContext)
			if err != nil {
				return err
			}

			document.Content = append(document.Content, block)
		}
	}
	return nil
}

func (cs *ContentSet) addLanguageToDoc(document *doc.Document, idf *Idf) {
	if idf != nil && idf.Lang != "" {
		document.Language = idf.Lang
	}
}
