package navigadoc_test

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/navigacontentlab/navigadoc"
	"github.com/navigacontentlab/navigadoc/doc"
)

func TestValidateDocumentUUIDs(t *testing.T) {
	tests := []TestData{
		{json: "./testdata/text.json", expectError: false},
		{json: "./testdata/uuids-invalid.json", expectError: true},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.xml, func(t *testing.T) {
			testData, err := ioutil.ReadFile(test.json)
			must(t, err, "could not open testfile")

			document := &doc.Document{}
			err = json.Unmarshal(testData, document)
			must(t, err, "could not unmarshal doc")

			if document.UUID != "" {
				_, err := uuid.Parse(document.UUID)
				if err != nil {
					t.Errorf("invalid document uuid %s", err)
				}

				document.UUID = strings.ToLower(document.UUID)
			}

			err = navigadoc.WalkDocument(document, nil, navigadoc.ValidateAndLowercaseDocumentUUIDs)
			if !test.expectError && err != nil {
				t.Error(err)
			} else if test.expectError && err == nil {
				t.Error("error was expected")
			}
		})
	}
}
