package newsml

import (
	"encoding/xml"
	"fmt"
	"strings"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type IDFElement struct {
	XMLName   xml.Name `xml:"element"`
	ID        string   `xml:"id,attr,omitempty"`
	Type      string   `xml:"type,attr"`
	Format    string   `xml:"format,attr,omitempty"`
	Variation string   `xml:"variation,attr,omitempty"`
	Text      string   `xml:",innerxml"`
}

func (ie *IDFElement) fromBlock(block doc.Block, opts *Options) error {
	ie.XMLName = xml.Name{
		Space: "http://www.infomaker.se/idf/1.0",
		Local: "element",
	}

	elementType := toXMLType(block.Type, opts)
	ie.ID = block.ID
	ie.Type = elementType

	for key, value := range block.Data {
		switch key {
		case "format":
			ie.Format = value
		case "text":
			ie.Text = value
		case "variation":
			ie.Variation = value
		}
	}

	if strings.Contains(ie.Type, "ordered-list") && block.Content != nil {
		var innerXML strings.Builder
		for _, para := range block.Content {
			if listItem, ok := para.Data["text"]; ok {
				innerXML.WriteString(fmt.Sprintf("<list-item>%s</list-item>", listItem))
			}
		}
		ie.Text = innerXML.String()
	}

	return nil
}

func (ie *IDFElement) toBlock(block *doc.Block, opts *Options) error {
	block.Type = fromXMLType(ie.Type, opts)
	block.ID = ie.ID
	block.Data = make(map[string]string)
	if ie.Text != "" {
		if !strings.Contains(ie.Type, "ordered-list") {
			text, err := SanitizeHTML(ie.Text, opts)
			if err != nil {
				return err
			}
			block.Data["text"] = text
		} else {
			block.Content = buildList(ie.Text, opts)
		}
	}
	block.Data["format"] = "html"
	if ie.Variation != "" {
		block.Data["variation"] = ie.Variation
	}

	return nil
}
