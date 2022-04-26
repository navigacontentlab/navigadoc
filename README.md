# CCA Format Converters

The purpose of this package is to provide support for marshaling/unmarshaling of
NavigaDoc JSON and Infomaker NewsML (IMNML), plus converting documents between XML
and Navigadoc JSON.

## Converter Functions

The doc-format package has the following components

* Package bitbucket.org/infomaker/doc-format/newsml

      * contains support for converting between the NewsML format and NavigaDoc, 
        e.g. root tags newsItem, conceptItem, planningItem, and planning (assignment)
      * contains functions for retrieving and merging conversion options
        * newsml.DefaultOptions() newsml.Options
        * docformat.MergeOptions(opts newsml.Options, custom *newsml.Options
  
* Package bitbucket.org/infomaker/doc-format/listpackage
      
      * contains support for converting between List/Package and NavigaDoc
    
* Package bitbucket.org/infomaker/doc-format
      
      * converter.go contains these functions for converting from one format to the other
        * AssignmentToNavigaDoc(*newsml.Assignment, *newsml.Options) (*doc.Document, error)
        * NavigaDocToAssignment(*doc.Document, *newsml.Options) (*newsml.Assignment, error)
        * NavigaDocToConceptItem(*doc.Document, *newsml.Options)
        * NavigaDocToItem(*doc.Document, *newsml.Options) (interface{}, error)
        * NavigaDocToList(*doc.Document, *newsml.Options) (*listpackage.List, error)
        * NavigaDocToNewsItem(*doc.Document, *newsml.Options) (*newsml.NewsItem, error)
        * NavigaDocToPackage(*doc.Document, *newsml.Options) (*listpackage.Package, error)
        * NavigaDocToPlanningItem(*doc.Document, *newsml.Options) (*newsml.PlanningItem, error)
        * ConceptItemToNavigaDoc(*newsml.ConceptItem, *newsml.Options) (*doc.Document, error)
        * ItemToNavigaDoc(interface{}, *newsml.Options) (*doc.Document, error)
        * ListToNavigaDoc(*listpackage.List, *newsml.Options) (*doc.Document, error)
        * NewsItemToNavigaDoc(*newsml.NewsItem, *newsml.Options) (*doc.Document, error)
        * PackageToNavigaDoc(*listpackage.Package, *newsml.Options) (*doc.Document, error)
        * PlanningItemToNavigaDoc(*newsml.PlanningItem, *newsml.Options) (*doc.Document, error)
      * converter.go also contains these helper functions for XML
        * XMLToNavigaDoc(string, *newsml.Options) (*doc.Document, error)
        * XMLToItem(io.Reader) (interface{}, error)
      * helper.go contains the following helper functions
        * For NEWSML Documents
            * WalkXMLDocumentElements(document *etree.Document, fns ...ElementVisitor) error
            * FixNewsMLUUIDsAndNamespaces(document *etree.Document) error
            * FixNewsMLNamespaces(element *etree.Element) error
            * ValidateAndLowercaseNewsMLUUID(element *etree.Element, theUUID string) error
            * ValidateAndLowercaseNewsMLUUIDs(element *etree.Element) error
            * ValidateNewsMLDates(xmlDoc *etree.Document, dateConfig newsml.DateConfig) error
            * ValidateDate(dateConfig newsml.DateConfig, eType string, name string, date string) error
        * For Navigadoc Documents
            * WalkDocument(document *doc.Document, fns ...BlockVisitor) error
            * WalkBlocks(s []doc.Block, fns ...BlockVisitor) error
            * WalkBlock(block doc.Block, fns ...BlockVisitor) (doc.Block, error)
            * ValidateAndLowercaseDocumentUUIDs(block doc.Block) (doc.Block, error)
            * ValidateDocumentDates(document *doc.Document, dateConfig newsml.DateConfig) error

## Conversion mappings

As the Navigadoc JSON structure does not match that of the XML-based documents, correctly transforming NavigaDoc to/from XML requires mappings for designating which field values go where. Going from NewsML to Navigadoc targets are hard-coded, so the config file is used mostly for transforming Navigadoc to XML.

For example, properties defined as itemMetaExtProperty and contentMetaExtProperty are all stored under Properties in Navigadoc. When transforming from Navigadoc to XML, by default all properties are copied into itemMetaExtaProperty elements. By using a configuration file exceptions may be defined that designate where the properties should actually reside, e.g. contentMetaExtProperty.

The docformat module has internal values which are used if the *newsml.Options parameter of the transformation functions is left nil. The default configuration may be obtained using newsml.DefaultOptions().

A custom configuration file may be supplied, however its values must be merged with the defaults in order to be used. This may be accomplished by using docformat.MergeOptions as follows

```go
    tmp := newsml.DefaultOptions()
    opts = &tmp
    
    customConfig := newsml.Options{}
    err = json.Unmarshal([]byte("./custom-config.json"), &customConfig)
    if err != nil {
        return err
    }
    
    docformat.MergeOptions(opts, &customConfig)
```

For those options that are stored as arrays, MergeOptions appends the custom values after the defaults. For _text-options_, _date-elements_ and _html-sanitize-options_ default values will be replaced by custom values of the same name.

## Custom Configuration File

The custom configuration file allows for defining certain rules for transforming between XML and Navigadoc.

The file is divided into six basic parts, 1) Options, 2) Mappings, 3) Exceptions, 4) Data Conversions, 5) Date Validations and 6) HTML Sanitize Options.

