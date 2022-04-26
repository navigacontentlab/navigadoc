package docformat_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	docformat "bitbucket.org/infomaker/doc-format/v2"
	"bitbucket.org/infomaker/doc-format/v2/doc"
	"bitbucket.org/infomaker/doc-format/v2/listpackage"
	"bitbucket.org/infomaker/doc-format/v2/newsml"
	"bitbucket.org/infomaker/doc-format/v2/rpc"
	"github.com/Infomaker/etree"
	"github.com/google/go-cmp/cmp"
)

const configFile = "testdata/custom-config.json"
const configWithMetaFile = "testdata/custom-config-use-meta.json"

type TestData struct {
	xml          string
	json         string
	root         string
	customConfig string
	typeConfig   string
	expectError  bool
}

func TestNewsItemToNavigaDoc_Regressions(t *testing.T) {
	const dir = "testdata/regression"

	files, err := ioutil.ReadDir(dir)
	must(t, err, "failed to list regression tests")

	for _, i := range files {
		i := i
		if !strings.HasSuffix(i.Name(), ".newsitem.xml") {
			continue
		}

		t.Run(i.Name(), func(t *testing.T) {
			newsItemPath := filepath.Join(dir, i.Name())
			item, err := loadNewsItem(t, newsItemPath)
			if err != nil {
				must(t, err, "failed to load newsitem")
			}

			opts := newsml.DefaultOptions()
			_, _ = docformat.NewsItemToNavigaDoc(item, &opts)
		})
	}
}

func loadNewsItem(t *testing.T, name string) (*newsml.NewsItem, error) {
	t.Helper()

	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read %q: %w", name, err)
	}
	defer safeClose(t, name, f)

	dec := xml.NewDecoder(f)
	var item newsml.NewsItem

	if err := dec.Decode(&item); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal NewsItem %q: %v", name, err)
	}

	return &item, nil
}

func safeClose(t *testing.T, name string, c io.Closer) {
	t.Helper()

	if err := c.Close(); err != nil {
		t.Errorf("failed to close %s: %v", name, err)
	}
}

func checkNamespaces(t *testing.T, testFile string, xmlDocBytes []byte) {
	t.Helper()

	xmlDoc := etree.NewDocument()
	err := xmlDoc.ReadFromBytes(xmlDocBytes)
	must(t, err, "failed etree.ReadFromBytes")

	xmlns := xmlDoc.Root().SelectAttr("xmlns")
	if xmlns == nil {
		t.Errorf(fmt.Sprintf("generated root is missing xmlns: %s", testFile))
	} else if strings.Compare(xmlns.Value, "http://iptc.org/std/nar/2006-10-01/") != 0 {
		t.Errorf(fmt.Sprintf("generated root does not have expected xmlns: %s", testFile))
	}

	allElements := xmlDoc.FindElements(fmt.Sprintf("/%s//", xmlDoc.Root().Tag))
	for _, element := range allElements {
		switch element.Tag {
		case "links":
			fallthrough
		case "metadata":
			parent := element.Parent().Tag
			if parent == "itemMeta" || parent == "contentMeta" {
				if xmlns := element.SelectAttr("xmlns"); xmlns == nil || xmlns.Value == "" {
					t.Errorf("%s missing namespace: %s", element.Tag, element.GetPath())
				}
			}
		case "idf":
			if xmlns := element.SelectAttr("xmlns"); xmlns == nil || xmlns.Value == "" {
				t.Errorf("idf is missing namespace: %s", element.GetPath())
			}
		}
	}
}

func TestConvertingNewsmlRoundtrip(t *testing.T) {
	testData, err := ioutil.ReadFile(configFile)
	must(t, err, "could not open file")
	customConfig := string(testData)

	testData, err = ioutil.ReadFile(configWithMetaFile)
	must(t, err, "could not open file")
	customConfigWithMeta := string(testData)

	var tests = []TestData{
		{xml: "testdata/nb-710.xml", root: "newsItem", customConfig: customConfig},
		{xml: "testdata/objecttexttocontent.xml", root: "newsItem"},
		{xml: "testdata/image.xml"},
		{xml: "testdata/image2.xml"},
		{xml: "testdata/pdf.xml"},
		{xml: "testdata/text.xml"},
		{json: "testdata/text.json"},
		{xml: "examples/full-article.xml", customConfig: customConfig},
		{xml: "examples/full-pdf.xml"},
		{xml: "examples/full-picture.xml"},
		// With additional fields defined in new XSD
		{xml: "examples/modified-full-article.xml"},
		// With custom config
		{xml: "examples/full-article-custom-link.xml", customConfig: customConfig},
		// Without root namespace
		{xml: "examples/full-article-no-namespace.xml"},
		{xml: "testdata/invalid-data-xml.xml", expectError: true},
		{xml: "testdata/image-empty-pubstatus.xml"},
		{xml: "examples/full-article-custom-tests.xml", customConfig: customConfig},
		// Data with duplicate tags and attributes
		{xml: "testdata/custom-data.xml", root: "newsItem", customConfig: customConfig},
		{xml: "testdata/custom-data.xml", root: "newsItem", customConfig: customConfigWithMeta},
		{json: "testdata/custom-asset.json"},
		{xml: "testdata/custom-asset.xml", root: "newsItem"},
		{json: "testdata/ampersand-image.json"},
		{json: "testdata/ampersand-article.json"},
		{xml: "testdata/text-empty-elements.xml"},
		{xml: "testdata/missing-contentset.xml"},
		{xml: "testdata/geo-link.xml"},
	}

	for i := range tests {
		test := tests[i]
		if test.xml != "" {
			t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
				newsitemConversionRoundtrip(t, test, nil)
			})
			t.Run("WithDefaultOptions/"+test.xml, func(t *testing.T) {
				opts := newsml.DefaultOptions()
				newsitemConversionRoundtrip(t, test, &opts)
			})
		}

		if test.json != "" {
			roundtripJSON(t, &test)
		}
	}
}

