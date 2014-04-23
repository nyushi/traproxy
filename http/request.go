package http

import (
	"bytes"
	"strconv"
)

// ReadRequestHeader reads header information from bytes
func ReadRequestHeader(rb []byte) ([]byte, *RequestHeader, error) {
	headerEnd := bytes.Index(rb, eoh)
	if headerEnd == -1 {
		return rb, nil, nil
	}
	boundary := headerEnd + len(eoh)
	reqBytes := rb[:boundary]
	rest := rb[boundary:]

	req, err := NewRequestHeader(reqBytes)
	return rest, req, err
}

// ReadRequestBody reads request body from bytes
func ReadRequestBody(rb []byte, req *RequestHeader) ([]byte, []byte) {
	var body []byte
	var rest []byte
	s := req.BodySize - req.BodyRead
	if len(rb) > s {
		body = rb[:s]
		rest = rb[s:]
	} else {
		body = rb
		rest = []byte{}
	}
	req.BodyRead += len(body)
	return rest, body
}

var (
	eol = []byte("\r\n")
	eoh = append(eol, eol...)
)

// RequestHeader represents HTTP Request Header
type RequestHeader struct {
	ReqLineTokens [][]byte
	Headers       [][][]byte
	BodySize      int
	BodyRead      int
}

// NewRequestHeader returns RequestHeader from bytes
func NewRequestHeader(b []byte) (*RequestHeader, error) {
	lines := bytes.Split(b, eol)
	reqline := bytes.Split(lines[0], []byte{' '})

	headers := [][][]byte{}
	bodySize := 0

	headerLines := lines[1:]
	for _, l := range headerLines {
		tokens := bytes.SplitN(l, []byte{':', ' '}, 2)
		if len(tokens) == 2 {
			headers = append(headers, tokens)
		}

		if bytes.Equal(bytes.ToLower(tokens[0]), []byte("content-length")) {
			size, err := strconv.Atoi(string(tokens[1]))
			if err != nil {
				return nil, err
			}
			bodySize = size
		}
	}

	r := &RequestHeader{
		ReqLineTokens: reqline,
		Headers:       headers,
		BodySize:      bodySize,
		BodyRead:      0,
	}
	return r, nil
}

// Bytes returns byte slice of RequestHeader
func (r *RequestHeader) Bytes() []byte {
	lines := [][]byte{}
	lines = append(lines, bytes.Join(r.ReqLineTokens, []byte{' '}))
	for _, h := range r.Headers {
		hline := bytes.Join(h, []byte{':', ' '})
		lines = append(lines, hline)
	}
	out := bytes.Join(lines, eol)
	out = append(out, eoh...)
	return out
}

// SetRequestURI sets uri to RequestHeader
func (r *RequestHeader) SetRequestURI(uri string) {
	r.ReqLineTokens[1] = []byte(uri)
}

// ReqLine returns request line bytes
func (r *RequestHeader) ReqLine() []byte {
	return bytes.Join(r.ReqLineTokens, []byte{' '})
}

// HeadersStr returns header strings
func (r *RequestHeader) HeadersStr() [][]string {
	headerStr := [][]string{}
	for _, header := range r.Headers {
		headerStr = append(
			headerStr,
			[]string{
				string(header[0]),
				string(header[1])})
	}
	return headerStr
}

// IsCompleted returns request status
func (r *RequestHeader) IsCompleted() bool {
	return r.BodySize == r.BodyRead
}
