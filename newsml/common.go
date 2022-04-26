package newsml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/microcosm-cc/bluemonday"

	"bitbucket.org/infomaker/doc-format/v2/doc"
	"github.com/Infomaker/etree"
)

type CatalogRef struct {
	Href string `xml:"href,attr"`
}

// ItemClass represents the <itemMeta><itemClass> element
type ItemClass struct {
	QCode string `xml:"qcode,attr"`
}

type CopyRightHolder struct {
	Name string `xml:"name,omitempty"`
}
type RightsInfo struct {
	CopyRightHolder *CopyRightHolder `xml:"copyrightHolder,omitempty"`
}

type Description struct {
	Role string `xml:"role,attr,omitempty"`
	Text string `xml:",chardata"`
}

func buildXMLLink(link doc.Block, opts *Options) (Link, error) {
	xmlLink := Link{
		Rel:   link.Rel,
		UUID:  link.UUID,
		Title: link.Title,
		Type:  link.Type,
		URI:   link.URI,
		URL:   link.URL,
		Role:  link.Role,
	}

	if link.Data != nil {
		xmlLink.Data = &Data{}
		s, err := transformDataToRaw(&link, opts, LinkContext)
		if err != nil {
			return Link{}, err
		}
		xmlLink.Data.Raw = s
	}

	for i := range link.Links {
		link, err := buildXMLLink(link.Links[i], opts)
		if err != nil {
			return Link{}, fmt.Errorf("failed to convert link %d: %w", i, err)
		}
		xmlLink.Links = append(xmlLink.Links, link)
	}

	option := opts.BlockOptions(link.Type, LinkContext)

	for i := range link.Meta {
		if option.HasFlags && link.Meta[i].Type == "x-im/flags" {
			continue
		}

		object, err := objectFromBlock(link.Meta[i], opts, LinkContext)
		if err != nil {
			return Link{}, fmt.Errorf("failed to convert meta %d: %w", i, err)
		}
		xmlLink.Meta = append(xmlLink.Meta, object)
	}

	return xmlLink, nil
}

func buildXMLService(link doc.Block) Service {
	xmlService := Service{
		QCode: link.Value,
		Name:  link.Title,
	}
	if link.Data != nil {
		if why, ok := link.Data["why"]; ok {
			xmlService.Why = why
		}
	}

	return xmlService
}

func extractContentToIDF(block doc.Block, opts *Options) (string, error) {
	content := block.Content
	type XMLText struct {
		XMLName  xml.Name         `xml:"text"`
		Format   string           `xml:"format,attr"`
		Children []*ElementObject `xml:",any"`
	}

	cs := NewContentSet("")
	err := cs.fromBlocks(content, opts)
	if err != nil {
		return "", err
	}

	xmlText := XMLText{
		Format:   "idf",
		Children: cs.children(),
	}

	xmlData, err := xml.Marshal(&xmlText)
	if err != nil {
		return "", err
	}

	return string(xmlData), nil
}

func extractContentToIDFData(content []doc.Block, opts *Options) (string, error) {
	cs := NewContentSet("")
	err := cs.fromBlocks(content, opts)
	if err != nil {
		return "", err
	}

	xmlData, err := xml.Marshal(cs.children())
	if err != nil {
		return "", err
	}

	return string(xmlData), nil
}

func isPropertyException(docType string, in string, opts *Options) bool {
	sections := make(map[string]bool)
	for _, prop := range opts.PropertyException {
		if strings.EqualFold(prop.Type, in) {
			for _, section := range prop.Section {
				sections[section] = true
			}
			switch docType {
			case "x-im/article", "x-im/image", "x-im/graphic", "x-im/pdf":
				return !sections["itemmeta-newsitem"]
			case "x-im/newscoverage":
				return !sections["itemmeta-planning"]
			case "x-im/assignment":
				return !sections["itemmeta-assignment"]
			default:
				// FIXME Is it safe to assume concept here?
				return !sections["itemmeta-concept"] || sections["itemmeta-conceptitem"]
			}
		}
	}

	return false
}