func TestCheckSchemaInIDFLinks(t *testing.T) {
	testData, err := ioutil.ReadFile("testdata/text.xml")
	must(t, err, "could not open file")

	var original newsml.NewsItem
	err = xml.Unmarshal(testData, &original)
	must(t, err, "could not unmarshal file")

	customOptions := newsml.Options{}
	o := newsml.DefaultOptions()

	docformat.MergeOptions(&o, &customOptions)

	navigadoc, err := docformat.NewsItemToNavigaDoc(&original, &o)
	if err != nil {
		t.Fatal(err)
	}
	if len(navigadoc.Content) > 5 {
		p := navigadoc.Content[5]

		if !strings.Contains(p.Data["text"], "mailto") {
			t.Errorf("%s failed, did not find mailto: link in text", t.Name())
		}
	}
}
func TestInvalidDatesShouldNotBeDefaulted(t *testing.T) {
	cases := []struct {
		desc string
		file string
	}{
		{"TestInvalidPublishDate", "testdata/invalid-publish-date.xml"},
		{"TestInvalidUnpublishDate", "testdata/invalid-unpublish-date.xml"},
		{"TestInvalidModified", "testdata/invalid-modified-date.xml"},
		{"TestInvalidCreated", "testdata/invalid-created-date.xml"},
		{"TestMissingUTC", "testdata/date-missing-utc.xml"},
		{"TestInvalidUTC", "testdata/date-invalid-utc.xml"},
	}

	// in all these cases we expect an error
	for _, tc := range cases {
		testXML, err := ioutil.ReadFile(tc.file)
		must(t, err, "could not read file")

		var original newsml.NewsItem
		err = xml.Unmarshal(testXML, &original)
		must(t, err, "could not unmarshal file")

		opts := newsml.DefaultOptions()
		_, err = docformat.NewsItemToNavigaDoc(&original, &opts)

		if err == nil {
			t.Fatalf("%s, expected error, got nil", tc.desc)
		}
	}
}

func TestDataConversionOptions(t *testing.T) {
	testData, err := ioutil.ReadFile(configFile)
	must(t, err, "could not open file")
	customConfig := string(testData)

	testData, err = ioutil.ReadFile(configWithMetaFile)
	must(t, err, "could not open file")
	customConfigWithMeta := string(testData)

	var tests = []TestData{
		{xml: "testdata/custom-data.xml", customConfig: customConfig},
		{xml: "testdata/custom-data.xml", customConfig: customConfigWithMeta},
	}

	for i := range tests {
		test := tests[i]
		testXML, err := ioutil.ReadFile(test.xml)
		must(t, err, "could not read file")

		var original newsml.NewsItem
		err = xml.Unmarshal(testXML, &original)
		must(t, err, "could not unmarshal file")

		o := newsml.DefaultOptions()
		opts := &o

		if test.customConfig != "" {
			customConfig := newsml.Options{}

			err = json.Unmarshal([]byte(test.customConfig), &customConfig)
			must(t, err, "could not unmarshal file")

			docformat.MergeOptions(opts, &customConfig)
		}

		navigadoc, err := docformat.NewsItemToNavigaDoc(&original, opts)
		if test.expectError && err != nil {
			return
		}
		must(t, err, "failed NewsItemToNavigaDoc")

		var destination newsml.DataDestination
		for _, dc := range opts.DataElementConversion {
			if dc.Type == "alma/categories" {
				destination = dc.Destination
			}
		}

	metaloop:
		for _, meta := range navigadoc.Meta {
			if meta.Type == "alma/categories" {
				switch destination {
				case "link":
					for _, l := range meta.Links {
						if l.Type == "alma/categories" {
							break metaloop
						}
					}
					t.Fatal("link destination")
				default:
					for _, m := range meta.Meta {
						if m.Type == "alma/categories" {
							break metaloop
						}
					}
					t.Fatal("meta destination")
				}
			}
		}
	}
}

func TestDefaultConfigOverride(t *testing.T) {
	testData, err := ioutil.ReadFile("./testdata/custom-config-overrides.json")
	must(t, err, "could not open file")
	customConfig := string(testData)

	var tests = []TestData{
		{xml: "testdata/newsml-config-overrides.xml", customConfig: customConfig, root: "newsItem"},
		{xml: "testdata/concept-config-overrides.xml", customConfig: customConfig, root: "conceptItem"},
		{xml: "testdata/planningitem-config-overrides.xml", customConfig: customConfig, root: "planningItem"},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.xml, func(t *testing.T) {
			testXML, err := ioutil.ReadFile(test.xml)
			must(t, err, "could not read file")

			o := newsml.DefaultOptions()
			opts := &o

			if test.customConfig != "" {
				customConfig := newsml.Options{}

				err = json.Unmarshal([]byte(test.customConfig), &customConfig)
				must(t, err, "could not unmarshal file")

				docformat.MergeOptions(opts, &customConfig)
			}

			navigadoc, err := docformat.XMLToNavigaDoc(string(testXML), opts)
			if test.expectError && err != nil {
				return
			}
			must(t, err, "failed ItemToNavigaDoc")

			xmlDoc, err := docformat.NavigaDocToItem(navigadoc, opts)
			must(t, err, "failed NavigaDocToItem")

			generatedXML, err := xml.Marshal(xmlDoc)
			must(t, err, "failed Marshal")

			compareXMLFiles(t, test.root, test.xml, testXML, generatedXML, navigadoc)
		})
	}
}

func newsitemConversionRoundtrip(t *testing.T, test TestData, opts *newsml.Options) {
	testXML, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not read file")

	var original newsml.NewsItem
	err = xml.Unmarshal(testXML, &original)
	if test.expectError && err != nil {
		return
	}
	must(t, err, "could not unmarshal file")

	if test.customConfig != "" {
		if opts == nil {
			o := newsml.DefaultOptions()
			opts = &o
		}
		customConfig := newsml.Options{}

		// Read the xml into the ConceptItem
		err = json.Unmarshal([]byte(test.customConfig), &customConfig)
		must(t, err, "could not unmarshal file")

		docformat.MergeOptions(opts, &customConfig)
	}

	navigadoc, err := docformat.NewsItemToNavigaDoc(&original, opts)
	if test.expectError && err != nil {
		return
	}
	must(t, err, "failed NewsItemToNavigaDoc")

	generatedXML, err := docformat.NavigaDocToNewsItem(navigadoc, opts)
	must(t, err, "failed NavigaDocToNewsItem")

	// Check that the item has a root namespace
	if generatedXML.XMLNamespace == "" {
		t.Errorf("missing root namespace %s", test.xml)
	}

	xmlDoc, err := xml.MarshalIndent(generatedXML, "", "    ")
	must(t, err, "could not parse generated xmlDoc")

	checkNamespaces(t, test.xml, xmlDoc)
	compareXMLFiles(t, "newsItem", test.xml, testXML, xmlDoc, navigadoc)
}

func stripchars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if !strings.ContainsRune(chr, r) {
			return r
		}
		return -1
	}, str)
}

