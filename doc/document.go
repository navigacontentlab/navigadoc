// Code generated by doc-generator. DO NOT EDIT.
package doc

import (
	"time"
)

type Document struct {
	// UUID is a unique ID for the document, this can be a random v4
	// UUID, or a URI-derived v5 UUID.
	UUID string `json:"uuid,omitempty"`
	// Type is the content type of the document.
	Type string `json:"type,omitempty"`
	// URI identifies the document (in a more human-readable way than
	// the UUID)
	URI string `json:"uri,omitempty"`
	// URL is the browseable location of the document (if any)
	URL string `json:"url,omitempty"`
	// Title is the title of the document, often used as the headline
	// when the document is displayed.
	Title string `json:"title,omitempty"`
	// Path is the path on which the document can be exposed when
	// consumed through a website.
	Path string `json:"path,omitempty"`
	// Products is a list of products that the document should be used
	// in.
	//
	// Deprecated: Do not use.
	Products []string `json:"products,omitempty"`
	// Created is the initial creation time of the document.
	Created *time.Time `json:"created,omitempty"`
	// Modified is the modified time as is should be presented to end
	// users, the actual modified timestamp is recorded in the document
	// commit. There is probably no reason not to update this timestamp
	// when doing manual edits.  Automated tools and systems should
	// probably leave it alone tho.
	Modified *time.Time `json:"modified,omitempty"`
	// Published is the published timestamp as it should be presented to
	// end users. The actual published timestamp is recorded in the
	// document commits in the "usable" branch. This shouldn't be
	// touched after the initial publishing of the document.
	Published *time.Time `json:"published,omitempty"`
	// Content is the content of the documen, this is essentially what
	// gets rendered on the page when you view a document.
	Content []Block `json:"content,omitempty"`
	// Meta is the metadata for a document, this could be stuff like
	// open graph tags and content profile information.
	Meta []Block `json:"meta,omitempty"`
	// Links are links to other resources and entities. This could be
	// links to categories and subject for the document, or authors.
	Links []Block `json:"links,omitempty"`
	// Properties are header-like properties for a document. This is
	// mainly used as a bucket for document-level stuff that needs to be
	// preserved when converting to and from other document formats.
	Properties []Property `json:"properties,omitempty"`
	// Source is the name of the source of the document, usually the
	// name of the application that generated it (or allowed a user to
	// generate it).
	Source string `json:"source,omitempty"`
	// Language is the language used in the document as an IETF language
	// tag. F.ex. "en", "en-UK", "es", or "sv-SE".
	Language string `json:"language,omitempty"`
	// A free form field detailing the status for the document, for example:
	// "draft" or "withheld".
	Status      string     `json:"status,omitempty"`
	Unpublished *time.Time `json:"unpublished,omitempty"`
	Provider    string     `json:"provider,omitempty"`
}
type Property struct {
	Name       string            `json:"name,omitempty"`
	Value      string            `json:"value,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}
type Block struct {
	// ID is the block ID
	ID string `json:"id,omitempty"`
	// UUID is used to reference another Document in a block.
	UUID string `json:"uuid,omitempty"`
	// URI is used to reference another entity in a document.
	URI string `json:"uri,omitempty"`
	// URL is a browseable URL for the the block.
	URL string `json:"url,omitempty"`
	// Type is the type of the block
	Type string `json:"type,omitempty"`
	// Title is the title/headline of the block, typically used in the
	// presentation of the block.
	Title string `json:"title,omitempty"`
	// Data contains block data
	Data map[string]string `json:"data,omitempty"`
	// Relationship describes the relationship to the document/parent
	// entity
	Rel string `json:"rel,omitempty"`
	// Name is a name for the block. An alternative to "rel" when
	// relationship is a term that doesn't fit.
	Name string `json:"name,omitempty"`
	// Value is a value for the block. Useful when we want to store a
	// primitive value.
	Value string `json:"value,omitempty"`
	// ContentType is used to describe the content type of the
	// block/linked entity if it differs from the type of the block.
	ContentType string `json:"contentType,omitempty"`
	// Links are used to link to other resources and documents.
	Links []Block `json:"links,omitempty"`
	// Content is used to embed content blocks.
	Content []Block `json:"content,omitempty"`
	// Meta is used to embed metadata
	Meta []Block `json:"meta,omitempty"`
	// Role is used for
	Role string `json:"role,omitempty"`
}
