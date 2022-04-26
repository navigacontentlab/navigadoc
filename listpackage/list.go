package listpackage

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/infomaker/doc-format/v2/doc"
	"bitbucket.org/infomaker/doc-format/v2/newsml"
)

type List struct {
	XMLName     xml.Name         `xml:"list"`
	UUID        string           `xml:"uuid,attr"`
	Name        string           `xml:"name,omitempty"`
	Description string           `xml:"description,omitempty"`
	Type        string           `xml:"type,omitempty"`
	Products    []*Product       `xml:"products>product,omitempty"`
	Items       *Items           `xml:"items,omitempty"`
	ItemMeta    *newsml.ItemMeta `xml:"itemMeta,omitempty"`
}

type Items struct {
	Limit int     `xml:"limit,attr"`
	Item  []*Item `xml:"item"`
}

type Item struct {
	UUID string `xml:"uuid,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

// AsList returns a new list based on the document contents
func AsList(document *doc.Document, opts *newsml.Options) (*List, error) {
	if document == nil {
		return nil, errors.New("nil document")
	}

	list := &List{}

	err := list.FromDoc(document, opts)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// TODO What validations to use?
func (l *List) validateFields() error {
	// if l.Description == "" {
	//	return errors.New("blank description")
	//}
	return nil
}

func (l *List) FromDoc(document *doc.Document, opts *newsml.Options) error {
	if document == nil {
		return errors.New("nil document")
	}
	if document.Type != "x-im/list" {
		return errors.New("wrong document type")
	}

	// Remove uri from document since it should not be reflected in xml
	docURI := document.URI
	document.URI = ""

	l.Type = "list"
	l.UUID = document.UUID

	l.ItemMeta = &newsml.ItemMeta{}
	err := l.ItemMeta.FromDoc(document, opts)
	if err != nil {
		return err
	}
	if l.ItemMeta.PubStatus != nil && l.ItemMeta.PubStatus.QCode == "" {
		l.ItemMeta.PubStatus = nil
	}

	links2 := make([]newsml.Link, 0)
	for _, link := range l.ItemMeta.Links {
		switch link.Rel {
		case "channel":
			l.Products = append(l.Products,
				&Product{
					Text: link.Title,
					UUID: link.UUID,
				})
		case "item":
			if l.Items == nil {
				l.Items = &Items{
					Item: make([]*Item, 0),
				}
			}

			item := &Item{
				Type: link.Type,
				UUID: link.UUID,
			}

			l.Items.Item = append(l.Items.Item, item)
		default:
			links2 = append(links2, link)
		}
	}
	l.ItemMeta.Links = links2

	if l.ItemMeta.Title != "" {
		l.ItemMeta.Title = ""
	}

	if l.ItemMeta.ItemClass != nil {
		l.ItemMeta.ItemClass = nil
	}

	// TODO External Source
	extTypes := map[string]string{
		"imext:description": "",
		"imext:product":     "",
		"imext:itemLimit":   "",
		"imext:type":        "",
	}

	// Rebuild the itemMetaExtProperty values minus the ones above
	ext2 := make([]newsml.MetaExtProperty, 0)
	for _, ext := range l.ItemMeta.ItemMetaExtProperty {
		if _, ok := extTypes[ext.Type]; !ok {
			ext2 = append(ext2, ext)
		}
	}
	l.ItemMeta.ItemMetaExtProperty = ext2

	l.Name = document.Title

	for _, meta := range document.Meta {
		if meta.Type == "x-im/list" {
			for key, value := range meta.Data {
				if key == "description" {
					l.Description = value
				} else if key == "limit" {
					limit, err := strconv.Atoi(value)
					if err == nil {
						if l.Items == nil {
							l.Items = &Items{}
						}
						l.Items.Limit = limit
					} else {
						return fmt.Errorf("invalid limit in document: %s", value)
					}
				}
			}
		}
	}

	if document.Content != nil {
		if l.Items == nil {
			l.Items = &Items{
				Item: make([]*Item, 0),
			}
		} else if l.Items.Item == nil {
			l.Items.Item = make([]*Item, 0)
		}
		for _, content := range document.Content {
			l.Items.Item = append(l.Items.Item, &Item{
				Type: strings.Title(strings.Replace(content.Type, "x-im/", "", 1)),
				UUID: content.UUID,
			})
		}
	}

	// Put back URI in document
	document.URI = docURI

	return nil
}

func (l *List) toDoc(document *doc.Document, opts *newsml.Options) error {
	if err := l.validateFields(); err != nil {
		return err
	}

	document.Type = "x-im/list"
	document.UUID = l.UUID

	for _, product := range l.Products {
		link := doc.Block{
			Rel:   "channel",
			Type:  "x-im/channel",
			Title: product.Text,
		}

		if product.UUID != "" {
			link.UUID = product.UUID
		}

		if document.Links == nil {
			document.Links = make([]doc.Block, 0)
		}

		document.Links = append(document.Links, link)
	}

	if l.Items != nil && l.Items.Item != nil && len(l.Items.Item) > 0 {
		if document.Content == nil {
			document.Content = make([]doc.Block, 0)
		}
		for _, item := range l.Items.Item {
			content := doc.Block{
				UUID: item.UUID,
				Type: fmt.Sprintf("x-im/%s", strings.ToLower(item.Type)),
			}
			document.Content = append(document.Content, content)
		}
	}

	if l.ItemMeta != nil {
		err := l.ItemMeta.ToDoc(document, opts)
		if err != nil {
			return err
		}
	}

	if document.Meta == nil {
		document.Meta = make([]doc.Block, 0)
	}

	meta := doc.Block{
		Type: "x-im/list",
		Data: map[string]string{
			"description": l.Description,
			"limit":       strconv.Itoa(l.Items.Limit),
		},
	}

	document.Meta = append(document.Meta, meta)

	// ItemMeta.ToDoc sets title so this must be done here
	document.Title = l.Name

	// Create and set URI
	document.URI = fmt.Sprintf("im://list/%s", l.UUID)

	return nil
}