### Options (text-options)

Options is used to define whether a _content_ member of a Navigadoc is an element or object in IDF, and how attributes are to be handled.

Each text-option entry has a name corresponding to the object or element type. The JSON object has the members "type", "isElement", "attributes", and "options". 

The _type_ member should match the object type

The _isElement_ member is a boolean defining whether the Navigadoc memmber gets transformed into an IDF <object> or <element>

The _attributes_ member defines handling for the data child elements. It consists of an object with two members, value-handling and value-attribute.

The _value-handling_ defines whether the element is handled as text, XML, CDATA or an attribute.

The _value-attribute_ member defines the name of the attribute holding the value, with the default being "value".

In Newsml given the following contentMeta/metadata/object

```xml
    <object id="test-1553ed0004d140fdffc2a31058b2597c" type="x-im/test-plugin">
        <data>
            <format>html</format>
            <astext>Some text here</astext>
            <asxml><h1>Some xml here</h1></asxml>
            <ascdata><![CDATA[<h1>Some html here</h1>]]></ascdata>
            <asattribute value="attrvalue"/>
        </data>
    </object>
```

When converting from Navigadoc the custom config options define x-im/test-plugin
as being an <object>
The _astext_ member contains a simple text value (the default)
The _asxml_ member has innerXML
The _ascdata_ member has CDATA
The _asattribute_ member value will map into an attribute named _value_

The following would be the _text-options_ configuration

```json5
{
  "text-options": {
    "x-im/test-plugin": {
      "type": "x-im/test-plugin",
      "isElement": false,
      "attributes": {
        "astext": {
          "value-handling": "",
          "value-attribute": ""
        },
        "asxml": {
          "value-handling": "xml",
          "value-attribute": ""
        },
        "ascdata": {
          "value-handling": "cdata",
          "value-attribute": ""
        },
        "asattribute": {
          "value-handling": "attribute",
          "value-attribute": ""
        }
      },
      "options": null
    }
  }
}
```

In Navigadoc this is stored under _meta_. Note that the ascdata and asattribute values are stored as plain text. The config is used for transforming those into CDATA and XML attribute values. 

```json5
    {
      "id": "test-1553ed0004d140fdffc2a31058b2597c",
      "type": "x-im/test-plugin",
      "data": {
        "asattribute": "attrvalue",
        "ascdata": "<h1>Some html here</h1>",
        "astext": "Some text here",
        "asxml": "<h1>Some xml here</h1>",
        "format": "html"
      }
    }
```

### Mappings

Mappings are used to define one-to-one mappings of values between XML and Navigadoc.

The examples below show selections from the defaults

#### QCode Mappings (qcode-type)

The qcode-type field maps values between //itemMeta/itemClass[@qcode]
and navigadoc.Type.

```json5
{
  "qcode-type": [
    {
      "qcode": "ninat:picture",
      "type": "x-im/image"
    },
    {
      "qcode": "ninat:graphic",
      "type": "x-im/pdf"
    }
  ]
}
```

#### Assignment QCode Mappings (assignment-qcode-type)

The assignment-qcode-type field maps values between /planning/itemClass[@qcode]
and navigadoc.Meta.Data["type"].

```json5
{
  "assignment-qcode-type": [
    {
      "qcode": "ninat:picture",
      "type": "x-im/image"
    },
    {
      "qcode": "ninat:graphic",
      "type": "x-im/pdf"
    }
  ]
}
```

