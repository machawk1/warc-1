package warc

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

const testRecordId = "<urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>"

func init() {
	for _, t := range []*[]byte{
		&WARCINFO_RECORD,
		&RESPONSE_RECORD,
		&RESPONSE_RECORD_2,
		&REQUEST_RECORD,
		&REQUEST_RECORD_2,
		&REVISIT_RECORD_1,
		&REVISIT_RECORD_2,
		&RESOURCE_RECORD,
		&METADATA_RECORD,
		&DNS_RESPONSE_RECORD,
		&DNS_RESOURCE_RECORD,
	} {
		// need to replace '\r' from raw string literals with actual
		// carriage return character
		*t = bytes.Replace(*t, []byte{'\\', 'r'}, []byte{0x0d}, -1)
	}
}

func TestWarcWrite(t *testing.T) {
	f, err := os.Open("testdata/test.warc")
	// data, err := ioutil.ReadFile("testdata/test.warc")
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer f.Close()

	rdr, err := NewReader(f)
	if err != nil {
		t.Error(err)
		return
	}

	records, err := rdr.ReadAll()
	if err != nil {
		t.Error(err)
		return
	}

	out, err := os.Create("testdata/out.warc")
	if err != nil {
		t.Error(err)
		return
	}
	defer out.Close()
	if err := WriteRecords(out, records); err != nil {
		t.Error(err)
		return
	}
}

func TestWarcinfoRecord(t *testing.T) {
	rec := &Record{
		Format: RecordFormatWarc,
		Type:   RecordTypeWarcInfo,
		Headers: map[string]string{
			warcRecordId:  testRecordId,
			warcType:      RecordTypeWarcInfo.String(),
			warcFilename:  "testfile.warc.gz",
			warcDate:      "2000-01-01T00:00:00Z",
			contentType:   "application/warc-fields",
			contentLength: "86",
		},
		Content: bytes.NewBuffer([]byte("software: recorder test\r\n" +
			"format: WARC File Format 1.0\r\n" +
			"json-metadata: {\"foo\": \"bar\"}\r\n")),
	}

	if err := testWriteRecord(rec, WARCINFO_RECORD); err != nil {
		t.Error(err)
	}
}

func TestRequestRecord(t *testing.T) {
	rec := &Record{
		Format: RecordFormatWarc,
		Type:   RecordTypeRequest,
		Headers: map[string]string{
			warcType:          RecordTypeRequest.String(),
			warcRecordId:      testRecordId,
			warcTargetUri:     "http://example.com/",
			warcDate:          "2000-01-01T00:00:00Z",
			warcPayloadDigest: "sha1:3I42H3S6NNFQ2MSVX7XZKYAYSCX5QBYJ",
			warcBlockDigest:   "sha1:ONEHF6PTXPTTHE3333XHTD2X45TZ3DTO",
			contentType:       "application/http; msgtype=request",
			contentLength:     "54",
		},
		Content: bytes.NewBuffer([]byte("GET / HTTP/1.0\r\n" +
			"User-Agent: foo\r\n" +
			"Host: example.com\r\n" +
			"\r\n")),
	}

	if err := testWriteRecord(rec, REQUEST_RECORD); err != nil {
		t.Error(err)
	}
}

func TestResponseRecord(t *testing.T) {
	rec := &Record{
		Format: RecordFormatWarc,
		Type:   RecordTypeResponse,
		Headers: map[string]string{
			contentLength:     "97",
			contentType:       "application/http; msgtype=response",
			warcBlockDigest:   "sha1:OS3OKGCWQIJOAOC3PKXQOQFD52NECQ74",
			warcDate:          "2000-01-01T00:00:00Z",
			warcPayloadDigest: "sha1:B6QJ6BNJ3R4B23XXMRKZKHLPGJY2VE4O",
			warcRecordId:      "<urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>",
			warcTargetUri:     "http://example.com/",
			warcType:          RecordTypeResponse.String(),
		},
		Content: bytes.NewBuffer([]byte("HTTP/1.0 200 OK\r\n" +
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
			"Custom-Header: somevalue\r\n" +
			"\r\n" +
			"some\n" +
			"text")),
	}

	if err := testWriteRecord(rec, RESPONSE_RECORD); err != nil {
		t.Error(err)
	}
}

