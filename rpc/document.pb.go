// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.19.4
// source: rpc/document.proto

package rpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Document is the content, doh!
type Document struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// UUID is a unique ID for the document, this can be a random v4
	// UUID, or a URI-derived v5 UUID.
	Uuid string `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	// Type is the content type of the document.
	Type string `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	// URI identifies the document (in a more human-readable way than
	// the UUID)
	Uri string `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	// URL is the browseable location of the document (if any)
	Url string `protobuf:"bytes,5,opt,name=url,proto3" json:"url,omitempty"`
	// Title is the title of the document, often used as the headline
	// when the document is displayed.
	Title string `protobuf:"bytes,6,opt,name=title,proto3" json:"title,omitempty"`
	// Path is the path on which the document can be exposed when
	// consumed through a website.
	Path string `protobuf:"bytes,8,opt,name=path,proto3" json:"path,omitempty"`
	// Products is a list of products that the document should be used
	// in.
	//
	// Deprecated: Do not use.
	Products []string `protobuf:"bytes,9,rep,name=products,proto3" json:"products,omitempty"`
	// Created is the initial creation time of the document.
	Created *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=created,proto3" json:"created,omitempty"`
	// Modified is the modified time as is should be presented to end
	// users, the actual modified timestamp is recorded in the document
	// commit. There is probably no reason not to update this timestamp
	// when doing manual edits.  Automated tools and systems should
	// probably leave it alone tho.
	Modified *timestamppb.Timestamp `protobuf:"bytes,11,opt,name=modified,proto3" json:"modified,omitempty"`
	// Published is the published timestamp as it should be presented to
	// end users. The actual published timestamp is recorded in the
	// document commits in the "usable" branch. This shouldn't be
	// touched after the initial publishing of the document.
	Published *timestamppb.Timestamp `protobuf:"bytes,12,opt,name=published,proto3" json:"published,omitempty"`
	// Content is the content of the documen, this is essentially what
	// gets rendered on the page when you view a document.
	Content []*Block `protobuf:"bytes,13,rep,name=content,proto3" json:"content,omitempty"`
	// Meta is the metadata for a document, this could be stuff like
	// open graph tags and content profile information.
	Meta []*Block `protobuf:"bytes,14,rep,name=meta,proto3" json:"meta,omitempty"`
	// Links are links to other resources and entities. This could be
	// links to categories and subject for the document, or authors.
	Links []*Block `protobuf:"bytes,15,rep,name=links,proto3" json:"links,omitempty"`
	// Properties are header-like properties for a document. This is
	// mainly used as a bucket for document-level stuff that needs to be
	// preserved when converting to and from other document formats.
	Properties []*Property `protobuf:"bytes,16,rep,name=properties,proto3" json:"properties,omitempty"`
	// Source is the name of the source of the document, usually the
	// name of the application that generated it (or allowed a user to
	// generate it).
	Source string `protobuf:"bytes,17,opt,name=source,proto3" json:"source,omitempty"`
	// Language is the language used in the document as an IETF language
	// tag. F.ex. "en", "en-UK", "es", or "sv-SE".
	Language string `protobuf:"bytes,18,opt,name=language,proto3" json:"language,omitempty"`
	// A free form field detailing the status for the document, for example:
	// "draft" or "withheld".
	Status      string                 `protobuf:"bytes,19,opt,name=status,proto3" json:"status,omitempty"`
	Unpublished *timestamppb.Timestamp `protobuf:"bytes,20,opt,name=unpublished,proto3" json:"unpublished,omitempty"`
	Provider    string                 `protobuf:"bytes,21,opt,name=provider,proto3" json:"provider,omitempty"` // string infoSource = 22;
}

func (x *Document) Reset() {
	*x = Document{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_document_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Document) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Document) ProtoMessage() {}

func (x *Document) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_document_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Document.ProtoReflect.Descriptor instead.
func (*Document) Descriptor() ([]byte, []int) {
	return file_rpc_document_proto_rawDescGZIP(), []int{0}
}

func (x *Document) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Document) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Document) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *Document) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Document) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Document) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

// Deprecated: Do not use.
func (x *Document) GetProducts() []string {
	if x != nil {
		return x.Products
	}
	return nil
}

func (x *Document) GetCreated() *timestamppb.Timestamp {
	if x != nil {
		return x.Created
	}
	return nil
}

func (x *Document) GetModified() *timestamppb.Timestamp {
	if x != nil {
		return x.Modified
	}
	return nil
}

func (x *Document) GetPublished() *timestamppb.Timestamp {
	if x != nil {
		return x.Published
	}
	return nil
}

func (x *Document) GetContent() []*Block {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *Document) GetMeta() []*Block {
	if x != nil {
		return x.Meta
	}
	return nil
}

