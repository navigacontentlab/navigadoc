package listpackage

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/infomaker/doc-format/v2/doc"
	"bitbucket.org/infomaker/doc-format/v2/newsml"
)

type Package struct {
	XMLName   xml.Name         `xml:"package"`
	UUID      string           `xml:"uuid,attr"`
	Published bool             `xml:"published,attr"`
	Cover     *Cover           `xml:"cover"`
	Name      string           `xml:"name,omitempty"`
	Type      string           `xml:"type,omitempty"`
	Products  []*Product       `xml:"products>product,omitempty"`
	Category  string           `xml:"category,omitempty"`
	PubStart  string           `xml:"pubStart,omitempty"`
	PubStop   string           `xml:"pubStop,omitempty"`
	PubStatus string           `xml:"pubStatus,omitempty"`
	ItemList  *ItemList        `xml:"itemList,omitempty"`
	ItemMeta  *newsml.ItemMeta `xml:"itemMeta,omitempty"`
}

type Cover struct {
	UUID string `xml:"uuid,attr"`
}

type ItemList struct {
	UUID string `xml:"uuid,attr"`
}

// AsNewsItem returns a new newsitem based on the document contents
func AsPackage(document *doc.Document, opts *newsml.Options) (*Package, error) {
	pkg := &Package{}

	err := pkg.fromDoc(document, opts)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

func (p *Package) validateFilds() error {
	return nil
}

func (p *Package) fromDoc(document *doc.Document, opts *newsml.Options) error {
	if document == nil {
		return errors.New("nil document")
	}

	p.UUID = document.UUID

	var pubStatuses = map[string]string{
		"draft":    "draft",
		"usable":   "public",
		"canceled": "private",
		"":         "draft",
	}

	if status, ok := pubStatuses[document.Status]; ok {
		p.PubStatus = status
	} else {
		return fmt.Errorf("can't convert status %s", document.Status)
	}

	p.ItemMeta = &newsml.ItemMeta{}
	err := p.ItemMeta.FromDoc(document, opts)
	if err != nil {
		return err
	}

	p.ItemMeta.PubStatus = nil

	ext2 := make([]newsml.MetaExtProperty, 0)
	for _, ext := range p.ItemMeta.ItemMetaExtProperty {
		var isForPackage = false
		for _, prop := range opts.PropertyException {
			if ext.Type == prop.Type {
				for _, section := range prop.Section {
					if section == "package" {
						isForPackage = true
					}
				}
				break
			}
		}
		if !isForPackage {
			ext2 = append(ext2, ext)
		}
	}

	p.ItemMeta.ItemMetaExtProperty = ext2

	p.Name = document.Title
	if document.Published != nil {
		p.PubStart = document.Published.Format(time.RFC3339Nano)
	}
	if document.Unpublished != nil {
		p.PubStop = document.Unpublished.Format(time.RFC3339Nano)
	}

	p.ItemMeta.Title = ""
	p.ItemMeta.ItemClass = nil

	for _, prop := range document.Properties {
		switch strings.ToLower(prop.Name) {
		case "type":
			p.Type = prop.Value
		case "category":
			p.Category = prop.Value
		case "published":
			isPublished, err := strconv.ParseBool(prop.Value)
			if err != nil {
				return err
			}
			p.Published = isPublished
		}
	}

	for _, link := range document.Links {
		switch link.Rel {
		case "list":
			p.ItemList = &ItemList{
				UUID: link.UUID,
			}
		case "channel":
			p.Products = append(p.Products,
				&Product{
					Text: link.Title,
					UUID: link.UUID,
				})
		case "cover":
			if p.Cover == nil {
				p.Cover = &Cover{
					UUID: link.UUID,
				}
			}
		}
	}

	// Remove the list and channel links from itemMeta
	for i := len(p.ItemMeta.Links) - 1; i >= 0; {
		link := p.ItemMeta.Links[i]
		if link.Rel == "list" || link.Rel == "channel" || link.Rel == "cover" {
			p.ItemMeta.Links = append(p.ItemMeta.Links[:i], p.ItemMeta.Links[i+1:]...)
			i = len(p.ItemMeta.Links) - 1
			continue
		}
		i--
	}

	// Remove the published property
	for i := len(p.ItemMeta.ItemMetaExtProperty) - 1; i >= 0; {
		prop := p.ItemMeta.ItemMetaExtProperty[i]
		if prop.Type == "published" {
			p.ItemMeta.ItemMetaExtProperty = append(p.ItemMeta.ItemMetaExtProperty[:i], p.ItemMeta.ItemMetaExtProperty[i+1:]...)
			break
		}
		i--
	}

	return nil
}

func (p *Package) toDoc(document *doc.Document, opts *newsml.Options) error {
	if err := p.validateFilds(); err != nil {
		return err
	}
	document.UUID = p.UUID
	document.Type = "x-im/package"

	var pubStatuses = map[string]string{
		"draft":   "draft",
		"public":  "usable",
		"private": "canceled",
		"":        "draft",
	}
	if _, ok := pubStatuses[p.PubStatus]; !ok {
		return fmt.Errorf("can't convert status %s", p.PubStatus)
	}

	document.Status = pubStatuses[p.PubStatus]
	newsml.AddPropertyToDoc(document, "published", strconv.FormatBool(p.Published))

	if p.PubStart != "" {
		pubdate, err := time.Parse(time.RFC3339Nano, p.PubStart)
		if err != nil {
			return err
		}
		document.Published = &pubdate
	}
	if p.PubStop != "" {
		pubdate, err := time.Parse(time.RFC3339Nano, p.PubStop)
		if err != nil {
			return err
		}
		document.Unpublished = &pubdate
	}

	if document.Meta == nil {
		document.Meta = make([]doc.Block, 0)
	}
	document.Meta = append(document.Meta, doc.Block{
		Type: "x-im/package",
		Data: nil,
	})

	for _, product := range p.Products {
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

	if p.Cover != nil {
		link := doc.Block{
			Rel:  "cover",
			Type: "x-im/article",
			UUID: p.Cover.UUID,
		}
		if document.Links == nil {
			document.Links = make([]doc.Block, 0)
		}

		document.Links = append(document.Links, link)
	}

	newsml.AddPropertyToDoc(document, "type", p.Type)
	newsml.AddPropertyToDoc(document, "category", p.Category)

	if p.ItemMeta != nil {
		err := p.ItemMeta.ToDoc(document, opts)
		if err != nil {
			return err
		}
	}

	if p.ItemList != nil {
		link := doc.Block{
			Type: "x-im/list",
			Rel:  "list",
			UUID: p.ItemList.UUID,
		}

		if document.Links == nil {
			document.Links = make([]doc.Block, 0)
		}
		document.Links = append(document.Links, link)
	}

	// ItemMeta.ToDoc sets title so this must be done here
	document.Title = p.Name

	return nil
}
