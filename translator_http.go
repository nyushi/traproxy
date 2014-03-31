package traproxy

import (
	"bytes"
	"github.com/nyushi/traproxy/http"
	"sync"
)

type HTTPTranslator struct {
	TranslatorBase

	buf               []byte
	processingRequest *http.RequestHeader
}

func (t *HTTPTranslator) filterRequest(in []byte) []byte {
	t.buf = append(t.buf, in...)
	out := []byte{}
	rb := http.RequestBuffer(t.buf)
	for {
		if t.processingRequest == nil {
			req, err := rb.ReadRequestHeader()
			if err != nil {
				break
			}
			if req == nil {
				break
			}
			t.processingRequest = req
			for _, h := range req.Headers {
				if bytes.Equal(bytes.ToLower(h[0]), []byte("host")) {
					req.SetRequestURI("http://" + string(h[1]) + string(req.ReqLineTokens[1]))
					break
				}
			}
			out = append(out, req.Bytes()...)
		}

		if t.processingRequest != nil {
			body := rb.ReadRequestBody(t.processingRequest)
			if t.processingRequest.IsCompleted() {
				t.processingRequest = nil
			}
			out = append(out, body...)
		}
	}
	return out
}

func (t *HTTPTranslator) Start() error {
	t.buf = []byte{}

	client, proxy, err := t.CheckSockets()
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer t.HandlePanic()

		Pipe(client, proxy, nil)
	}()
	go func() {
		defer wg.Done()
		defer t.HandlePanic()

		f := t.filterRequest
		Pipe(proxy, client, &f)
	}()
	wg.Wait()
	return nil
}