func (x *Document) GetLinks() []*Block {
	if x != nil {
		return x.Links
	}
	return nil
}

func (x *Document) GetProperties() []*Property {
	if x != nil {
		return x.Properties
	}
	return nil
}

func (x *Document) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *Document) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *Document) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Document) GetUnpublished() *timestamppb.Timestamp {
	if x != nil {
		return x.Unpublished
	}
	return nil
}

func (x *Document) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

// Property is a key-value pair
type Property struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value      string            `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Parameters map[string]string `protobuf:"bytes,3,rep,name=parameters,proto3" json:"parameters,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Property) Reset() {
	*x = Property{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_document_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Property) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Property) ProtoMessage() {}

func (x *Property) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_document_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Property.ProtoReflect.Descriptor instead.
func (*Property) Descriptor() ([]byte, []int) {
	return file_rpc_document_proto_rawDescGZIP(), []int{1}
}

func (x *Property) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Property) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Property) GetParameters() map[string]string {
	if x != nil {
		return x.Parameters
	}
	return nil
}

// Block is the building block for data embedded in documents. It is
// used for both content, links and metadata. Blocks have can be
// nested, but that's nothing to strive for, keep it simple.
type Block struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID is the block ID
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// UUID is used to reference another Document in a block.
	Uuid string `protobuf:"bytes,2,opt,name=uuid,proto3" json:"uuid,omitempty"`
	// URI is used to reference another entity in a document.
	Uri string `protobuf:"bytes,3,opt,name=uri,proto3" json:"uri,omitempty"`
	// URL is a browseable URL for the the block.
	Url string `protobuf:"bytes,4,opt,name=url,proto3" json:"url,omitempty"`
	// Type is the type of the block
	Type string `protobuf:"bytes,5,opt,name=type,proto3" json:"type,omitempty"`
	// Title is the title/headline of the block, typically used in the
	// presentation of the block.
	Title string `protobuf:"bytes,6,opt,name=title,proto3" json:"title,omitempty"`
	// Data contains block data
	Data map[string]string `protobuf:"bytes,7,rep,name=data,proto3" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Relationship describes the relationship to the document/parent
	// entity
	Rel string `protobuf:"bytes,8,opt,name=rel,proto3" json:"rel,omitempty"`
	// Name is a name for the block. An alternative to "rel" when
	// relationship is a term that doesn't fit.
	Name string `protobuf:"bytes,9,opt,name=name,proto3" json:"name,omitempty"`
	// Value is a value for the block. Useful when we want to store a
	// primitive value.
	Value string `protobuf:"bytes,10,opt,name=value,proto3" json:"value,omitempty"`
	// ContentType is used to describe the content type of the
	// block/linked entity if it differs from the type of the block.
	ContentType string `protobuf:"bytes,11,opt,name=contentType,proto3" json:"contentType,omitempty"`
	// Links are used to link to other resources and documents.
	Links []*Block `protobuf:"bytes,13,rep,name=links,proto3" json:"links,omitempty"`
	// Content is used to embed content blocks.
	Content []*Block `protobuf:"bytes,14,rep,name=content,proto3" json:"content,omitempty"`
	// Meta is used to embed metadata
	Meta []*Block `protobuf:"bytes,15,rep,name=meta,proto3" json:"meta,omitempty"`
	// Role is used for
	Role string `protobuf:"bytes,16,opt,name=role,proto3" json:"role,omitempty"`
}

func (x *Block) Reset() {
	*x = Block{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpc_document_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_document_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
	return file_rpc_document_proto_rawDescGZIP(), []int{2}
}

func (x *Block) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Block) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *Block) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *Block) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Block) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Block) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Block) GetData() map[string]string {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Block) GetRel() string {
	if x != nil {
		return x.Rel
	}
	return ""
}

func (x *Block) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Block) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Block) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *Block) GetLinks() []*Block {
	if x != nil {
		return x.Links
	}
	return nil
}

func (x *Block) GetContent() []*Block {
	if x != nil {
		return x.Content
	}
	return nil
}

func (x *Block) GetMeta() []*Block {
	if x != nil {
		return x.Meta
	}
	return nil
}

func (x *Block) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

var File_rpc_document_proto protoreflect.FileDescriptor

var file_rpc_document_proto_rawDesc = []byte{
	0x0a, 0x12, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x1a, 0x1f, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x91, 0x05,
	0x0a, 0x08, 0x44, 0x6f, 0x63, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x75, 0x72, 0x69, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x70, 0x61, 0x74, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68,
	0x12, 0x1e, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x18, 0x09, 0x20, 0x03,
	0x28, 0x09, 0x42, 0x02, 0x18, 0x01, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73,
	0x12, 0x34, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x36, 0x0a, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x12, 0x38,
	0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x64, 0x18, 0x0c, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x65, 0x64, 0x12, 0x27, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6e, 0x61, 0x76, 0x69,
	0x67, 0x61, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x12, 0x21, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x0e, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0d, 0x2e, 0x6e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x04,
	0x6d, 0x65, 0x74, 0x61, 0x12, 0x23, 0x0a, 0x05, 0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x18, 0x0f, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x2e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x52, 0x05, 0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x12, 0x30, 0x0a, 0x0a, 0x70, 0x72, 0x6f,
	0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x18, 0x10, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e,
	0x6e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x52,
	0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x18,
	0x12, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x3c, 0x0a, 0x0b, 0x75, 0x6e, 0x70, 0x75, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x65, 0x64, 0x18, 0x14, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0b, 0x75, 0x6e, 0x70, 0x75, 0x62, 0x6c,
	0x69, 0x73, 0x68, 0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65,
	0x72, 0x18, 0x15, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65,
	0x72, 0x22, 0xb5, 0x01, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x40, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61,
	0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x6e,
	0x61, 0x76, 0x69, 0x67, 0x61, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x2e, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a,
	0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x1a, 0x3d, 0x0a, 0x0f, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xc8, 0x03, 0x0a, 0x05, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x69, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x2b, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x07, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x2e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x72, 0x65, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x20,
	0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x23, 0x0a, 0x05, 0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0d, 0x2e, 0x6e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x05,
	0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x12, 0x27, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x18, 0x0e, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6e, 0x61, 0x76, 0x69, 0x67, 0x61, 0x2e,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x21,
	0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6e,
	0x61, 0x76, 0x69, 0x67, 0x61, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x04, 0x6d, 0x65, 0x74,
	0x61, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x72, 0x6f, 0x6c, 0x65, 0x1a, 0x37, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x4a, 0x04,
	0x08, 0x0c, 0x10, 0x0d, 0x42, 0x2b, 0x5a, 0x29, 0x62, 0x69, 0x74, 0x62, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x69, 0x6e, 0x66, 0x6f, 0x6d, 0x61, 0x6b, 0x65, 0x72, 0x2f,
	0x64, 0x6f, 0x63, 0x2d, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2f, 0x76, 0x32, 0x2f, 0x72, 0x70,
	0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rpc_document_proto_rawDescOnce sync.Once
	file_rpc_document_proto_rawDescData = file_rpc_document_proto_rawDesc
)

func file_rpc_document_proto_rawDescGZIP() []byte {
	file_rpc_document_proto_rawDescOnce.Do(func() {
		file_rpc_document_proto_rawDescData = protoimpl.X.CompressGZIP(file_rpc_document_proto_rawDescData)
	})
	return file_rpc_document_proto_rawDescData
}

var file_rpc_document_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_rpc_document_proto_goTypes = []interface{}{
	(*Document)(nil),              // 0: naviga.Document
	(*Property)(nil),              // 1: naviga.Property
	(*Block)(nil),                 // 2: naviga.Block
	nil,                           // 3: naviga.Property.ParametersEntry
	nil,                           // 4: naviga.Block.DataEntry
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
}
var file_rpc_document_proto_depIdxs = []int32{
	5,  // 0: naviga.Document.created:type_name -> google.protobuf.Timestamp
	5,  // 1: naviga.Document.modified:type_name -> google.protobuf.Timestamp
	5,  // 2: naviga.Document.published:type_name -> google.protobuf.Timestamp
	2,  // 3: naviga.Document.content:type_name -> naviga.Block
	2,  // 4: naviga.Document.meta:type_name -> naviga.Block
	2,  // 5: naviga.Document.links:type_name -> naviga.Block
	1,  // 6: naviga.Document.properties:type_name -> naviga.Property
	5,  // 7: naviga.Document.unpublished:type_name -> google.protobuf.Timestamp
	3,  // 8: naviga.Property.parameters:type_name -> naviga.Property.ParametersEntry
	4,  // 9: naviga.Block.data:type_name -> naviga.Block.DataEntry
	2,  // 10: naviga.Block.links:type_name -> naviga.Block
	2,  // 11: naviga.Block.content:type_name -> naviga.Block
	2,  // 12: naviga.Block.meta:type_name -> naviga.Block
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_rpc_document_proto_init() }
func file_rpc_document_proto_init() {
	if File_rpc_document_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rpc_document_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Document); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rpc_document_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Property); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rpc_document_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Block); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rpc_document_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rpc_document_proto_goTypes,
		DependencyIndexes: file_rpc_document_proto_depIdxs,
		MessageInfos:      file_rpc_document_proto_msgTypes,
	}.Build()
	File_rpc_document_proto = out.File
	file_rpc_document_proto_rawDesc = nil
	file_rpc_document_proto_goTypes = nil
	file_rpc_document_proto_depIdxs = nil
}
