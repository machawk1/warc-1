package warc

import (
	"bytes"
	"io"
	"time"
)

// A Record consists of a version indicator (eg: WARC/1.0), zero or more headers,
// and possibly a content block.
// Upgrades to specific types of records can be done using type assertions
// and/or the Type method.
type Record struct {
	Version string
	Headers map[string]string
	Content []byte
}

// Return the type of record
func (r *Record) Type() RecordType {
	return recordType(r.Headers[warc_type])
}

// The ID for this record
func (r *Record) Id() string {
	return r.Headers[warc_record_id]
}

// Datestamp of record creation
func (r *Record) Date() time.Time {
	// TODO
	return time.Now()
}

// Length of content block in bytes
func (r *Record) ContentLength() int64 {
	// TODO
	return 0
}

// Write this record to a given writer
func (r *Record) Write(w io.Writer) error {
	if err := writeHeader(w, r); err != nil {
		return err
	}
	return writeBlock(w, r.Content)
}

// Bytes returns the record formatted as a byte slice
func (r *Record) Bytes() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := r.Write(buf)
	return buf.Bytes(), err
}

// A WARC format file is the simple concatenation of one or more WARC
// records. The first record usually describes the records to follow. In
// general, record content is either the direct result of a retrieval
// attempt — web pages, inline images, URL redirection information, DNS
// hostname lookup results, standalone files, etc. — or is synthesized
// material (e.g., metadata, transformed content) that provides additional
// information about archived content.
type Records []Record

// RecordType enumerates different types of WARC Records
type RecordType int

const (
	RecordTypeUnknown RecordType = iota
	// A 'warcinfo' record describes the records that follow it, up through end
	// of file, end of input, or until next 'warcinfo' record. Typically, this
	// appears once and at the beginning of a WARC file. For a web archive, it
	// often contains information about the web crawl which generated the
	// following records.
	// The format of this descriptive record block may vary, though the use of
	// the "application/warc-fields" content-type is recommended. Allowable
	// fields include, but are not limited to, all \[DCMI\] plus the following
	// field definitions. All fields are optional.
	RecordTypeWarcInfo
	// A 'response' record should contain a complete scheme-specific response,
	// including network protocol information where possible. The exact
	// contents of a 'response' record are determined not just by the record
	// type but also by the URI scheme of the record's target-URI, as described
	// below.
	RecordTypeResponse
	// A 'resource' record contains a resource, without full protocol response
	// information. For example: a file directly retrieved from a locally
	// accessible repository or the result of a networked retrieval where the
	// protocol information has been discarded. The exact contents of a
	// 'resource' record are determined not just by the record type but also by
	// the URI scheme of the record's target-URI, as described below.
	// For all 'resource' records, the payload is defined as the record block.
	// A 'resource' record, with a synthesized target-URI, may also be used to
	// archive other artefacts of a harvesting process inside WARC files.
	RecordTypeResource
	// A 'request' record holds the details of a complete scheme-specific
	// request, including network protocol information where possible. The
	// exact contents of a 'request' record are determined not just by the
	// record type but also by the URI scheme of the record's target-URI, as
	// described below.
	RecordTypeRequest
	// A 'metadata' record contains content created in order to further
	// describe, explain, or accompany a harvested resource, in ways not
	// covered by other record types. A 'metadata' record will almost always
	// refer to another record of another type, with that other record holding
	// original harvested or transformed content. (However, it is allowable for
	// a 'metadata' record to refer to any record type, including other
	// 'metadata' records.) Any number of metadata records may reference one
	// specific other record.
	// The format of the metadata record block may vary. The
	// "application/warc-fields" format, defined earlier, may be used.
	// Allowable fields include all \[DCMI\] plus the following field
	// definitions. All fields are optional.
	RecordTypeMetadata
	// A 'revisit' record describes the revisitation of content already
	// archived, and might include only an abbreviated content body which has
	// to be interpreted relative to a previous record. Most typically, a
	// 'revisit' record is used instead of a 'response' or 'resource' record to
	// indicate that the content visited was either a complete or substantial
	// duplicate of material previously archived.
	// Using a 'revisit' record instead of another type is optional, for when
	// benefits of reduced storage size or improved cross-referencing of
	// material are desired.
	RecordTypeRevisit
	// A 'conversion' record shall contain an alternative version of another
	// record's content that was created as the result of an archival process.
	// Typically, this is used to hold content transformations that maintain
	// viability of content after widely available rendering tools for the
	// originally stored format disappear. As needed, the original content may
	// be migrated (transformed) to a more viable format in order to keep the
	// information usable with current tools while minimizing loss of
	// information (intellectual content, look and feel, etc). Any number of
	// 'conversion' records may be created that reference a specific source
	// record, which may itself contain transformed content. Each
	// transformation should result in a freestanding, complete record, with no
	// dependency on survival of the original record.
	// Metadata records may be used to further describe transformation records.
	// Wherever practical, a 'conversion' record should contain a
	// 'WARC-Refers-To' field to identify the prior material converted.
	RecordTypeConversion
	// Record blocks from 'continuation' records must be appended to
	// corresponding prior record block(s) (e.g., from other WARC files) to
	// create the logically complete full-sized original record. That is,
	// 'continuation' records are used when a record that would otherwise cause
	// a WARC file size to exceed a desired limit is broken into segments. A
	// continuation record shall contain the named fields
	// 'WARC-Segment-Origin-ID' and 'WARC-Segment-Number', and the last
	// 'continuation' record of a series shall contain a
	// 'WARC-Segment-Total-Length' field. The full details of WARC record
	// segmentation are described in the below section Record Segmentation. See
	// also annex C.8 below for an example of a ‘continuation’ record.
	RecordTypeContinuation
)

// RecordType satisfies the stringer interface
func (r RecordType) String() string {
	switch r {
	case RecordTypeWarcInfo:
		return "warcinfo"
	case RecordTypeResponse:
		return "response"
	case RecordTypeResource:
		return "resource"
	case RecordTypeRequest:
		return "request"
	case RecordTypeMetadata:
		return "metadata"
	case RecordTypeRevisit:
		return "revisit"
	case RecordTypeConversion:
		return "conversion"
	case RecordTypeContinuation:
		return "continuation"
	default:
		return ""
	}
}

func recordType(s string) RecordType {
	switch s {
	case RecordTypeWarcInfo.String():
		return RecordTypeWarcInfo
	case RecordTypeResponse.String():
		return RecordTypeResponse
	case RecordTypeResource.String():
		return RecordTypeResource
	case RecordTypeRequest.String():
		return RecordTypeRequest
	case RecordTypeMetadata.String():
		return RecordTypeMetadata
	case RecordTypeRevisit.String():
		return RecordTypeRevisit
	case RecordTypeConversion.String():
		return RecordTypeConversion
	case RecordTypeContinuation.String():
		return RecordTypeContinuation
	default:
		return RecordTypeUnknown
	}
}
