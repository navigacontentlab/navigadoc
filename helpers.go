package docformat

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/xeipuuv/gojsonschema"

	"bitbucket.org/infomaker/doc-format/v2/newsml"

	"bitbucket.org/infomaker/doc-format/v2/doc"
	"github.com/beevik/etree"
	"github.com/google/uuid"
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

// WalkXMLDocumentElements For each element in the document call the ElementVistor functions provided
func WalkXMLDocumentElements(document *etree.Document, args []interface{}, fns ...ElementVisitor) error {
	allElements := document.FindElements(fmt.Sprintf("/%s//", document.Root().Tag))
	for _, element := range allElements {
		for i := range fns {
			if err := fns[i](element, args...); err != nil {
				return err
			}
		}
	}

	return nil
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

// ValidateDocumentDates Walks each document block and validates dates.
// Requires external configuration
func ValidateDocumentDates(document *doc.Document, dateConfig newsml.DateConfig) error {
	var err error

	// First the document root members
	val := reflect.ValueOf(*document)
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		if dateConfig.IsDate(newsml.Block, typeField.Name) {
			valueField := val.Field(i)
			switch valueField.Interface().(type) {
			case *time.Time:
				tim := valueField.Interface().(*time.Time)
				err = ValidateDate(dateConfig, newsml.Block, typeField.Name, tim.Format(time.RFC3339))
			case string:
				err = ValidateDate(dateConfig, newsml.Block, typeField.Name, fmt.Sprintf("%v", valueField.Interface()))
			}
		}
		if err != nil {
			return fmt.Errorf("error: %s", err)
		}
	}

	err = WalkDocument(document, nil, func(block doc.Block, args ...interface{}) (doc.Block, error) {
		val := reflect.ValueOf(block)
		for i := 0; i < val.NumField(); i++ {
			typeField := val.Type().Field(i)
			if dateConfig.IsDate(newsml.Block, typeField.Name) {
				valueField := val.Field(i)
				switch valueField.Interface().(type) {
				case *time.Time:
					tim := valueField.Interface().(*time.Time)
					err = ValidateDate(dateConfig, newsml.Block, typeField.Name, tim.Format(time.RFC3339))
				case string:
					err = ValidateDate(dateConfig, newsml.Block, typeField.Name, fmt.Sprintf("%v", valueField.Interface()))
				}
			}
			if err != nil {
				return block, fmt.Errorf("error: %s", err)
			}
		}

		if len(block.Data) > 0 {
			for key, value := range block.Data {
				if dateConfig.IsDate(newsml.Block, key) {
					err = ValidateDate(dateConfig, newsml.Block, key, value)
				}
				if err != nil {
					return block, fmt.Errorf("error in block data %s: %s", block.Name, err.Error())
				}
			}
		}

		return block, nil
	})

	return nil
}

// ValidateNewsMLDates Used to validate dates
// Requires external configuration
func ValidateNewsMLDates(xmlDoc *etree.Document, dateConfig newsml.DateConfig) error {
	var err error

	err = WalkXMLDocumentElements(xmlDoc, nil, func(element *etree.Element, args ...interface{}) error {
		var typeAttribute string
		a := element.SelectAttr("type")
		if a != nil {
			typeAttribute = a.Value
		}

		if dateConfig.IsDate(newsml.TagNode, element.Tag) {
			err = ValidateDate(dateConfig, newsml.TagNode, element.Tag, element.Text())
		} else if typeAttribute != "" && dateConfig.IsDate(newsml.TypeNode, typeAttribute) {
			valueAttr := dateConfig.GetAttribute(newsml.TypeNode, typeAttribute)
			d := element.SelectAttr(valueAttr)
			if d != nil {
				err = ValidateDate(dateConfig, newsml.TypeNode, typeAttribute, d.Value)
			} else {
				err = fmt.Errorf("no attribute found for %s", valueAttr)
			}
		}
		if err != nil {
			return fmt.Errorf("error %s: %s", element.GetPath(), err.Error())
		}

		return nil
	})

	return err
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
// FixNewsMLUUIDsAndNamespaces Validate UUID format and convert to lowercase.
// Add namespace to <links>, <object>, <metadata> and <idf> where missing
// Returns RequiredArgumentError if link is missing required arguments
func FixNewsMLUUIDsAndNamespaces(document *etree.Document) error {
	err := WalkXMLDocumentElements(document, nil,
		ValidateAndLowercaseNewsMLUUIDs,
		FixNewsMLNamespaces,
	)

	return err
}

// ValidateNewsMLUUIDsAndFixNamespaces Validate UUID format.
// Add namespace to <links>, <object>, <metadata> and <idf> where missing
// Returns RequiredArgumentError if link is missing required arguments
func ValidateNewsMLUUIDsAndFixNamespaces(document *etree.Document) error {
	err := WalkXMLDocumentElements(document, nil,
		ValidateNewsMLUUIDs,
		FixNewsMLNamespaces,
	)

	return err
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

// ValidateDate Validate element date. Requires external config.
func ValidateDate(dateConfig newsml.DateConfig, eType string, name string, date string) error {
	format := dateConfig.GetFormat(eType, name)
	if format == "" {
		format = "RFC3339"
	}

	if date == "" && dateConfig.AllowBlank(eType, name) {
		return nil
	}

	var err error
	hasOffset := regexp.MustCompile("(Z|[+-][0-1][0-5]:[0-5][0-9])$")

	switch format {
	case "RFC3339", "RFC3339Nano":
		if !hasOffset.MatchString(date) {
			err = fmt.Errorf("time \"%s\" UTC offset is missing or invalid", date)
			break
		}
		_, err = time.Parse(time.RFC3339Nano, date)
	default:
		// if not blank assume a regex pattern for now
		if format != "" {
			validDate, err := regexp.Compile(format)
			if err != nil {
				return InvalidArgumentError{
					Msg: fmt.Sprintf("invalid date format: %s", format),
				}
			}
			if !validDate.MatchString(date) {
				return InvalidArgumentError{
					Msg: fmt.Sprintf("invalid date %s %s: %s", name, date, err),
				}
			}
		}
	}

	if err != nil {
		if dateConfig.AllowString(eType, name) {
			return nil
		}
		return InvalidArgumentError{
			Msg: fmt.Sprintf("invalid date %s %s: %s", name, date, err),
		}
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
