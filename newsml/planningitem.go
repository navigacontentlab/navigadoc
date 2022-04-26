package newsml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type PlanningItem struct {
	XMLName         xml.Name         `xml:"planningItem"`
	XMLNamespace    string           `xml:"xmlns,attr,omitempty"`
	Conformance     string           `xml:"conformance,attr,omitempty"`
	GUID            string           `xml:"guid,attr,omitempty"`
	Standard        string           `xml:"standard,attr,omitempty"`
	StandardVersion string           `xml:"standardversion,attr,omitempty"`
	Version         string           `xml:"version,attr,omitempty"`
	CatalogRef      []CatalogRef     `xml:"catalogRef,omitempty"`
	RightsInfo      *RightsInfo      `xml:"rightsInfo,omitempty"`
	ItemMeta        *ItemMeta        `xml:"itemMeta,omitempty"`
	ContentMeta     *ContentMeta     `xml:"contentMeta,omitempty"`
	NewsCoverageSet *NewsCoverageSet `xml:"newsCoverageSet,omitempty"`
}

func (pi *PlanningItem) fromDoc(document *doc.Document, opts *Options) error {
	if document == nil {
		return errors.New("nil document")
	}
	pi.XMLName = xml.Name{
		Local: "planningItem",
	}
	pi.XMLNamespace = "http://iptc.org/std/nar/2006-10-01/"
	pi.GUID = document.UUID
	pi.Conformance = "power"
	pi.Standard = "NewsML-G2"
	pi.StandardVersion = "2.26"
	pi.Version = "1"
	pi.CatalogRef = []CatalogRef{
		{
			Href: "http://www.iptc.org/std/catalog/catalog.IPTC-G2-Standards_30.xml",
		},
		{
			Href: "http://infomaker.se/spec/catalog/catalog.infomaker.g2.1_0.xml",
		},
	}

	for _, property := range document.Properties {
		if strings.EqualFold(property.Name, "copyrightholder") {
			pi.RightsInfo = &RightsInfo{
				CopyRightHolder: &CopyRightHolder{
					Name: property.Value,
				},
			}
		}
	}

	pi.ItemMeta = &ItemMeta{
		ItemMetaExtProperty: make([]MetaExtProperty, 0),
	}
	err := pi.ItemMeta.FromDoc(document, opts)
	if err != nil {
		return err
	}

	// Bypass qcodeToType as itemMeta map can't have every type
	pi.ItemMeta.ItemClass = &ItemClass{
		QCode: "plinat:newscoverage",
	}

	// TODO Which to use for Created and Modified?
	if document.Created != nil {
		pi.ItemMeta.FirstCreated = document.Created.Format(time.RFC3339)
	}

	if document.Modified != nil {
		pi.ItemMeta.VersionCreated = document.Modified.Format(time.RFC3339)
	}

	if document.Created != nil {
		pi.ItemMeta.ItemMetaExtProperty = append(pi.ItemMeta.ItemMetaExtProperty, MetaExtProperty{
			Type:  "imext:created",
			Value: document.Created.Format(time.RFC3339Nano),
		})
	}

	if document.Modified != nil {
		pi.ItemMeta.ItemMetaExtProperty = append(pi.ItemMeta.ItemMetaExtProperty, MetaExtProperty{
			Type:  "imext:modified",
			Value: document.Modified.Format(time.RFC3339Nano),
		})
	}

	// Fix type names
	var typeMap = make(map[string]string)
	for _, ptype := range opts.PropertyType {
		for _, section := range ptype.Section {
			if section == "planning" {
				typeMap[ptype.NavigaDoc] = ptype.XML
			}
		}
	}
	// TODO Remove since prefixes are no longer stripped?
	for i := 0; i < len(pi.ItemMeta.ItemMetaExtProperty); i++ {
		im := &pi.ItemMeta.ItemMetaExtProperty[i]
		if fix, ok := typeMap[im.Type]; ok {
			im.Type = fix
		}
	}

	var xtraFields []string
	for _, prop := range opts.PropertyException {
		for _, section := range prop.Section {
			if section == "planning" {
				xtraFields = append(xtraFields, prop.Type)
			}
		}
	}
	// Remove itemMetaExtProperty with the above types
	for i := len(pi.ItemMeta.ItemMetaExtProperty) - 1; i >= 0; i-- {
		im := &pi.ItemMeta.ItemMetaExtProperty[i]
		for _, fv := range xtraFields {
			if strings.EqualFold(fv, im.Type) {
				pi.ItemMeta.ItemMetaExtProperty = append(pi.ItemMeta.ItemMetaExtProperty[:i], pi.ItemMeta.ItemMetaExtProperty[i+1:]...)
				i = len(pi.ItemMeta.ItemMetaExtProperty)
				break
			}
		}
	}

	// Remove ItemMeta links with type x-im/assignment
	for i := len(pi.ItemMeta.Links) - 1; i >= 0; i-- {
		l := pi.ItemMeta.Links[i]
		if l.Type == "x-im/assignment" {
			pi.ItemMeta.Links = append(pi.ItemMeta.Links[:i], pi.ItemMeta.Links[i+1:]...)
			i = len(pi.ItemMeta.Links)
		}
	}

	pi.ContentMeta = &ContentMeta{}
	err = pi.ContentMeta.fromDoc(document, opts)
	if err != nil {
		return err
	}
	pi.ContentMeta.MetaData = nil

	for _, meta := range document.Meta {
		if meta.Type == "x-im/newscoverage" && meta.Data != nil {
			for key, value := range meta.Data {
				switch key {
				case "slug":
					pi.ContentMeta.Slugline = value
				case "priority":
					pi.ContentMeta.Urgency = value
				case "start":
					fallthrough
				case "end":
					if pi.ItemMeta.ItemMetaExtProperty == nil {
						pi.ItemMeta.ItemMetaExtProperty = make([]MetaExtProperty, 0)
					}
					pi.ItemMeta.ItemMetaExtProperty = append(
						pi.ItemMeta.ItemMetaExtProperty, MetaExtProperty{
							Type:  fmt.Sprintf("nrpdate:%s", key),
							Value: value,
							Why:   fmt.Sprintf("nrpwhy:%s", meta.Data["dateGranularity"]),
						})
				case "description":
					if pi.ContentMeta.Description == nil {
						pi.ContentMeta.Description = make([]*Description, 0)
					}
					pi.ContentMeta.Description = append(
						pi.ContentMeta.Description,
						&Description{
							Role: "nrpdesc:intern",
							Text: value,
						})
				case "publicDescription":
					if pi.ContentMeta.Description == nil {
						pi.ContentMeta.Description = make([]*Description, 0)
					}
					pi.ContentMeta.Description = append(
						pi.ContentMeta.Description,
						&Description{
							Role: "nrpdesc:extern",
							Text: value,
						})
				}
			}
		}
	}

	pi.NewsCoverageSet = &NewsCoverageSet{}
	err = pi.NewsCoverageSet.fromDoc(document, opts)
	if err != nil {
		return err
	}

	for _, newsCoverage := range pi.NewsCoverageSet.NewsCoverage {
		for i := len(newsCoverage.Links) - 1; i >= 0; i-- {
			link := newsCoverage.Links[i]
			for _, le := range opts.LinkException {
				if le.Type == link.Type {
					for _, s := range le.Section {
						if s != "planning" {
							newsCoverage.Links = append(newsCoverage.Links[:i], newsCoverage.Links[i+1:]...)
							i = len(newsCoverage.Links)
						}
					}
				}
			}
		}
	}
	return nil
}

