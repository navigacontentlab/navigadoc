package newsml

// contentmeta och itemmeta är samma för alla

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"bitbucket.org/infomaker/doc-format/v2/doc"
	"github.com/Infomaker/etree"
)

// ContentMeta represents <newsItem><contentMeta>
type ContentMeta struct {
	ContentCreated         string             `xml:"contentCreated,omitempty"`
	ContentModified        string             `xml:"contentModified,omitempty"`
	InfoSource             *InfoSource        `xml:"infoSource,omitempty"`
	AltID                  string             `xml:"altId,omitempty"`
	Slugline               string             `xml:"slugline,omitempty"`
	By                     string             `xml:"by,omitempty"`
	Headline               string             `xml:"headline,omitempty"`
	Description            []*Description     `xml:"description,omitempty"`
	Creator                *Creator           `xml:"creator,omitempty"`
	Language               *Language          `xml:"language,omitempty"`
	Urgency                string             `xml:"urgency,omitempty"`
	ContentMetaExtProperty []*MetaExtProperty `xml:"contentMetaExtProperty,omitempty"`
	MetaData               NSObjects          `xml:"metadata"`
	Links                  NSLinks            `xml:"links"`
}

type Objects []Object

func (l Objects) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(l) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for i := range l {
		element := xml.StartElement{
			Name: xml.Name{
				Local: "object",
			},
		}
		if err := e.EncodeElement(l[i], element); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (l *Objects) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var internal internalObjects

	if err := d.DecodeElement(&internal, &start); err != nil {
		return err
	}

	*l = (*l)[:]

	for i := range internal.Objects {
		*l = append(*l, internal.Objects[i])
	}

	return nil
}

type NSObjects []Object

type internalObjects struct {
	XMLName xml.Name
	Objects []Object `xml:"object"`
}

func (l NSObjects) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(l) == 0 {
		return nil
	}

	start.Name.Space = "http://www.infomaker.se/newsml/1.0"

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for i := range l {
		element := xml.StartElement{
			Name: xml.Name{
				Local: "object",
			},
		}
		if err := e.EncodeElement(l[i], element); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (l *NSObjects) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var internal internalObjects

	if err := d.DecodeElement(&internal, &start); err != nil {
		return err
	}

	*l = (*l)[:]

	for i := range internal.Objects {
		*l = append(*l, internal.Objects[i])
	}

	return nil
}

func (cm *ContentMeta) fromDoc(document *doc.Document, opts *Options) error {
	cm.addDates(document)
	cm.addProperties(document.Properties, opts)

	for _, block := range document.Meta {
		err := cm.addMetaBlock(document.Type, block, opts)
		if err != nil {
			return err
		}
	}

	err := cm.addLinks(document.Links, opts)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ContentMeta) addDates(document *doc.Document) {
	if document.Created != nil {
		cm.ContentCreated = document.Created.Format(time.RFC3339Nano)
	}
	if document.Modified != nil {
		cm.ContentModified = document.Modified.Format(time.RFC3339Nano)
	}
}

func (cm *ContentMeta) addProperties(properties []doc.Property, opts *Options) {
	for _, property := range properties {
		switch strings.ToLower(property.Name) {
		case "infosource":
			cm.InfoSource = &InfoSource{Text: property.Value}
			if l, ok := property.Parameters["literal"]; ok {
				cm.InfoSource.Literal = l
			}
		case "altid":
			cm.AltID = property.Value
		case "slugline":
			cm.Slugline = property.Value
		case "headline":
			cm.Headline = property.Value
		case "by":
			cm.By = property.Value
		case "description":
			if cm.Description == nil {
				cm.Description = make([]*Description, 0)
			}
			descr := &Description{
				Text: property.Value,
			}
			if role, ok := property.Parameters["role"]; ok {
				descr.Role = role
			}
			cm.Description = append(cm.Description, descr)
		case "creator":
			cm.Creator = &Creator{Text: property.Value}
			if literal, ok := property.Parameters["literal"]; ok {
				cm.Creator.Literal = literal
			}
		case "urgency":
			cm.Urgency = property.Value
		case "language":
			cm.addLanguage(property.Value)
		case "nrp:sector":
			if cm.ContentMetaExtProperty == nil {
				cm.ContentMetaExtProperty = make([]*MetaExtProperty, 0)
			}
			cmxp := &MetaExtProperty{
				Type:  "nrp:sector",
				Value: property.Value,
			}
			if literal, ok := property.Parameters["literal"]; ok {
				cmxp.Literal = literal
			}
			cm.ContentMetaExtProperty = append(cm.ContentMetaExtProperty, cmxp)
		default:
			if hasPropertyExceptionType(property.Name, "contentmeta", opts) {
				cm.ContentMetaExtProperty = append(cm.ContentMetaExtProperty,
					&MetaExtProperty{Type: property.Name, Value: property.Value})
			}
		}
	}
}