func hasPropertyExceptionType(str string, section string, opts *Options) bool {
	for _, prop := range opts.PropertyException {
		if strings.EqualFold(prop.Type, str) {
			switch section {
			case "contentmeta":
				for _, s := range prop.Section {
					if s == section || s == "itemmeta-newsml" {
						return true
					}
				}
			case "planning":
				for _, s := range prop.Section {
					if s == section || s == "itemmeta-planning" {
						return true
					}
				}
			}
		}
	}

	return false
}

func fromXMLStatus(xmlstatus string, opts *Options) string {
	for _, status := range opts.Status {
		if status.XML == xmlstatus {
			return status.NavigaDoc
		}
	}

	return xmlstatus
}

func toXMLStatus(jsonstatus string, opts *Options) string {
	for _, status := range opts.Status {
		if status.NavigaDoc == jsonstatus {
			return status.XML
		}
	}

	return jsonstatus
}

func convertTimestamp(tss string) (*time.Time, error) {
	tss = strings.TrimSpace(tss)

	// Special ("graceful") handling of legacy dates
	if tss == "" || tss == "null" || tss == "undefined" {
		return nil, nil
	}

	hasOffset := regexp.MustCompile("(Z|[+-][0-1][0-5]:[0-5][0-9])$")
	if !hasOffset.MatchString(tss) {
		return nil, fmt.Errorf("time \"%s\" UTC offset is missing or invalid", tss)
	}

	t, err := time.Parse(time.RFC3339, tss)
	if err != nil {
		return nil, fmt.Errorf("time \"%s\" parse error: %s", tss, err)
	}
	t = t.UTC()
	return &t, nil
}

// Recursive function to build doc Links from NewsItem Links
func buildDocLinks(docLink *doc.Block, xmlLinks []Link, opts *Options, context ContextType) error {
	for i, xmlLink := range xmlLinks {
		link, err := blockFromLink(xmlLink, opts, context)
		if err != nil {
			return fmt.Errorf("failed to convert link %d: %w", i, err)
		}

		docLink.Links = append(docLink.Links, link)
	}

	return nil
}

func blockFromLink(xmlLink Link, opts *Options, context ContextType) (doc.Block, error) {
	link := doc.Block{
		Title: xmlLink.Title,
		Type:  xmlLink.Type,
		URI:   xmlLink.URI,
		URL:   xmlLink.URL,
		UUID:  xmlLink.UUID,
		Rel:   xmlLink.Rel,
		Role:  xmlLink.Role,
	}

	if xmlLink.Data != nil {
		data, blocks, err := transformDataFromRaw(link.Type, xmlLink.Data.Raw, opts, context)
		if err != nil {
			return doc.Block{}, err
		}

		link.Data = data
		dest := opts.GetElementDestination(link.Type)
		switch dest {
		case DestinationLink:
			link.Links = append(link.Links, blocks...)
		case DestinationMeta:
			link.Meta = append(link.Meta, blocks...)
		default:
			return doc.Block{}, fmt.Errorf("invalid destination configured: %s", dest)
		}
	}

	err := buildDocLinks(&link, xmlLink.Links, opts, context)
	if err != nil {
		return doc.Block{}, err
	}

	for i := range xmlLink.Meta {
		child, err := blockFromObject(xmlLink.Meta[i], opts, context)
		if err != nil {
			return doc.Block{}, fmt.Errorf("failed to convert metadata object %d: %w", i, err)
		}
		link.Meta = append(link.Meta, child)
	}

	return link, nil
}

