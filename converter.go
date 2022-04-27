// Package docformat provides helper functions for converting
// between CCA Doc and Infomaker NewsML (IMNML) format
package navigadoc

import (
	"bufio"
	"bytes"

	// embed navigadoc schema
	_ "embed"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/navigacontentlab/navigadoc/doc"
)

//go:embed schema/navigadoc-schema.json
// NavigaDocSchema embedded schema
var NavigaDocSchema string

const peekSize = 1024

func peekRootTag(r *bufio.Reader) (string, error) {
	// Peek at the first bytes of the document so that we can
	// sniff the document type from the root tag.
	head, err := r.Peek(peekSize)
	if err != nil && err != bufio.ErrBufferFull && err != io.EOF {
		return "", fmt.Errorf("failed to read the beginning of the XML document: %w", err)
	}

	dec := xml.NewDecoder(bytes.NewReader(head))

	for {
		// Read tokens from the XML document in a stream.
		t, err := dec.Token()
		if err == io.EOF {
			return "", fmt.Errorf("root tag was not found within the first %d bytes of the document", peekSize)
		}
		if err != nil {
			return "", fmt.Errorf("failed to parse XML: %w", err)
		}

		root, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		return root.Name.Local, nil
	}
}

func checkForEmptyBlocks(document *doc.Document) error {
	var err error

	value := reflect.ValueOf(*document)
	if value.IsZero() {
		return ErrEmptyDoc
	}

	for i, block := range document.Meta {
		err = checkForEmptyBlocksRecursive(block, "meta", i, nil)
		if err != nil {
			return err
		}
	}
	for i, block := range document.Links {
		err = checkForEmptyBlocksRecursive(block, "links", i, nil)
		if err != nil {
			return err
		}
	}
	for i, block := range document.Content {
		err = checkForEmptyBlocksRecursive(block, "content", i, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkForEmptyBlocksRecursive(block doc.Block, kind string, idx int, path []string) error {
	path = append(path, kind, strconv.Itoa(idx))

	value := reflect.ValueOf(block)
	if value.IsZero() {
		return fmt.Errorf("%w: %s", ErrEmptyBlock, strings.Join(path, "/"))
	}

	var err error
	for i, block := range block.Meta {
		err = checkForEmptyBlocksRecursive(block, "meta", i, path)
		if err != nil {
			return err
		}
	}

	for i, block := range block.Links {
		err = checkForEmptyBlocksRecursive(block, "links", i, path)
		if err != nil {
			return err
		}
	}
	for i, block := range block.Content {
		err = checkForEmptyBlocksRecursive(block, "content,", i, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func mergeSections(dflt []string, cust []string) []string {
	allSections := append(dflt, cust...)

	noDups := map[string]bool{}
	for _, s := range allSections {
		noDups[s] = true
	}

	var keys []string

	for key := range noDups {
		switch key {
		case "itemmeta-newsitem":
			delete(noDups, "contentmeta")
		case "itemmeta-concept":
			delete(noDups, "concept")
		case "itemmeta-conceptitem":
			delete(noDups, "conceptitem")
		case "itemmeta-planning":
			delete(noDups, "planning")
		}
	}
	for key := range noDups {
		keys = append(keys, key)
	}

	return keys
}