func TestConvertingConceptRoundtrip(t *testing.T) {
	// we unmarshal an newsml file then converts into naviga doc, and then back to newsml
	var tests = []TestData{
		{xml: "testdata/author.xml"},
		{xml: "testdata/category.xml"},
		{xml: "testdata/channel.xml"},
		{xml: "testdata/content-profile.xml"},
		{xml: "testdata/content-profile-uri.xml"},
		{xml: "testdata/event.xml"},
		{xml: "testdata/organisation.xml"},
		{xml: "testdata/person.xml"},
		{xml: "testdata/place-point.xml"},
		{xml: "testdata/place-polygon.xml"},
		{xml: "testdata/section.xml"},
		{xml: "testdata/story.xml"},
		{xml: "testdata/topic.xml"},
		{xml: "examples/full-concept.xml"},
		{xml: "examples/full-concept-bad-uri.xml"},
		// custom x-im/editor
		{xml: "testdata/custom-concept.xml", customConfig: "testdata/custom-config.json"},
		{xml: "examples/full-concept-no-namespace.xml"},
		{xml: "examples/full-concept-empty-dates.xml", expectError: true},
		{json: "examples/concept-invalid-xml.json"},
		{json: "testdata/story-concept.json"},
	}

	for i := range tests {
		test := tests[i]
		if test.xml != "" {
			t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
				conceptConversionRoundtrip(t, test, nil)
			})
			t.Run("WithDefaultOptions/"+test.xml, func(t *testing.T) {
				opts := newsml.DefaultOptions()
				conceptConversionRoundtrip(t, test, &opts)
			})
		}

		if test.json != "" {
			roundtripJSON(t, &test)
		}
	}
}

func conceptConversionRoundtrip(t *testing.T, test TestData, opts *newsml.Options) {
	want, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not open file")

	original := newsml.ConceptItem{}

	// Read the xml into the ConceptItem
	err = xml.Unmarshal(want, &original)
	must(t, err, "could not unmarshal file")

	if test.customConfig != "" {
		o := newsml.DefaultOptions()
		opts = &o
		testData, err := ioutil.ReadFile(test.customConfig)
		must(t, err, "could not open file")
		cfg := &newsml.Options{}

		err = json.Unmarshal(testData, &cfg)
		must(t, err, "could not unmarshal file")

		docformat.MergeOptions(opts, cfg)
	}

	// Convert the ConceptItem to NavigaDoc format
	document, err := docformat.ConceptItemToNavigaDoc(&original, opts)
	if test.expectError && err != nil {
		return
	}
	must(t, err, "failed ConceptItemToNavigaDoc")

	// Concert the NavigaDoc back to a ConceptItem
	got, err := docformat.NavigaDocToConceptItem(document, opts)
	must(t, err, "failed NavigaDocToConceptItem")

	// Turn the ConceptItem into xml
	xmlDoc, err := xml.MarshalIndent(got, "", "    ")
	must(t, err, "could not marshal document")

	checkNamespaces(t, test.xml, xmlDoc)

	compareXMLFiles(t, "conceptItem", test.xml, want, xmlDoc, document)
}

func TestGeneratedDocument(t *testing.T) {
	var test = struct {
		xml string
	}{
		xml: "testdata/text.xml",
	}

	testData, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not open file")

	var original newsml.NewsItem
	err = xml.Unmarshal(testData, &original)
	must(t, err, "could not unmarshal file")

	opts := newsml.DefaultOptions()
	document, err := docformat.ItemToNavigaDoc(&original, &opts)
	must(t, err, "failed ItemToNavigaDoc")

	rpcDoc := rpc.Document{}
	err = rpcDoc.FromDocDocument(document)
	must(t, err, "failed FromDocDocument")

	newDoc := doc.Document{}
	err = rpcDoc.ToDocDocument(&newDoc)
	must(t, err, "failed ToDocDocument")

	got, err := docformat.NavigaDocToItem(&newDoc, &opts)
	must(t, err, "failed NavigaDocToItem")

	// Turn the NewsItem into xmlDoc
	xmlDoc, err := xml.MarshalIndent(got, "", "    ")
	must(t, err, "failed to marshal newsitem")

	// Compare XML to original to check that no data was lost
	err = newsml.CompareXML("original", "text.xml", "newsItem", string(testData), string(xmlDoc))
	if err != nil {
		t.Fail()
		t.Log(fmt.Sprintf("%s\n", err))
	}

	doStructComparison(t, &original, got)
}

func TestConvertingListRoundtrip(t *testing.T) {
	const testFile = "testdata/type-config.json"
	testData, err := ioutil.ReadFile(testFile)
	must(t, err, "could not open file")

	typeConfig := string(testData)

	var tests = []TestData{
		{xml: "examples/full-list.xml"},
		{xml: "examples/full-list.xml", typeConfig: typeConfig},
	}

	opts := newsml.DefaultOptions()

	for i := range tests {
		test := tests[i]
		t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
			listConversionRoundtrip(t, test, nil)
		})
		t.Run("WithDefaultOptions/"+test.xml, func(t *testing.T) {
			listConversionRoundtrip(t, test, &opts)
		})
	}
}

func listConversionRoundtrip(t *testing.T, test TestData, opts *newsml.Options) {
	t.Helper()

	want, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not open file")

	original := listpackage.List{}

	// Read the xmlDoc into the List
	err = xml.Unmarshal(want, &original)
	must(t, err, "could not unmarshal file")

	// Convert the List to NavigaDoc format
	document, err := docformat.ListToNavigaDoc(&original, opts)
	must(t, err, "failed ListToNavigaDoc")

	// Concert the NavigaDoc back to a List
	got, err := docformat.NavigaDocToList(document, opts)
	must(t, err, "failed NavigaDocToList")

	got.ItemMeta.PubStatus = nil

	// Turn the List into xmlDoc
	xmlDoc, err := xml.MarshalIndent(got, "", "    ")
	must(t, err, "could not marshal XML")

	compareXMLFiles(t, "list", test.xml, want, xmlDoc, document)
}

func TestConvertingPackageRoundtrip(t *testing.T) {
	const testFile = "testdata/type-config.json"
	testData, err := ioutil.ReadFile(testFile)
	must(t, err, "could not open file")

	typeConfig := string(testData)

	var tests = []TestData{
		{xml: "examples/full-package.xml"},
		{xml: "testdata/empty-package.xml"},
		{xml: "examples/full-package.xml", typeConfig: typeConfig},
	}

	opts := newsml.DefaultOptions()

	for _, test := range tests {
		test := test
		t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
			packageConversionRoundtrip(t, test, nil)
		})
		t.Run("WithDefaultOptions/"+test.xml, func(t *testing.T) {
			packageConversionRoundtrip(t, test, &opts)
		})
	}
}