func (pi *PlanningItem) toDoc(document *doc.Document, opts *Options) error {
	document.UUID = pi.GUID
	document.Properties = []doc.Property{}

	if pi.RightsInfo != nil {
		pi.addPropertyToDoc(document, "copyrightholder", pi.RightsInfo.CopyRightHolder.Name)
	}

	if pi.ItemMeta != nil {
		err := pi.ItemMeta.ToDoc(document, opts)
		if err != nil {
			return err
		}
	}

	if pi.ContentMeta != nil {
		err := pi.ContentMeta.toDoc(document, opts)
		if err != nil {
			return err
		}
	}

	// Remove extProperties which get mapped below to meta
	var xtraFields = []string{
		"nrpdate:start",
		"nrpdate:end",
		"nrpdate:created",
		"nrpdate:modified",
		"description",
		"slugline",
		"urgency",
	}
	for i := len(document.Properties) - 1; i >= 0; i-- {
		p := document.Properties[i]
		for _, fv := range xtraFields {
			if strings.EqualFold(fv, p.Name) {
				document.Properties = append(document.Properties[:i], document.Properties[i+1:]...)
				i = len(document.Properties)
				break
			}
		}
	}

	meta := doc.Block{
		Type: "x-im/newscoverage",
		Data: make(map[string]string),
	}
	meta.Data["priority"] = pi.ContentMeta.Urgency
	meta.Data["slug"] = pi.ContentMeta.Slugline

	if pi.ContentMeta.Description != nil {
		for _, descr := range pi.ContentMeta.Description {
			switch descr.Role {
			case "nrpdesc:intern":
				meta.Data["description"] = descr.Text
			case "nrpdesc:extern":
				meta.Data["publicDescription"] = descr.Text
			}
		}
	}

	for _, prop := range pi.ItemMeta.ItemMetaExtProperty {
		switch prop.Type {
		case "nrpdate:start":
			meta.Data["start"] = prop.Value
			meta.Data["dateGranularity"] = strings.TrimPrefix(prop.Why, "nrpwhy:")
		case "nrpdate:end":
			meta.Data["end"] = prop.Value
			meta.Data["dateGranularity"] = strings.TrimPrefix(prop.Why, "nrpwhy:")
		case "nrpdate:created":
			createdTime, err := time.Parse(time.RFC3339, prop.Value)
			if err != nil {
				return fmt.Errorf("failed to parse created: %s", prop.Value)
			}
			document.Created = &createdTime
		case "nrpdate:modified":
			modifiedTime, err := time.Parse(time.RFC3339, prop.Value)
			if err != nil {
				return fmt.Errorf("failed to parse modified: %s", prop.Value)
			}
			document.Modified = &modifiedTime
		}
	}

	document.Meta = append(document.Meta, meta)

	if pi.NewsCoverageSet != nil {
		err := pi.NewsCoverageSet.toDoc(document, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pi *PlanningItem) NavigaDocToPlanningItem(document *doc.Document, opts *Options) error {
	err := pi.fromDoc(document, opts)
	if err != nil {
		return err
	}

	return nil
}

func (pi *PlanningItem) PlanningItemToNavigaDoc(planningItem *PlanningItem, opts *Options) (*doc.Document, error) {
	document := &doc.Document{}

	err := planningItem.toDoc(document, opts)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func (pi *PlanningItem) addPropertyToDoc(document *doc.Document, name string, value string) {
	if value != "" {
		document.Properties = append(document.Properties, doc.Property{Name: name, Value: value})
	}
}