func testWriteRecord(r *Record, expect []byte) error {
	if r.ContentLength() != r.Content.Len() {
		return fmt.Errorf("Record Content-Length mistmatch: %d != %d", r.ContentLength(), r.Content.Len())
	}

	buf := &bytes.Buffer{}
	if err := r.Write(buf); err != nil {
		return fmt.Errorf("error writing record: %s", err.Error())
	}

	if len(buf.Bytes()) != len(expect) {
		dmp := dmp.New()
		diffs := dmp.DiffMain(buf.String(), string(expect), true)
		fmt.Println("error diff output:")
		fmt.Println(dmp.DiffPrettyText(diffs))

		for i, b := range buf.Bytes() {
			if i >= len(expect) || b != expect[i] {
				return fmt.Errorf("byte length mismatch. expected: %d, got: %d. first error at index %d: '%#v'", len(expect), len(buf.Bytes()), i, b)
			}
		}

		return fmt.Errorf("byte length mismatch. expected: %d, got: %d, ", len(expect), len(buf.Bytes()))
	}

	if !bytes.Equal(buf.Bytes(), expect) {
		return fmt.Errorf("byte mismatch: %s != %s", buf.String(), string(expect))
	}

	return nil
}

// func testRequestResponseConcur(t *testing.T) {
// }

// func testReadFromStreamNoContentLength(t *testing.T) {

// }

func validateResponse(r *Record) error {
	return nil
}

var WARCINFO_RECORD = []byte(`WARC/1.0\r
Content-Length: 86\r
Content-Type: application/warc-fields\r
Warc-Date: 2000-01-01T00:00:00Z\r
Warc-Filename: testfile.warc.gz\r
Warc-Record-Id: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
Warc-Type: warcinfo\r
\r
software: recorder test\r
format: WARC File Format 1.0\r
json-metadata: {"foo": "bar"}\r
\r
\r
`)

var RESPONSE_RECORD = []byte(`WARC/1.0\r
Content-Length: 97\r
Content-Type: application/http; msgtype=response\r
Warc-Block-Digest: sha1:OS3OKGCWQIJOAOC3PKXQOQFD52NECQ74\r
Warc-Date: 2000-01-01T00:00:00Z\r
Warc-Payload-Digest: sha1:B6QJ6BNJ3R4B23XXMRKZKHLPGJY2VE4O\r
Warc-Record-Id: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
Warc-Target-Uri: http://example.com/\r
Warc-Type: response\r
\r
HTTP/1.0 200 OK\r
Content-Type: text/plain; charset="UTF-8"\r
Custom-Header: somevalue\r
\r
some
text\r
\r
`)

var RESPONSE_RECORD_2 = []byte(`
WARC/1.0\r
WARC-Type: response\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: http://example.com/\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:B6QJ6BNJ3R4B23XXMRKZKHLPGJY2VE4O\r
WARC-Block-Digest: sha1:U6KNJY5MVNU3IMKED7FSO2JKW6MZ3QUX\r
Content-Type: application/http; msgtype=response\r
Content-Length: 145\r
\r
HTTP/1.0 200 OK\r
Content-Type: text/plain; charset="UTF-8"\r
Content-Length: 9\r
Custom-Header: somevalue\r
Content-Encoding: x-unknown\r
\r
some
text\r
\r
`)

var REQUEST_RECORD = []byte(`WARC/1.0\r
Content-Length: 54\r
Content-Type: application/http; msgtype=request\r
Warc-Block-Digest: sha1:ONEHF6PTXPTTHE3333XHTD2X45TZ3DTO\r
Warc-Date: 2000-01-01T00:00:00Z\r
Warc-Payload-Digest: sha1:3I42H3S6NNFQ2MSVX7XZKYAYSCX5QBYJ\r
Warc-Record-Id: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
Warc-Target-Uri: http://example.com/\r
Warc-Type: request\r
\r
GET / HTTP/1.0\r
User-Agent: foo\r
Host: example.com\r
\r
\r
\r
`)

