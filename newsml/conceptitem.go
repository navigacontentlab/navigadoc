package newsml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type ConceptItem struct {
	XMLName         xml.Name     `xml:"conceptItem"`
	XMLNamespace    string       `xml:"xmlns,attr,omitempty"`
	Conformance     string       `xml:"conformance,attr,omitempty"`
	GUID            string       `xml:"guid,attr,omitempty"`
	Standard        string       `xml:"standard,attr,omitempty"`
	StandardVersion string       `xml:"standardversion,attr,omitempty"`
	Version         string       `xml:"version,attr,omitempty"`
	CatalogRef      []CatalogRef `xml:"catalogRef,omitempty"`
	RightsInfo      *RightsInfo  `xml:"rightsInfo,omitempty"`
	ItemMeta        *ItemMeta    `xml:"itemMeta,omitempty"`
	ContentMeta     *ContentMeta `xml:"contentMeta,omitempty"`
	Concept         *Concept     `xml:"concept"`
}

// AsNewsItem returns a new newsitem based on the document contents
func AsConceptItem(document *doc.Document, opts *Options) (*ConceptItem, error) {
	conceptItem := &ConceptItem{}

	err := conceptItem.fromDoc(document, opts)
	if err != nil {
		return nil, err
	}
	return conceptItem, nil
}