#### Concept QCode Mappings (concept-qcode-type)

The concept-qcode-type field maps values between //concept/type[@qcode]
and navigadoc.Type.

```json5
{
  "concept-qcode-type": [
    {
      "qcode": "cpnat:person",
      "type": "x-im/author"
    },
    {
      "qcode": "cpnat:object",
      "type": "x-im/category"
    }
  ]
}
```
 
### Status Mappings (status)

The statuses field maps values between //itemMeta/pubStatus[@qcode] and
navigadoc.Status.

```json5
{
  "status": [
    {
      "newsml": "imext:draft",
      "navigadoc": "draft"
    },
    {
      "newsml": "imext:done",
      "navigadoc": "done"
    }
  ]
}
```

### Element Type Mappings (element-type)

The element-type field maps values between //contentSet/inlineXML/idf/group/element[@type]
and navigadoc.Content.Type

```json5
{
  "element-type": [
    {
      "newsml": "body",
      "navigadoc": "x-im/paragraph"
    },
    {
      "newsml": "headline",
      "navigadoc": "x-im/header"
    }
  ]
}
```

### Property Type Mappings (property-type)

The property-type field maps values between //itemMeta/itemMetaExtProperty[@type]
and navigadoc.Property.Type

Note that if the type is entered in the property exceptions (see below) the mapping
may be to //contentMeta/contentMetaExtProperty[@type]

```json5
{
  "property-type": [
    {
      "section": [
        "planning"
      ],
      "newsml": "nrpdate:start",
      "navigadoc": "imext:start"
    },
    {
      "section": [
        "planning"
      ],
      "newsml": "nrpdate:end",
      "navigadoc": "imext:end"
    }
  ]
}
```

### Exceptions

Where there are no hard-coded default mappings, all links from the navigadoc are transformed into links under itemMeta, properties are transformed into itemMetaExtProperty values, and all objects are transformed into objects under contentMeta. The exceptions field allows for defining exceptions to those rules, including where they should go.

Default mappings may be overridden by using one of the following Section values. See each exception type for details.

| Document Type | Section | Description |
|---|---|---|
|newsml|itemmeta-newsml| newsitem/contentMeta |
|concept|itemmeta-concept|  conceptItem/concept |
|concept|itemmeta-conceptItem| conceptItem/contentMeta |
|concept|itemmeta-planning| planningItem/planning |

#### Link Exceptions (link-exceptions)

By default, links from NavigaDoc get stored under itemMeta.

In the following example any navigadoc link with the type x-im/articlesource will be stored under //contentMeta during transformation. Where rel is supplied the combination of type and rel is used to perform the mapping.

```json5
{
  "exceptions": {
    "links": [
      {
        "type": "x-im/articlesource",
        "rel": [],
        "section": ["contentmeta"]
      },
      {
        "type": "x-im/category",
        "rel": ["category"],
        "section": ["contentmeta"]
      }
    ],
  }
}
```

Links refers only to links that are not under other elements, e.g. //itemMeta/links annd not //contentMeta/metadata/objects/links. The
section values for links are _contentmeta_, _concept_, and _planning_

The section values for objects are _contentmeta_, _concept_, and _planning_. Sections are context-sensitive, so if working with a Concept, for example, exceptions having section _planning_ would not apply.

##### Default Link Exceptions

| Type | Rel | Section | Description |
|------|---|--------|---|
|x-im/articlesource| |contentmeta| |
|x-im/premium| |contentmeta| |
|x-im/articletype| |contentmeta| |
|x-geo/point| |contentmeta| |
|x-im/contenttype| |contentmeta| |
|x-im/articlecontent| |contentmeta| |
|x-im/articlecontenttype| |contentmeta| |
|x-im/plus| |contentmeta| |
|x-im/articleoptions| |contentmeta| |
|x-im/articleoptions/plus| |contentmeta| |
|x-im/category| category | planning| Context planning & rel=category |
|x-im/category| category | contentmeta| Others & rel=category, else itemMeta |
|x-im/articleoptions/comments| comment | contentmeta| Only if rel=comment, else itemMeta|

Example: by default, a link with type x-im/premuim is stored in contentMeta. To have it stored in itemMeta in newsitem documents the following would be used

```json5
{
  "link-exceptions": [
    {
      "type": "x-im/premium",
      "rel": [],
      "section": [
        "itemmeta-newsitem"
      ]
    }
  ]
}

```
#### Property Exceptions (property-exceptions)

