package navigadoc

import (
	"errors"
	"fmt"

	"strings"

	"github.com/google/uuid"
	"github.com/navigacontentlab/navigadoc/doc"
	"github.com/xeipuuv/gojsonschema"
)

type BlockVisitor func(block doc.Block, args ...interface{}) (doc.Block, error)

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
