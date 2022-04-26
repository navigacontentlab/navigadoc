package newsml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type Concept struct {
	Text       string        `xml:",chardata"`
	ConceptID  *ConceptID    `xml:"conceptId"`
	Type       *ConceptType  `xml:"type"`
	Name       string        `xml:"name"`
	Definition []*Definition `xml:"definition"`
	MetaData   NSObjects     `xml:"metadata"`
}

type ConceptID struct {
	URI      string `xml:"uri,attr"`
	Creator  string `xml:"creator,attr,omitempty"`
	Created  string `xml:"created,attr,omitempty"`
	Modified string `xml:"modified,attr,omitempty"`
	QCode    string `xml:"qcode,attr,omitempty"`
	Text     string `xml:",chardata"`
}

type ConceptType struct {
	Text  string `xml:",chardata"`
	Qcode string `xml:"qcode,attr"`
}

type Definition struct {
	Text string `xml:",innerxml"`
	Role string `xml:"role,attr"`
}

func (c *Concept) toDoc(document *doc.Document, opts *Options) error {
	document.Title = c.Name
	document.URI = c.ConceptID.URI

	for i, object := range c.MetaData {
		block, err := blockFromObject(object, opts, MetaContext)
		if err != nil {
			return fmt.Errorf("failed to convert metadata object %d: %w", i, err)
		}

		if block.Type == "x-im/event-details" {
			delete(block.Data, "registration")
		}

		document.Meta = append(document.Meta, block)
	}

	for _, def := range c.Definition {
		noTags := bluemonday.StrictPolicy().Sanitize(def.Text)
		text := html.UnescapeString(noTags)

		document.Properties = append(document.Properties,
			doc.Property{Name: "definition", Value: text,
				Parameters: map[string]string{"role": def.Role}})
	}

	document.Properties = append(document.Properties,
		doc.Property{Name: "conceptid", Value: c.ConceptID.Text,
			Parameters: map[string]string{
				"uri":      c.ConceptID.URI,
				"creator":  c.ConceptID.Creator,
				"created":  c.ConceptID.Created,
				"qcode":    c.ConceptID.QCode,
				"modified": c.ConceptID.Modified,
			}})

	return nil
}

func (c *Concept) fromDoc(document *doc.Document, opts *Options) error {
	if document == nil {
		return errors.New("nil document")
	}

	c.Name = document.Title
	c.ConceptID = &ConceptID{
		URI: document.URI,
	}

	err := c.addFromProperties(document)
	if err != nil {
		return err
	}

	c.addConceptMetadata(document, opts)

	qcode, err := concepttypeToQCode(document.Type, opts)
	if err != nil {
		return err
	}

	c.Type = &ConceptType{Qcode: qcode}
	return nil
}

func (c *Concept) addConceptMetadata(document *doc.Document, opts *Options) {
	for _, prop := range document.Properties {
		if prop.Name == "concepttypes" {
			for id, ctype := range prop.Parameters {
				for i := range document.Meta {
					meta := document.Meta[i]
					if meta.ID == id {
						object := Object{
							ID:   id,
							Type: ctype,
						}
						s, err := transformDataToRaw(&meta, opts, MetaContext)
						if err != nil {
							return
						}
						object.Data = &Data{Raw: s}
						c.MetaData = append(c.MetaData, object)
						break
					}
				}
			}
		}
	}

	// TODO External Source
	// TODO How to distinguish /contentMeta/metadata from /concept/metadata?
	for i := range document.Meta {
		block := document.Meta[i]
		if isConceptObjectType(block.Type, opts) {
			object := Object{
				ID:   block.ID,
				Type: block.Type,
			}
			s, err := transformDataToRaw(&block, opts, MetaContext)
			if err != nil {
				return
			}
			object.Data = &Data{Raw: s}
			c.MetaData = append(c.MetaData, object)
		}
	}
}

func (c *Concept) addFromProperties(document *doc.Document) error {
	for _, prop := range document.Properties {
		switch strings.ToLower(prop.Name) {
		case "definition":
			role := prop.Parameters["role"]
			if role == "" {
				break
			}

			var buf bytes.Buffer

			err := xml.EscapeText(&buf, []byte(prop.Value))
			if err != nil {
				return fmt.Errorf("failed to XML escape definition value: %v", err)
			}

			c.Definition = append(c.Definition, &Definition{Role: role, Text: buf.String()})
		case "conceptid":
			conceptURI := prop.Parameters["uri"]

			// The document.URI is the authoritative
			// source of the URI, only populate from the
			// property if it was missing.
			if c.ConceptID.URI == "" {
				c.ConceptID.URI = conceptURI
			}

			c.ConceptID.Creator = prop.Parameters["creator"]
			c.ConceptID.Created = prop.Parameters["created"]
			c.ConceptID.Modified = prop.Parameters["modified"]
			c.ConceptID.QCode = prop.Parameters["qcode"]
			c.Type = &ConceptType{Qcode: c.ConceptID.QCode}
		}
	}
	return nil
}