var REQUEST_RECORD_2 = []byte(`
WARC/1.0\r
WARC-Type: request\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: http://example.com/\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:R5VZAKIE53UW5VGK43QJIFYS333QM5ZA\r
WARC-Block-Digest: sha1:L7SVBUPPQ6RH3ANJD42G5JL7RHRVZ5DV\r
Content-Type: application/http; msgtype=request\r
Content-Length: 92\r
\r
POST /path HTTP/1.0\r
Content-Type: application/json\r
Content-Length: 17\r
\r
{"some": "value"}\r
\r
`)

var REVISIT_RECORD_1 = []byte(`
WARC/1.0\r
WARC-Type: revisit\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: http://example.com/\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Profile: http://netpreserve.org/warc/1.0/revisit/identical-payload-digest\r
WARC-Refers-To-Target-URI: http://example.com/foo\r
WARC-Refers-To-Date: 1999-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:B6QJ6BNJ3R4B23XXMRKZKHLPGJY2VE4O\r
WARC-Block-Digest: sha1:3I42H3S6NNFQ2MSVX7XZKYAYSCX5QBYJ\r
Content-Type: application/http; msgtype=response\r
Content-Length: 0\r
\r
\r
\r
`)

var REVISIT_RECORD_2 = []byte(`
WARC/1.0\r
WARC-Type: revisit\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: http://example.com/\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Profile: http://netpreserve.org/warc/1.0/revisit/identical-payload-digest\r
WARC-Refers-To-Target-URI: http://example.com/foo\r
WARC-Refers-To-Date: 1999-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:B6QJ6BNJ3R4B23XXMRKZKHLPGJY2VE4O\r
WARC-Block-Digest: sha1:A6J5UTI2QHHCZFCFNHQHCDD3JJFKP53V\r
Content-Type: application/http; msgtype=response\r
Content-Length: 88\r
\r
HTTP/1.0 200 OK\r
Content-Type: text/plain; charset="UTF-8"\r
Custom-Header: somevalue\r
\r
\r
\r
`)

var RESOURCE_RECORD = []byte(`
WARC/1.0\r
WARC-Type: resource\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: ftp://example.com/\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:B6QJ6BNJ3R4B23XXMRKZKHLPGJY2VE4O\r
WARC-Block-Digest: sha1:B6QJ6BNJ3R4B23XXMRKZKHLPGJY2VE4O\r
Content-Type: text/plain\r
Content-Length: 9\r
\r
some
text\r
\r
`)

var METADATA_RECORD = []byte(`
WARC/1.0\r
WARC-Type: metadata\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: http://example.com/\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:ZOLBLKAQVZE5DXH56XE6EH6AI6ZUGDPT\r
WARC-Block-Digest: sha1:ZOLBLKAQVZE5DXH56XE6EH6AI6ZUGDPT\r
Content-Type: application/json\r
Content-Length: 67\r
\r
{"metadata": {"nested": "obj", "list": [1, 2, 3], "length": "123"}}\r
\r
`)

var DNS_RESPONSE_RECORD = []byte(`
WARC/1.0\r
WARC-Type: response\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: dns:google.com\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:2AAVJYKKIWK5CF6EWE7PH63EMNLO44TH\r
WARC-Block-Digest: sha1:2AAVJYKKIWK5CF6EWE7PH63EMNLO44TH\r
Content-Type: application/http; msgtype=response\r
Content-Length: 147\r
\r
20170509000739
google.com.     185 IN  A   209.148.113.239
google.com.     185 IN  A   209.148.113.238
google.com.     185 IN  A   209.148.113.250
\r\r
`)

var DNS_RESOURCE_RECORD = []byte(`
WARC/1.0\r
WARC-Type: resource\r
WARC-Record-ID: <urn:uuid:12345678-feb0-11e6-8f83-68a86d1772ce>\r
WARC-Target-URI: dns:google.com\r
WARC-Date: 2000-01-01T00:00:00Z\r
WARC-Payload-Digest: sha1:2AAVJYKKIWK5CF6EWE7PH63EMNLO44TH\r
WARC-Block-Digest: sha1:2AAVJYKKIWK5CF6EWE7PH63EMNLO44TH\r
Content-Type: application/warc-record\r
Content-Length: 147\r
\r
20170509000739
google.com.     185 IN  A   209.148.113.239
google.com.     185 IN  A   209.148.113.238
google.com.     185 IN  A   209.148.113.250
\r\r
`)
