package navigadoc_test

import (
	"fmt"
	"testing"

	"github.com/navigacontentlab/navigadoc"
	"github.com/navigacontentlab/navigadoc/doc"
)

var blockFields = []string{"ID", "UUID", "URI", "URL", "Type", "Title", "Rel", "Name", "Value", "ContentType", "Role"}

type blocksToDeleteTestCase struct {
	description    string
	document       doc.Document
	blocksToDelete []doc.Block
	asserter       func(t *testing.T, doc *doc.Document, description string)
}

func createAsserter(numberOfLinks int, numberOfMeta int, numberOfContent int) func(t *testing.T, doc *doc.Document, description string) {
	return func(t *testing.T, document *doc.Document, description string) {
		if len(document.Links) != numberOfLinks {
			t.Errorf("expected %d links blocks in document, found %d; %s", numberOfLinks, len(document.Links), description)
		}

		if len(document.Meta) != numberOfMeta {
			t.Errorf("expected %d meta blocks in document, found %d; %s", numberOfMeta, len(document.Meta), description)
		}

		if len(document.Content) != numberOfContent {
			t.Errorf("expected %d content blocks in document, found %d; %s", numberOfContent, len(document.Content), description)
		}
	}
}

var blockWithValues = []doc.Block{
	{
		ID:          "ID",
		UUID:        "UUID",
		URI:         "URI",
		URL:         "URL",
		Type:        "Type",
		Title:       "Title",
		Rel:         "Rel",
		Name:        "Name",
		Value:       "Value",
		ContentType: "ContentType",
		Role:        "Role",
	},
}

func createBlockSliceWithOneValue(key, value string) []doc.Block {
	return []doc.Block{createBlockWithOneValue(key, value)}
}

func createBlockWithOneValue(key, value string) doc.Block {
	switch key {
	case "ID":
		return doc.Block{ID: value}
	case "UUID":
		return doc.Block{UUID: value}
	case "URI":
		return doc.Block{URI: value}
	case "URL":
		return doc.Block{URL: value}
	case "Type":
		return doc.Block{Type: value}
	case "Title":
		return doc.Block{Title: value}
	case "Rel":
		return doc.Block{Rel: value}
	case "Name":
		return doc.Block{Name: value}
	case "Value":
		return doc.Block{Value: value}
	case "ContentType":
		return doc.Block{ContentType: value}
	case "Role":
		return doc.Block{Role: value}
	}
	return doc.Block{}
}

func TestDeleteBlocks(t *testing.T) {
	cases := []blocksToDeleteTestCase{
		{
			description: "all links should be removed from document, matching all fields in block",
			document: doc.Document{
				Links:   blockWithValues,
				Meta:    blockWithValues,
				Content: blockWithValues,
			},
			blocksToDelete: blockWithValues,
			asserter:       createAsserter(0, 0, 0),
		},
		{
			description: "delete only with ID and more values in document block",
			document: doc.Document{
				Links:   blockWithValues,
				Meta:    blockWithValues,
				Content: blockWithValues,
			},
			blocksToDelete: []doc.Block{
				{
					ID: "ID",
				},
			},
			asserter: createAsserter(0, 0, 0),
		},
		{
			description: "no match no delete ID",
			document: doc.Document{
				Links:   blockWithValues,
				Meta:    blockWithValues,
				Content: blockWithValues,
			},
			blocksToDelete: []doc.Block{
				{
					ID: "---",
				},
			},
			asserter: createAsserter(1, 1, 1),
		},
		{
			description: "no match no delete UUID",
			document: doc.Document{
				Links:   blockWithValues,
				Meta:    blockWithValues,
				Content: blockWithValues,
			},
			blocksToDelete: []doc.Block{
				{
					UUID: "---",
				},
			},
			asserter: createAsserter(1, 1, 1),
		},
		{
			description: "no match no delete Type",
			document: doc.Document{
				Links:   blockWithValues,
				Meta:    blockWithValues,
				Content: blockWithValues,
			},
			blocksToDelete: []doc.Block{
				{
					Type: "---",
				},
			},
			asserter: createAsserter(1, 1, 1),
		},
	}

	for _, a := range blockFields {
		cases = append(cases, blocksToDeleteTestCase{
			description: fmt.Sprintf("no match no delete %s", a),
			document: doc.Document{
				Links:   blockWithValues,
				Meta:    blockWithValues,
				Content: blockWithValues,
			},
			blocksToDelete: createBlockSliceWithOneValue(a, "----"),
			asserter:       createAsserter(1, 1, 1),
		})
	}

	for _, a := range blockFields {
		cases = append(cases, blocksToDeleteTestCase{
			description: fmt.Sprintf("match delete %s", a),
			document: doc.Document{
				Links:   blockWithValues,
				Meta:    blockWithValues,
				Content: blockWithValues,
			},
			blocksToDelete: createBlockSliceWithOneValue(a, a),
			asserter:       createAsserter(0, 0, 0),
		})
	}

	for _, test := range cases {
		test.asserter(t, navigadoc.DeleteBlocks(test.document, test.blocksToDelete), test.description)
	}
}

