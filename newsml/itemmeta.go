package newsml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

// ItemMeta represents the <itemMeta? element
type ItemMeta struct {
	ItemClass           *ItemClass        `xml:"itemClass,omitempty"`
	FileName            *string           `xml:"fileName,omitempty"`
	VersionCreated      string            `xml:"versionCreated"`
	FirstCreated        string            `xml:"firstCreated"`
	Provider            *Provider         `xml:"provider,omitempty"`
	PubStatus           *PubStatus        `xml:"pubStatus"`
	Role                *ItemClass        `xml:"role,omitempty"`
	Title               string            `xml:"title,omitempty"`
	Links               NSLinks           `xml:"links"`
	EdNote              []string          `xml:"edNote,omitempty"`
	ItemMetaExtProperty []MetaExtProperty `xml:"itemMetaExtProperty,omitempty"`
	Service             []Service         `xml:"service,omitempty"`
	MetaData            NSObjects         `xml:"metadata"`
}

type Links []Link

type internalLinks struct {
	XMLName xml.Name
	Links   []Link `xml:"link"`
}

func (l Links) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(l) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for i := range l {
		element := xml.StartElement{
			Name: xml.Name{
				Local: "link",
			},
		}
		if err := e.EncodeElement(l[i], element); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (l *Links) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var internal internalLinks

	if err := d.DecodeElement(&internal, &start); err != nil {
		return err
	}

	*l = (*l)[:]

	for i := range internal.Links {
		*l = append(*l, internal.Links[i])
	}

	return nil
}

type NSLinks []Link

func (l NSLinks) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
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
				Local: "link",
			},
		}
		if err := e.EncodeElement(l[i], element); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

func (l *NSLinks) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var internal internalLinks

	if err := d.DecodeElement(&internal, &start); err != nil {
		return err
	}

	*l = (*l)[:]

	for i := range internal.Links {
		*l = append(*l, internal.Links[i])
	}

	return nil
}

