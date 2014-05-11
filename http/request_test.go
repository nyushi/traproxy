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
		return fmt.Errorf("expected=%v, got=%v", expected, got)
	}
	if expected.BodyRead != got.BodyRead {
		return fmt.Errorf("var BodyRead not match, expected=%v, got=%v",
			expected.BodyRead, got.BodyRead)
	}
	if expected.BodySize != got.BodySize {
		return fmt.Errorf("var BodySize not match, expected=%v, got=%v",
			expected.BodySize, got.BodySize)
	}

	expectedReqLine := bytes.Join(expected.ReqLineTokens, []byte{})
	gotReqLine := bytes.Join(got.ReqLineTokens, []byte{})
	if !bytes.Equal(expectedReqLine, gotReqLine) {
		return fmt.Errorf("var ReqLineTokens not match, expected=%v, got=%v",
			string(expectedReqLine), string(gotReqLine))
	}

	if len(expected.Headers) != len(got.Headers) {
		return fmt.Errorf("var Headers not match, expected=%v, got=%v",
			expected.HeadersStr(), got.HeadersStr(),
		)
	}
	return nil
}

func TestReadRequestHeader(t *testing.T) {
	b := []byte{}
	rest, req, err := ReadRequestHeader(b)
	if err != nil {
		t.Error(err)
	}
	if req != nil {
		t.Error("request is not nil")
	}
	if len(b) != 0 {
		t.Error("rest size is not 0")
	}

	b = rest
	b = append(b, []byte("GET / HTTP/1.1\r\n\r\nrest")...)
	rest, req, err = ReadRequestHeader(b)
	if err != nil {
		t.Error(err)
	}
	if req == nil {
		t.Error("request is nil")
	}
	err = checkRequest(req, &RequestHeader{
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
	if !bytes.Equal(rest, []byte("rest")) {
		t.Errorf("invalid rest: expected='rest', got='%s'", string(rest))
	}
}

func TestReadRequestBody(t *testing.T) {
	b := []byte{'1', '2', '3', '4'}
	header := &RequestHeader{BodySize: 1, BodyRead: 0}
	expected := "1"
	rest, got := ReadRequestBody(b, header)
	if string(got) != expected {
		t.Errorf("error at ReadRequestBody: expected=%s, got=%s", expected, string(got))
	}

	expected = "234"
	got = rest
	if string(got) != expected {
		t.Errorf("error at ReadRequestBody: expected=%s, got=%s", expected, string(got))
	}

	b = rest
	header.BodySize = 1000
	header.BodyRead = 0
	expected = "234"
	rest, got = ReadRequestBody(b, header)
	if string(got) != expected {
		t.Errorf("error at ReadRequestBody: expected=%s, got=%s", expected, string(got))
	}

	expected = ""
	got = rest
	if string(got) != expected {
		t.Errorf("error at ReadRequestBody: expected=%s, got=%s", expected, string(got))
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

func TestRequestHeader(t *testing.T) {
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

func TestRequestHeaderReqLine(t *testing.T) {
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

func TestRequestHeaderBytes(t *testing.T) {
	in := "GET / HTTP/1.1\r\n" +
		"A: 1\r\n" +
		"B: 2\r\n" +
		"\r\n"
	r, err := NewRequestHeader([]byte(string(in)))
	if err != nil {
		t.Error(err)
	}

	got := r.Bytes()
	expected := "GET / HTTP/1.1\r\nA: 1\r\nB: 2\r\n\r\n"
	if string(got) != expected {
		t.Errorf("error at Bytes: expected=%s, got=%s",
			string(expected),
			string(got),
		)

	}
}

func TestRequestHeaderIsCompleted(t *testing.T) {
	in := "GET / HTTP/1.1\r\n" +
		"A: 1\r\n" +
		"B: 2\r\n" +
		"\r\n"
	r, err := NewRequestHeader([]byte(string(in)))
	if err != nil {
		t.Error(err)
	}
	if !r.IsCompleted() {
		t.Error("IsCompleted error: expected=true, got=false")
	}

	r.BodySize = 1
	if r.IsCompleted() {
		t.Error("IsCompleted error: expected=false, got=true")
	}

	r.BodyRead = 1
	if !r.IsCompleted() {
		t.Error("IsCompleted error: expected=true, got=false")
	}
}

func TestRequestHeaderSetRequestURI(t *testing.T) {
	in := "GET / HTTP/1.1\r\n" +
		"A: 1\r\n" +
		"B: 2\r\n" +
		"\r\n"
	r, err := NewRequestHeader([]byte(string(in)))
	if err != nil {
		t.Error(err)
	}
	r.SetRequestURI("/test")
	if !bytes.Equal(r.ReqLineTokens[1], []byte("/test")) {
		t.Errorf("error at SetRequestURI: expected=/test, got=%s", string(r.ReqLineTokens[1]))
	}
}