func packageConversionRoundtrip(t *testing.T, test TestData, opts *newsml.Options) {
	t.Helper()

	want, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not open file")

	original := listpackage.Package{}

	// Read the xmlDoc into the Package
	err = xml.Unmarshal(want, &original)
	must(t, err, "could not unmarshal file")

	// Convert the Package to NavigaDoc format
	document, err := docformat.PackageToNavigaDoc(&original, opts)
	must(t, err, "failed PackageToNavigaDoc")

	// Concert the NavigaDoc back to a Package
	got, err := docformat.NavigaDocToPackage(document, opts)
	must(t, err, "failed NavigaDocToPackage")

	// Fudge the PubStatus for the test
	if original.ItemMeta.PubStatus == nil {
		got.ItemMeta.PubStatus = nil
	}

	// Turn the Package into xmlDoc
	xmlDoc, err := xml.MarshalIndent(got, "", "    ")
	must(t, err, "could marshal XML")

	compareXMLFiles(t, "package", test.xml, want, xmlDoc, document)
}

func TestConvertAssignmentRoundTrip(t *testing.T) {
	const testFile = "testdata/type-config.json"
	testData, err := ioutil.ReadFile(testFile)
	must(t, err, "could not open file")

	typeConfig := string(testData)

	testData, err = ioutil.ReadFile(configFile)
	must(t, err, "could not open file")
	customConfig := string(testData)

	var tests = []TestData{
		{xml: "examples/full-assignment.xml", customConfig: customConfig},
		{xml: "examples/full-assignment-no-namespace.xml"},
		{xml: "examples/full-assignment-empty-dates.xml", expectError: true},
		{xml: "examples/full-assignment.xml", customConfig: customConfig, typeConfig: typeConfig},
	}

	opts := newsml.DefaultOptions()

	for i := range tests {
		test := tests[i]
		if test.xml != "" {
			t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
				assignmentConversionRoundtrip(t, test, nil)
			})
			t.Run("WithDefaultOptions/"+test.xml, func(t *testing.T) {
				assignmentConversionRoundtrip(t, test, &opts)
			})
		}

		if test.json != "" {
			// Add empty nrpdate:modified to match original
			opts := []cmp.Option{
				cmp.Transformer("AddModified", func(in []*doc.Property) []*doc.Property {
					in = append(in, &doc.Property{Name: "nrpdate:modified"})
					sort.Slice(in, func(i, j int) bool {
						return in[i].Name > in[j].Name
					})
					return in
				}),
			}
			roundtripJSON(t, &test, opts...)
		}
	}
}

func assignmentConversionRoundtrip(t *testing.T, test TestData, opts *newsml.Options) {
	t.Helper()

	want, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not open file")

	var original newsml.Assignment
	err = xml.Unmarshal(want, &original)
	must(t, err, "could not unmarshal XML")

	if test.customConfig != "" {
		o := newsml.DefaultOptions()
		opts = &o
		customConfig := newsml.Options{}

		// Read the xml into the ConceptItem
		err = json.Unmarshal([]byte(test.customConfig), &customConfig)
		must(t, err, "could not unmarshal file")

		docformat.MergeOptions(opts, &customConfig)
	}

	document, err := docformat.AssignmentToNavigaDoc(&original, opts)
	if test.expectError && err != nil {
		return
	}
	must(t, err, "failed AssignmentToNavigaDoc")

	// Concert the NavigaDoc back to an Assignment
	got, err := docformat.NavigaDocToAssignment(document, opts)
	must(t, err, "failed NavigaDocToAssignment")

	// Turn the Assignment into xmlDoc
	xmlDoc, err := xml.MarshalIndent(got, "", "    ")
	must(t, err, "could not marshal XML document")

	checkNamespaces(t, test.xml, xmlDoc)

	compareXMLFiles(t, "planning", test.xml, want, xmlDoc, document)
}

func TestConvertPlanningItemRoundTrip(t *testing.T) {
	const testFile = "testdata/type-config.json"

	testData, err := ioutil.ReadFile(testFile)
	must(t, err, "could not open file")

	typeConfig := string(testData)

	var tests = []TestData{
		{xml: "examples/full-planningitem.xml"},
		{xml: "examples/full-planningitem.xml", typeConfig: typeConfig},
	}

	opts := newsml.DefaultOptions()

	for i := range tests {
		test := tests[i]
		t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
			planningItemConversionRoundtrip(t, test, nil)
		})
		t.Run("WithDefaultOptions/"+test.xml, func(t *testing.T) {
			planningItemConversionRoundtrip(t, test, &opts)
		})
	}
}

func planningItemConversionRoundtrip(t *testing.T, test TestData, opts *newsml.Options) {
	t.Helper()

	want, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not open file")

	var original newsml.PlanningItem
	err = xml.Unmarshal(want, &original)
	must(t, err, "could not unmarshal file")

	document, err := docformat.PlanningItemToNavigaDoc(&original, opts)
	must(t, err, "failed PlanningItemToNavigaDoc")

	// Concert the NavigaDoc back to a List
	got, err := docformat.NavigaDocToPlanningItem(document, opts)
	must(t, err, "failed NavigaDocToPlanningItem")

	// Turn the PlanningItem into xmlDoc
	xmlDoc, err := xml.MarshalIndent(got, "", "    ")
	must(t, err, "could not marshal XML document")

	checkNamespaces(t, test.xml, xmlDoc)

	compareXMLFiles(t, "planningItem", test.xml, want, xmlDoc, document)
}

func TestConceptItemCreatingExtraMetaEvent(t *testing.T) {
	concept := &newsml.ConceptItem{
		ItemMeta: &newsml.ItemMeta{
			Title: "some bogus title",
		},
	}
	opts := newsml.DefaultOptions()
	document, err := docformat.ConceptItemToNavigaDoc(concept, &opts)
	must(t, err, "failed ConceptItemToNavigaDoc")

	if len(document.Meta) != 0 {
		t.Fatal("meta should only be added if used")
	}
}

func TestConceptNoMetadata(t *testing.T) {
	test := TestData{
		xml: "examples/full-concept-no-contentmeta.xml",
	}

	want, err := ioutil.ReadFile(test.xml)
	must(t, err, "could not open file")

	concept := newsml.ConceptItem{}

	// Read the xml into the ConceptItem
	err = xml.Unmarshal(want, &concept)
	must(t, err, "could not unmarshal file")

	opts := newsml.DefaultOptions()
	document, err := docformat.ConceptItemToNavigaDoc(&concept, &opts)
	must(t, err, "failed ConceptItemToNavigaDoc")

	// Convert the NavigaDoc back to a ConceptItem
	got, err := docformat.NavigaDocToConceptItem(document, &opts)
	must(t, err, "failed NavigaDocToConceptItem")

	doStructComparison(t, &concept, got)

	// Just check a couple of fields known to have fallen out
	if got.Concept.Name != concept.Concept.Name || got.Concept.Type.Qcode != concept.Concept.Type.Qcode {
		t.Fatalf("concept mismatch")
	}
}