func (cm *ContentMeta) addLinks(links []doc.Block, opts *Options) error {
	for i, link := range links {
		allowedTypes := getAllowedContentMetaLinks(opts)

		if !allowedTypes[link.Type] {
			continue
		}

		if link.Type == "x-im/category" && isItemMetaCategoryRel(link.Rel, opts) {
			continue
		}

		cmLink, err := buildXMLLink(link, opts)
		if err != nil {
			return fmt.Errorf("failed to convert link %d: %w", i, err)
		}

		cm.Links = append(cm.Links, cmLink)
	}

	return nil
}

func transformElementToBlock(element *etree.Element, dataElement *DataElement, opts *Options, context ContextType) (doc.Block, error) {
	block := doc.Block{
		Name: dataElement.Name,
		Type: dataElement.Type,
		Rel:  dataElement.Rel,
	}
	var innerXML bytes.Buffer
	var writeSettings etree.WriteSettings

	for _, child := range element.Child {
		child.WriteTo(&innerXML, &writeSettings)
	}

	xmldata, meta, err := transformDataFromRaw(dataElement.Type, innerXML.String(), opts, context)
	if err != nil {
		return block, err
	}

	for _, attr := range element.Attr {
		xmldata[attr.Key] = attr.Value
	}

	block.Data = xmldata
	dest := opts.GetElementDestination(block.Type)
	switch dest {
	case DestinationLink:
		block.Links = append(block.Links, meta...)
	case DestinationMeta:
		block.Meta = append(block.Meta, meta...)
	default:
		return block, fmt.Errorf("invalid destination configured: %s", dest)
	}

	return block, nil
}

func transformDataFromRaw(objectType string, raw string, opts *Options, context ContextType) (map[string]string, []doc.Block, error) {
	var dc *DataElementConversion
	var blocks = make([]doc.Block, 0)

	if dc = opts.GetDataConversionForType(objectType); dc != nil {
		switch dc.Datatype {
		case DataConversionAsIDF:
			idf, err := getIDFContentFromRaw("<text format=\"idf\">"+raw+"</text>", opts)
			if err != nil {
				return nil, nil, err
			}
			blocks = append(blocks, idf...)
			return nil, blocks, nil
		case DataConversionAsString:
			blocks = append(blocks, doc.Block{
				Data: map[string]string{
					"format": "xml",
					"text":   raw,
				},
			})
			return nil, blocks, nil
		}
	}

	tree := etree.NewDocument()
	err := tree.ReadFromString(raw)
	if err != nil {
		return nil, nil, err
	}

	var innerXML bytes.Buffer
	var writeSettings etree.WriteSettings
	data := map[string]string{}

	option := opts.BlockOptions(objectType, context)

	for _, element := range tree.ChildElements() {
		if element.Tag == "" || element.SelectAttrValue("format", "") == "idf" {
			continue
		}

		if element.Tag == "flags" && option.HasFlags {
			flagsBlock := doc.Block{
				Type: "x-im/flags",
				Data: map[string]string{},
			}

			for _, childFlag := range element.ChildElements() {
				flagsBlock.Data[childFlag.Text()] = "true"
			}

			blocks = append(blocks, flagsBlock)

			continue
		}

		if dc != nil && dc.Datatype == DataConversionAsXML {
			if me := opts.GetDataConversionElement(objectType, element.Tag); me != nil {
				block, err := transformElementToBlock(element, me, opts, context)
				if err != nil {
					return nil, nil, err
				}
				blocks = append(blocks, block)
				continue
			}
		}

		attr := option.Attribute(element.Tag)

		if attr.ValueHandling == ValueAsCData {
			v, ok := getCDATA(element)
			if ok {
				data[element.Tag] = v
				continue
			}
		}

		switch attr.ValueHandling {
		case ValueAsText:
			data[element.Tag] = element.Text()
		case ValueAsAttribute:
			name := attr.ValueAttribute
			if name == "" {
				name = DefaultValueAttribute
			}
			data[element.Tag] = element.SelectAttrValue(name, "")
		case ValueAsXML, ValueAsCData: // ...we let CDATA fall through to XML.
			innerXML.Reset()

			for _, child := range element.Child {
				child.WriteTo(&innerXML, &writeSettings)
			}

			if objectType != "x-im/htmlembed" {
				data[element.Tag], err = SanitizeHTML(innerXML.String(), opts)
				if err != nil {
					return nil, nil, err
				}
			} else {
				data[element.Tag] = innerXML.String()
			}
		default:
			return nil, nil, fmt.Errorf("unknown value handling %q configured for %q",
				attr.ValueHandling, element.Tag)
		}
	}

	return data, blocks, nil
}