func TestDeDuplicateLinks(t *testing.T) {
	links := []doc.Block{
		{Type: "Type"},
		{Type: "Type"},
		{Type: "Type"},
		{UUID: "UUID"},
		{Type: "Type"},
		{Type: "Type"},
		{Type: "Type", UUID: "UUID"},
	}
	linksToDeduplicate := []doc.Block{
		{Type: "Type"},
	}
	if len(navigadoc.DeDuplicateLinks(links, linksToDeduplicate)) != 2 {
		t.Errorf("expected 2 links got %d", len(navigadoc.DeDuplicateLinks(links, linksToDeduplicate)))
	}
}

type BlocksToReplaceTestCase struct {
	description     string
	document        doc.Document
	blocksToReplace []navigadoc.BlockReplacement
	asserter        func(t *testing.T, document doc.Document, description string)
}

func getValueFromBlockSlice(field string, blocks []doc.Block, i int) (string, error) {
	switch field {
	case "ID":
		return blocks[i].ID, nil
	case "UUID":
		return blocks[i].UUID, nil
	case "URI":
		return blocks[i].URI, nil
	case "URL":
		return blocks[i].URL, nil
	case "Type":
		return blocks[i].Type, nil
	case "Title":
		return blocks[i].Title, nil
	case "Rel":
		return blocks[i].Rel, nil
	case "Name":
		return blocks[i].Name, nil
	case "Value":
		return blocks[i].Value, nil
	case "ContentType":
		return blocks[i].ContentType, nil
	case "Role":
		return blocks[i].Role, nil
	default:
		return "", fmt.Errorf("no such field %s", field)
	}
}

func createReplaceBlocksAsserter(field string) func(t *testing.T, d doc.Document, description string) {
	return func(t *testing.T, d doc.Document, description string) {
		if len(d.Links) != 1 {
			t.Errorf("expected 1 links in document was %d ; %s", len(d.Links), description)
		}

		if len(d.Meta) != 1 {
			t.Errorf("expected 1 meta in document was %d ; %s", len(d.Links), description)
		}

		if len(d.Content) != 1 {
			t.Errorf("expected 1 content in document was %d ; %s", len(d.Links), description)
		}

		// links
		actualValue, err := getValueFromBlockSlice(field, d.Links, 0)
		if err != nil {
			t.Errorf("failed to get value from block slice, %v", err)
		}
		if actualValue != field+"-replaced" {
			t.Errorf("in doc.Links[] expected %s '%s', was %s; %s", field, field+"-replaced", actualValue, description)
		}

		// meta
		actualValue, err = getValueFromBlockSlice(field, d.Meta, 0)
		if err != nil {
			t.Errorf("failed to get value from block slice, %v", err)
		}
		if actualValue != field+"-replaced" {
			t.Errorf("in doc.Meta[] expected %s '%s', was %s; %s", field, field+"-replaced", actualValue, description)
		}

		// content
		actualValue, err = getValueFromBlockSlice(field, d.Content, 0)
		if err != nil {
			t.Errorf("failed to get value from block slice, %v", err)
		}
		if actualValue != field+"-replaced" {
			t.Errorf("in doc.Content[] expected  %s '%s', was %s; %s", field, field+"-replaced", actualValue, description)
		}
	}
}

type TestBlockToReplace struct {
	OldBlock doc.Block
	NewBlock doc.Block
}