Properties in NavigaDoc may transform to itemMetaExtProperty, contentMetaExtProperty or directly to XML elements. By default, properties from navigaDoc get stored as itemMeta/itemMetaExtProperty. The section values for properties are _contentmeta_, _concept_, and _planning_ 

```json5
{
  "exceptions": {
    "property-exceptions": [
      {
        "type": "test-property",
        "rel": [],
        "section": [
          "contentmeta"
        ]
      },
      {
        "type": "asn:property",
        "rel": [],
        "section": [
          "planning"
        ]
      }
    ]
  }
}
```
##### Default Property Exceptions

| Type | Section | Description |
|---|---|---|
|uri|contentmeta| |
|altId|contentmeta| |
|infosource|contentmeta| |
|contentcreated|contentmeta| |
|contentmodified|contentmeta| |
|type|contentmeta| |
|by|contentmeta| |
|headline|contentmeta| |
|description|concept| |
|description|contentmeta| |
|definition|concept| |
|slugline|planning| |
|slugline|contentmeta| |
|provider|contentmeta| |
|nrp:sector|concept| |
|nrp:sector|contentmeta| |
|imext:header|contentmeta| |
|imext:subheader|contentmeta| |
|imext:deck|contentmeta| |
|imext:simplebyline|contentmeta| |
|copyrightholder|planning| |
|copyrightholder|concept| |
|headline|planning| |
|imext:headline|concept| |
|imext:slugline|concept| |
|urgency|concept| |
|urgency|planning| |
|concepttypes|concept| |
|imext:qcode|concept| |
|nrptype:evtyp|concept| |
|infoSource|contentmeta| |
|language|contentmeta| |

Example: by default, imext:simplebyline is stored under contentMeta. To store it under itemMeta the following would be used

```json5
{
  "property-exceptions": [
    {
      "type": "imext:simplebyline",
      "section": [
        "itemmeta-newsitem"
      ]
    }
  ]
}
```

#### Object Exceptions (object-exceptions)

by default, objects from NavigaDoc are stored under contentMeta/metadata.

For NewsML, objects may be under itemMeta or contentMeta, with contentMeta being the most prevalent.

For Concepts, the <concept> element of the Concept XML has <metadata> as does the <contentMeta> element. This metadata is stored in the Navigadoc under _meta_. The default is to store the Navigadoc _meta_ under _contentMeta/metadata_, and the object exceptions are used to designate objects that belong under _content/metadata_.

```json5
{
  "object-exceptions": [
    {
      "type": "x-im/contact-info",
      "section": [
        "concept"
      ]
    },
    {
      "type": "x-im/position",
      "section": [
        "concept"
      ]
    }
  ]
}
```
##### Default Object Exceptions

The internal default object Exceptions are as follows, all of which are related to Concepts.

|Type|Rel|Section|Description|
|---|---|---|---|
|x-im/contact-info| |concept| |
|x-im/position | |concept| |
|x-im/event-details | |concept| |
|cpnat:person | |conceptitem| |
|cpnat:object | |conceptitem| |
|cpnat:event | |conceptitem| |
|cpnat:abstract | |conceptitem| |
|cpnat:organisation | |conceptitem| |
|cpnat:poi | |conceptitem| |
|x-im/event | |conceptitem| |
|x-im/polygon | |concept| |

Example: by default, an object with type x-tt/internalnote would be stored under contentMeta/metadata. To store it under itemMeta/metadata  

```json5
{
  "object-exceptions": [
    {
      "type": "x-tt/internalnote",
      "rel": [],
      "section": [
        "itemmeta-newsitem"
      ]
    }
  ]
}
```

### Data Conversions (data-conversions)

The _data-conversions_ section is used to define handling of custom types

The _type_ member is the value of the type attribute
The _destination_ member defines where in the Navigadoc to place the data, the allowable values being "link" and "meta"
The _data-type_ member defines special handling for <data> elements. The allowable values are blank, "idf" and "blob"

By default the <data> children are extracted as XML
Use "idf" when the <data> children are comprised of IDF <element> and <object> elements
Use "blob" when the <data> child elements are complex, having multiple child layers, or where child elements have attributes

#### Examples for XML (data-type blank or not supplied)

In this example the data is simple and requires no special configuration. There is only one level of children being simple elements with no attributes.

```xml
<object id="metadata-d9223e1" type="ford/fast-metadata">
    <data>
        <storytype>news</storytype>
        <category>sports</category>
        <subject>Sports</subject>
    </data>
</object>
```

