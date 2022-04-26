package newsml

import (
	"reflect"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

type DateConfig map[string]map[string]map[string]interface{}

type ContextType string

const (
	MetaContext    ContextType = "meta"
	LinkContext    ContextType = "link"
	ContentContext ContextType = "content"
)

const (
	TagNode  string = "tags"
	TypeNode string = "types"
	Block    string = "blocks"
)

func (dc DateConfig) IsDate(key string, name string) bool {
	if _, ok := dc[key][name]; ok {
		return ok
	}
	return false
}

func (dc DateConfig) IsRequired(key string, name string) bool {
	return dc[key][name]["required"] == true
}

func (dc DateConfig) AllowBlank(key string, name string) bool {
	return dc[key][name]["allow-blank"] == true
}

func (dc DateConfig) AllowString(key string, name string) bool {
	return dc[key][name]["allow-string"] == true
}

func (dc DateConfig) GetFormat(key string, name string) string {
	if format, ok := dc[key][name]["format"]; ok {
		if reflect.TypeOf(format).Kind() == reflect.String {
			return format.(string)
		}
	}
	return "RFC3339"
}

func (dc DateConfig) GetAttribute(key string, name string) string {
	if attribute, ok := dc[key][name]["use-attribute"]; ok {
		if reflect.TypeOf(attribute).Kind() == reflect.String {
			return attribute.(string)
		}
	}
	return "value"
}

type Options struct {
	Option                map[string]Option         `json:"text-options"`
	NewsmlQcodeType       []Qcode2Type              `json:"qcode-type,omitempty"`
	AssignmentQcodeType   []AssignmentQcode2Type    `json:"assignment-qcode-type,omitempty"`
	ConceptQcodeType      []ConceptQcode2Type       `json:"concept-qcode-type,omitempty"`
	Status                []StatusOrElementType     `json:"status,omitempty"`
	ElementType           []StatusOrElementType     `json:"element-type,omitempty"`
	PropertyType          []PropertyType            `json:"property-type,omitempty"`
	LinkException         []LinkObjectExceptionType `json:"link-exceptions,omitempty"`
	ObjectException       []LinkObjectExceptionType `json:"object-exceptions,omitempty"`
	PropertyException     []PropertyExceptionType   `json:"property-exceptions,omitempty"`
	DataElementConversion []DataElementConversion   `json:"data-conversions,omitempty"`
	DateElements          DateConfig                `json:"date-elements,omitempty"`
	HTMLSanitizeOptions   HTMLSanitizeOptions       `json:"html-sanitize-options,omitempty"`
}

func (opts *Options) BlockOptions(blockType string, context ContextType) Option {
	attr := opts.Option[blockType]
	overrides, ok := attr.Overrides[string(context)]
	if !ok {
		return attr
	}

	newAttrs := map[string]AttributeOptions{}
	for k, v := range attr.Attributes {
		newAttrs[k] = v
	}
	for k, v := range overrides {
		newAttrs[k] = v
	}
	attr.Attributes = newAttrs

	return attr
}

type Qcode2Type struct {
	Qcode string `json:"qcode"`
	Type  string `json:"type"`
}

type AssignmentQcode2Type struct {
	Qcode string `json:"qcode"`
	Type  string `json:"type"`
}

type ConceptQcode2Type struct {
	Qcode string `json:"qcode"`
	Type  string `json:"type"`
}

type StatusOrElementType struct {
	XML       string `json:"newsml"`
	NavigaDoc string `json:"navigadoc"`
}

type PropertyType struct {
	Section   []string `json:"section"`
	XML       string   `json:"newsml"`
	NavigaDoc string   `json:"navigadoc"`
}

type RelType struct {
	Name    string `json:"name"`
	Section string `json:"section"`
}

type LinkObjectExceptionType struct {
	Type    string    `json:"type"`
	Rel     []RelType `json:"rel,omitempty"`
	Section []string  `json:"section"`
}

type PropertyExceptionType struct {
	Type    string   `json:"type"`
	Section []string `json:"section"`
}

type DataElementConversion struct {
	Type        string                 `json:"type,omitempty"`
	Destination DataDestination        `json:"destination,omitempty"`
	Elements    []DataElement          `json:"elements,omitempty"`
	Datatype    DataConversionDataType `json:"data-type,omitempty"`
	Flags       bool                   `json:"flags"`
}

type DataElement struct {
	Name       string   `json:"name,omitempty"`
	Type       string   `json:"type,omitempty"`
	Rel        string   `json:"rel,omitempty"`
	Attributes []string `json:"attributes,omitempty"`
}

type HTMLSanitizeOptions struct {
	AllowStandardAttributes *bool                   `json:"allow-standard-attributes,omitempty"`
	AllowImages             *bool                   `json:"allow-images,omitempty"`
	AllowLists              *bool                   `json:"allow-lists,omitempty"`
	AllowTables             *bool                   `json:"allow-tables,omitempty"`
	AllowRelativeURLs       *bool                   `json:"allow-relative-urls,omitempty"`
	AllowableURLSchemes     string                  `json:"allowed-url-schemes,omitempty"`
	ElementsAttributes      []HTMLElementAttributes `json:"allowed-elements-attributes,omitempty"`
}

type HTMLElementAttributes struct {
	Elements   string `json:"element,omitempty"`
	Attributes string `json:"attributes,omitempty"`
}

func (s HTMLSanitizeOptions) SetAllowables(p *bluemonday.Policy) {
	if s.AllowStandardAttributes != nil {
		if *s.AllowStandardAttributes {
			p.AllowStandardAttributes()
		}
	}
	if s.AllowLists != nil {
		if *s.AllowLists {
			p.AllowLists()
		}
	}

	if s.AllowImages != nil {
		if *s.AllowImages {
			p.AllowImages()
		}
	}

	if s.AllowTables != nil {
		if *s.AllowTables {
			p.AllowTables()
		}
	}

	p.RequireParseableURLs(true)
	p.RequireNoFollowOnFullyQualifiedLinks(false)
	p.RequireNoFollowOnLinks(false)

	if s.AllowRelativeURLs != nil {
		if *s.AllowRelativeURLs {
			p.AllowRelativeURLs(true)
		}
	}

	if s.AllowableURLSchemes != "" {
		p.AllowURLSchemes(strings.Split(s.AllowableURLSchemes, ",")...)
	}
}

func (s HTMLSanitizeOptions) SetElements(p *bluemonday.Policy) {
	for _, e := range s.ElementsAttributes {
		if strings.TrimSpace(e.Attributes) != "" {
			p.AllowAttrs(strings.Split(e.Attributes, ",")...).OnElements(strings.Split(e.Elements, ",")...)
		} else {
			elm := strings.Split(e.Elements, ",")
			p.AllowNoAttrs().OnElements(elm...)
			p.AllowElements(elm...)
		}
	}
}

func NewOptions() Options {
	return Options{
		Option: make(map[string]Option),
	}
}

func (opts *Options) Elements(types ...string) {
	for _, t := range types {
		o := opts.Option[t]

		o.Type = t
		o.IsElement = true

		opts.Option[t] = o
	}
}

func (opts *Options) SupportFlags(types ...string) {
	for _, t := range types {
		o := opts.Option[t]

		o.Type = t
		o.HasFlags = true

		opts.Option[t] = o
	}
}

func (opts *Options) Attributes(t string, attr map[string]AttributeOptions) {
	o := opts.Option[t]

	o.Type = t
	o.Attributes = attr

	opts.Option[t] = o
}

func (opts *Options) OverrideAttributes(t, context string, attr map[string]AttributeOptions) {
	o := opts.Option[t]

	o.Type = t
	if o.Overrides == nil {
		o.Overrides = map[string]map[string]AttributeOptions{}
	}
	o.Overrides[context] = attr

	opts.Option[t] = o
}

func (opts *Options) NewsmlQcodes(qcodes []Qcode2Type) {
	opts.NewsmlQcodeType = qcodes
}

func (opts *Options) AssignmentQcodes(qcodes []AssignmentQcode2Type) {
	opts.AssignmentQcodeType = qcodes
}

func (opts *Options) ConceptQcodes(qcodes []ConceptQcode2Type) {
	opts.ConceptQcodeType = qcodes
}

func (opts *Options) Statuses(statuses []StatusOrElementType) {
	opts.Status = statuses
}

func (opts *Options) ElementTypes(elementTypes []StatusOrElementType) {
	opts.ElementType = elementTypes
}

func (opts *Options) PropertyTypes(propertyTypes []PropertyType) {
	opts.PropertyType = propertyTypes
}

func (opts *Options) LinkExceptions(linkExceptionType []LinkObjectExceptionType) {
	opts.LinkException = linkExceptionType
}

func (opts *Options) ObjectExceptions(objectExceptionTypes []LinkObjectExceptionType) {
	opts.ObjectException = objectExceptionTypes
}

func (opts *Options) PropertieExceptions(propertyExceptionTypes []PropertyExceptionType) {
	opts.PropertyException = propertyExceptionTypes
}

func (opts *Options) DataConversions(dataElementConversions []DataElementConversion) {
	opts.DataElementConversion = dataElementConversions
}

func (opts *Options) GetDataConversionForType(objectType string) *DataElementConversion {
	for _, dc := range opts.DataElementConversion {
		if dc.Type == objectType {
			return &dc
		}
	}
	return nil
}

func (opts *Options) GetDataConversionElement(objectType string, tag string) *DataElement {
	for _, dc := range opts.DataElementConversion {
		if dc.Type == objectType {
			for _, e := range dc.Elements {
				if e.Type == objectType && e.Name == tag {
					return &e
				}
			}
		}
	}
	return nil
}

func (opts *Options) GetElementDestination(objectType string) DataDestination {
	for _, dc := range opts.DataElementConversion {
		if dc.Type == objectType {
			return dc.Destination
		}
	}
	return DestinationMeta
}

func (opts *Options) DateValidations(dc DateConfig) {
	opts.DateElements = dc
}

func (opts *Options) SanitizeOptions(so HTMLSanitizeOptions) {
	opts.HTMLSanitizeOptions = so
}

// TODO: perhaps this should live in an option package?
type Option struct {
	Type       string                                 `json:"type"`
	IsElement  bool                                   `json:"isElement"`
	HasFlags   bool                                   `json:"flags"`
	Attributes map[string]AttributeOptions            `json:"attributes"`
	Overrides  map[string]map[string]AttributeOptions `json:"overrides"`
	Options    map[string]string                      `json:"options"`
}

func (o Option) Attribute(n string) AttributeOptions {
	if o.Attributes == nil {
		return AttributeOptions{}
	}
	return o.Attributes[n]
}

type AttributeOptions struct {
	ValueHandling  ValueHandling `json:"value-handling"`
	ValueAttribute string        `json:"value-attribute"`
}

const DefaultValueAttribute = "value"

type ValueHandling string

const (
	ValueAsText      ValueHandling = ""
	ValueAsCData     ValueHandling = "cdata"
	ValueAsXML       ValueHandling = "xml"
	ValueAsAttribute ValueHandling = "attribute"
)

type DataDestination string

const (
	DestinationMeta DataDestination = "meta"
	DestinationLink DataDestination = "link"
)

type DataConversionDataType string

const (
	DataConversionAsXML    DataConversionDataType = ""
	DataConversionAsString DataConversionDataType = "blob"
	DataConversionAsIDF    DataConversionDataType = "idf"
)

func isElementType(typeStr string, opts *Options) bool {
	return opts.Option[typeStr].IsElement
}

func DefaultOptions() Options {
	opts := NewOptions()

	opts.Elements(
		"x-im/paragraph", "preamble", "leadin", "body", "dateline",
		"headline", "x-im/header", "subheadline", "subheadline1", "subheadline2",
		"subheadline3", "subheadline4", "subheadline5", "headline",
		"drophead", "fact-body", "pagedateline", "preleadin",
		"preamble", "madmansrow", "blockquote", "x-im/unordered-list",
		"x-im/ordered-list", "monospace", "byline", "attribution",
	)

	opts.SupportFlags("x-im/image", "x-im/teaser")

	opts.Attributes("x-im/iframely", map[string]AttributeOptions{
		"embedCode": {ValueHandling: ValueAsCData},
	})

	opts.Attributes("x-im/youtube", map[string]AttributeOptions{
		"embedCode": {ValueHandling: ValueAsCData},
	})

	opts.Attributes("x-im/htmlembed", map[string]AttributeOptions{
		"embedCode": {ValueHandling: ValueAsCData},
	})

	opts.Attributes("x-im/table", map[string]AttributeOptions{
		"thead": {ValueHandling: ValueAsXML},
		"tbody": {ValueHandling: ValueAsXML},
		"tfoot": {ValueHandling: ValueAsXML},
	})

	opts.Attributes("x-im/teaser", map[string]AttributeOptions{
		"title": {ValueHandling: ValueAsXML},
		"text":  {ValueHandling: ValueAsXML},
	})

	opts.Attributes("x-im/content-part", map[string]AttributeOptions{
		"caption": {ValueHandling: ValueAsXML},
		"title":   {ValueHandling: ValueAsXML},
		"subject": {ValueHandling: ValueAsXML},
	})

	opts.Attributes("x-im/pdf", map[string]AttributeOptions{
		"title": {ValueHandling: ValueAsXML},
		"text":  {ValueHandling: ValueAsXML},
	})

	opts.Attributes("x-im/imagegallery", map[string]AttributeOptions{
		"text": {ValueHandling: ValueAsXML},
	})

	opts.OverrideAttributes("x-im/image", "content", map[string]AttributeOptions{
		"text": {ValueHandling: ValueAsXML},
	})

	opts.NewsmlQcodes([]Qcode2Type{
		{"ninat:picture", "x-im/image"},
		{"ninat:graphic", "x-im/pdf"},
		{"ninat:text", "x-im/article"},
		{"ninat:video", "x-im/video"},
		{"cinat:concept", "x-im/concept"},
	})

	opts.AssignmentQcodes([]AssignmentQcode2Type{
		{"ninat:picture", "x-im/image"},
		{"ninat:graphic", "x-im/pdf"},
		{"ninat:text", "x-im/article"},
		{"ninat:video", "x-im/video"},
	})

	opts.ConceptQcodes([]ConceptQcode2Type{
		{"cpnat:person", "x-im/author"},
		{"cpnat:object", "x-im/category"},
		{"cpnat:abstract", "x-im/channel"},
		{"cpnat:object", "x-im/content-profile"},
		{"cpnat:event", "x-im/event"},
		{"cpnat:organisation", "x-im/organisation"},
		{"cpnat:person", "x-im/person"},
		{"cpnat:poi", "x-im/place"},
		{"cpnat:object", "x-im/section"},
		{"cpnat:object", "x-im/topic"},
		{"cpnat:abstract", "x-im/story"},
	})

	opts.Statuses([]StatusOrElementType{
		{"imext:draft", "draft"},
		{"imext:done", "done"},
		{"imext:approved", "approved"},
		{"stat:usable", "usable"},
		{"stat:canceled", "canceled"},
		{"stat:withheld", "withheld"},
	})

	opts.ElementTypes([]StatusOrElementType{
		{"body", "x-im/paragraph"},
		{"headline", "x-im/header"},
	})

	opts.PropertyTypes([]PropertyType{
		{[]string{"planning"}, "nrpdate:start", "imext:start"},
		{[]string{"planning"}, "nrpdate:end", "imext:end"},
		{[]string{"planning"}, "nrpdate:created", "imext:created"},
		{[]string{"planning"}, "nrpdate:modified", "imext:modified"},
		{[]string{"planning"}, "nrp:sector", "imext:sector"},
	})

	opts.LinkExceptions([]LinkObjectExceptionType{
		{"x-im/articlesource", []RelType{}, []string{"contentmeta"}},
		{"x-im/premium", []RelType{}, []string{"contentmeta"}},
		{"x-im/articletype", []RelType{}, []string{"contentmeta"}},
		{"x-geo/point", []RelType{}, []string{"contentmeta"}},
		{"x-im/contenttype", []RelType{}, []string{"contentmeta"}},
		{"x-im/articlecontent", []RelType{}, []string{"contentmeta"}},
		{"x-im/articlecontenttype", []RelType{}, []string{"contentmeta"}},
		{"x-im/category", []RelType{
			{"category", "planning"},
			{"category", "contentmeta"},
		}, []string{"contentmeta"}},
		{"x-im/articleoptions", []RelType{}, []string{"contentmeta"}},
		{"x-im/articleoptions/plus", []RelType{}, []string{"contentmeta"}},
		{"x-im/articleoptions/comments", []RelType{
			{"comment", "contentmeta"},
		}, []string{"contentmeta"}},
		{"x-im/plus", []RelType{}, []string{"contentmeta"}},
	})

	opts.ObjectExceptions([]LinkObjectExceptionType{
		{"x-im/contact-info", []RelType{}, []string{"concept"}},
		{"x-im/position", []RelType{}, []string{"concept"}},
		{"x-im/event-details", []RelType{}, []string{"concept"}},
		{"cpnat:person", []RelType{}, []string{"conceptitem"}},
		{"cpnat:object", []RelType{}, []string{"conceptitem"}},
		{"cpnat:event", []RelType{}, []string{"conceptitem"}},
		{"cpnat:abstract", []RelType{}, []string{"conceptitem"}},
		{"cpnat:organisation", []RelType{}, []string{"conceptitem"}},
		{"cpnat:poi", []RelType{}, []string{"conceptitem"}},
		{"x-im/event", []RelType{}, []string{"conceptitem"}},
		{"x-im/polygon", []RelType{}, []string{"concept"}},
	})

	opts.PropertieExceptions([]PropertyExceptionType{
		{"imext:description", []string{"list"}},
		{"imext:product", []string{"list", "package"}},
		{"imext:itemLimit", []string{"list"}},
		{"imext:type", []string{"list", "package"}},
		{"imext:pubstart", []string{"package"}},
		{"imext:pubstop", []string{"package"}},
		{"imext:cover", []string{"package"}},
		{"category", []string{"package"}},
		{"uri", []string{"contentmeta"}},
		{"altId", []string{"contentmeta"}},
		{"infosource", []string{"contentmeta"}},
		{"contentcreated", []string{"contentmeta"}},
		{"contentmodified", []string{"contentmeta"}},
		{"type", []string{"contentmeta"}},
		{"by", []string{"contentmeta"}},
		{"headline", []string{"contentmeta"}},
		{"description", []string{"contentmeta", "concept"}},
		{"definition", []string{"concept"}},
		{"slugline", []string{"contentmeta", "planning"}},
		{"provider", []string{"contentmeta"}},
		{"nrp:sector", []string{"contentmeta", "concept"}},
		{"imext:header", []string{"contentmeta"}},
		{"imext:subheader", []string{"contentmeta"}},
		{"imext:deck", []string{"contentmeta"}},
		{"imext:simplebyline", []string{"contentmeta"}},
		{"copyrightholder", []string{"planning", "concept"}},
		{"headline", []string{"planning"}},
		{"imext:headline", []string{"concept"}},
		{"imext:slugline", []string{"concept"}},
		{"urgency", []string{"planning", "concept"}},
		{"concepttypes", []string{"concept"}},
		{"imext:qcode", []string{"concept"}},
		{"nrptype:evtyp", []string{"concept"}},
		{"infoSource", []string{"contentmeta"}},
		{"creator", []string{"contentmeta"}},
		{"language", []string{"contentmeta"}},
		{"conceptid", []string{"concept"}},
	})

	opts.DateValidations(DateConfig{
		"tags": {
			"firstCreated": map[string]interface{}{
				"allow-blank":  false,
				"allow-string": false,
				"format":       "RFC3339Nano",
			},
			"versionCreated": map[string]interface{}{
				"allow-blank":  false,
				"allow-string": false,
				"format":       "RFC3339Nano",
			},
			"contentCreated": map[string]interface{}{
				"allow-blank":  false,
				"allow-string": false,
				"format":       "RFC3339Nano",
			},
			"contentModified": map[string]interface{}{
				"allow-blank":  false,
				"allow-string": false,
				"format":       "RFC3339Nano",
			}},
	})

	opts.SanitizeOptions(HTMLSanitizeOptions{
		AllowStandardAttributes: newTrue(),
		AllowImages:             newTrue(),
		AllowLists:              newTrue(),
		AllowTables:             newTrue(),
		AllowRelativeURLs:       newFalse(),
		AllowableURLSchemes:     "http,https,mailto,sms,tel,callto",
		ElementsAttributes: []HTMLElementAttributes{
			{
				Elements:   "strong",
				Attributes: "id,type",
			},
			{
				Elements:   "em",
				Attributes: "",
			},
			{
				Elements:   "mark",
				Attributes: "id",
			},
			{
				Elements:   "a",
				Attributes: "href,rel,target,download,title",
			},
			{
				Elements:   "element",
				Attributes: "id,type",
			},
			{
				Elements:   "ins",
				Attributes: "id,collapsed,creationdate,modificationdate,username",
			},
			{
				Elements:   "code",
				Attributes: "id",
			},
			{
				Elements:   "x-person",
				Attributes: "id,tag,tagdescription,tagid,tagimageurl,taglongdescription",
			},
		},
	})

	return opts
}
func newTrue() *bool {
	b := true
	return &b
}
func newFalse() *bool {
	b := false
	return &b
}
