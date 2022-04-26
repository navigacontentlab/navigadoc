package newsml

import (
	"encoding/xml"
	"fmt"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type IDFObjectBase struct {
	ID          string `xml:"id,attr,omitempty"`
	UUID        string `xml:"uuid,attr,omitempty"`
	URI         string `xml:"uri,attr,omitempty"`
	URL         string `xml:"url,attr,omitempty"`
	Type        string `xml:"type,attr,omitempty"`
	Title       string `xml:"title,attr,omitempty"`
	Name        string `xml:"name,attr,omitempty"`
	Value       string `xml:"value,attr,omitempty"`
	Rel         string `xml:"rel,attr,omitempty"`
	ContentType string `xml:"contenttype,attr,omitempty"`
}

// IDFObject ...
type IDFObject struct {
	IDFObjectBase
	Content       IDFObjects     `xml:"content,omitempty"`
	Data          *Data          `xml:"data,omitempty"`
	IDFLinks      IDFLinks       `xml:"links,omitempty"`
	IDFProperties *IDFProperties `xml:"properties,omitempty"`
	IDFMeta       IDFObjects     `xml:"meta,omitempty"`
}

type IDFLinks struct {
	IDFLink []IDFObject `xml:"link,omitempty"`
}

type IDFObjects []IDFObject

type internalIDFObjects struct {
	XMLName xml.Name
	Objects []IDFObject `xml:"object"`
}

func (o IDFObjects) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(o) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for i := range o {
		element := xml.StartElement{
			Name: xml.Name{
				Local: "object",
			},
		}
		if err := e.EncodeElement(o[i], element); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (o *IDFObjects) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var internal internalIDFObjects

	if err := d.DecodeElement(&internal, &start); err != nil {
		return err
	}

	*o = (*o)[:]

	for i := range internal.Objects {
		*o = append(*o, internal.Objects[i])
	}

	return nil
}

type IDFProperties struct {
	IDFProperty []*IDFProperty `xml:"property,omitempty"`
}

type IDFProperty struct {
	Name  string `xml:"name,attr,omitempty"`
	Value string `xml:"value,attr,omitempty"`
	Data  *Data  `xml:"parameters,omitempty"`
}

// Build a doc.Block from an newsml.IDFObject
func (io *IDFObject) toBlock(block *doc.Block, opts *Options, context ContextType) error {
	block.ID = io.ID
	block.URL = io.URL
	block.URI = io.URI
	block.Title = io.Title
	block.UUID = io.UUID
	block.Type = fromXMLType(io.Type, opts)
	block.Data = make(map[string]string)

	if io.Data != nil {
		raw := io.Data.Raw

		content, err := getIDFContentFromRaw(raw, opts)
		if err != nil {
			return err
		}

		block.Content = content

		xmldata, meta, err := transformDataFromRaw(block.Type, raw, opts, context)
		if err != nil {
			return err
		}
		if val, ok := xmldata["text"]; ok && val == "" {
			delete(xmldata, "text")
		}
		if val, ok := xmldata["format"]; ok && val == "" {
			delete(xmldata, "format")
		}
		if val, ok := xmldata["variation"]; ok && val == "" {
			delete(xmldata, "variation")
		}
		if len(xmldata) > 0 {
			block.Data = xmldata
		}

		dest := opts.GetElementDestination(block.Type)
		switch dest {
		case DestinationLink:
			block.Links = append(block.Links, meta...)
		case DestinationMeta:
			block.Meta = append(block.Meta, meta...)
		default:
			return fmt.Errorf("invalid destination configured: %s", dest)
		}
	}

	if len(io.IDFLinks.IDFLink) > 0 {
		result, err := io.buildBlocksFromIDFObjects(io.IDFLinks.IDFLink, opts, context)
		if err != nil {
			return err
		}
		block.Links = result
	}

	if io.IDFProperties != nil && len(io.IDFProperties.IDFProperty) > 0 {
		meta := doc.Block{
			Type: "x-im/generic-property",
			Data: make(map[string]string),
		}
		for _, p := range io.IDFProperties.IDFProperty {
			meta.Data[p.Name] = p.Value
		}
		block.Meta = append(block.Meta, meta)
	}

	if io.Content != nil && len(io.Content) > 0 {
		result, err := io.buildBlocksFromIDFObjects(io.Content, opts, context)
		if err != nil {
			return err
		}
		block.Content = result
	}

	if io.IDFMeta != nil && len(io.IDFMeta) > 0 {
		result, err := io.buildBlocksFromIDFObjects(io.IDFMeta, opts, context)
		if err != nil {
			return err
		}
		block.Meta = append(block.Meta, result...)
	}

	return nil
}

// Build a newsml.IDFObject from a doc.Block
func (io *IDFObject) fromBlock(block doc.Block, opts *Options, context ContextType) error {
	io.IDFObjectBase = IDFObjectBase{
		Type:        block.Type,
		ID:          block.ID,
		UUID:        block.UUID,
		Title:       block.Title,
		URL:         block.URL,
		URI:         block.URI,
		Rel:         block.Rel,
		Name:        block.Name,
		Value:       block.Value,
		ContentType: block.ContentType,
	}

	rawData := ""

	// Check if object has Content
	if block.Content != nil {
		rawContent, err := extractContentToIDF(block, opts)
		if err != nil {
			return err
		}
		rawData += rawContent
	}

	if block.Data != nil {
		blockXMLData, err := transformDataToRaw(&block, opts, context)
		if err != nil {
			return err
		}
		rawData += blockXMLData
	}

	if rawData != "" {
		io.Data = &Data{
			Raw: rawData,
		}
	}
	if len(block.Links) > 0 {
		for _, link := range block.Links {
			xmlLink, err := io.buildIDFObjectFromBlock(link, opts, context)
			if err != nil {
				return err
			}
			io.IDFLinks.IDFLink = append(io.IDFLinks.IDFLink, xmlLink)
		}
	}

	if block.Meta != nil && len(block.Meta) > 0 {
		for _, meta := range block.Meta {
			if meta.Type == "x-im/generic-property" {
				if io.IDFProperties == nil {
					io.IDFProperties = &IDFProperties{
						IDFProperty: make([]*IDFProperty, 0),
					}
					for key, val := range meta.Data {
						prop := &IDFProperty{
							Name:  key,
							Value: val,
						}
						io.IDFProperties.IDFProperty = append(io.IDFProperties.IDFProperty, prop)
					}
				}
			} else {
				xmlMeta, err := io.buildIDFObjectFromBlock(meta, opts, context)
				if err != nil {
					return err
				}
				io.IDFMeta = append(io.IDFMeta, xmlMeta)
			}
		}
	}

	return nil
}

// Recursive function to build doc blocks from IDFObjects
func (io *IDFObject) buildBlocksFromIDFObjects(idfObjects IDFObjects, opts *Options, context ContextType) ([]doc.Block, error) {
	blocksOut := make([]doc.Block, 0)

	for _, idfObject := range idfObjects {
		obj := doc.Block{
			ID:          idfObject.ID,
			Title:       idfObject.Title,
			Type:        idfObject.Type,
			URI:         idfObject.URI,
			URL:         idfObject.URL,
			UUID:        idfObject.UUID,
			Rel:         idfObject.Rel,
			Name:        idfObject.Name,
			ContentType: idfObject.ContentType,
			Value:       idfObject.Value,
		}

		if idfObject.Data != nil {
			data, blocks, err := transformDataFromRaw(obj.Type, idfObject.Data.Raw, opts, context)
			if err != nil {
				return nil, err
			}

			obj.Data = data
			dest := opts.GetElementDestination(obj.Type)
			switch dest {
			case DestinationLink:
				obj.Links = append(obj.Links, blocks...)
			case DestinationMeta:
				obj.Meta = append(obj.Meta, blocks...)
			default:
				return nil, fmt.Errorf("invalid destination configured: %s", dest)
			}
		}

		if idfObject.IDFProperties != nil {
			meta := doc.Block{
				Type: "x-im/generic-property",
			}
			for _, prop := range idfObject.IDFProperties.IDFProperty {
				dp := &doc.Block{
					Name:  prop.Name,
					Value: prop.Value,
				}
				if prop.Data != nil {
					data, m, err := transformDataFromRaw(meta.Type, prop.Data.Raw, opts, context)
					if err != nil {
						return nil, err
					}
					dp.Data = data
					dest := opts.GetElementDestination(meta.Type)
					switch dest {
					case DestinationLink:
						dp.Links = append(dp.Links, m...)
					case DestinationMeta:
						dp.Meta = append(dp.Meta, m...)
					default:
						return nil, fmt.Errorf("invalid destination configured: %s", dest)
					}
				}
			}
			obj.Meta = append(obj.Meta, meta)
		}

		if len(idfObject.IDFLinks.IDFLink) > 0 {
			blocks, err := io.buildBlocksFromIDFObjects(idfObject.IDFLinks.IDFLink, opts, context)
			if err != nil {
				return nil, err
			}
			obj.Links = blocks
		}

		if idfObject.Content != nil && len(idfObject.Content) > 0 {
			blocks, err := io.buildBlocksFromIDFObjects(idfObject.Content, opts, context)
			if err != nil {
				return nil, err
			}
			obj.Content = blocks
		}

		if idfObject.IDFMeta != nil && len(idfObject.IDFMeta) > 0 {
			blocks, err := io.buildBlocksFromIDFObjects(idfObject.IDFMeta, opts, context)
			if err != nil {
				return nil, err
			}
			obj.Meta = blocks
		}

		blocksOut = append(blocksOut, obj)
	}

	return blocksOut, nil
}

func (io *IDFObject) buildIDFObjectFromBlock(block doc.Block, opts *Options, context ContextType) (IDFObject, error) {
	idfObject := IDFObject{
		IDFObjectBase: IDFObjectBase{
			Type:        block.Type,
			ID:          block.ID,
			UUID:        block.UUID,
			Title:       block.Title,
			URL:         block.URL,
			URI:         block.URI,
			Rel:         block.Rel,
			Name:        block.Name,
			Value:       block.Value,
			ContentType: block.ContentType,
		},
	}

	if block.Data != nil {
		idfObject.Data = &Data{}
		s, err := transformDataToRaw(&block, opts, context)
		if err != nil {
			return IDFObject{}, err
		}
		idfObject.Data.Raw = s
	}

	for _, link := range block.Links {
		obj, err := io.buildChildIDFObject(link, opts, context)
		if err != nil {
			return IDFObject{}, err
		}
		idfObject.IDFLinks.IDFLink = append(idfObject.IDFLinks.IDFLink, obj)
	}

	for _, content := range block.Content {
		obj, err := io.buildChildIDFObject(content, opts, context)
		if err != nil {
			return IDFObject{}, err
		}
		idfObject.Content = append(idfObject.Content, obj)
	}

	for _, meta := range block.Meta {
		if meta.Type == "x-im/flags" {
			continue
		}
		if meta.Type == "x-im/generic-property" {
			idfObject.IDFProperties = &IDFProperties{}
			for key, val := range meta.Data {
				idfProperty := &IDFProperty{
					Name:  key,
					Value: val,
				}
				idfObject.IDFProperties.IDFProperty = append(idfObject.IDFProperties.IDFProperty, idfProperty)
			}
		} else {
			obj, err := io.buildChildIDFObject(meta, opts, context)
			if err != nil {
				return IDFObject{}, err
			}
			idfObject.IDFMeta = append(idfObject.IDFMeta, obj)
		}
	}

	return idfObject, nil
}

// Recursive function for building children of IDFObject
func (io *IDFObject) buildChildIDFObject(block doc.Block, opts *Options, context ContextType) (IDFObject, error) {
	idfObject := IDFObject{
		IDFObjectBase: IDFObjectBase{
			Type:        block.Type,
			ID:          block.ID,
			UUID:        block.UUID,
			Title:       block.Title,
			URL:         block.URL,
			URI:         block.URI,
			Rel:         block.Rel,
			Name:        block.Name,
			Value:       block.Value,
			ContentType: block.ContentType,
		},
	}

	if block.Data != nil {
		idfObject.Data = &Data{}

		s, err := transformDataToRaw(&block, opts, context)
		if err != nil {
			return idfObject, err
		}

		idfObject.Data.Raw = s
	}

	for _, link := range block.Links {
		obj, err := io.buildChildIDFObject(link, opts, context)
		if err != nil {
			return IDFObject{}, err
		}
		idfObject.IDFLinks.IDFLink = append(idfObject.IDFLinks.IDFLink, obj)
	}

	for _, content := range block.Content {
		obj, err := io.buildChildIDFObject(content, opts, context)
		if err != nil {
			return IDFObject{}, err
		}
		idfObject.Content = append(idfObject.Content, obj)
	}

	for _, meta := range block.Meta {
		if meta.Type == "x-im/flags" {
			continue
		} else if meta.Type == "x-im/generic-property" {
			idfObject.IDFProperties = &IDFProperties{}
			for key, val := range meta.Data {
				idfProperty := &IDFProperty{
					Name:  key,
					Value: val,
				}
				idfObject.IDFProperties.IDFProperty = append(idfObject.IDFProperties.IDFProperty, idfProperty)
			}
			continue
		}
		obj, err := io.buildChildIDFObject(meta, opts, context)
		if err != nil {
			return IDFObject{}, err
		}
		idfObject.IDFMeta = append(idfObject.IDFMeta, obj)
	}

	return idfObject, nil
}