func TestXMLToDOCToXML(t *testing.T) {
	var tests = []TestData{
		{xml: "testdata/author.xml", root: "conceptItem"},
		{xml: "testdata/category.xml", root: "conceptItem"},
		{xml: "testdata/channel.xml", root: "conceptItem"},
		{xml: "testdata/content-profile.xml", root: "conceptItem"},
		{xml: "testdata/content-profile-uri.xml", root: "conceptItem"},
		{xml: "testdata/event.xml", root: "conceptItem"},
		{xml: "testdata/organisation.xml", root: "conceptItem"},
		{xml: "testdata/person.xml", root: "conceptItem"},
		{xml: "testdata/place-point.xml", root: "conceptItem"},
		{xml: "testdata/place-polygon.xml", root: "conceptItem"},
		{xml: "testdata/section.xml", root: "conceptItem"},
		{xml: "testdata/story.xml", root: "conceptItem"},
		{xml: "testdata/topic.xml", root: "conceptItem"},
		{xml: "examples/full-article.xml", root: "newsItem"},
		{xml: "examples/full-assignment.xml", root: "planning"},
		{xml: "examples/full-concept.xml", root: "conceptItem"},
		{xml: "examples/full-list.xml", root: "list"},
		{xml: "examples/full-package.xml", root: "package"},
		{xml: "examples/full-pdf.xml", root: "newsItem"},
		{xml: "examples/full-picture.xml", root: "newsItem"},
		{xml: "examples/full-planningitem.xml", root: "planningItem"},
	}

	opts := newsml.DefaultOptions()
	// Needs custom config for x-gm/textsize
	testData, err := ioutil.ReadFile(configFile)
	must(t, err, "could not open file")

	customConfig := newsml.Options{}
	err = json.Unmarshal(testData, &customConfig)
	must(t, err, "could not unmarshal file")

	docformat.MergeOptions(&opts, &customConfig)

	for i := range tests {
		test := tests[i]
		t.Run(test.xml, func(t *testing.T) {
			want, err := ioutil.ReadFile(test.xml)
			must(t, err, "could not open file")

			document, err := docformat.XMLToNavigaDoc(string(want), &opts)
			must(t, err, "failed XMLToNavigaDoc")

			xmlDoc, err := docformat.NavigaDocToItem(document, &opts)
			must(t, err, "failed NavigaDocToItem")

			got, err := xml.Marshal(xmlDoc)
			must(t, err, "failed to marshal XML")

			compareXMLFiles(t, test.root, test.xml, want, got, document)
		})
	}
}

func TestDocDataTags(t *testing.T) {
	const testDoc = "testdata/text.json"

	testData, err := ioutil.ReadFile(testDoc)
	must(t, err, "could not open file")

	navigaDoc := doc.Document{}
	err = json.Unmarshal(testData, &navigaDoc)
	must(t, err, "could not unmarshal file")

	// Tag can't start with a number
	navigaDoc.Meta[0].Data = map[string]string{"0_tag": "value"}
	opts := newsml.DefaultOptions()

	_, err = docformat.NavigaDocToItem(&navigaDoc, &opts)
	if err == nil {
		t.Fatalf("expected error when converting document %s", testDoc)
	}

	// Tags must be A-Za-z0-9_
	navigaDoc.Meta[0].Data = map[string]string{"a#$tag": "value"}
	_, err = docformat.NavigaDocToItem(&navigaDoc, &opts)
	if err == nil {
		t.Fatalf("expected error when converting document %s", testDoc)
	}

	navigaDoc.Meta[0].Data = map[string]string{"a-0_tag": "value"}
	_, err = docformat.NavigaDocToItem(&navigaDoc, &opts)
	must(t, err, "failed NavigaDocToItem")
}

func TestIDFMissingNamespace(t *testing.T) {
	const xmlFile = "testdata/text_missing_idfns.xml"

	xmlDoc, err := ioutil.ReadFile(xmlFile)
	must(t, err, "could not open file")

	opts := newsml.DefaultOptions()
	document, err := docformat.XMLToNavigaDoc(string(xmlDoc), &opts)
	must(t, err, "failed XMLToNavigaDoc")

	got := len(document.Content)
	want := 10
	if got != want {
		t.Fatalf("could not convert idf to content for, got %d items, wanted %d items %s", got, want, xmlFile)
	}

	item, err := docformat.NavigaDocToItem(document, &opts)
	must(t, err, "failed NavigaDocToItem")

	newsItem, ok := item.(*newsml.NewsItem)
	if !ok {
		t.Fatalf("expected conversion to item to yield a *NewsItem, got %T", item)
	}

	wantIDFNS := "http://www.infomaker.se/idf/1.0"
	gotIDFNS := newsItem.ContentSet.InlineXML.Idf.Xmlns
	if gotIDFNS != wantIDFNS {
		t.Errorf("wrong idf namespace, wanted %q, got %q",
			wantIDFNS, gotIDFNS)
	}
}

func TestEmptyBlocks(t *testing.T) {
	var tests = []struct {
		testfile string
		error    error
	}{
		{"testdata/empty-blocks-example.json", docformat.ErrEmptyBlock},
		{"testdata/empty-document-example.json", docformat.ErrEmptyDoc},
		{"testdata/text.json", nil},
	}

	for _, test := range tests {
		jsonData, err := ioutil.ReadFile(test.testfile)
		if err != nil {
			t.Fatalf("could not open file %s", test.testfile)
		}

		var navigaDoc doc.Document

		err = json.Unmarshal(jsonData, &navigaDoc)
		if err != nil {
			t.Fatalf("could not unmarshal doc %s", err)
		}
		opts := newsml.DefaultOptions()
		_, err = docformat.NavigaDocToNewsItem(&navigaDoc, &opts)
		if test.error == nil {
			if err != nil {
				t.Fatalf("Correct formated document should NOT return error. Testfile: %s", test.testfile)
			}
		} else {
			if err == nil {
				t.Fatalf("Empty blocks should return error. Testfile: %s", test.testfile)
			}
			if !errors.Is(err, test.error) {
				t.Fatalf("Test should have returned %T error. Testfile: %s", test.error, test.testfile)
			}
		}
	}
}