In this example the category element is duplicated and has an attribute

```xml
<object id="metadata-d9223e1" type="ford/fast-metadata">
    <data>
        <storytype>news</storytype>
        <category primary="false">
            <name>EU</name>
            <slug>eu</slug>
        </category>
        <category primary="true">
            <name>Football</name>
            <slug>football</slug>
        </category>
        <subject>Sports</subject>
    </data>
</object>
```

The custom configuration for the above would be as follows.

```json5
{
  "data-conversions": [
    {
      "type": "ford/fast-metadata",
      "destination": "link",
      "elements": [
        {
          "name": "category",
          "type": "ford/fast-metadata",
          "rel": "category",
          "attributes": [
            "primary"
          ]
        }
      ]
    }
  ]

}
```

The Navigadoc would be as follows, with the category values stored under _links_.

```json5
{
  "content": [
    {
      "id": "metadata-d9223e1",
      "type": "ford/fast-metadata",
      "data": {
        "storytype": "news",
        "subject": "Sports",
      },
      "links": [
        {
          "type": "ford/fast-metadata",
          "data": {
            "name": "EU",
            "primary": "false",
            "slug": "eu"
          },
          "rel": "category",
          "name": "category"
        },
        {
          "type": "ford/fast-metadata",
          "data": {
            "name": "Football",
            "primary": "true",
            "slug": "policyfootball"
          },
          "rel": "category",
          "name": "category"
        }
      ]
    }
  ]
}
```

#### Example for "idf"

In this example the IDF <object> data contains IDF elements

Newsml

```xml
  <object id="sidebar-d2643e115" type="ford/sidebar">
      <data>
          <element id="subheadline-d2643e117" type="subheadline">SubHead</element>
          <element id="paragraph-d2643e120" type="body">First paragraph</element>
          <element id="paragraph-d2643e123" type="body">Second paragraph</element>
          <element id="paragraph-d2643e126" type="body">Third paragraph</element>
      </data>
  </object>
```

Configuration

```json5
{
  "data-conversions": [
    {
      "type": "ford/sidebar",
      "destination": "link",
      "data-type": "idf",
      "elements": []
    }
  ]
}
```

Navigadoc

```json5
{
  "content": [
    {
      "id": "sidebar-d2643e115",
      "type": "ford/sidebar",
      "links": [
        {
          "id": "subheadline-d2643e117",
          "type": "subheadline",
          "data": {
            "format": "html",
            "text": "SubHead"
          }
        },
        {
          "id": "paragraph-d2643e120",
          "type": "x-im/paragraph",
          "data": {
            "format": "html",
            "text": "First paragraph"
          }
        },
        {
          "id": "paragraph-d2643e123",
          "type": "x-im/paragraph",
          "data": {
            "format": "html",
            "text": "Second paragraph"
          }
        },
        {
          "id": "paragraph-d2643e126",
          "type": "x-im/paragraph",
          "data": {
            "format": "html",
            "text": "Third paragraph"
          }
        }
      ]
    }
  ]
}
```

#### Example for "blob"

In this example the data is complex with multiple levels of children, some having attributes

Newsml

```xml
<object id="table-d318949e26" type="ford/table">
    <data>
        <table format="html">
            <tr id="tr-d318949e38">
                <td id="td-d318949e41"><a href="">data</a></td>
            </tr>
            <tr id="tr-d318949e60">
                <td id="td-d318949e65"><strong>data</strong></td>
            </tr>
        </table>
    </data>
</object>
```

Configuration

```json5
{
  "data-conversions": [
    {
      "type": "ford/table",
      "destination": "link",
      "data-type": "blob",
      "elements": []
    }
  ]
}
```

Navigadoc

```json5
{
  "content": [
    {
      "id": "table-d318949e26",
      "type": "ford/table",
      "links": [
        {
          "data": {
            "format": "xml",
            "text": "<table format=\"html\"><tr id=\"tr-d318949e38\"><td id=\"td-d318949e41\"><a href=\"\">data</a></td></tr><tr id=\"tr-d318949e60\"><td id=\"td-d318949e65\"><strong>data</strong></td></tr></table>"
          }
        }
      ]
    }
  ]
}
```

### Date Validation (date-elements)

The _date-elements_ section is used to define rules for NewsML elements and Navigadoc fields which should be validated as dates. 