func (c *ConceptItem) fromDoc(document *doc.Document, opts *Options) error {
	if document == nil {
		return errors.New("nil document")
	}

	c.XMLName = xml.Name{
		Local: "conceptItem",
	}
	c.XMLNamespace = "http://iptc.org/std/nar/2006-10-01/"
	c.GUID = document.UUID
	c.Conformance = "power"
	c.Standard = "NewsML-G2"
	c.StandardVersion = "2.26"
	c.Version = "1"
	c.CatalogRef = []CatalogRef{
		{
			Href: "http://www.iptc.org/std/catalog/catalog.IPTC-G2-Standards_30.xml",
		},
		{
			Href: "http://infomaker.se/spec/catalog/catalog.infomaker.g2.1_0.xml",
		},
	}

	for _, property := range document.Properties {
		if strings.EqualFold(property.Name, "copyrightHolder") {
			c.RightsInfo = &RightsInfo{
				CopyRightHolder: &CopyRightHolder{
					Name: property.Value,
				},
			}
		}
	}
	c.ItemMeta = &ItemMeta{}
	err := c.ItemMeta.FromDoc(document, opts)
	if err != nil {
		return err
	}

	// Remove duplicated entries
	for i := len(c.ItemMeta.ItemMetaExtProperty) - 1; i >= 0; i-- {
		switch c.ItemMeta.ItemMetaExtProperty[i].Type {
		case "imext:uri", "nrpdate:start", "nrpdate:end", "conceptid":
			c.ItemMeta.ItemMetaExtProperty = append(c.ItemMeta.ItemMetaExtProperty[:i], c.ItemMeta.ItemMetaExtProperty[i+1:]...)
		default:
		}
	}

	// Bypass qcodeToType as map can't have every type
	c.ItemMeta.ItemClass = &ItemClass{
		QCode: "cinat:concept",
	}
	// Title gets mapped to Concept.Name
	c.ItemMeta.Title = ""

	c.ContentMeta = &ContentMeta{}
	err = c.ContentMeta.fromDoc(document, opts)
	if err != nil {
		return err
	}

	for i := len(c.ContentMeta.MetaData) - 1; i >= 0; i-- {
		cmObject := c.ContentMeta.MetaData[i]
		for _, object := range opts.ObjectException {
			if cmObject.Type == object.Type {
				for _, section := range object.Section {
					if section == "conceptitem" {
						c.ContentMeta.MetaData = append(c.ContentMeta.MetaData[:i], c.ContentMeta.MetaData[i+1:]...)
						i = len(c.ContentMeta.MetaData)
					}
				}
			}
		}
	}

	var registration string
	var dateGranularity string
	for _, meta := range document.Meta {
		if (meta.Type == "x-im/newscoverage" || meta.Type == "x-im/event") && meta.Data != nil {
			for key, value := range meta.Data {
				if key == "dateGranularity" {
					dateGranularity = value
				}
			}
			for key, value := range meta.Data {
				switch key {
				case "slug":
					c.ContentMeta.Slugline = value
				case "priority":
					c.ContentMeta.Urgency = value
				case "description":
					if c.ContentMeta.Description == nil {
						c.ContentMeta.Description = make([]*Description, 0)
					}
					c.ContentMeta.Description = append(
						c.ContentMeta.Description,
						&Description{
							Role: "nrpdesc:intern",
							Text: value,
						})
				case "start":
					fallthrough
				case "end":
					if c.ItemMeta.ItemMetaExtProperty == nil {
						c.ItemMeta.ItemMetaExtProperty = make([]MetaExtProperty, 0)
					}
					c.ItemMeta.ItemMetaExtProperty = append(
						c.ItemMeta.ItemMetaExtProperty, MetaExtProperty{
							Type:  fmt.Sprintf("nrpdate:%s", key),
							Value: value,
							Why:   fmt.Sprintf("nrpwhy:%s", dateGranularity),
						})
				case "publicDescription":
					if c.ContentMeta.Description == nil {
						c.ContentMeta.Description = make([]*Description, 0)
					}
					c.ContentMeta.Description = append(
						c.ContentMeta.Description,
						&Description{
							Role: "nrpdesc:extern",
							Text: value,
						})
				case "registration":
					registration = value
				}
			}
		}
	}

	var xtraFields []string
	for _, prop := range opts.PropertyException {
		for _, section := range prop.Section {
			if section == "concept" {
				xtraFields = append(xtraFields, prop.Type)
			}
		}
	}
	// Remove itemMetaExtProperties having the above types
	for i := len(c.ItemMeta.ItemMetaExtProperty) - 1; i >= 0; i-- {
		im := &c.ItemMeta.ItemMetaExtProperty[i]
		for _, fv := range xtraFields {
			if strings.EqualFold(fv, im.Type) {
				c.ItemMeta.ItemMetaExtProperty = append(c.ItemMeta.ItemMetaExtProperty[:i], c.ItemMeta.ItemMetaExtProperty[i+1:]...)
				i = len(c.ItemMeta.ItemMetaExtProperty)
			}
		}
	}

	if document.Links != nil {
		for _, link := range document.Links {
			if link.Rel == "section" {
				extProp := &MetaExtProperty{
					Type:    "nrp:sector",
					Value:   link.Value,
					Literal: link.Title,
				}
				if c.ContentMeta.ContentMetaExtProperty == nil {
					c.ContentMeta.ContentMetaExtProperty = make([]*MetaExtProperty, 0)
				}
				c.ContentMeta.ContentMetaExtProperty = append(c.ContentMeta.ContentMetaExtProperty, extProp)
			}
		}

		// Remove duplicates with contentMeta
		for i := len(c.ItemMeta.Links) - 1; i >= 0; i-- {
			link := c.ItemMeta.Links[i]
			if link.Type == "x-geo/point" || link.Rel == "section" {
				c.ItemMeta.Links = append(c.ItemMeta.Links[:i], c.ItemMeta.Links[i+1:]...)
			}
		}
	}

	c.Concept = &Concept{}
	err = c.Concept.fromDoc(document, opts)
	if err != nil {
		return err
	}

	blk := doc.Block{
		Data: map[string]string{
			"registration": registration,
		},
	}
	data, err := transformDataToRaw(&blk, opts, MetaContext)
	if err != nil {
		return err
	}
	if registration != "" {
		var found bool
		for _, object := range c.Concept.MetaData {
			if object.Type == "x-im/event-details" {
				object.Data.Raw = data
				found = true
				break
			}
		}

		if !found {
			co := Object{
				ID:   "d1e25",
				Type: "x-im/event-details",
				Data: &Data{data},
			}
			c.Concept.MetaData = append(c.Concept.MetaData, co)
		}
	}

	return nil
}