// TestReadLegacyDateValues tests values like "null", "undefined" etc that
// might be encountered in the repo due to old legacy
func TestReadLegacyDateValues(t *testing.T) {
	// Setup and validate testdata
	b, err := ioutil.ReadFile(path.Join("testdata", "article-minimum.xml"))
	if err != nil {
		t.Fatal("could not open test file article-minimum.xml")
	}

	var item newsml.NewsItem

	err = xml.NewDecoder(bytes.NewReader(b)).Decode(&item)
	if err != nil {
		t.Fatalf("failed to parse test data: %v", err)
	}

	x, err := xml.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	opts := newsml.DefaultOptions()

	_, err = docformat.XMLToNavigaDoc(string(x), &opts)
	if err != nil {
		t.Fatalf("failed to convert test file: %v", err)
	}

	// Setup test cases for "publication" dates and run them
	var pubTests = []struct {
		prop        string
		val         string
		shouldError bool
		shouldExist bool
	}{
		{"imext:pubstop", "null", false, false},
		{"imext:pubstop", "", false, false},
		{"imext:pubstop", "2020-03-20T12:24:31Z", false, true},
		{"imext:pubstop", "2020:03:20", true, false},
		{"imext:pubstart", "null", false, false},
		{"imext:pubstart", "", false, false},
		{"imext:pubstart", "2020-03-20T12:24:31Z", false, true},
		{"imext:pubstart", "2020:03:20", true, false},
	}

	for _, test := range pubTests {
		t.Run(fmt.Sprintf(
			"%s with value %s", test.prop, test.val,
		), func(t *testing.T) {
			// Setup test data
			item.ItemMeta.ItemMetaExtProperty = []newsml.MetaExtProperty{
				{Type: "imext:type", Value: "x-im/article"},
				{Type: test.prop, Value: test.val},
			}

			got := dateTestHelper(t, item, test.shouldError)
			if got == nil {
				return
			}

			switch test.prop {
			case "imext:pubstop":
				if test.shouldExist && got.Unpublished == nil {
					t.Error(`expected "unpublished" but got nil`)
				}
				if !test.shouldExist && got.Unpublished != nil {
					t.Error(`expected "unpublished" to be nil but got value`)
				}
			case "imext:pubstart":
				if test.shouldExist && got.Published == nil {
					t.Error(`expected "published" but got nil`)
				}
				if !test.shouldExist && got.Published != nil {
					t.Error(`expected "published" to be nil but got value`)
				}
			default:
				t.Fatalf("invalid test case, prop %s not supported", test.prop)
			}
		})
	}

	// Reset itemMetaExtProperties
	item.ItemMeta.ItemMetaExtProperty = []newsml.MetaExtProperty{
		{Type: "imext:type", Value: "x-im/article"},
	}

	// Setup tests for "created" and "modified"
	var tests = []struct {
		val         string
		shouldError bool
		shouldExist bool
	}{
		{val: "null", shouldError: true, shouldExist: false},
		{val: "undefined", shouldError: true, shouldExist: false},
		{val: "2020:03:20", shouldError: true, shouldExist: false},
		{val: "", shouldError: false, shouldExist: false},
		{val: "2020-03-20T12:24:31Z", shouldError: false, shouldExist: true},
	}

	// Test FirstCreated
	for _, test := range tests {
		t.Run(fmt.Sprintf(
			"test firstCreated with value %s", test.val,
		), func(t *testing.T) {
			item.ItemMeta.FirstCreated = test.val
			item.ContentMeta.ContentCreated = ""

			got := dateTestHelper(t, item, test.shouldError)
			if got == nil {
				return
			}

			if test.shouldExist && got.Created == nil {
				t.Error(`expected "created" but got nil`)
			}
			if !test.shouldExist && got.Created != nil {
				t.Error(`expected "created" to be nil but got value`)
			}
		})
	}

	// Test VersionCreated
	for _, test := range tests {
		t.Run(fmt.Sprintf(
			"test versionCreated with value %s", test.val,
		), func(t *testing.T) {
			item.ItemMeta.VersionCreated = test.val
			item.ContentMeta.ContentCreated = ""

			got := dateTestHelper(t, item, test.shouldError)
			if got == nil {
				return
			}

			if test.shouldExist && got.Modified == nil {
				t.Error(`expected "modified" but got nil`)
			}
			if !test.shouldExist && got.Modified != nil {
				t.Error(`expected "modified" to be nil but got value`)
			}
		})
	}
}

func TestDoubleCDATAEmbedCode(t *testing.T) {
	testData, err := ioutil.ReadFile(configFile)
	must(t, err, "could not open file")
	customConfig := string(testData)

	var tests = []TestData{
		{xml: "testdata/text-embedcode.xml", root: "newsItem", customConfig: customConfig},
	}

	for i := range tests {
		test := tests[i]
		if test.xml != "" {
			t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
				testXML, err := ioutil.ReadFile(test.xml)
				must(t, err, "could not read file")

				xmlDocOriginal := etree.NewDocument()
				err = xmlDocOriginal.ReadFromBytes(testXML)
				must(t, err, "could not parse original xml")

				var original newsml.NewsItem
				err = xml.Unmarshal(testXML, &original)
				must(t, err, "could not unmarshal file")

				var opts *newsml.Options
				if test.customConfig != "" {
					o := newsml.DefaultOptions()
					opts = &o
					customConfig := newsml.Options{}

					err = json.Unmarshal([]byte(test.customConfig), &customConfig)
					must(t, err, "could not unmarshal file")

					docformat.MergeOptions(opts, &customConfig)
				}

				navigadoc, err := docformat.NewsItemToNavigaDoc(&original, opts)
				if test.expectError && err != nil {
					return
				}
				must(t, err, "failed NewsItemToNavigaDoc")

				generatedNewsItem, err := docformat.NavigaDocToNewsItem(navigadoc, opts)
				must(t, err, "failed NavigaDocToNewsItem")

				originalEmbedCode, err := getEmbedCode(&original)
				must(t, err, "failed getting embedcode for original")
				generatedEmbedCode, err := getEmbedCode(generatedNewsItem)
				must(t, err, "failed getting embedcode for generated")

				originalCDATA := getCDATAString(originalEmbedCode)
				generatedCDATA := getCDATAString(generatedEmbedCode)

				if strings.Compare(originalCDATA, generatedCDATA) != 0 {
					t.Fatalf("embedcode mismatch \"%v\" vs \"%v\"", originalCDATA, generatedCDATA)
				}

				if t.Failed() || testing.Verbose() {
					generatedXML, err := xml.MarshalIndent(generatedNewsItem, "", "  ")
					must(t, err, "failed marshaling newsItem for generated")
					dumpDocs(t, navigadoc, nil, testXML, generatedXML, nil, nil)
				}
			})
		}
	}
}

