package docformat_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"unicode"

	"bitbucket.org/infomaker/doc-format/v2/newsml"

	"github.com/google/uuid"

	"bitbucket.org/infomaker/doc-format/v2/doc"

	docformat "bitbucket.org/infomaker/doc-format/v2"
	"github.com/beevik/etree"
)

func TestValidateNewsMLDates(t *testing.T) {
	var custom newsml.Options

	file, err := os.Open("./testdata/custom-config.json")
	if err != nil {
		t.Fatal("unable to open ./testdata/dates-config.json")
	}

	dec := json.NewDecoder(file)
	err = dec.Decode(&custom)
	if err != nil {
		t.Fatal("unable to decode ./testdata/dates-config.json")
	}

	tests := []TestData{
		{xml: "./testdata/good-dates.xml", expectError: false},
		{xml: "./testdata/bad-dates.xml", expectError: true},
		{xml: "./examples/full-article.xml", expectError: false},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.xml, func(t *testing.T) {
			xmlDoc := etree.NewDocument()
			err := xmlDoc.ReadFromFile(test.xml)
			if err != nil {
				t.Fatal(err)
			}
			err = docformat.ValidateNewsMLDates(xmlDoc, custom.DateElements)
			if !test.expectError && err != nil {
				t.Error(err)
			} else if test.expectError && err == nil {
				t.Error("error was expected")
			}
			err = docformat.ValidateNewsMLDates(xmlDoc, newsml.DefaultOptions().DateElements)
			if !test.expectError && err != nil {
				t.Error(err)
			} else if test.expectError && err == nil {
				t.Error("error was expected")
			}
		})
	}
}