func (im *ItemMeta) FromDoc(document *doc.Document, opts *Options) error {
	if document == nil {
		return errors.New("nil document")
	}
	if document.Type == "" {
		return errors.New("document is missing type")
	}

	err := im.addQCodeType(document.Type, opts)
	if err != nil {
		return err
	}
	im.addPubStatus(document.Published, document.Status, opts)
	im.addPubStop(document.Unpublished)

	im.Title = document.Title
	if document.URL != "" {
		im.addExtProperty("imext:url", document.URL)
	}
	if document.URI != "" {
		im.addExtProperty("imext:uri", document.URI)
	}
	if document.Type != "" {
		im.addExtProperty("imext:type", document.Type)
	}
	if document.Path != "" {
		im.addExtProperty("imext:path", document.Path)
	}
	if document.Provider != "" {
		im.Provider = &Provider{Literal: document.Provider}
	}

	for _, property := range document.Properties {
		switch strings.ToLower(property.Name) {
		case "filename":
			filename := property.Value
			im.FileName = &filename
		case "section":
			imsection := fmt.Sprintf("imsection:%s", property.Value)
			service := Service{
				QCode: imsection,
				Name:  property.Parameters["name"],
			}
			im.Service = append(im.Service, service)
		case "role":
			// TODO Awaiting feedback from Dashboard Team
			im.Role = &ItemClass{QCode: property.Value}
		default:
			if !isPropertyException(document.Type, property.Name, opts) {
				im.addExtProperty(property.Name, property.Value, property.Parameters)
			}
		}
	}

	im.sortExtProperties()

	if document.Created != nil {
		im.FirstCreated = document.Created.Format(time.RFC3339Nano)
	}
	if document.Modified != nil {
		im.VersionCreated = document.Modified.Format(time.RFC3339Nano)
	}

	err = im.addLinks(document.Links, opts)
	if err != nil {
		return err
	}

	err = im.addGeoLinks(document)
	if err != nil {
		return err
	}

	for _, meta := range document.Meta {
		if meta.Type == "x-im/note" {
			im.addEdNote(meta.Data["text"])
		}
	}

	for _, block := range document.Meta {
		err := im.addMetaBlock(document.Type, block, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (im *ItemMeta) addEdNote(note string) {
	im.EdNote = append(im.EdNote, note)
}

func (im *ItemMeta) addLinks(links []doc.Block, opts *Options) error {
	//contentmetaTypes := getAllowedContentMetaLinks(opts)
	for i, link := range links {
		if isItemMetaLinkException(link.Type, link.Rel, opts) {
			continue
		}
		switch link.Type {
		case "x-im/related-geo":
			continue
		case "x-im/service":
			xmlService := buildXMLService(link)
			im.Service = append(im.Service, xmlService)
		default:
			xmlLink, err := buildXMLLink(link, opts)
			if err != nil {
				return fmt.Errorf("failed to convert link %d: %w", i, err)
			}
			im.Links = append(im.Links, xmlLink)
		}
	}
	return nil
}

func (im *ItemMeta) addPubStatus(published *time.Time, status string, opts *Options) {
	if published != nil {
		im.addExtProperty("imext:pubstart", published.Format(time.RFC3339Nano))
	}

	if status == "" {
		im.PubStatus = nil
		return
	}

	im.PubStatus = &PubStatus{}
	im.PubStatus.QCode = toXMLStatus(status, opts)
}

func (im *ItemMeta) addPubStop(unpublished *time.Time) {
	if unpublished != nil {
		im.addExtProperty("imext:pubstop", unpublished.Format(time.RFC3339Nano))
	}
}

func (im *ItemMeta) addQCodeType(docType string, opts *Options) error {
	var qcode string
	var err error

	switch docType {
	case "x-im/article":
		fallthrough
	case "x-im/image":
		fallthrough
	case "x-im/pdf":
		qcode, err = documenttypeToQCode(docType, opts)
		if err != nil {
			return err
		}
	case "x-im/newscoverage":
		qcode = "x-im/assignment"
	case "x-im/list":
		qcode = "x-im/list"
	case "x-im/package":
		qcode = "x-im/package"
	default:
		// first we look to see if there is a concept type configured in options
		qcode, err = concepttypeToQCode(docType, opts)

		// if no such is found we fallback to assume it's an asset document
		if err != nil {
			qcode = "imext:newsitem"
		}
	}

	im.ItemClass = &ItemClass{
		QCode: qcode,
	}

	return nil
}

func (im *ItemMeta) addExtProperty(itemType string, itemValue string, params ...map[string]string) {
	prop := MetaExtProperty{
		Type:  itemType,
		Value: itemValue,
	}

	if len(params) > 0 {
		if creator, ok := params[0]["creator"]; ok {
			prop.Creator = creator
		}
		if why, ok := params[0]["why"]; ok {
			prop.Why = why
		}
	}

	im.ItemMetaExtProperty = append(im.ItemMetaExtProperty, prop)
}

func (im *ItemMeta) sortExtProperties() {
	sort.SliceStable(im.ItemMetaExtProperty, func(i, j int) bool {
		return im.ItemMetaExtProperty[i].Type < im.ItemMetaExtProperty[j].Type
	})
}

func (im *ItemMeta) ToDoc(document *doc.Document, opts *Options) error {
	if im.FirstCreated != "" {
		converted, err := convertTimestamp(im.FirstCreated)
		if err != nil {
			return fmt.Errorf("firstCreated %s", err)
		}
		if converted == nil {
			return fmt.Errorf("firstCreated has invalid value: %s", im.FirstCreated)
		}
		document.Created = converted
	}
	if im.VersionCreated != "" {
		converted, err := convertTimestamp(im.VersionCreated)
		if err != nil {
			return fmt.Errorf("versionCreated %s", err)
		}
		if converted == nil {
			return fmt.Errorf("versionCreated has invalid value: %s", im.VersionCreated)
		}
		document.Modified = converted
	}
	document.Title = im.Title
	if im.ItemClass != nil {
		// Ignore the error as type for concepts comes from ext property
		doctype, _ := qcodeToDocumentType(im.ItemClass.QCode, opts)
		document.Type = doctype
	}
	if im.PubStatus != nil && im.PubStatus.QCode != "" {
		document.Status = fromXMLStatus(im.PubStatus.QCode, opts)
	}

	if im.FileName != nil {
		im.addPropertyToDoc(document, "filename", *im.FileName)
	}

	// TODO Awaiting word from Dashboard Team?
	if im.Role != nil {
		im.addPropertyToDoc(document, "role", im.Role.QCode)
	}

	if im.Provider != nil && im.Provider.Literal != "" {
		document.Provider = im.Provider.Literal
	}

	for _, extProp := range im.ItemMetaExtProperty {
		prop := extProp
		err := im.addExtPropertyToDoc(document, &prop)
		if err != nil {
			return err
		}
	}

	for i, newsmlLink := range im.Links {
		if newsmlLink.Rel == "related-geo" {
			// Turn the data into links, one to many
			err := im.addGeoLinksToDoc(document, newsmlLink)
			if err != nil {
				return fmt.Errorf("failed to convert geo link %d: %w", i, err)
			}
			continue
		}

		block, err := blockFromLink(newsmlLink, opts, LinkContext)
		if err != nil {
			return fmt.Errorf("failed to convert link %d: %w", i, err)
		}
		document.Links = append(document.Links, block)
	}

	im.addEdnoteToDoc(document)
	im.addServicesToDoc(document)

	err := im.addMetaToDoc(document, opts)
	if err != nil {
		return err
	}

	return nil
}

func (im *ItemMeta) addExtPropertyToDoc(document *doc.Document, extprop *MetaExtProperty) error {
	switch extprop.Type {
	case "imext:uri":
		document.URI = extprop.Value
	case "imext:url":
		document.URL = extprop.Value
	case "imext:path":
		document.Path = extprop.Value
	case "imext:type":
		document.Type = extprop.Value
	case "imext:pubstart":
		converted, err := convertTimestamp(extprop.Value)
		if err != nil {
			return fmt.Errorf("imext:pubstart %s", err)
		}
		if converted == nil {
			return nil
		}
		document.Published = converted
	case "imext:pubstop":
		converted, err := convertTimestamp(extprop.Value)
		if err != nil {
			return fmt.Errorf("imext:pubstop %s", err)
		}
		if converted == nil {
			return nil
		}
		document.Unpublished = converted
	default:
		im.addPropertyToDoc(document, extprop.Type, extprop.Value)
		im.addCreator(document, extprop)
		im.addWhy(document, extprop)
	}

	return nil
}

func (im *ItemMeta) addCreator(document *doc.Document, extprop *MetaExtProperty) {
	if extprop.Creator != "" {
		params := document.Properties[len(document.Properties)-1].Parameters
		if params == nil {
			document.Properties[len(document.Properties)-1].Parameters = map[string]string{}
		}
		document.Properties[len(document.Properties)-1].Parameters["creator"] = extprop.Creator
	}
}

func (im *ItemMeta) addWhy(document *doc.Document, extprop *MetaExtProperty) {
	if extprop.Why != "" {
		params := document.Properties[len(document.Properties)-1].Parameters
		if params == nil {
			document.Properties[len(document.Properties)-1].Parameters = map[string]string{}
		}
		document.Properties[len(document.Properties)-1].Parameters["why"] = extprop.Why
	}
}

func (im *ItemMeta) addPropertyToDoc(document *doc.Document, name string, value string) {
	document.Properties = append(document.Properties, doc.Property{Name: name, Value: value})
}

func (im *ItemMeta) addServicesToDoc(document *doc.Document) {
	for _, service := range im.Service {
		block := doc.Block{
			Type:  "x-im/service",
			Title: service.Name,
			Value: service.QCode,
			Data: map[string]string{
				"why": service.Why,
			},
		}
		document.Links = append(document.Links, block)
	}
}

func (im *ItemMeta) addEdnoteToDoc(document *doc.Document) {
	for _, note := range im.EdNote {
		document.Meta = append(document.Meta, doc.Block{
			ID:   strconv.Itoa(len(document.Meta)),
			Type: "x-im/note",
			Data: map[string]string{
				"text": note,
			},
		})
	}
}

type RelatedgeoUUID struct {
	Text  string `xml:",chardata"`
	Title string `xml:"title,attr"`
}

type RelatedgeoData struct {
	XMLName         xml.Name          `xml:"data"`
	RelatedgeoUUIDs []*RelatedgeoUUID `xml:"uuid,omitempty"`
}

func (im *ItemMeta) addGeoLinks(document *doc.Document) error {
	var relatedGeoLinks []doc.Block
	for _, link := range document.Links {
		if link.Rel == "related-geo" {
			relatedGeoLinks = append(relatedGeoLinks, link)
		}
	}

	if len(relatedGeoLinks) > 0 {
		xmlLink := Link{
			Rel: "related-geo",
		}

		data := &RelatedgeoData{}

		for _, link := range relatedGeoLinks {
			related := &RelatedgeoUUID{}
			if len(link.Title) > 0 {
				related.Title = link.Title
			}
			if len(link.UUID) > 0 {
				related.Text = link.UUID
			}

			if len(link.UUID) > 0 || len(link.Title) > 0 {
				data.RelatedgeoUUIDs = append(data.RelatedgeoUUIDs, related)
			}
		}

		xmlBytes, err := xml.Marshal(data)
		if err != nil {
			return err
		}
		xmlDoc := strings.ReplaceAll(string(xmlBytes), "<data>", "")
		xmlDoc = strings.ReplaceAll(xmlDoc, "</data>", "")
		xmlLink.Data = &Data{Raw: xmlDoc}

		im.Links = append(im.Links, xmlLink)
	}

	return nil
}

func (im *ItemMeta) addGeoLinksToDoc(document *doc.Document, xmlLink Link) error {
	if xmlLink.Rel != "related-geo" {
		return nil
	}

	if xmlLink.Data == nil {
		link := doc.Block{
			Type: "x-im/related-geo",
			Rel:  "related-geo",
		}
		document.Links = append(document.Links, link)
		return nil
	}

	data := &RelatedgeoData{}
	xmlFragment := fmt.Sprintf("<data>%s</data>", xmlLink.Data.Raw)
	err := xml.Unmarshal([]byte(xmlFragment), data)
	if err != nil {
		return err
	}

	for _, uuid := range data.RelatedgeoUUIDs {
		link := doc.Block{
			Type:  "x-im/related-geo",
			Rel:   "related-geo",
			UUID:  uuid.Text,
			Title: uuid.Title,
		}
		document.Links = append(document.Links, link)
	}

	return nil
}

func (im *ItemMeta) addMetaToDoc(document *doc.Document, opts *Options) error {
	for i, object := range im.MetaData {
		block, err := blockFromObject(object, opts, MetaContext)
		if err != nil {
			return fmt.Errorf("failed to convert metadata object %d: %w", i, err)
		}

		document.Meta = append(document.Meta, block)
	}
	return nil
}

func (im *ItemMeta) addMetaBlock(docType string, block doc.Block, opts *Options) error {
	// already taken care off elsewhere
	if block.Type == "x-im/note" || !isItemMetaObjectType(docType, block.Type, opts) {
		return nil
	}

	object, err := objectFromBlock(block, opts, MetaContext)
	if err != nil {
		return err
	}

	im.MetaData = append(im.MetaData, object)

	return nil
}