// TestMixedCaseHTML tests that mixed case html tags are lowercased
// through the bluemonday sanitize
func TestMixedCaseHTML(t *testing.T) {
	testData, err := ioutil.ReadFile(configFile)
	must(t, err, "could not open file")
	customConfig := string(testData)

	var tests = []TestData{
		{xml: "testdata/mixedcase-tags.xml", root: "newsItem", customConfig: customConfig},
	}

	for i := range tests {
		test := tests[i]
		if test.xml != "" {
			t.Run("WithoutOptions/"+test.xml, func(t *testing.T) {
				testXML, err := ioutil.ReadFile(test.xml)
				must(t, err, "could not read file")

				xmlDocOriginal := etree.NewDocument()
				err = xmlDocOriginal.ReadFromBytes(testXML)
				must(t, err, "could not parse original xml")

				var original newsml.NewsItem
				err = xml.Unmarshal(testXML, &original)
				must(t, err, "could not unmarshal file")

				var opts *newsml.Options
				if test.customConfig != "" {
					o := newsml.DefaultOptions()
					opts = &o
					customConfig := newsml.Options{}

					err = json.Unmarshal([]byte(test.customConfig), &customConfig)
					must(t, err, "could not unmarshal file")

					docformat.MergeOptions(opts, &customConfig)
				}

				navigadoc, err := docformat.NewsItemToNavigaDoc(&original, opts)
				if test.expectError && err != nil {
					return
				}
				must(t, err, "failed NewsItemToNavigaDoc")

				generatedNewsItem, err := docformat.NavigaDocToNewsItem(navigadoc, opts)
				must(t, err, "failed NavigaDocToNewsItem")

				xmlTextGenerated, err := xml.Marshal(generatedNewsItem)
				if err != nil {
					must(t, err, "error marshaling xml")
				}

				xmlDocGenerated := etree.NewDocument()
				err = xmlDocGenerated.ReadFromBytes(xmlTextGenerated)
				if err != nil {
					must(t, err, "error parsing generated xml")
				}

				const authorXpath = "//itemMeta/links/link[@type='x-im/author']"
				elementOriginal := xmlDocOriginal.FindElement(authorXpath)
				if elementOriginal == nil {
					t.Fatal("can't find author link in original")
				}

				elementGenerated := xmlDocGenerated.FindElement(authorXpath)
				if elementGenerated == nil {
					t.Fatal("can't find author link in generated")
				}

				if strings.Compare(elementOriginal.SelectAttr("title").Value, elementGenerated.SelectAttr("title").Value) != 0 {
					t.Fatal("author titles don't match")
				}

				xpaths := []string{
					"//contentMeta/metadata/object[@type='x-im/mixedcase']/data/caption",
					"//contentSet/inlineXML/idf/group[@type='body']/element[@type='body']",
				}

				for _, xpath := range xpaths {
					var innerOriginalXML bytes.Buffer
					var innerGeneratedXML bytes.Buffer

					elementOriginal = xmlDocOriginal.FindElement(xpath)
					if elementOriginal == nil {
						t.Fatal("can't find author link in original")
					}

					for _, child := range elementOriginal.Child {
						child.WriteTo(&innerOriginalXML, &xmlDocOriginal.WriteSettings)
					}

					elementGenerated = xmlDocGenerated.FindElement(xpath)
					if elementGenerated == nil {
						t.Fatal("can't find author link in generated")
					}

					for _, child := range elementGenerated.Child {
						child.WriteTo(&innerGeneratedXML, &xmlDocGenerated.WriteSettings)
					}

					textOriginal := innerOriginalXML.String()
					textGenerated := innerGeneratedXML.String()

					if strings.Compare(textOriginal, textGenerated) != 0 {
						t.Fatalf("expected case sensitive compare to match: %s", xpath)
					}
				}
			})
		}
	}
}

func TestConceptDuplicateProperties(t *testing.T) {
	testFile := "examples/full-concept.xml"
	xmlDoc, err := ioutil.ReadFile(testFile)
	must(t, err, "could not open file")

	conceptItem := &newsml.ConceptItem{}

	// Read the xml into the ConceptItem
	err = xml.Unmarshal(xmlDoc, conceptItem)
	must(t, err, "could not unmarshal file")

	o := newsml.DefaultOptions()
	opts := &o

	// Convert the ConceptItem to NavigaDoc format
	document, err := docformat.ConceptItemToNavigaDoc(conceptItem, opts)
	must(t, err, "failed ConceptItemToNavigaDoc")

	// Check the properties for duplicates
	var propKeys = make(map[string]map[string]string)
	for _, prop := range document.Properties {
		if _, ok := propKeys[prop.Name]; ok {
			switch prop.Name {
			case "definition":
				if prop.Parameters["role"] == propKeys["definition"]["role"] {
					t.Fatalf("duplicate property: %s", prop.Name)
				}
			default:
				t.Fatalf("duplicate property: %s", prop.Name)
			}
		}
		propKeys[prop.Name] = prop.Parameters
	}

	// Convert the NavigaDoc back to a ConceptItem
	conceptItem, err = docformat.NavigaDocToConceptItem(document, opts)
	must(t, err, "failed NavigaDocToConceptItem")

	// Back to a navigadoc
	document, err = docformat.ConceptItemToNavigaDoc(conceptItem, opts)
	must(t, err, "failed ConceptItemToNavigaDoc")

	// Check the properties for duplicates
	propKeys = make(map[string]map[string]string)
	for _, prop := range document.Properties {
		if _, ok := propKeys[prop.Name]; ok {
			switch prop.Name {
			case "definition":
				if prop.Parameters["role"] == propKeys["definition"]["role"] {
					t.Fatalf("duplicate property: %s", prop.Name)
				}
			default:
				t.Fatalf("duplicate property: %s", prop.Name)
			}
		}
		propKeys[prop.Name] = prop.Parameters
	}
}

func dateTestHelper(
	t *testing.T, item newsml.NewsItem, shouldError bool,
) *doc.Document {
	b, err := xml.Marshal(item)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	opts := newsml.DefaultOptions()

	navigadoc, err := docformat.XMLToNavigaDoc(string(b), &opts)
	if shouldError {
		if err == nil {
			t.Fatal("expected error but got nil")
		}
		return nil // We are happy, continue please
	} else if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	return navigadoc
}