// Deprecated: UUID handling in OpenContent to be case-insensitive
func TestValidateAndLowercaseNewsMLUUIDs(t *testing.T) {
	tests := []TestData{
		{xml: "./examples/full-article.xml", expectError: false},
		{xml: "./testdata/uuids-invalid.xml", expectError: true},
		{xml: "./testdata/uuids-uppercase.xml", expectError: false},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.xml, func(t *testing.T) {
			xmlDoc := etree.NewDocument()
			err := xmlDoc.ReadFromFile(test.xml)
			if err != nil {
				t.Fatal(err)
			}
			err = docformat.FixNewsMLUUIDsAndNamespaces(xmlDoc)
			if !test.expectError && err != nil {
				t.Error(err)
			} else if test.expectError && err == nil {
				t.Error("error was expected")
			}

			if test.expectError && err != nil {
				if errors.Is(err, docformat.InvalidArgumentError{}) {
					t.Logf("InvalidArgumentError %s", err)
				} else if errors.Is(err, docformat.RequiredArgumentError{}) {
					t.Logf("RequiredArgumentError %s", err)
				}
			}

			// Test that all UUIDs are lowercase
			err = docformat.WalkXMLDocumentElements(xmlDoc, nil, func(element *etree.Element, args ...interface{}) error {
				uuidAttr := element.SelectAttr("uuid")
				if element.Tag != "" && uuidAttr == nil {
					return nil
				}

				var uuidValue string
				if uuidAttr != nil {
					uuidValue = uuidAttr.Value
				} else {
					uuidValue = element.Text()
				}

				for _, r := range uuidValue {
					if unicode.IsUpper(r) {
						return fmt.Errorf("uuid contains uppercase: %s", uuidValue)
					}
				}

				return nil
			})
			if err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

func TestValidateNewsMLUUIDs(t *testing.T) {
	tests := []TestData{
		{xml: "./examples/full-article.xml", expectError: false},
		{xml: "./testdata/uuids-invalid.xml", expectError: true},
		{xml: "./testdata/uuids-uppercase.xml", expectError: false, customConfig: "uc"},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.xml, func(t *testing.T) {
			xmlDoc := etree.NewDocument()
			err := xmlDoc.ReadFromFile(test.xml)
			if err != nil {
				t.Fatal(err)
			}
			err = docformat.ValidateNewsMLUUIDsAndFixNamespaces(xmlDoc)
			if !test.expectError && err != nil {
				t.Error(err)
			} else if test.expectError && err == nil {
				t.Error("error was expected")
			}

			if test.expectError && err != nil {
				if errors.Is(err, docformat.InvalidArgumentError{}) {
					t.Logf("InvalidArgumentError %s", err)
				} else if errors.Is(err, docformat.RequiredArgumentError{}) {
					t.Logf("RequiredArgumentError %s", err)
				}
			}

			if test.customConfig == "uc" {
				err = docformat.WalkXMLDocumentElements(xmlDoc, nil, func(element *etree.Element, args ...interface{}) error {
					uuidAttr := element.SelectAttr("uuid")
					if element.Tag != "" && uuidAttr == nil {
						return nil
					}

					var uuidValue string
					if uuidAttr != nil {
						uuidValue = uuidAttr.Value
					} else {
						uuidValue = element.Text()
					}

					for _, r := range uuidValue {
						if unicode.IsLower(r) {
							return fmt.Errorf("uppercase uuid expected: %s", uuidValue)
						}
					}

					return nil
				})
				if err != nil {
					t.Errorf("%s", err)
				}
			}
		})
	}
}

func TestValidateDocumentDates(t *testing.T) {
	var dc newsml.DateConfig
	file, err := os.Open("./testdata/dates-config.json")
	if err != nil {
		t.Fatal("unable to open ./testdata/dates-config.json")
	}

	dec := json.NewDecoder(file)
	err = dec.Decode(&dc)
	if err != nil {
		t.Fatal("unable to decode ./testdata/dates-config.json")
	}

	tests := []TestData{
		{json: "./testdata/text.json", expectError: false},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.xml, func(t *testing.T) {
			testData, err := ioutil.ReadFile(test.json)
			must(t, err, "could not open testfile")

			var document doc.Document
			err = json.Unmarshal(testData, &document)
			must(t, err, "could not unmarshal doc")

			err = docformat.ValidateDocumentDates(&document, dc)
			if !test.expectError && err != nil {
				t.Error(err)
			} else if test.expectError && err == nil {
				t.Error("error was expected")
			}
		})
	}
}

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

			err = docformat.WalkDocument(document, nil, docformat.ValidateAndLowercaseDocumentUUIDs)
			if !test.expectError && err != nil {
				t.Error(err)
			} else if test.expectError && err == nil {
				t.Error("error was expected")
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	var custom newsml.Options

	file, err := os.Open("./testdata/custom-config.json")
	if err != nil {
		t.Fatal("unable to open ./testdata/dates-config.json")
	}

	dec := json.NewDecoder(file)
	err = dec.Decode(&custom)
	if err != nil {
		t.Fatal("unable to decode ./testdata/dates-config.json")
	}

	date := "2021-06-01T04:42:00-05:00"
	err = docformat.ValidateDate(custom.DateElements, newsml.Block, "firstCreated", date)
	if err != nil {
		t.Fatalf("date %s should have errored", date)
	}

	date = "2021-06-01T04:42:00Z"
	err = docformat.ValidateDate(custom.DateElements, newsml.Block, "firstCreated", date)
	if err != nil {
		t.Fatalf("date %s should have errored", date)
	}

	date = ""
	err = docformat.ValidateDate(custom.DateElements, newsml.Block, "firstCreated", date)
	if err == nil {
		t.Fatalf("date %s missing UTC offset should have errored", date)
	}
	// t.Logf("Expected error: %s\n", err)

	date = "2021-99-99T04:42:00Z"
	err = docformat.ValidateDate(custom.DateElements, newsml.Block, "firstCreated", date)
	if err == nil {
		t.Fatalf("date %s missing UTC offset should have errored", date)
	}
	// t.Logf("Expected error: %s\n", err)

	date = "2021-06-01T04:42:00"
	err = docformat.ValidateDate(custom.DateElements, newsml.Block, "firstCreated", date)
	if err == nil {
		t.Fatalf("date %s missing UTC offset should have errored", date)
	}
	// t.Logf("Expected error: %s\n", err)

	date = "2021-06-01T04:42:00-16:00"
	err = docformat.ValidateDate(custom.DateElements, newsml.Block, "firstCreated", date)
	if err == nil {
		t.Fatalf("date %s invalid UTC offset should have errored", date)
	}
	// t.Logf("Expected error: %s\n", err)
}