func getCDATA(e *etree.Element) (string, bool) {
	if len(e.Child) == 0 {
		return "", false
	}

	buf := strings.Builder{}
	for i := range e.Child {
		switch c := e.Child[i].(type) {
		case *etree.CharData:
			if c.IsWhitespace() {
				continue
			}
			buf.WriteString(c.Data)
		case *etree.Element:
			return buf.String(), buf.Len() > 0
		}
	}

	return buf.String(), true
}

func checkDataKey(key string) error {
	if key == "" {
		return errors.New("keys may not be empty")
	}

	for i, c := range key {
		if c == '_' || c == '-' {
			continue
		}
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			continue
		}
		if c >= '0' && c <= '9' {
			if i == 0 {
				return errors.New("keys may not start with numbers")
			}
			continue
		}

		return fmt.Errorf("keys may not contain %q", string(c))
	}

	return nil
}

func transformDataToRaw(block *doc.Block, opts *Options, context ContextType) (string, error) {
	var dc *DataElementConversion

	destination, blocks, err := getBlocksForDestination(block, opts)
	if err != nil {
		return "", err
	}

	if dc = opts.GetDataConversionForType(block.Type); dc != nil && dc.Datatype != DataConversionAsXML {
		return doDataConversion(block, opts, dc, blocks, destination)
	}

	data := block.Data
	blockType := block.Type

	tree := etree.NewDocument()

	if len(blocks) > 0 {
		for _, b := range blocks {
			if element := opts.GetDataConversionElement(b.Type, b.Name); element != nil {
				el := tree.CreateElement(element.Name)
				for key, value := range b.Data {
					attr := opts.Option[blockType].Attribute(key)
					switch attr.ValueHandling {
					case ValueAsAttribute:
						el.CreateAttr(key, value)
					default:
						newEl := tree.CreateElement(key)
						el.AddChild(newEl)
						newEl.SetText(value)
					}
				}
			}
		}
		if destination == DestinationLink {
			block.Links = block.Links[:0]
		}
	}

	var keys = make([]string, 0)
	for key := range data {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	option := opts.BlockOptions(blockType, context)

	for _, key := range keys {
		if err := checkDataKey(key); err != nil {
			return "", fmt.Errorf("invalid data key %q: %w", key, err)
		}

		attr := option.Attribute(key)

		switch attr.ValueHandling {
		case ValueAsCData:
			el := tree.CreateElement(key)
			el.CreateCData(data[key])
		case ValueAsText:
			el := tree.CreateElement(key)
			el.SetText(data[key])
		case ValueAsAttribute:
			name := attr.ValueAttribute
			if name == "" {
				name = DefaultValueAttribute
			}

			el := tree.CreateElement(key)
			el.CreateAttr(name, data[key])
		case ValueAsXML:
			innerXML := etree.NewDocument()
			err := innerXML.ReadFromString(fmt.Sprintf("<%s>%s</%s>", key, data[key], key))
			if err != nil {
				return "", fmt.Errorf("invalid XML in value for %q: %w", key, err)
			}

			if innerXML.Root() == nil {
				return "", fmt.Errorf("invalid XML in value for %q", key)
			}

			tree.AddChild(innerXML.Root().Copy())
		default:
			return "", fmt.Errorf("unknown value handling %q configured for %q",
				attr.ValueHandling, key)
		}
	}

	if option.HasFlags && block.Meta != nil {
		for _, meta := range block.Meta {
			if meta.Type != "x-im/flags" {
				continue
			}

			flags := tree.CreateElement("flags")

			for key, enabled := range meta.Data {
				if enabled != "true" {
					continue
				}

				flag := tree.CreateElement("flag")
				flag.SetText(key)
				flags.AddChild(flag)
			}

			tree.AddChild(flags)
		}
	}

	s, err := tree.WriteToString()
	if err != nil {
		return "", err
	}

	return s, nil
}

func (cm *ContentMeta) addMetaBlock(docType string, block doc.Block, opts *Options) error {
	// already taken care off elsewhere
	if block.Type == "x-im/note" || isItemMetaObjectType(docType, block.Type, opts) || isConceptObjectType(block.Type, opts) {
		return nil
	}

	object, err := objectFromBlock(block, opts, MetaContext)
	if err != nil {
		return err
	}

	cm.MetaData = append(cm.MetaData, object)

	return nil
}

func objectFromBlock(block doc.Block, opts *Options, context ContextType) (Object, error) {
	object := Object{
		ID:    block.ID,
		Type:  block.Type,
		Title: block.Title,
		UUID:  block.UUID,
	}

	rawData := ""

	if block.Content != nil {
		rawContent, err := extractContentToIDF(block, opts)
		if err != nil {
			return Object{}, err
		}
		rawData += rawContent
	}

	if block.Data != nil {
		if object.Data == nil {
			object.Data = &Data{}
		}

		s, err := transformDataToRaw(&block, opts, context)
		if err != nil {
			return Object{}, err
		}
		rawData += s
	}

	if rawData != "" {
		object.Data = &Data{
			Raw: rawData,
		}
	}

	for _, link := range block.Links {
		if link.Type == "x-im/imchn" {
			continue
		}
		xmlLink, err := buildXMLLink(link, opts)
		if err != nil {
			return Object{}, err
		}
		object.Links = append(object.Links, xmlLink)
	}

	for i, meta := range block.Meta {
		if element := opts.GetDataConversionElement(meta.Type, meta.Name); element != nil {
			continue
		}
		if meta.Type == "x-im/flags" {
			continue
		}
		if meta.Type == "x-im/generic-property" && meta.Data != nil {
			if object.Properties == nil {
				object.Properties = &Properties{
					Property: make([]*Property, 0),
				}
			}
			for key, val := range meta.Data {
				object.Properties.Property = append(object.Properties.Property, &Property{
					Name:  key,
					Value: val,
				})
			}
			continue
		}
		child, err := objectFromBlock(meta, opts, context)
		if err != nil {
			return Object{}, fmt.Errorf("failed to convert meta block %d: %w", i, err)
		}

		object.Meta = append(object.Meta, child)
	}

	return object, nil
}

func (cm *ContentMeta) addLanguage(language string) {
	if language != "" {
		cm.Language = &Language{
			Tag: language,
		}
	}
}

func (cm *ContentMeta) toDoc(document *doc.Document, opts *Options) error {
	if cm.AltID != "" {
		cm.addPropertyToDoc(document, "altId", cm.AltID)
	}
	if cm.Slugline != "" {
		cm.addPropertyToDoc(document, "slugline", cm.Slugline)
	}
	if cm.Headline != "" {
		cm.addPropertyToDoc(document, "headline", cm.Headline)
	}
	if cm.By != "" {
		cm.addPropertyToDoc(document, "by", cm.By)
	}
	if cm.Urgency != "" {
		cm.addPropertyToDoc(document, "urgency", cm.Urgency)
	}
	if cm.Language != nil {
		cm.addPropertyToDoc(document, "language", cm.Language.Tag)
	}
	if cm.Description != nil {
		for _, descr := range cm.Description {
			cm.addDescriptionToDoc(document, descr)
		}
	}
	cm.addInfoSourceToDoc(document, cm.InfoSource)
	if cm.Creator != nil {
		cm.addCreatorToDoc(document, cm.Creator)
	}
	err := cm.addMetaToDoc(document, opts)
	if err != nil {
		return err
	}
	err = cm.addLinksToDoc(document, opts)
	if err != nil {
		return err
	}

	for _, extProp := range cm.ContentMetaExtProperty {
		cm.addExtPropertyToDoc(document, extProp)
	}

	if document.Created == nil {
		contentCreated, err := time.Parse(time.RFC3339Nano, cm.ContentCreated)
		// ContentCreated may be empty, so don't return error
		if err == nil {
			document.Created = &contentCreated
		}
	}

	return nil
}

func (cm *ContentMeta) addExtPropertyToDoc(document *doc.Document, extprop *MetaExtProperty) {
	cm.addPropertyToDoc(document, extprop.Type, extprop.Value)

	if extprop.Literal != "" {
		params := document.Properties[len(document.Properties)-1].Parameters
		if params == nil {
			document.Properties[len(document.Properties)-1].Parameters = map[string]string{}
		}
		if extprop.Value == "" {
			// Property was not added to document
			prop := doc.Property{
				Name:       extprop.Type,
				Value:      extprop.Value,
				Parameters: map[string]string{},
			}
			prop.Parameters["literal"] = extprop.Literal
			document.Properties = append(document.Properties, prop)
		} else {
			document.Properties[len(document.Properties)-1].Parameters["literal"] = extprop.Literal
		}
	}
}

func (cm *ContentMeta) addMetaToDoc(document *doc.Document, opts *Options) error {
	for i, object := range cm.MetaData {
		block, err := blockFromObject(object, opts, MetaContext)
		if err != nil {
			return fmt.Errorf("failed to convert metadata object %d: %w", i, err)
		}

		document.Meta = append(document.Meta, block)
	}
	return nil
}

func blockFromObject(object Object, opts *Options, context ContextType) (doc.Block, error) {
	block := doc.Block{
		Type:  object.Type,
		ID:    object.ID,
		Title: object.Title,
		UUID:  object.UUID,
	}

	if len(object.Links) != 0 {
		err := buildDocLinks(&block, object.Links, opts, context)
		if err != nil {
			return doc.Block{}, err
		}
	}

	var blocks []doc.Block
	var xmldata map[string]string
	if object.Data != nil {
		dc := opts.GetDataConversionForType(block.Type)
		if dc != nil && dc.Datatype == DataConversionAsString {
			block.Data = map[string]string{
				"format": "xml",
				"text":   object.Data.Raw,
			}
			blocks = append(blocks, block)
		} else {
			content, err := getIDFContentFromRaw(object.Data.Raw, opts)
			if err != nil {
				return doc.Block{}, err
			}
			block.Content = content
			xmldata, blocks, err = transformDataFromRaw(block.Type, object.Data.Raw, opts, context)
			if err != nil {
				return doc.Block{}, err
			}
			block.Data = xmldata
		}
		dest := opts.GetElementDestination(block.Type)
		switch dest {
		case DestinationLink:
			block.Links = append(block.Links, blocks...)
		case DestinationMeta:
			block.Meta = append(block.Meta, blocks...)
		default:
			return doc.Block{}, fmt.Errorf("invalid destination configured: %s", dest)
		}
	}

	if object.Properties != nil {
		meta := doc.Block{
			Type: "x-im/generic-property",
			Data: make(map[string]string),
		}
		for _, prop := range object.Properties.Property {
			meta.Data[prop.Name] = prop.Value
		}
		block.Meta = append(block.Meta, meta)
	}

	for i := range object.Meta {
		child, err := blockFromObject(object.Meta[i], opts, context)
		if err != nil {
			return doc.Block{}, fmt.Errorf("failed to convert metadata object %d: %w", i, err)
		}
		block.Meta = append(block.Meta, child)
	}

	return block, nil
}

func (cm *ContentMeta) addLinksToDoc(document *doc.Document, opts *Options) error {
	for i, link := range cm.Links {
		allowedTypes := getAllowedContentMetaLinks(opts)
		if !allowedTypes[link.Type] {
			continue
		}

		block, err := blockFromLink(link, opts, LinkContext)
		if err != nil {
			return fmt.Errorf("failed to convert link %d: %w", i, err)
		}

		document.Links = append(document.Links, block)
	}

	return nil
}

func (cm *ContentMeta) addPropertyToDoc(document *doc.Document, name string, value string) {
	document.Properties = append(document.Properties, doc.Property{Name: name, Value: value})
}

func (cm *ContentMeta) addInfoSourceToDoc(document *doc.Document, source *InfoSource) {
	if source != nil {
		infoSource := doc.Property{
			Name:  "infoSource",
			Value: source.Text,
		}
		if source.Literal != "" {
			infoSource.Parameters = map[string]string{"literal": source.Literal}
		}
		document.Properties = append(document.Properties, infoSource)
	}
}

func (cm *ContentMeta) addCreatorToDoc(document *doc.Document, creator *Creator) {
	if creator != nil {
		prop := doc.Property{
			Name:  "creator",
			Value: creator.Text,
		}
		if creator.Literal != "" {
			prop.Parameters = map[string]string{
				"literal": creator.Literal,
			}
		}
		document.Properties = append(document.Properties, prop)
	}
}

func (cm *ContentMeta) addDescriptionToDoc(document *doc.Document, description *Description) {
	if description != nil {
		prop := doc.Property{
			Name:  "description",
			Value: description.Text,
		}
		if description.Role != "" {
			prop.Parameters = map[string]string{
				"role": description.Role,
			}
		}

		document.Properties = append(document.Properties, prop)
	}
}

func (cm *ContentMeta) empty() bool {
	return cm.ContentCreated == "" &&
		cm.ContentModified == "" &&
		cm.InfoSource == nil &&
		cm.Creator == nil &&
		cm.AltID == "" &&
		cm.Slugline == "" &&
		cm.By == "" &&
		len(cm.Description) == 0 &&
		len(cm.MetaData) == 0 &&
		cm.Links == nil
}

func getBlocksForDestination(block *doc.Block, opts *Options) (DataDestination, []doc.Block, error) {
	destination := opts.GetElementDestination(block.Type)
	switch destination {
	case DestinationLink:
		return destination, block.Links, nil
	case DestinationMeta:
		return destination, block.Meta, nil
	default:
		return "", nil, fmt.Errorf("invalid destination configured: %s", destination)
	}
}

func doDataConversion(block *doc.Block, opts *Options, dc *DataElementConversion,
	blocks []doc.Block, destination DataDestination) (string, error) {
	var output string
	var err error
	switch dc.Datatype {
	case DataConversionAsIDF:
		output, err = extractContentToIDFData(blocks, opts)
		if err != nil {
			return "", fmt.Errorf("error extracting content to idf: %w", err)
		}
	case DataConversionAsString:
		switch dc.Destination {
		case DestinationLink:
			if text, ok := block.Links[0].Data["text"]; ok {
				output = text
			}
		case DestinationMeta:
			if text, ok := block.Meta[0].Data["text"]; ok {
				output = text
			}
		}
	}
	switch destination {
	case DestinationLink:
		block.Links = []doc.Block{}
	case DestinationMeta:
		block.Meta = []doc.Block{}
	}
	return output, nil
}
