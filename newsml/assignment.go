package newsml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type Assignment struct {
	XMLName             xml.Name           `xml:"planning"`
	XMLNamespace        string             `xml:"xmlns,attr,omitempty"`
	Conformance         string             `xml:"conformance,attr,omitempty"`
	GUID                string             `xml:"guid,attr,omitempty"`
	Standard            string             `xml:"standard,attr,omitempty"`
	StandardVersion     string             `xml:"standardversion,attr,omitempty"`
	Version             string             `xml:"version,attr,omitempty"`
	ItemClass           *ItemClass         `xml:"itemClass,omitempty"`
	Headline            string             `xml:"headline,omitempty"`
	Description         []*Description     `xml:"description,omitempty"`
	PlanningExtProperty []*MetaExtProperty `xml:"planningExtProperty,omitempty"`
	Links               NSLinks            `xml:"links"`
}

func (a *Assignment) validateFields(opts *Options) error {
	if a.ItemClass == nil || a.ItemClass.QCode == "" {
		return errors.New("missing ItemClass.QCode")
	} else if _, err := qcodeToSubtype(a.ItemClass.QCode, opts); err != nil {
		return err
	}
	return nil
}

func (a *Assignment) fromDoc(document *doc.Document, opts *Options) error {
	if document == nil {
		return errors.New("nil document")
	}
	a.XMLName = xml.Name{
		Local: "planning",
	}
	a.XMLNamespace = "http://iptc.org/std/nar/2006-10-01/"
	a.GUID = document.UUID
	a.Headline = document.Title
	a.Conformance = "power"
	a.Standard = "NewsML-G2"
	a.StandardVersion = "2.26"
	a.Version = "1"

	a.PlanningExtProperty = append(a.PlanningExtProperty,
		&MetaExtProperty{
			Type:  "imext:type",
			Value: document.Type,
		})

	a.PlanningExtProperty = append(a.PlanningExtProperty,
		&MetaExtProperty{
			Type:  "imext:status",
			Value: document.Status,
		})

	if document.Created != nil {
		a.PlanningExtProperty = append(a.PlanningExtProperty,
			&MetaExtProperty{
				Type:  "nrpdate:created",
				Value: document.Created.Format(time.RFC3339Nano),
			})
	}

	if document.Modified != nil {
		a.PlanningExtProperty = append(a.PlanningExtProperty,
			&MetaExtProperty{
				Type:  "nrpdate:modified",
				Value: document.Modified.Format(time.RFC3339Nano),
			})
	}

	for _, property := range document.Properties {
		if hasPropertyExceptionType(property.Name, "planning", opts) {
			a.PlanningExtProperty = append(a.PlanningExtProperty,
				&MetaExtProperty{
					Type:  property.Name,
					Value: property.Value,
				})
		}
	}

	for _, link := range document.Links {
		xmlLink, err := buildXMLLink(link, opts)
		if err != nil {
			return err
		}
		a.Links = append(a.Links, xmlLink)
	}

	if document.Meta != nil {
		for _, meta := range document.Meta {
			if meta.Type == "x-im/assignment" && meta.Data != nil {
				for key, data := range meta.Data {
					switch key {
					case "type":
						qcode, err := subtypeToQCode(data, opts)
						if err != nil {
							return err
						}
						if a.ItemClass == nil {
							a.ItemClass = &ItemClass{}
						}
						a.ItemClass.QCode = qcode
					case "start":
						if a.PlanningExtProperty == nil {
							a.PlanningExtProperty = make([]*MetaExtProperty, 1)
						}
						a.PlanningExtProperty = append(a.PlanningExtProperty,
							&MetaExtProperty{
								Type:  "nrpdate:start",
								Value: data,
							})
					case "end":
						if a.PlanningExtProperty == nil {
							a.PlanningExtProperty = make([]*MetaExtProperty, 1)
						}
						a.PlanningExtProperty = append(a.PlanningExtProperty,
							&MetaExtProperty{
								Type:  "nrpdate:end",
								Value: data,
							})
					case "description":
						if a.Description == nil {
							a.Description = []*Description{}
						}
						a.Description = append(a.Description, &Description{
							Text: data,
							Role: "nrpdesc:intern",
						})
					case "publicDescription":
						if a.Description == nil {
							a.Description = []*Description{}
						}
						a.Description = append(a.Description, &Description{
							Text: data,
							Role: "nrpdesc:extern",
						})
					}
				}
			}
		}
	}

	return nil
}

func (a *Assignment) toDoc(document *doc.Document, opts *Options) error {
	if err := a.validateFields(opts); err != nil {
		return err
	}
	document.UUID = a.GUID
	document.Properties = []doc.Property{}
	document.Type = "x-im/assignment"
	document.Title = a.Headline

	var startDate string
	var endDate string
	for _, prop := range a.PlanningExtProperty {
		switch prop.Type {
		case "imext:status":
			document.Status = prop.Value
		case "nrpdate:created":
			created, err := time.Parse(time.RFC3339, prop.Value)
			if err != nil {
				return err
			}
			document.Created = &created
		case "nrpdate:modified":
			modified, err := time.Parse(time.RFC3339, prop.Value)
			if err != nil {
				return err
			}
			document.Modified = &modified
		case "nrpdate:start":
			startDate = prop.Value
		case "nrpdate:end":
			endDate = prop.Value
		default:
			a.addPropertyToDoc(document, prop.Type, prop.Value)
		}
	}

	if len(a.Links) != 0 {
		jLink := doc.Block{}
		err := buildDocLinks(&jLink, a.Links, opts, LinkContext)
		if err != nil {
			return err
		}
		document.Links = append(document.Links, jLink.Links...)
	}

	subtype, err := qcodeToSubtype(a.ItemClass.QCode, opts)
	if err != nil {
		return err
	}
	typeMeta := doc.Block{
		Type: "x-im/assignment",
		Data: map[string]string{
			"type":  subtype,
			"start": startDate,
			"end":   endDate,
		},
	}

	if a.Description != nil {
		for _, descr := range a.Description {
			switch descr.Role {
			case "nrpdesc:intern":
				typeMeta.Data["description"] = descr.Text
			case "nrpdesc:extern":
				typeMeta.Data["publicDescription"] = descr.Text
			}
		}
	}

	if document.Meta == nil {
		document.Meta = make([]doc.Block, 0)
	}
	document.Meta = append(document.Meta, typeMeta)

	return nil
}

func qcodeToSubtype(qcode string, opts *Options) (string, error) {
	for _, q2s := range opts.AssignmentQcodeType {
		if q2s.Qcode == qcode {
			return q2s.Type, nil
		}
	}

	return "", fmt.Errorf("missing subtype for qcode %s", qcode)
}

func subtypeToQCode(subtype string, opts *Options) (string, error) {
	for _, q2s := range opts.AssignmentQcodeType {
		if q2s.Type == subtype {
			return q2s.Qcode, nil
		}
	}

	return "", fmt.Errorf("missing qcode for subtype %s", subtype)
}

func (a *Assignment) NavigaDocToAssignment(document *doc.Document, opts *Options) error {
	err := a.fromDoc(document, opts)
	if err != nil {
		return err
	}

	return nil
}

func (a *Assignment) AssignmentToNavigaDoc(assignment *Assignment, opts *Options) (*doc.Document, error) {
	document := &doc.Document{}

	err := assignment.toDoc(document, opts)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func (a *Assignment) addPropertyToDoc(document *doc.Document, name string, value string) {
	if value != "" {
		document.Properties = append(document.Properties, doc.Property{Name: name, Value: value})
	}
}