The _date-elements_ contains three sub-sections _tags_, _types_ and _blocks_.

The _tags_ section defines which XML tag names are date values.

The _types_ section defines for XML elements having a _type_ attribute which are date values.

The _blocks_ section defines for Navigadoc which Block fields are date values.

The _allow-blank_ member defines whether the value is allowed to be an empty string.

The _allow-string_ member defines whether the value may be a string, such as "future"

The _format_ member defines the date format. It may be one of "RFC3339", "RFC3339Nano" or a regex pattern may be used for other dates and times, e.g. "^([0-9]|0[0-9]|1[0-9]|2[0-3]):([0-9]|[0-5][0-9]):([0-9]|[0-5][0-9])$" for HH:MM:SS. 

For _types_ the default is to look for an attribute named _value_ with the date. The member _use-attribute_ may be defined to provide the name of another attribute that holds the date value.

```json5
{
  "date-elements": {
    "tags": {
      "firstCreated": {
        "allow-blank": false,
        "allow-string": false,
        "format": "RFC3339"
      }
    },
    "types": {
      "imext:pubstart": {
        "allow-blank": true,
        "allow-string": false,
        "format": "RFC3339",
        "use-attribute": "value"
      }
    },
    "blocks": {
      "Modified": {
        "allow-blank": false,
        "allow-string": false,
        "format": "RFC3339",
        "use-attribute": "value"
      }
    }
  }
}
```

#### HTML Sanitize Options (html-sanitize-options)

For list types, i.e. x-im/ordered-list, doc-format use the package blue-monday to sanitize HTML in the list data. The _html-sanitize-options_ section allows for setting some bluemonday policy options in the configuration.

The _allow-standard-attributes_ setting permits or denies the standard HTML attributes globally

The _allow-images_ setting permits or denies the "img" element and its attributes

The _allow-lists_ setting permits or denies lists

The _allow-tables_ setting permits or denies tables

The _allow-relative-urls_ setting permits or denies relative urls in hrefs

The _allowed-url-schemes_ setting whitelists which protocols are allowed in hrefs

The _allowed-elements-attributes_ setting defines which elements and attributes are allowed

```json5
{
  "html-sanitize-options": {
    "allow-standard-attributes": true,
    "allow-images": true,
    "allow-lists": true,
    "allow-tables": true,
    "allow-relative-urls": false,
    "allowed-url-schemes": "http,https,mailto",
    "allowed-elements-attributes": [
      {
        "element": "strong",
        "attributes": "id,type"
      },
      {
        "element": "em",
        "attributes": ""
      },
      {
        "element": "mark",
        "attributes": ""
      },
      {
        "element": "a",
        "attributes": "href,rel,target"
      },
      {
        "element": "company",
        "attributes": "id,tag,tagId"
      }
    ]
  }
}
```

## Example NavigaDoc to Infomaker NewsML (IMNML)

```go

    import (
        "fmt"
	    "io/ioutil"

	    df "bitbucket.org/infomaker/doc-format"
	    "bitbucket.org/infomaker/doc-format/newsml"
	    "bitbucket.org/infomaker/doc-format/doc"
	    "github.com/golang/protobuf/jsonpb"
    )
    
	jsonIn, _ := ioutil.ReadFile("navigadoc-example.json")

    // Unmarshal the NavigaDoc into doc.Document
    navigaDoc := doc.Document{}
    err := jsonpb.UnmarshalString(string(jsonIn), &navigaDoc)
	if err != nil {
		panic(err)
    }

    // Convert to newsml.NewsItem
    opts := newsml.DefaultOptions()
    newsitem, err := df.NavigaDocToNewsItem(&navigaDoc, &opts)
	if err != nil {
		panic(err)
	}

```

## Example Infomaker NewsML (IMNML) to NavigaDoc

```go

    import (
        df "bitbucket.org/infomaker/doc-format"
        "bitbucket.org/infomaker/doc-format/newsml"
    )

	xmlIn, _ := ioutil.ReadFile("newsitem-example.xml")

    // Unmarshal the newsml.NewsItem into doc.Document
	newsItem := newsml.NewsItem{}
	err = xml.Unmarshal(xmlIn, &newsItem)
	if err != nil {
		panic(err)
	}

    // Convert to NavigaDoc JSON
    opts := newsml.DefaultOptions()
	navigaDoc, err := df.NewsItemToNavigaDoc(&newsItem, &opts)
	if err != nil {
		panic(err)
	}



```

