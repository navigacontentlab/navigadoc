package newsml

import (
	"errors"
	"fmt"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type NewsCoverageSet struct {
	NewsCoverage []*NewsCoverage `xml:"newsCoverage,omitempty"`
}

type NewsCoverage struct {
	ID    string  `xml:"id,attr,omitempty"`
	Links NSLinks `xml:"links"`
}

func (ncs *NewsCoverageSet) fromDoc(document *doc.Document, opts *Options) error {
	if document == nil {
		return errors.New("nil document")
	}

	for _, block := range document.Links {
		if block.Type != "x-im/assignment" {
			continue
		}

		nc := &NewsCoverage{
			ID: block.ID,
		}

		for _, link := range block.Links {
			if link.Type == "x-im/imchn" || link.Type == "x-im/articlesource" {
				continue
			}

			xmlLink, err := buildXMLLink(link, opts)
			if err != nil {
				return err
			}
			nc.Links = append(nc.Links, xmlLink)
		}

		ncs.NewsCoverage = append(ncs.NewsCoverage, nc)
	}

	return nil
}

func (ncs *NewsCoverageSet) toDoc(document *doc.Document, opts *Options) error {
	for i, nc := range ncs.NewsCoverage {
		block := doc.Block{
			Type: "x-im/assignment",
			ID:   nc.ID,
		}

		err := buildDocLinks(&block, nc.Links, opts, LinkContext)
		if err != nil {
			return fmt.Errorf("failed to convert newscoverage %d: %w", i, err)
		}

		document.Links = append(document.Links, block)
	}

	return nil
}