func getIDFContentFromRaw(raw string, opts *Options) ([]doc.Block, error) {
	var blocks []doc.Block

	newdoc := etree.NewDocument()
	err := newdoc.ReadFromString(raw)
	if err != nil {
		return nil, err
	}
	format := newdoc.FindElement("//text[@format='idf']")

	if format != nil {
		blocks = []doc.Block{}
		newdoc.SetRoot(format)
		// Handle special use case for content-part were text is in IDF format
		idfContent := Group{}

		textAsXMLRoot, _ := newdoc.WriteToString()
		err := xml.Unmarshal([]byte(textAsXMLRoot), &idfContent)
		if err != nil {
			return nil, err
		}
		for _, content := range idfContent.Child {
			childBlock, err := handleGroupChild(*content, opts, ContentContext)
			if err != nil {
				return blocks, err
			}
			blocks = append(blocks, childBlock)
		}
	}

	return blocks, nil
}

func qcodeToDocumentType(qcode string, opts *Options) (string, error) {
	for _, q2s := range opts.NewsmlQcodeType {
		if q2s.Qcode == qcode {
			return q2s.Type, nil
		}
	}

	return "", fmt.Errorf("missing doctype for qcode %s", qcode)
}

func documenttypeToQCode(typ string, opts *Options) (string, error) {
	for _, q2s := range opts.NewsmlQcodeType {
		if q2s.Type == typ {
			return q2s.Qcode, nil
		}
	}

	return "", fmt.Errorf("missing qcode for doctype %s", typ)
}

func concepttypeToQCode(typ string, opts *Options) (string, error) {
	for _, q2s := range opts.ConceptQcodeType {
		if q2s.Type == typ {
			return q2s.Qcode, nil
		}
	}

	return "", fmt.Errorf("missing qcode for doctype %s", typ)
}

func fromXMLType(str string, opts *Options) string {
	for _, etype := range opts.ElementType {
		if etype.XML == str {
			return etype.NavigaDoc
		}
	}

	return str
}

func toXMLType(str string, opts *Options) string {
	for _, etype := range opts.ElementType {
		if etype.NavigaDoc == str {
			return etype.XML
		}
	}

	return str
}

func AddPropertyToDoc(document *doc.Document, name string, value string) {
	document.Properties = append(document.Properties, doc.Property{Name: name, Value: value})
}

func getAllowedContentMetaLinks(opts *Options) map[string]bool {
	allowedTypes := map[string]bool{}
	for _, link := range opts.LinkException {
		for _, section := range link.Section {
			if section == "contentmeta" {
				allowedTypes[link.Type] = true
			}
		}
	}
	return allowedTypes
}

func isItemMetaLinkException(linkType string, linkRel string, opts *Options) bool {
	for _, link := range opts.LinkException {
		if link.Type == linkType {
			for _, s := range link.Section {
				if s == "itemmeta-newsitem" {
					return false
				}
			}
			if len(link.Rel) == 0 {
				return true
			}
			for _, rel := range link.Rel {
				if rel.Name == linkRel {
					return true
				}
			}
		}
	}
	return false
}

func isItemMetaCategoryRel(relIn string, opts *Options) bool {
	for _, link := range opts.LinkException {
		if link.Type == "x-im/category" {
			for _, rel := range link.Rel {
				if rel.Name == relIn {
					return false
				}
			}
		}
	}
	return true
}

var ignoreFields = []string{"catalogref", "catalogRef", "contentCreated", "contentModified", "firstCreated", "versionCreated"}

