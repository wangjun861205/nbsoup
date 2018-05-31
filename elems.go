package nbsoup

import (
	"bytes"
	"io"
)

type elemType int

const (
	htmlTag elemType = iota
	htmlText
)

var voidTags = map[string]bool{
	"area":     true,
	"base":     true,
	"basefont": true,
	"bgsound":  true,
	"br":       true,
	"col":      true,
	"command":  true,
	"embed":    true,
	"frame":    true,
	"hr":       true,
	"image":    true,
	"img":      true,
	"input":    true,
	"isindex":  true,
	"keygen":   true,
	"link":     true,
	"menuitem": true,
	"meta":     true,
	"nextid":   true,
	"param":    true,
	"source":   true,
	"track":    true,
	"wbr":      true,
	"META":     true,
}

type element interface {
	getName() []byte
	getAttrList() [][]byte
}

type elemProcessor struct {
	htmlProcessor *htmlProcessor
	elemChan      chan element
	errChan       chan error
	stopChan      chan struct{}
}

func newElemProcessor(hp *htmlProcessor) *elemProcessor {
	return &elemProcessor{
		htmlProcessor: hp,
		elemChan:      make(chan element),
		errChan:       make(chan error),
		stopChan:      make(chan struct{}),
	}
}

func (ep *elemProcessor) process() {
	for {
		select {
		case <-ep.stopChan:
			close(ep.htmlProcessor.stopChan)
			continue
		case err := <-ep.htmlProcessor.errChan:
			ep.errChan <- err
			close(ep.elemChan)
			return
		case b, ok := <-ep.htmlProcessor.byteChan:
			if !ok {
				close(ep.elemChan)
				return
			}
			if !isTag(b) {
				ep.elemChan <- content(b)
				continue
			}
			switch {
			case isCommentTag(b):
				ep.elemChan <- ep.parseCommentTag(b)
			case isVoidTag(b):
				ep.elemChan <- ep.parseVoidTag(b)
			case isEndTag(b):
				ep.elemChan <- ep.parseEndTag(b)
			default:
				ep.elemChan <- ep.parseStartTag(b)
			}
		}
	}
}

func (ep *elemProcessor) parseCommentTag(b []byte) *commentTag {
	return &commentTag{bytes.Trim(b, "<>! ")}
}

func (ep *elemProcessor) parseVoidTag(b []byte) *voidTag {
	var attrList [][]byte
	b = bytes.Trim(b, "<>/ ")
	l := bytes.SplitN(b, []byte(" "), 2)
	if len(l) > 1 {
		attrList = ep.parseAttr(l[1])
	}
	return &voidTag{l[0], attrList}
}

func (ep *elemProcessor) parseStartTag(b []byte) *startTag {
	var attrList [][]byte
	b = bytes.Trim(b, "<>")
	l := bytes.SplitN(b, []byte(" "), 2)
	if len(l) > 1 {
		attrList = ep.parseAttr(l[1])
	}
	return &startTag{l[0], attrList}
}

func (ep *elemProcessor) parseEndTag(b []byte) *endTag {
	return &endTag{bytes.Trim(b, "<>/ ")}
}

func (ep *elemProcessor) parseAttr(b []byte) [][]byte {
	b = bytes.Trim(b, " ")
	attrList := make([][]byte, 0, 16)
	reader := bytes.NewReader(b)
	bAttr := make([]byte, 0, 64)
	var inQuote bool
	for {
		c, err := reader.ReadByte()
		if err != nil {
			if len(bAttr) > 0 {
				attrList = append(attrList, bAttr)
			}
			return attrList
		}
		switch c {
		case '"':
			inQuote = !inQuote
			bAttr = append(bAttr, c)
		case ' ':
			if inQuote {
				bAttr = append(bAttr, c)
				continue
			}
			attr := make([]byte, len(bAttr))
			copy(attr, bAttr)
			attrList = append(attrList, attr)
			bAttr = bAttr[:0]
		default:
			bAttr = append(bAttr, c)
		}
	}
}

func splitElems(html []byte) ([]string, error) {
	elems := make([]string, 0, 1024)
	reader := bytes.NewReader(html)
	tag := make([]byte, 0, 512)
	content := make([]byte, 0, 1<<16)
	var inTag bool
	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				return elems, nil
			} else {
				return nil, err
			}
		}
		switch {
		case ignoreChars[b]:
			continue
		case b == '<':
			if ct := bytes.Trim(content, " \t"); len(ct) > 0 {
				elems = append(elems, string(ct))
				content = content[:0]
			}
			inTag = true
			tag = append(tag, b)
		case b == '>':
			inTag = false
			tag = append(tag, b)
			elems = append(elems, string(tag))
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

func isTag(elem []byte) bool {
	if elem[0] == '<' && elem[len(elem)-1] == '>' {
		return true
	}
	return false
}

func isCommentTag(elem []byte) bool {
	if elem[1] == '!' {
		return true
	}
	return false
}

func isEndTag(elem []byte) bool {
	if elem[1] == '/' {
		return true
	}
	return false
}

func isVoidTag(elem []byte) bool {
	name := bytes.Split(bytes.Trim(elem, "<>/ "), []byte(" "))[0]
	return voidTags[string(name)]
}
