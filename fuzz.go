//go:build gofuzz
// +build gofuzz

package navigadoc

import (
	"bytes"
	"encoding/xml"

	"bitbucket.org/infomaker/doc-format/v2/newsml"
)

func Fuzz(data []byte) int {
	reader := bytes.NewReader(data)
	dec := xml.NewDecoder(reader)

	var article newsml.NewsItem
	if err := dec.Decode(&article); err != nil {
		return 0
	}

	doc, err := NewsItemToNavigaDoc(&article, GetDefaultConversionOptions())
	if err != nil {
		if doc != nil {
			panic("document should be nil on error")
		}
		return 0
	}

	item, err := NavigaDocToNewsItem(doc, GetDefaultConversionOptions())
	if err != nil {
		if item != nil {
			panic("newsitem should be nil on error")
		}
		return 0
	}

	return 1
}
