package traproxy

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
)

type HTTPSTranslator struct {
	TranslatorBase
}

func (t *HTTPSTranslator) isConnectSucceeded(resp []byte) bool {
	lines := bytes.Split(resp, []byte("\r\n"))
	tokens := bytes.Split(lines[0], []byte(" "))
	if bytes.Equal(tokens[1], []byte("200")) {
		return true
	}
	return false
}

func (t *HTTPSTranslator) prepare() error {
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n", t.Dst)
	_, err := t.Proxy.Write([]byte(req))
	if err != nil {
		return errors.New(fmt.Sprintf("failed to write at CONNECT: %s", err.Error()))
	}

	buf := make([]byte, 1024)

	size, err := t.Proxy.Read(buf)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to read at CONNECT: %s", err.Error()))
	}
	ok := t.isConnectSucceeded(buf[:size])
	if !ok {
		return errors.New(fmt.Sprintf("error response at CONNECT request: %s", string(buf[:size])))
	}
	return nil
}

func (t *HTTPSTranslator) Start() error {
	client, proxy, err := t.CheckSockets()
	if err != nil {
		return err
	}

	err = t.prepare()
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

		Pipe(proxy, client, nil)
	}()
	wg.Wait()
	return nil
}
