package nbsoup

import (
	"bytes"
	"errors"
	"io"
	"regexp"
)

var ErrHTMLProcessorStop = errors.New("htmlProcessor has been stopped")

var ignoreChars = map[byte]bool{
	'\r': true,
	'\n': true,
}

var replaceRe = regexp.MustCompile(`(\r+|\n+| +|\t+)`)

type htmlProcessor struct {
	byteChan chan []byte
	errChan  chan error
	stopChan chan struct{}
}

func newHTMLProcessor() *htmlProcessor {
	return &htmlProcessor{
		byteChan: make(chan []byte),
		errChan:  make(chan error),
		stopChan: make(chan struct{}),
	}
}

func (hp *htmlProcessor) process(html []byte) {
	html = replaceRe.ReplaceAll(html, []byte(" "))
	html = replaceRe.ReplaceAll(html, []byte(" "))
	reader := bytes.NewReader(html)
	tag := make([]byte, 0, 512)
	content := make([]byte, 0, 1<<16)
	var inTag bool
	for {
		select {
		case <-hp.stopChan:
			hp.errChan <- ErrHTMLProcessorStop
			return
		default:
			b, err := reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					close(hp.byteChan)
					return
				} else {
					hp.errChan <- err
					close(hp.byteChan)
					return
				}
			}
			switch {
			case b == '<':
				if ct := bytes.Trim(content, " \t"); len(ct) > 0 {
					hp.byteChan <- ct
				}
				content = content[:0]
				inTag = true
				tag = append(tag, b)
			case b == '>':
				inTag = false
				tag = append(tag, b)
				bs := make([]byte, len(tag))
				copy(bs, tag)
				hp.byteChan <- bs
				tag = tag[:0]
			default:
				if inTag {
					tag = append(tag, b)
				} else {
					content = append(content, b)
				}
			}
		}
	}
}
