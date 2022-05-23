// Package navigadoc provides utility functions to the Naviga Doc format
package navigadoc

import (

	// embed navigadoc schema
	_ "embed"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/navigacontentlab/navigadoc/doc"
)

//go:embed schema/navigadoc-schema.json
// NavigaDocSchema embedded schemas
var NavigaDocSchema string

func CheckForEmptyBlocks(document *doc.Document) error {
	var err error

	value := reflect.ValueOf(*document)
	if value.IsZero() {
		return ErrEmptyDoc
	}

	for i, block := range document.Meta {
		err = CheckForEmptyBlocksRecursive(block, "meta", i, nil)
		if err != nil {
			return err
		}
	}
	for i, block := range document.Links {
		err = CheckForEmptyBlocksRecursive(block, "links", i, nil)
		if err != nil {
			return err
		}
	}
	for i, block := range document.Content {
		err = CheckForEmptyBlocksRecursive(block, "content", i, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckForEmptyBlocksRecursive(block doc.Block, kind string, idx int, path []string) error {
	path = append(path, kind, strconv.Itoa(idx))

	value := reflect.ValueOf(block)
	if value.IsZero() {
		return fmt.Errorf("%w: %s", ErrEmptyBlock, strings.Join(path, "/"))
	}

	var err error
	for i, block := range block.Meta {
		err = CheckForEmptyBlocksRecursive(block, "meta", i, path)
		if err != nil {
			return err
		}
	}

	for i, block := range block.Links {
		err = CheckForEmptyBlocksRecursive(block, "links", i, path)
		if err != nil {
			return err
		}
	}
	for i, block := range block.Content {
		err = CheckForEmptyBlocksRecursive(block, "content,", i, path)
		if err != nil {
			return err
		}
	}
	return nil
}