func (t TestBlockToReplace) GetOldBlock() doc.Block {
	return t.OldBlock
}

func (t TestBlockToReplace) GetNewBlock() doc.Block {
	return t.NewBlock
}

func TestReplaceBlocks(t *testing.T) {
	blocksToReplaceTestCase := []BlocksToReplaceTestCase{
		{
			description: "empty doc.Type replaces to '-----'",
			document: doc.Document{
				Links: []doc.Block{
					{Type: "any"},
					{Type: "any"},
					{Type: "any"},
					{Type: "any"},
					{Type: "any"},
				},
			},
			blocksToReplace: []navigadoc.BlockReplacement{
				TestBlockToReplace{
					OldBlock: doc.Block{Type: ""},
					NewBlock: doc.Block{
						Type: "-----",
						Data: map[string]string{
							"key": "value",
						},
					},
				},
			},
			asserter: func(t *testing.T, d doc.Document, description string) {
				if len(d.Links) != 5 {
					t.Errorf("expected 5 links in document was %d ; %s", len(d.Links), description)
				}
				for _, link := range d.Links {
					if link.Type != "-----" {
						t.Errorf("expected doc.Links[0].Type = '------', was %s; %s", d.Links[0].Type, description)
					}

					// assert that data blocks in new block is copied
					if len(link.Data) != 1 {
						t.Errorf("expected len(doc.Links[0].Data) = %d, was %d; %s", 1, len(d.Links[0].Data), description)
					}
				}
			},
		},
	}

	for _, field := range blockFields {
		testCase := BlocksToReplaceTestCase{
			description: fmt.Sprintf("replace block %[1]s.%[1]s with %[1]s-replaced", field),
			document: doc.Document{
				Links:   createBlockSliceWithOneValue(field, field),
				Content: createBlockSliceWithOneValue(field, field),
				Meta:    createBlockSliceWithOneValue(field, field),
			},
			blocksToReplace: []navigadoc.BlockReplacement{
				TestBlockToReplace{
					OldBlock: createBlockWithOneValue(field, field),
					NewBlock: createBlockWithOneValue(field, field+"-replaced"),
				},
			},
			asserter: createReplaceBlocksAsserter(field),
		}
		blocksToReplaceTestCase = append(blocksToReplaceTestCase, testCase)
	}

	for _, tc := range blocksToReplaceTestCase {
		result := navigadoc.ReplaceBlocks(tc.document, tc.blocksToReplace)
		tc.asserter(t, *result, tc.description)
	}
}

type MergePropertiesTestCase struct {
	description        string
	existingProperties []doc.Property
	newProperties      []doc.Property
	asserter           func(t *testing.T, mergedProperties []doc.Property, description string)
}

func createMergePropertiesAsserter(numberOfExpectedProps int) func(t *testing.T, mergedProperties []doc.Property, description string) {
	return func(t *testing.T, mergedProps []doc.Property, description string) {
		if len(mergedProps) != numberOfExpectedProps {
			t.Errorf("expected %d properties, but was %d; %s", numberOfExpectedProps, len(mergedProps), description)
		}
	}
}

func TestMergeProperties(t *testing.T) {
	testCases := []MergePropertiesTestCase{
		{
			description: "test add new property",
			existingProperties: []doc.Property{
				{
					Name:  "one",
					Value: "value",
				},
			},
			newProperties: []doc.Property{
				{
					Name:  "two",
					Value: "value",
				},
			},
			asserter: createMergePropertiesAsserter(2),
		},
		{
			description: "test add same property",
			existingProperties: []doc.Property{
				{
					Name:  "one",
					Value: "value",
				},
			},
			newProperties: []doc.Property{
				{
					Name:  "one",
					Value: "value",
				},
			},
			asserter: createMergePropertiesAsserter(1),
		},
		{
			description: "test add same name different value",
			existingProperties: []doc.Property{
				{
					Name:  "one",
					Value: "value",
				},
			},
			newProperties: []doc.Property{
				{
					Name:  "one",
					Value: "value-changed",
				},
			},
			asserter: createMergePropertiesAsserter(1),
		},
	}

	for _, tc := range testCases {
		mergedProperties := navigadoc.MergeProperties(tc.existingProperties, tc.newProperties)
		tc.asserter(t, mergedProperties, tc.description)
	}
}
