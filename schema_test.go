package docformat_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"bitbucket.org/infomaker/doc-format/v2/doc"

	docformat "bitbucket.org/infomaker/doc-format/v2"
	"bitbucket.org/infomaker/doc-format/v2/newsml"
)

func TestValidateNavigadoc(t *testing.T) {
	testData, err := ioutil.ReadFile(configFile)
	must(t, err, "could not open file")
	customConfig := string(testData)

	var tests = []TestData{
		{xml: "examples/full-article.xml", customConfig: customConfig, expectError: false},
		{xml: "examples/full-picture.xml", customConfig: customConfig, expectError: false},
		{xml: "examples/article-template.xml", customConfig: customConfig, expectError: false},
		{xml: "examples/example-pdf.xml", customConfig: customConfig, expectError: false},
		{xml: "examples/example-text.xml", customConfig: customConfig, expectError: false},
		// FIXME? example-wire.xml winds up with properties missing value, so removed from required list
		{xml: "examples/example-wire.xml", customConfig: customConfig, expectError: false},
		{json: "examples/cca-example.json", expectError: false},
		{json: "testdata/empty-blocks-example.json", expectError: false},
		{json: "testdata/empty-document-example.json", expectError: true},
		{json: "testdata/custom-config.json", expectError: true},
		{json: "testdata/uuids-uppercase.json", expectError: false},
	}

	opts := newsml.DefaultOptions()
	for i := range tests {
		test := tests[i]
		if test.xml != "" {
			t.Run(test.xml, func(t *testing.T) {
				testXML, err := ioutil.ReadFile(test.xml)
				must(t, err, "could not read file")

				var original newsml.NewsItem
				err = xml.Unmarshal(testXML, &original)
				must(t, err, "could not unmarshal file")

				if test.customConfig != "" {
					customConfig := newsml.Options{}

					err = json.Unmarshal([]byte(test.customConfig), &customConfig)
					must(t, err, "could not unmarshal file")

					docformat.MergeOptions(&opts, &customConfig)
				}

				navigadoc, err := docformat.NewsItemToNavigaDoc(&original, &opts)
				if test.expectError && err != nil {
					return
				}
				must(t, err, "failed NewsItemToNavigaDoc")

				docBytes, err := json.MarshalIndent(navigadoc, "", " ")
				must(t, err, "failed marshal")

				docText := string(docBytes)
				errs, err := docformat.ValidateNavigadocJSON(docText)
				if err != nil || len(errs) > 0 || testing.Verbose() {
					if err != nil || len(errs) > 0 {
						t.Errorf("unexpected errors %s", test.xml)
					}
					dumpDocs(t, navigadoc, nil, testXML, nil, errs...)
				}

				// Test the UUID pattern
				docUUID := navigadoc.UUID
				navigadoc.UUID = "JKjhsnajuSH"
				docBytes, err = json.MarshalIndent(navigadoc, "", " ")
				must(t, err, "failed marshal")

				test.expectError = true
				errs, err = docformat.ValidateNavigadocJSON(string(docBytes))
				if err == nil && len(errs) == 0 {
					t.Errorf("expected error on bad uuid %s", test.xml)
				}

				navigadoc.UUID = docUUID

				// Test the date-time
				rgxp := regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}")
				dates := rgxp.FindAllString(docText, -1)
				badJSON := docText
				for _, d := range dates {
					badJSON = strings.ReplaceAll(badJSON, d, "xxxx-xx:xx:xxT")
					break
				}
				errs, err = docformat.ValidateNavigadocJSON(badJSON)
				if err == nil && len(errs) == 0 {
					t.Errorf("expected error on bad date %s", test.xml)
				}

				// Test required ID
				rgxp = regexp.MustCompile("\"type\".*?:.*?\".*?\".*,")
				types := rgxp.FindAllString(docText, -1)
				badJSON = docText
				for _, t := range types {
					badJSON = strings.ReplaceAll(badJSON, t, "")
					break
				}
				errs, err = docformat.ValidateNavigadocJSON(badJSON)
				if err == nil && len(errs) == 0 {
					t.Errorf("expected error on required id %s", test.xml)
				}
			})
		}

		if test.json != "" {
			t.Run(test.json, func(t *testing.T) {
				testJSON, err := ioutil.ReadFile(test.json)
				must(t, err, "could not read file")

				errs, err := docformat.ValidateNavigadocJSON(string(testJSON))
				if !test.expectError && err != nil || len(errs) > 0 {
					navigadoc := doc.Document{}
					err := json.Unmarshal(testJSON, &navigadoc)
					if err != nil {
						errs = append(errs, fmt.Errorf("error unmarshaling %s", test.json))
						navigadoc = doc.Document{}
					}
					dumpDocs(t, &navigadoc, nil, nil, nil, errs...)
				}
			})
		}
	}
}