func compareXMLFiles(
	t *testing.T, root string, filename string, originalXML []byte,
	generatedXML []byte, document *doc.Document,
) {
	t.Helper()

	errOriginal := newsml.CompareXML("original", filename, root, string(originalXML), string(generatedXML))
	checkf(t, errOriginal, "[%s] comparing XML originalXML to generatedXML", filename)

	// Reverse the comparison to check for extraneous data
	errGenerated := newsml.CompareXML("generated", filename, root, string(generatedXML), string(originalXML))
	checkf(t, errGenerated, "[%s] comparing XML generatedXML to originalXML", filename)

	if t.Failed() || testing.Verbose() {
		dumpDocs(t, document, nil, originalXML, generatedXML, errOriginal, errGenerated)
	}
}

func roundtripJSON(t *testing.T, test *TestData, diffOptions ...cmp.Option) {
	t.Helper()

	t.Run(test.json, func(t *testing.T) {
		testData, err := ioutil.ReadFile(test.json)
		must(t, err, "could not open testfile")

		var want doc.Document
		err = json.Unmarshal(testData, &want)
		must(t, err, "could not unmarshal doc")

		defaultCase := func(document *doc.Document, opts *newsml.Options) (interface{}, error) {
			return docformat.NavigaDocToNewsItem(document, opts)
		}

		opts := newsml.DefaultOptions()
		xmlItem, err := docformat.NavigaDocToItemWithDefault(&want, &opts, defaultCase)
		must(t, err, "failed NavigaDocToItem")

		got, err := docformat.ItemToNavigaDoc(xmlItem, &opts)
		must(t, err, "failed ItemToNavigaDoc")

		xmlBytes, err := xml.MarshalIndent(xmlItem, "", "  ")
		must(t, err, "failed marshalling XML")

		doStructComparison(t, &want, got, diffOptions...)
		if t.Failed() || testing.Verbose() {
			dumpDocs(t, &want, got, nil, xmlBytes)
		}
	})
}

func ensureDebugDir(t *testing.T) string {
	t.Helper()

	outDir := filepath.Join("testdebug", t.Name())
	must(t, os.MkdirAll(outDir, 0770), "failed to create test debug output directory")

	return outDir
}

func doStructComparison(
	t *testing.T, want interface{}, got interface{}, opts ...cmp.Option,
) {
	structCompareOptions := []cmp.Option{
		cmp.Transformer("RemoveWhitespace", func(in *newsml.Data) *newsml.Data {
			if in == nil {
				return in
			}
			in.Raw = stripchars(in.Raw, "\n \t\r")
			return in
		}),
	}

	if len(opts) > 0 {
		structCompareOptions = append(structCompareOptions, opts...)
	}

	if testing.Verbose() && !cmp.Equal(want, got, structCompareOptions...) {
		dumpStructComparison(t, cmp.Diff(want, got))
	}
}

func dumpStructComparison(t *testing.T, diff string) {
	t.Helper()

	outDir := ensureDebugDir(t)
	diffPath := filepath.Join(outDir, "struct-diff.txt")

	err := ioutil.WriteFile(diffPath, []byte(diff), 0600)
	must(t, err, "failed to create struct diff output file")

	t.Logf("struct diff: %s", diffPath)
}

func dumpDocs(t *testing.T, document *doc.Document, generatedDoc *doc.Document, originalXML []byte, generatedXML []byte, errors ...error) {
	t.Helper()

	if !t.Failed() && os.Getenv("DUMP_DOCUMENTS") != "true" {
		return
	}

	outDir := ensureDebugDir(t)

	if len(originalXML) > 0 {
		originalPath := filepath.Join(outDir, "original.xml")
		err := ioutil.WriteFile(originalPath, originalXML, 0600)
		must(t, err, "failed to create original XML output file")
		t.Logf("Original XML: %s", originalPath)
	}

	if len(generatedXML) > 0 {
		generatedPath := filepath.Join(outDir, "generated.xml")
		err := ioutil.WriteFile(generatedPath, generatedXML, 0600)
		must(t, err, "failed to create generated XML output file")
		t.Logf("Generated XML: %s", generatedPath)
	}

	if len(errors) > 0 {
		errorsPath := filepath.Join(outDir, "errors.log")

		logFile, err := os.Create(errorsPath)
		must(t, err, "failed to create error log file")
		defer func() {
			must(t, logFile.Close(),
				"failed to close error log file")
		}()

		for _, lErr := range errors {
			if lErr == nil {
				continue
			}

			_, err := fmt.Fprintln(logFile, lErr.Error())
			must(t, err, "failed to write to error log file")
		}
	}

	if document != nil {
		docPath := filepath.Join(outDir, "navigadoc.json")
		dumpNavigaDoc(t, document, docPath)
		t.Logf("NavigaDoc: %s", docPath)
	}

	if generatedDoc != nil {
		docPath := filepath.Join(outDir, "generateddoc.json")
		dumpNavigaDoc(t, generatedDoc, docPath)
		t.Logf("GeneratedDoc: %s", docPath)
	}
}

func getEmbedCode(newsItem *newsml.NewsItem) (*etree.Element, error) {
	for _, e := range newsItem.ContentSet.InlineXML.Idf.Group[0].Child {
		if e.Object != nil && e.Object.Type == "x-im/iframely" {
			obj := e.Object

			tree := etree.NewDocument()
			err := tree.ReadFromString(obj.Data.Raw)
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling xml: %v", err)
			}

			for _, element := range tree.ChildElements() {
				if element.Tag == "embedCode" {
					return element, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("embedcode not found")
}

func getCDATAString(e *etree.Element) string {
	if len(e.Child) == 0 {
		return ""
	}

	buf := strings.Builder{}
	for i := range e.Child {
		switch c := e.Child[i].(type) {
		case *etree.CharData:
			if c.IsWhitespace() {
				continue
			}
			buf.WriteString(c.Data)
		case *etree.Element:
			return buf.String()
		}
	}

	return buf.String()
}

func dumpNavigaDoc(t *testing.T, document *doc.Document, docPath string) {
	docFile, err := os.Create(docPath)
	must(t, err, "failed to create navigadoc output file")

	defer func(t *testing.T, c io.Closer) {
		err := c.Close()
		if err != nil {
			t.Logf("failed to close navigadoc.json: %s", err.Error())
		}
	}(t, docFile)

	enc := json.NewEncoder(docFile)

	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	must(t, enc.Encode(document), "failed to write navigadoc output to file")
}
