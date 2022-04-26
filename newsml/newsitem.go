package newsml

import (
	"encoding/xml"
	"errors"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

// NewsItem represents the <newsItem> document
type NewsItem struct {
	XMLName         xml.Name     `xml:"newsItem"`
	XMLNamespace    string       `xml:"xmlns,attr,omitempty"`
	Conformance     string       `xml:"conformance,attr,omitempty"`
	GUID            string       `xml:"guid,attr,omitempty"`
	Standard        string       `xml:"standard,attr,omitempty"`
	StandardVersion string       `xml:"standardversion,attr,omitempty"`
	Version         string       `xml:"version,attr,omitempty"`
	CatalogRef      []CatalogRef `xml:"catalogRef,omitempty"`
	ItemMeta        *ItemMeta    `xml:"itemMeta,omitempty"`
	ContentMeta     *ContentMeta `xml:"contentMeta,omitempty"`
	ContentSet      *ContentSet  `xml:"contentSet,omitempty"`
}

// CatalogRef represents the <newsItem catalogRef> Attribute

// Language represents the language the text is written in
type Language struct {
	Tag string `xml:"tag,attr,omitempty"`
}

// Provider represents the <newsItem><provider> element
// The party (person or organisation) responsible for the management of the Item.
type Provider struct {
	Literal string `xml:"literal,attr,omitempty"`
}

// PubStatus represents the <newsItem><pubStatus> element
type PubStatus struct {
	QCode string `xml:"qcode,attr"`
}

// Service represents the <newsItem><service> element
type Service struct {
	QCode string `xml:"qcode,attr,omitempty"`
	Name  string `xml:"name,omitempty"`
	Why   string `xml:"why,attr,omitempty"`
}

// Link represents a <link> element
type Link struct {
	Rel   string  `xml:"rel,attr,omitempty"`
	Role  string  `xml:"role,attr,omitempty"`
	Title string  `xml:"title,attr,omitempty"`
	Type  string  `xml:"type,attr,omitempty"`
	QCode string  `xml:"qcode,attr,omitempty"`
	URI   string  `xml:"uri,attr,omitempty"`
	URL   string  `xml:"url,attr,omitempty"`
	UUID  string  `xml:"uuid,attr,omitempty"`
	Data  *Data   `xml:"data,omitempty"`
	Links Links   `xml:"links"`
	Meta  Objects `xml:"meta"`
}

// ItemMetaExtProperty represents the <newsItem><itemMeta><itemMetaExtProperty> element
type MetaExtProperty struct {
	Type    string `xml:"type,attr"`
	Value   string `xml:"value,attr"`
	Creator string `xml:"creator,attr,omitempty"`
	Why     string `xml:"why,attr,omitempty"`
	Literal string `xml:"literal,attr,omitempty"`
}

// InfoSource represents <newsItem><contentMeta><infoSource>
type InfoSource struct {
	Text    string `xml:",chardata"`
	Literal string `xml:"literal,attr,omitempty"`
}

// Creator represents <newsItem><contentMeta><creator>
type Creator struct {
	Literal string `xml:"literal,attr,omitempty"`
	Text    string `xml:",chardata"`
}

// Object represents <newsItem><contentMeta><metadata><object>
type Object struct {
	ID    string `xml:"id,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`
	Title string `xml:"title,attr,omitempty"`
	UUID  string `xml:"uuid,attr,omitempty"`
	Data  *Data  `xml:"data"`

	Links      Links       `xml:"links"`
	Properties *Properties `xml:"properties,omitempty"`
	Meta       Objects     `xml:"meta"`
}

type Properties struct {
	Property []*Property `xml:"property,omitempty"`
}

// Data represents <newsItem><contentMeta><metadata><object><data>
type Data struct {
	Raw string `xml:",innerxml"`
}

// MarshalXML handles transformation of the combined type of Object and Element
// in IDF.
// When Unmarshal all element and object becomes ElementObject struct.
// When we are going back to xml we need to identify and create correct type of element
func (elob ElementObject) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if elob.Object != nil {
		start.Name.Local = "object"
		return e.EncodeElement(elob.Object, start)
	}

	if elob.Element != nil {
		start.Name.Local = "element"
		return e.EncodeElement(elob.Element, start)
	}

	return nil
}

func (elob *ElementObject) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case "element":
		var e IDFElement
		if err := d.DecodeElement(&e, &start); err != nil {
			return err
		}
		elob.Element = &e
	case "object":
		var o IDFObject

		if err := d.DecodeElement(&o, &start); err != nil {
			return err
		}

		elob.Object = &o
	default:
		return d.Skip()
	}

	return nil
}

// AsDocument takes a newsitem or a conceptitem and returns a document
func AsDocument(v interface{}, opts *Options) (*doc.Document, error) {
	document := &doc.Document{}

	switch i := v.(type) {
	case *NewsItem:
		err := i.toDoc(document, opts)
		if err != nil {
			return nil, err
		}
	case *ConceptItem:
		err := i.toDoc(document, opts)
		if err != nil {
			return nil, err
		}
	}

	return document, nil
}

// AsNewsItem returns a new newsitem based on the document contents
func AsNewsItem(document *doc.Document, opts *Options) (*NewsItem, error) {
	newsitem := &NewsItem{}

	err := newsitem.fromDoc(document, opts)
	if err != nil {
		return nil, err
	}
	return newsitem, nil
}

func (n *NewsItem) fromDoc(document *doc.Document, opts *Options) error {
	if document == nil {
		return errors.New("nil document")
	}

	n.XMLName = xml.Name{
		Local: "newsItem",
	}
	n.XMLNamespace = "http://iptc.org/std/nar/2006-10-01/"
	n.GUID = document.UUID
	n.Conformance = "power"
	n.Standard = "NewsML-G2"
	n.StandardVersion = "2.26"
	n.Version = "1"
	n.CatalogRef = []CatalogRef{
		{
			Href: "http://www.iptc.org/std/catalog/catalog.IPTC-G2-Standards_30.xml",
		},
		{
			Href: "http://infomaker.se/spec/catalog/catalog.infomaker.g2.1_0.xml",
		},
	}

	n.ItemMeta = &ItemMeta{}
	err := n.ItemMeta.FromDoc(document, opts)
	if err != nil {
		return err
	}

	n.ContentMeta = &ContentMeta{}
	err = n.ContentMeta.fromDoc(document, opts)
	if err != nil {
		return err
	}

	n.ContentSet = NewContentSet(document.Language)
	err = n.ContentSet.fromBlocks(document.Content, opts)
	if err != nil {
		return err
	}

	if document.Type != "x-im/article" {
		n.ContentSet = nil
	}

	if n.ContentMeta.empty() {
		n.ContentMeta = nil
	}

	return nil
}

func (n *NewsItem) toDoc(document *doc.Document, opts *Options) error {
	document.UUID = n.GUID
	document.Properties = []doc.Property{}

	if n.ItemMeta != nil {
		err := n.ItemMeta.ToDoc(document, opts)
		if err != nil {
			return err
		}
	}

	if n.ContentMeta != nil {
		err := n.ContentMeta.toDoc(document, opts)
		if err != nil {
			return err
		}
		if document.Language != "" {
			n.ContentMeta.Language = &Language{
				document.Language,
			}
		}
	}

	if n.ContentSet != nil {
		err := n.ContentSet.toDoc(document, opts)
		if err != nil {
			return err
		}
		if document.Language != "" {
			n.ContentSet.InlineXML.Idf.Lang = document.Language
		}
	}

	for _, c := range document.Content {
		if len(c.Data) == 0 {
			c.Data = nil
		}
	}

	return nil
}
