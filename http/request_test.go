package http

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func checkRequest(expected, got *RequestHeader) error {
	if expected == got {
		return nil
	}
	if expected == nil || got == nil {
		return errors.New(fmt.Sprintf("expected=%v, got=%v", expected, got))
	}
	if expected.BodyRead != got.BodyRead {
		return errors.New(fmt.Sprintf("BodyRead not match, expected=%v, got=%v",
			expected.BodyRead, got.BodyRead))
	}
	if expected.BodySize != got.BodySize {
		return errors.New(fmt.Sprintf("BodySize not match, expected=%v, got=%v",
			expected.BodySize, got.BodySize))
	}

	expectedReqLine := bytes.Join(expected.ReqLineTokens, []byte{})
	gotReqLine := bytes.Join(got.ReqLineTokens, []byte{})
	if !bytes.Equal(expectedReqLine, gotReqLine) {
		return errors.New(fmt.Sprintf("ReqLineTokens not match, expected=%v, got=%v",
			string(expectedReqLine), string(gotReqLine)))
	}

	if len(expected.Headers) != len(got.Headers) {
		return errors.New(fmt.Sprintf("Headers not match, expected=%v, got=%v",
			expected.HeadersStr(), got.HeadersStr(),
		))
	}
	return nil
}

func TestRequestBuffer(t *testing.T) {
	testBytes := []byte{'1', '2', '3'}
	rb := RequestBuffer{}
	if len(rb) != 0 {
		t.Errorf("size error: expected=0, got=%d", len(rb))
	}
	rb = append(rb, testBytes...)
	if len([]byte(rb)) != 3 {
		t.Errorf("size error: expected=3, got=%d", len(rb))
	}
	if !bytes.Equal(rb, testBytes) {
		t.Error("data error: expected=%v, got=%v", testBytes, rb)
	}
}

func TestRequestBufferReadRequest(t *testing.T) {
	b := RequestBuffer([]byte{})
	r, err := b.ReadRequestHeader()
	if err != nil {
		t.Error(err)
	}
	if r != nil {
		t.Error("request is not nil")
	}
	if len(b) != 0 {
		t.Error("rest size is not 0")
	}

	b = append(b, []byte("GET / HTTP/1.1\r\n\r\nrest")...)
	r, err = b.ReadRequestHeader()
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Error("request is nil")
	}
	err = checkRequest(r, &RequestHeader{
		ReqLineTokens: [][]byte{
			[]byte("GET"),
			[]byte("/"),
			[]byte("HTTP/1.1"),
		},
		Headers:  [][][]byte{},
		BodySize: 0,
		BodyRead: 0,
	})
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(b, []byte("rest")) {
		t.Errorf("invalid rest: expected='rest', got='%s'", string(b))
	}
}

var newRequestTests = []struct {
	in  string
	out *RequestHeader
	err error
}{
	{
		"",
		&RequestHeader{},
		nil,
	},
	{
		"GET / HTTP/1.1\r\n" +
			"Head1: 1\r\n" +
			"Head2: 2\r\n" +
			"\r\n",
		&RequestHeader{
			ReqLineTokens: [][]byte{
				[]byte("GET"),
				[]byte("/"),
				[]byte("HTTP/1.1"),
			},
			Headers: [][][]byte{
				[][]byte{
					[]byte(string("Head1")),
					[]byte(string("1")),
				},
				[][]byte{
					[]byte(string("Head2")),
					[]byte(string("2")),
				},
			},
			BodySize: 0,
			BodyRead: 0,
		},
		nil,
	},
	{
		"GET / HTTP/1.1\r\n" +
			"Content-Length: 2\r\n" +
			"\r\n",
		&RequestHeader{
			ReqLineTokens: [][]byte{
				[]byte("GET"),
				[]byte("/"),
				[]byte("HTTP/1.1"),
			},
			Headers: [][][]byte{
				[][]byte{
					[]byte(string("Content-Length")),
					[]byte(string("2")),
				},
			},
			BodySize: 2,
			BodyRead: 0,
		},
		nil,
	},
	{
		"GET / HTTP/1.1\r\n" +
			"Content-Length: XXX\r\n" +
			"\r\n",
		nil,
		errors.New("strconv.ParseInt: parsing \"XXX\": invalid syntax"),
	},
}

func TestRequest(t *testing.T) {
	for _, v := range newRequestTests {
		r, err := NewRequestHeader([]byte(v.in))
		if err != nil {
			if err.Error() != v.err.Error() {
				t.Errorf("'%s' Request error not match: expected='%s', got='%s'",
					v.in, err.Error(), v.err.Error(),
				)
			}
		} else {
			err = checkRequest(v.out, r)
			if err != nil {
				t.Errorf("'%s' Request not match: %s", v.in, err.Error())
			}
		}
	}
}

func TestRequestReqLine(t *testing.T) {
	reqline := "GET / HTTP/1.0"
	r, err := NewRequestHeader([]byte(string(reqline + "\r\n\r\n")))
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(r.ReqLine(), []byte(reqline)) {
		t.Errorf("ReqLine not match: expected=%v, got=%v",
			string(r.ReqLine()), reqline)
	}

	reqline = "GET http://example.com/ HTTP/1.0"
	r.ReqLineTokens[1] = []byte("http://example.com/")
	if !bytes.Equal(r.ReqLine(), []byte(reqline)) {
		t.Errorf("ReqLine not match: expected=%v, got=%v",
			string(r.ReqLine()), reqline)
	}
}

func TestRequestHeaderStr(t *testing.T) {
	in := "GET / HTTP/1.1\r\n" +
		"A: 1\r\n" +
		"B: 2\r\n" +
		"\r\n"
	r, err := NewRequestHeader([]byte(string(in)))
	if err != nil {
		t.Error(err)
	}

	headers := r.HeadersStr()

	if len(headers) != 2 ||
		headers[0][0] != "A" || headers[0][1] != "1" ||
		headers[1][0] != "B" || headers[1][1] != "2" {

		t.Errorf("'%v' HeadersStr not match: got=%s",
			in,
			headers,
		)
	}
}