func (c *ConceptItem) toDoc(document *doc.Document, opts *Options) error {
	document.UUID = c.GUID
	document.Properties = []doc.Property{}

	if c.RightsInfo != nil {
		c.addPropertyToDoc(document, "copyrightHolder", c.RightsInfo.CopyRightHolder.Name)
	}

	metaHasAddedInfo := false
	metaBlock := doc.Block{
		Type: "x-im/event",
		Data: make(map[string]string),
	}

	if c.ItemMeta != nil {
		err := c.ItemMeta.ToDoc(document, opts)
		if err != nil {
			return err
		}
		for _, prop := range c.ItemMeta.ItemMetaExtProperty {
			switch prop.Type {
			case "nrpdate:start":
				metaHasAddedInfo = true
				metaBlock.Data["start"] = prop.Value
				metaBlock.Data["dateGranularity"] = strings.ReplaceAll(prop.Why, "nrpwhy:", "")
			case "nrpdate:end":
				metaHasAddedInfo = true
				metaBlock.Data["end"] = prop.Value
				metaBlock.Data["dateGranularity"] = strings.ReplaceAll(prop.Why, "nrpwhy:", "")
			case "nrpdate:created":
				createdTime, err := time.Parse(time.RFC3339, prop.Value)
				if err != nil {
					return fmt.Errorf("ailed to parse created: %s", prop.Value)
				}
				document.Created = &createdTime
			case "nrpdate:modified":
				modifiedTime, err := time.Parse(time.RFC3339, prop.Value)
				if err != nil {
					return fmt.Errorf("ailed to parse modified: %s", prop.Value)
				}
				document.Modified = &modifiedTime
			}
		}
	}

	if c.ContentMeta != nil {
		err := c.ContentMeta.toDoc(document, opts)
		if err != nil {
			return err
		}

		if c.ContentMeta.Urgency != "" {
			metaHasAddedInfo = true
			metaBlock.Data["priority"] = c.ContentMeta.Urgency
		}
		if c.ContentMeta.Slugline != "" {
			metaHasAddedInfo = true
			metaBlock.Data["slug"] = c.ContentMeta.Slugline
		}
		if c.ContentMeta.Description != nil {
			for _, descr := range c.ContentMeta.Description {
				switch descr.Role {
				case "nrpdesc:intern":
					metaHasAddedInfo = true
					metaBlock.Data["description"] = descr.Text
				case "nrpdesc:extern":
					metaHasAddedInfo = true
					metaBlock.Data["publicDescription"] = descr.Text
				}
			}
			// Remove description from properties
			for i := len(document.Properties) - 1; i >= 0; i-- {
				switch document.Properties[i].Name {
				case "description":
					document.Properties = append(document.Properties[:i], document.Properties[i+1:]...)
				default:
				}
			}
		}
		// sector is added as property by contentmeta, make it a link
		// and remove the property
		for _, prop := range c.ContentMeta.ContentMetaExtProperty {
			if prop.Type == "nrp:sector" {
				link := doc.Block{
					URI:   fmt.Sprintf("nrp://section/%s", strings.ToLower(prop.Literal)),
					Title: prop.Literal,
					Rel:   "section",
					Value: prop.Value,
				}
				if document.Links == nil {
					document.Links = make([]doc.Block, 0)
				}
				document.Links = append(document.Links, link)
				// Remove sector from properties
				for i := len(document.Properties) - 1; i >= 0; i-- {
					switch document.Properties[i].Name {
					case "nrp:sector":
						document.Properties = append(document.Properties[:i], document.Properties[i+1:]...)
					default:
					}
				}
				break
			}
		}
	}

	if c.Concept != nil {
		err := c.Concept.toDoc(document, opts)
		if err != nil {
			return err
		}
		for _, object := range c.Concept.MetaData {
			if object.Type == "x-im/event-details" && object.Data != nil {
				data, blocks, err := transformDataFromRaw(object.Type, object.Data.Raw, opts, MetaContext)
				if err != nil {
					return err
				}
				if registration, ok := data["registration"]; ok {
					metaHasAddedInfo = true
					metaBlock.Data["registration"] = registration
				}
				dest := opts.GetElementDestination(object.Type)
				switch dest {
				case DestinationLink:
					metaBlock.Links = append(metaBlock.Links, blocks...)
				case DestinationMeta:
					metaBlock.Meta = append(metaBlock.Meta, blocks...)
				default:
					return fmt.Errorf("invalid destination configured: %s", dest)
				}
			}
		}
	}

	if c.ContentMeta != nil {
		var xtraProps = []string{
			"sector",
		}
		for i := len(document.Properties) - 1; i >= 0; i-- {
			for _, prop := range xtraProps {
				if document.Properties[i].Name == prop {
					document.Properties = append(document.Properties[:i], document.Properties[i+1:]...)
				}
			}
		}
	}

	if metaHasAddedInfo {
		document.Meta = append(document.Meta, metaBlock)
	}

	return nil
}

func (c *ConceptItem) addPropertyToDoc(document *doc.Document, name string, value string) {
	if value != "" {
		document.Properties = append(document.Properties, doc.Property{Name: name, Value: value})
	}
}
