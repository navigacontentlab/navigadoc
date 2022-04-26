package listpackage

import (
	"errors"

	"bitbucket.org/infomaker/doc-format/v2/newsml"

	"bitbucket.org/infomaker/doc-format/v2/doc"
)

type Product struct {
	UUID string `xml:"uuid,attr,omitempty"`
	Text string `xml:",chardata"`
}

func AsDocument(v interface{}, opts *newsml.Options) (*doc.Document, error) {
	if v == nil {
		return nil, errors.New("nil input")
	}

	document := &doc.Document{}

	switch i := v.(type) {
	case *List:
		err := i.toDoc(document, opts)
		if err != nil {
			return nil, err
		}
	case *Package:
		err := i.toDoc(document, opts)
		if err != nil {
			return nil, err
		}
	}

	return document, nil
}
