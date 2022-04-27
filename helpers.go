package navigadoc

import (
	"errors"
	"fmt"

	"strings"

	"github.com/beevik/etree"
	"github.com/google/uuid"
	"github.com/navigacontentlab/navigadoc/doc"
	"github.com/xeipuuv/gojsonschema"
)

type BlockVisitor func(block doc.Block, args ...interface{}) (doc.Block, error)
type ElementVisitor func(element *etree.Element, args ...interface{}) error

// WalkDocument For each Block in the Document call the BlockVisitor functions
func WalkDocument(document *doc.Document, args []interface{}, fns ...BlockVisitor) error {
	if document == nil || fns == nil {
		return nil
	}

	if err := WalkBlocks(document.Content, args, fns...); err != nil {
		return err
	}

	if err := WalkBlocks(document.Meta, args, fns...); err != nil {
		return err
	}

	if err := WalkBlocks(document.Links, args, fns...); err != nil {
		return err
	}

	return nil
}

// WalkBlocks For each Block in the array call the BlockVisitor functions
func WalkBlocks(s []doc.Block, args []interface{}, fns ...BlockVisitor) error {
	for i := range s {
		block, err := WalkBlock(s[i], args, fns...)
		if err != nil {
			return err
		}
		s[i] = block
	}

	return nil
}

// WalkBlock For each block in the document call the BlockVisitor functions provided
func WalkBlock(block doc.Block, args []interface{}, fns ...BlockVisitor) (doc.Block, error) {
	for i := range fns {
		var err error
		block, err = fns[i](block, args...)
		if err != nil {
			return block, err
		}
	}

	if err := WalkBlocks(block.Content, args, fns...); err != nil {
		return block, err
	}
	if err := WalkBlocks(block.Meta, args, fns...); err != nil {
		return block, err
	}

	if err := WalkBlocks(block.Links, args, fns...); err != nil {
		return block, err
	}

	return block, nil
}

// Deprecated: UUID handling in OpenContent to be case-insensitive
// ValidateAndLowercaseDocumentUUIDs Walks each document block to validate and lowercase convert UUIDs
func ValidateAndLowercaseDocumentUUIDs(block doc.Block, args ...interface{}) (doc.Block, error) {
	err := validateUUID(block.UUID)
	if err != nil {
		return block, InvalidArgumentError{
			Msg: fmt.Sprintf("uuid error %s[%s]: %s", block.Type, block.UUID, err),
		}
	}

	block.UUID = strings.ToLower(block.UUID)

	return block, nil
}

// ValidateDocumentUUIDs Walks each document block to validate and lowercase convert UUIDs
func ValidateDocumentUUIDs(block doc.Block, args ...interface{}) (doc.Block, error) {
	err := validateUUID(block.UUID)
	if err != nil {
		return block, InvalidArgumentError{
			Msg: fmt.Sprintf("uuid error %s[%s]: %s", block.Type, block.UUID, err),
		}
	}

	return block, nil
}

func validateUUID(theUUID string) error {
	if theUUID == "" {
		return nil
	}

	_, err := uuid.Parse(theUUID)
	if err != nil {
		return fmt.Errorf("invalid uuid: %v", err)
	}

	return nil
}

// Deprecated: UUID handling in OpenContent to be case-insensitive
// ValidateAndLowercaseNewsMLUUIDs Validate UUID format and convert to lowercase
// Remove empty uuid attributes from link elements
func ValidateAndLowercaseNewsMLUUIDs(element *etree.Element, args ...interface{}) error {
	var err error

	uuidAttr := element.SelectAttr("uuid")
	// allows for tags in <data> named uuid to be checked
	if element.Tag != "uuid" && uuidAttr == nil {
		return nil
	}

	var theUUID string
	if uuidAttr != nil {
		theUUID = uuidAttr.Value
	} else {
		theUUID = element.Text()
	}

	if theUUID != "" {
		err = ValidateAndLowercaseNewsMLUUID(element, theUUID)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateNewsMLUUIDs Validate UUID format, and remove empty uuid attributes from link elements
func ValidateNewsMLUUIDs(element *etree.Element, args ...interface{}) error {
	var err error

	uuidAttr := element.SelectAttr("uuid")
	// allows for tags in <data> named uuid to be checked
	if element.Tag != "uuid" && uuidAttr == nil {
		return nil
	}

	var theUUID string
	if uuidAttr != nil {
		theUUID = uuidAttr.Value
	} else {
		theUUID = element.Text()
	}

	if theUUID != "" {
		err = validateUUID(theUUID)
		if err != nil {
			return InvalidArgumentError{
				Msg: fmt.Sprintf("uuid error %s[%s]: %s", element.GetPath(), theUUID, err),
			}
		}
	}

	return nil
}

// FixNewsMLNamespaces Add namespace to <links>, <object>, <metadata> and <idf> where missing
// Returns RequiredArgumentError if link is missing required arguments
func FixNewsMLNamespaces(element *etree.Element, args ...interface{}) error {
	switch element.Tag {
	case "link":
		uuidAttr := element.SelectAttr("uuid")
		if uuidAttr != nil && uuidAttr.Value == "" {
			element.RemoveAttr("uuid")
		}
		if rel := element.SelectAttr("rel"); rel == nil {
			return RequiredArgumentError{
				Msg: fmt.Sprintf("%s[@rel]", element.GetPath()),
			}
		}
	case "links", "object", "metadata":
		parent := element.Parent().Tag
		if parent == "itemMeta" || parent == "contentMeta" {
			if xmlns := element.SelectAttr("xmlns"); xmlns == nil || xmlns.Value == "" {
				element.CreateAttr("xmlns", "http://www.infomaker.se/newsml/1.0")
			}
		}
	case "idf":
		if xmlns := element.SelectAttr("xmlns"); xmlns == nil || xmlns.Value == "" {
			element.CreateAttr("xmlns", "http://www.infomaker.se/idf/1.0")
		}
	}
	return nil
}

// Deprecated: UUID handling in OpenContent to be case-insensitive
// ValidateAndLowercaseNewsMLUUID Validate the UUID format and convert to lowercase
func ValidateAndLowercaseNewsMLUUID(element *etree.Element, theUUID string) error {
	err := validateUUID(theUUID)
	if err != nil {
		return InvalidArgumentError{
			Msg: fmt.Sprintf("uuid error %s[%s]: %s", element.GetPath(), theUUID, err),
		}
	}

	uuidAttr := element.CreateAttr("uuid", strings.ToLower(theUUID))
	if uuidAttr == nil {
		return fmt.Errorf("lowercasing uuid returned nil")
	}

	return nil
}

// ValidateNavigadocJSON validates the Navigadoc against a JSON Schema
// The error array contains schema issues
// A non-nil error indicates an error with the validator
func ValidateNavigadocJSON(document string) ([]error, error) {
	schemaLoader := gojsonschema.NewStringLoader(NavigaDocSchema)
	documentLoader := gojsonschema.NewStringLoader(document)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	}

	var errs []error
	if result.Valid() {
		return nil, nil
	}

	for _, e := range result.Errors() {
		errs = append(errs, errors.New(e.String()))
	}
	return errs, nil
}