func CompareXML(targetName string, name string, root string, want string, got string) error {
	errorList := make([]string, 0)

	origXML := etree.NewDocument()
	if err := origXML.ReadFromString(want); err != nil {
		errorList = append(errorList, fmt.Sprintf("ERROR etree failed to parse %s %s: %s", targetName, name, err))
		return errors.New(strings.Join(errorList, "\n"))
	}

	newXML := etree.NewDocument()
	if err := newXML.ReadFromString(got); err != nil {
		errorList = append(errorList, fmt.Sprintf("ERROR etree failed to parse %s xml: %s", targetName, err))
		return errors.New(strings.Join(errorList, "\n"))
	}

	allElements := origXML.FindElements(fmt.Sprintf("/%s//", root))
	if len(allElements) == 0 {
		errorList = append(errorList, fmt.Sprintf("ERROR %s has no elements found for root: %s", targetName, root))
		return errors.New(strings.Join(errorList, "\n"))
	}
originLoop:
	for _, origElement := range allElements {
		for _, s := range ignoreFields {
			if strings.Contains(origElement.GetPath(), s) {
				continue originLoop
			}
		}

		path := origElement.GetPath()
		parentPath := origElement.Parent().GetPath()

		// We have a test case for random things being added
		// to the IDF, ignore that they are discarded.
		if parentPath == "/newsItem/contentSet/inlineXML/idf/group" {
			switch origElement.Tag {
			case "element", "object":
			default:
				continue originLoop
			}
		}

		if origElement.Attr != nil && len(origElement.Attr) > 0 {
			for _, a := range origElement.Attr {
				if a.Value == "" {
					origElement.RemoveAttr(a.Key)
				}
			}
		}

		copyElements := newXML.FindElements(path)
		emptyElement := false
		if origElement.Child == nil || len(origElement.Child) == 0 {
			emptyElement = true
		} else if len(origElement.Child) == 1 {
			switch child := origElement.Child[0].(type) {
			case *etree.CharData:
				elm := child
				if elm.IsWhitespace() {
					emptyElement = true
				}
			default:
			}
		}
		if emptyElement {
			origElement.RemoveAttr("xmlns")
		}

		if (len(origElement.Attr) > 0 || !emptyElement) && len(copyElements) == 0 {
			if !isPathException(origXML.Root().Tag, targetName, origElement.GetPath()) {
				errorList = append(errorList, fmt.Sprintf("[%s] ERROR %s is missing path %s", name, targetName, origElement.GetPath()))
			}
		}

		if origElement.Attr != nil && len(origElement.Attr) > 0 {
			// Check if the original has attributes
			// Try to find one that has matches all of the attributes
			var attrsMatch bool
			var attrsMatchCount int

			for _, copyElement := range copyElements {
				for _, a := range copyElement.Attr {
					if a.Value == "" {
						copyElement.RemoveAttr(a.Key)
					}
				}

				// Compare datetimes using time, ignore im://event
				origIsDatetime := false
				origIsEvent := false
				for _, a := range origElement.Attr {
					if a.Key == "type" {
						if attributeIsDatetime(a.Value) {
							origIsDatetime = true
						}
					} else if a.Key == "uri" && strings.HasPrefix(a.Value, "im://event") {
						origIsEvent = true
					}
				}

				copyIsDatetime := false
				copyIsEvent := false
				for _, a := range copyElement.Attr {
					if a.Key == "type" && attributeIsDatetime(a.Value) {
						copyIsDatetime = true
					} else if a.Key == "uri" && strings.HasPrefix(a.Value, "im://event") {
						copyIsEvent = true
					}
				}

				attrsMatch = false
				attrsMatchCount = 0

				if len(copyElement.Attr) >= len(origElement.Attr) {
					// If type="foo" value="" ignore it as it won't carry over
					for _, attr := range origElement.Attr {
						for _, attr2 := range copyElement.Attr {
							// NPS-252 Do case insensitive comparison for tags sanitized by bluemonday
							if strings.EqualFold(attr.Key, attr2.Key) {
								if attr.Key == "value" && origIsDatetime && copyIsDatetime {
									origDate, err := time.Parse(time.RFC3339Nano, attr.Value)
									if err != nil {
										errorList = append(errorList, fmt.Sprintf("[%s] %s error parsing date %s %v: %s", name, targetName, attr.Key, attr.Value, err))
										break
									}
									genDate, err := time.Parse(time.RFC3339Nano, attr2.Value)
									if err != nil {
										errorList = append(errorList, fmt.Sprintf("[%s] %s error parsing date %s %v: %s", name, targetName, attr2.Key, attr2.Value, err))
										break
									}
									if origDate.Equal(genDate) {
										attrsMatchCount++
										break
									}
								} else if origIsEvent && copyIsEvent || attr.Value == attr2.Value {
									attrsMatchCount++
									break
								}
							}
						}
					}
					if attrsMatchCount == len(origElement.Attr) {
						attrsMatch = true
						break
					}
				}
				if attrsMatch {
					break
				}
			}

			if !attrsMatch && !isAttributesException(origXML.Root().Tag, targetName, origElement, copyElements) {
				errorList = append(errorList, fmt.Sprintf("[%s] ERROR %s no attributes match found for %s %v", name, targetName, origElement.GetPath(), origElement.Attr))
			}
		}

		// Find elements with matching text for data elements
		if strings.HasSuffix(origElement.GetPath(), "/data") {
			var foundText bool
			for _, origChild := range origElement.ChildElements() {
				foundText = false
				origInner := etree.NewDocument()
				origInner.AddChild(origChild.Copy())
				origData, _ := origInner.WriteToString()
				ot := strings.TrimSpace(origChild.Text())
				for _, copyElement := range copyElements {
					for _, copyChild := range copyElement.ChildElements() {
						if origChild.Tag == copyChild.Tag {
							copyInner := etree.NewDocument()
							copyInner.AddChild(copyChild.Copy())
							copyData, _ := copyInner.WriteToString()
							nt := strings.TrimSpace(copyChild.Text())
							if ot == nt {
								foundText = true
								break
							} else if html.UnescapeString(origData) == html.UnescapeString(copyData) || stripSpace(ot) == stripSpace(nt) {
								errorList = append(errorList, fmt.Sprintf("[%s] WARN %s text not exact match %s %v", name, targetName, origElement.GetPath(), origElement.Attr))
								foundText = true
								break
							}
						}
					}
					if foundText {
						break
					}
				}
				if !foundText && !isDataException(origXML.Root().Tag, targetName, origChild.GetPath()) {
					errorList = append(errorList, fmt.Sprintf("[%s] ERROR %s no match found for %s: %s", name, targetName, origChild.GetPath(), ot))
				}
			}
		}
	}

	if len(errorList) > 0 {
		return errors.New(strings.Join(errorList, "\n"))
	}

	return nil
}

func stripSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func attributeIsDatetime(value string) bool {
	return strings.HasPrefix(value, "nrpdate") ||
		strings.HasSuffix(value, "pubstart") ||
		strings.HasSuffix(value, "proddate") ||
		strings.HasSuffix(value, "proddateend")
}

func isAttributesException(root string, target string, element *etree.Element, copyElements []*etree.Element) bool {
	path := element.GetPath()

	if target == "original" {
		// If original doesn't have these they are added
		for _, attr := range element.Attr {
			switch attr.Key {
			case "type":
				if attr.Value == "nrpdate:created" || attr.Value == "nrpdate:modified" ||
					attr.Value == "nrpdate:start" || attr.Value == "nrpdate:end" ||
					attr.Value == "imext:pubstart" {
					return true
				}
			case "dir", "ltr":
				return true
			default:
			}
		}
	}

	if target == "generated" {
		// Properties with type="foo" value="" do not get carried over
		if strings.HasSuffix(path, "Property") {
			typeAttr := element.SelectAttr("type")
			valAttr := element.SelectAttr("value")
			if typeAttr != nil && (valAttr == nil || valAttr.Value == "") {
				return true
			}
		}

		// For planning ignore metadata and itemMeta
		if root == "planning" && (strings.Contains(path, "metadata") ||
			strings.Contains(path, "itemMeta")) {
			return true
		}

		// FIXME better way of dealing with format="html"
		if strings.HasSuffix(path, "element") {
			return true
		}
	}

	if root == "conceptItem" {
		if path == "/conceptItem/concept/conceptId" {
			// conceptId.uri gets generated so just check that others match
			for _, c := range copyElements {
				if c.GetPath() == path {
					for _, attr := range element.Attr {
						for _, attr2 := range c.Attr {
							if attr.Key == attr2.Key && attr.Value != attr2.Value {
								return false
							}
						}
					}
				}
			}
			return true
		}

		if path == "/conceptItem/concept/type" {
			if target == "generated" {
				qcodeAttr := element.SelectAttr("qcode")
				if qcodeAttr == nil {
					return false // It's an error
				}
			}
			return true
		}
	}

	return false
}

