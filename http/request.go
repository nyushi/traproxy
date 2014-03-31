package http

import (
	"bytes"
	"strconv"
)

type RequestBuffer []byte

var (
	eol = []byte("\r\n")
	eoh = append(eol, eol...)
)

func (rb *RequestBuffer) ReadRequestHeader() (*RequestHeader, error) {
	headerEnd := bytes.Index(*rb, eoh)
	if headerEnd == -1 {
		return nil, nil
	}
	boundary := headerEnd + len(eoh)
	reqBytes := (*rb)[:boundary]
	rest := (*rb)[boundary:]
	*rb = rest

	req, err := NewRequestHeader(reqBytes)
	return req, err
}

func (rb *RequestBuffer) ReadRequestBody(req *RequestHeader) []byte {
	var body []byte
	s := req.BodySize - req.BodyRead
	if len(*rb) > s {
		body = (*rb)[:s]
		*rb = (*rb)[s:]
	} else {
		body = *rb
		*rb = []byte{}
	}
	req.BodyRead += len(body)
	return body
}

type RequestHeader struct {
	ReqLineTokens [][]byte
	Headers       [][][]byte
	BodySize      int
	BodyRead      int
}

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

func (r *RequestHeader) SetRequestURI(uri string) {
	r.ReqLineTokens[1] = []byte(uri)
}

func (r *RequestHeader) ReqLine() []byte {
	return bytes.Join(r.ReqLineTokens, []byte{' '})
}

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

func (r *RequestHeader) IsCompleted() bool {
	return r.BodySize == r.BodyRead
}