func isPathException(root string, target string, path string) bool {
	// For Planning rightsInfo and copyrightHolder may be
	// blank in original and won't be created during rt
	if target == "generated" &&
		(strings.HasSuffix(path, "rightsInfo") ||
			strings.HasSuffix(path, "copyrightHolder")) {
		return true
	}

	// For Image contentSet/inlineXML/idf may be blank
	if root == "newsItem" && target == "generated" &&
		(strings.HasSuffix(path, "contentSet") ||
			strings.HasSuffix(path, "inlineXML")) {
		return true
	}

	// For concept contentMeta is optional but being created
	// in generated item
	if target == "original" && root == "conceptItem" &&
		strings.HasSuffix(path, "contentMeta") {
		return true
	}

	// For planning ignore metadata and itemMeta
	if root == "planning" && target == "generated" &&
		(strings.Contains(path, "metadata") ||
			strings.Contains(path, "itemMeta")) {
		return true
	}

	return false
}

func isDataException(root string, target string, path string) bool {
	// Handle events having <data><registration/></data>
	if strings.HasSuffix(path, "concept/metadata/object/data/registration") {
		return true
	}

	// For planning ignore metadata and itemMeta
	if root == "planning" && target == "generated" &&
		(strings.Contains(path, "metadata") ||
			strings.Contains(path, "itemMeta")) {
		return true
	}

	return false
}

func isConceptObjectType(in string, opts *Options) bool {
	for _, typ := range opts.ObjectException {
		if typ.Type == in {
			for _, section := range typ.Section {
				if section == "concept" {
					return true
				}
			}
		}
	}

	return false
}

func SanitizeHTML(text string, opts *Options) (string, error) {
	p := bluemonday.NewPolicy()

	sanitizeOptions := opts.HTMLSanitizeOptions
	sanitizeOptions.SetAllowables(p)
	sanitizeOptions.SetElements(p)

	sanitizedText := p.SanitizeReader(strings.NewReader(text)).String()

	fragment := etree.NewDocument()
	err := fragment.ReadFromString("<root>" + sanitizedText + "</root>")
	if err != nil || fragment.Root() == nil {
		return "", fmt.Errorf("error parsing text: %v", err)
	}

	for _, h := range opts.HTMLSanitizeOptions.ElementsAttributes {
		hes := strings.Split(h.Elements, ",")
		for _, he := range hes {
			for _, e := range fragment.FindElements(fmt.Sprintf("//%s", strings.ToLower(he))) {
				e.Tag = he
			}
		}
	}

	var innerXML bytes.Buffer
	var writeSettings etree.WriteSettings
	for _, child := range fragment.Root().Child {
		child.WriteTo(&innerXML, &writeSettings)
	}

	return innerXML.String(), nil
}

func isItemMetaObjectType(docType string, in string, opts *Options) bool {
	sections := make(map[string]bool)
	for _, typ := range opts.ObjectException {
		if typ.Type == in {
			for _, section := range typ.Section {
				sections[section] = true
			}
			switch docType {
			case "x-im/article", "x-im/image", "x-im/graphic", "x-im/pdf":
				return sections["itemmeta-newsitem"]
			case "x-im/newscoverage":
				return sections["itemmeta-planning"]
			case "x-im/assignment":
				return sections["itemmeta-assignment"]
			default:
				// FIXME Is it safe to assume concept here?
				return sections["itemmeta-concept"] || sections["itemmeta-conceptitem"]
			}
		}
	}

	return false
}
