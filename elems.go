package nbsoup

import (
	"bytes"
	"fmt"
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
	fmt.Stringer
}

type elemProcessor struct {
	htmlProcessor *htmlProcessor
	elemChan      chan element
	errChan       chan error
	stopChan      chan struct{}
}

// type elemProcessor struct {
// 	htmlProcessor *htmlProcessor
// 	elemChan      *bufElemChan
// 	errChan       chan error
// 	stopChan      chan struct{}
// }

func newElemProcessor(hp *htmlProcessor) *elemProcessor {
	return &elemProcessor{
		htmlProcessor: hp,
		elemChan:      make(chan element),
		errChan:       make(chan error),
		stopChan:      make(chan struct{}),
	}
}

// func newElemProcessor(hp *htmlProcessor) *elemProcessor {
// 	return &elemProcessor{
// 		htmlProcessor: hp,
// 		elemChan:      newBufElemChan(),
// 		errChan:       make(chan error),
// 		stopChan:      make(chan struct{}),
// 	}
// }

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
				bc := make([]byte, len(b))
				copy(bc, b)
				ep.elemChan <- content(bc)
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

// func (ep *elemProcessor) process() {
// 	for {
// 		select {
// 		case <-ep.stopChan:
// 			close(ep.htmlProcessor.stopChan)
// 			continue
// 		case err := <-ep.htmlProcessor.errChan:
// 			ep.errChan <- err
// 			ep.elemChan.close()
// 			return
// 		case b, ok := <-ep.htmlProcessor.byteChan:
// 			if !ok {
// 				ep.elemChan.close()
// 				return
// 			}
// 			if !isTag(b) {
// 				fmt.Println("===================", string(b))
// 				c := content(b)
// 				ep.elemChan.write(c)
// 				continue
// 			}
// 			switch {
// 			case isCommentTag(b):
// 				ep.elemChan.write(ep.parseCommentTag(b))
// 			case isVoidTag(b):
// 				ep.elemChan.write(ep.parseVoidTag(b))
// 			case isEndTag(b):
// 				ep.elemChan.write(ep.parseEndTag(b))
// 			default:
// 				ep.elemChan.write(ep.parseStartTag(b))
// 			}
// 		}
// 	}
// }

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

type elemCorrector struct {
	ep             *elemProcessor
	elemChan       *bufElemChan
	startTagBuffer []*startTag
	buffer         []element
	errChan        chan error
}

func newElemCorrector(ep *elemProcessor) *elemCorrector {
	return &elemCorrector{
		ep,
		newBufElemChan(),
		make([]*startTag, 0, 64),
		make([]element, 0, 128),
		make(chan error),
	}
}

func (ec *elemCorrector) process() {
	allElems := make([]element, 0, 1024)
	for elem := range ec.ep.elemChan {
		allElems = append(allElems, elem)
	}
	allElemChan := make(chan element, len(allElems))
	for _, elem := range allElems {
		allElemChan <- elem
	}
	close(allElemChan)
	index := -1
OUTER:
	for {
		select {
		case err := <-ec.ep.errChan:
			ec.errChan <- err
			ec.elemChan.close()
			return
		// case elem, ok := <-ec.ep.elemChan:
		case elem, ok := <-allElemChan:
			if !ok {
				break OUTER
			}
			index += 1
			switch e := elem.(type) {
			case *voidTag, content:
				ec.buffer = append(ec.buffer, e)
			case *startTag:
				if string(e.name) == "center" {
					continue
				}
				ec.startTagBuffer = append(ec.startTagBuffer, e)
				ec.buffer = append(ec.buffer, e)
			case *endTag:
				if string(e.name) == "center" {
					continue
				}
				if len(ec.startTagBuffer) == 0 {
					continue
				}
				if string(e.name) == string(ec.startTagBuffer[len(ec.startTagBuffer)-1].name) {
					ec.startTagBuffer = ec.startTagBuffer[:len(ec.startTagBuffer)-1]
					ec.buffer = append(ec.buffer, e)
				} else {
					if index := ec.lastMatchedStartTagIndex(e); index == -1 {
						// fmt.Println("start tags now:", ec.startTagBuffer)
						// fmt.Println("drop tag:", e)
						continue
					} else {
						if ec.calculateBalance(allElems, e) >= 0 {
							// fmt.Println("current end tag:", e)
							// fmt.Println("before:", ec.startTagBuffer)
							l := make([]*startTag, len(ec.startTagBuffer[index:]))
							copy(l, ec.startTagBuffer[index:])
							ec.startTagBuffer = ec.startTagBuffer[:index]
							for i := len(l) - 1; i >= 1; i-- {
								fakeEndTag := &endTag{name: l[i].name}
								// fmt.Println("add fake end tag:", fakeEndTag)
								ec.buffer = append(ec.buffer, fakeEndTag)
								newList := make([]element, len(allElems)+1)
								copy(newList[:index], allElems[:index])
								newList[index] = fakeEndTag
								copy(newList[index+1:], allElems[index:])
								allElems = newList
								index += 1
							}
							// fmt.Println("after:", ec.startTagBuffer)
							ec.buffer = append(ec.buffer, e)
						} else {
							newList := make([]element, len(allElems)-1)
							copy(newList[:index], allElems[:index])
							copy(newList[index:], allElems[index+1:])
							allElems = newList
							index -= 1
							continue
						}
					}
					// if string(e.name) == string(ec.startTagBuffer[len(ec.startTagBuffer)-1].name) {
					// 	ec.startTagBuffer = ec.startTagBuffer[:len(ec.startTagBuffer)-1]
					// 	ec.buffer = append(ec.buffer, e)
					// } else {
					// 	if index := ec.lastMatchedStartTagIndex(e); index == -1 {
					// 		fmt.Println("start tags now:", ec.startTagBuffer)
					// 		fmt.Println("drop tag:", e)
					// 		continue
					// 	} else {
					// 		fmt.Println("current end tag:", e)
					// 		fmt.Println("before:", ec.startTagBuffer)
					// 		l := make([]*startTag, len(ec.startTagBuffer[index:]))
					// 		copy(l, ec.startTagBuffer[index:])
					// 		ec.startTagBuffer = ec.startTagBuffer[:index]
					// 		for i := len(l) - 1; i >= 1; i-- {
					// 			fakeEndTag := &endTag{name: l[i].name}
					// 			fmt.Println("add fake end tag:", fakeEndTag)
					// 			ec.buffer = append(ec.buffer, fakeEndTag)
					// 		}
					// 		fmt.Println("after:", ec.startTagBuffer)
					// 		ec.buffer = append(ec.buffer, e)
					// 	}

					// if !ec.isSkipMatch(e) {
					// 	fmt.Println("start tags now:", ec.startTagBuffer)
					// 	fmt.Println("drop tag:", e)
					// 	continue
					// } else {
					// 	fmt.Println("current end tag:", e)
					// 	fmt.Println("before:", ec.startTagBuffer)
					// 	fakeEndTag := &endTag{name: ec.startTagBuffer[len(ec.startTagBuffer)-1].name}
					// 	fmt.Println("add fake end tag:", fakeEndTag)
					// 	ec.buffer = append(ec.buffer, fakeEndTag)
					// 	fmt.Println("after:", ec.startTagBuffer)
					// 	ec.buffer = append(ec.buffer, e)
					// }
				}
			}
		}
	}
	if len(ec.startTagBuffer) != 0 {
		for i := len(ec.startTagBuffer) - 1; i >= 0; i-- {
			fakeEndTag := &endTag{name: ec.startTagBuffer[i].name}
			ec.buffer = append(ec.buffer, fakeEndTag)
		}
	}
	for _, e := range ec.buffer {
		ec.elemChan.write(e)
	}
	ec.elemChan.close()
}

func (ec *elemCorrector) lastMatchedStartTagIndex(et *endTag) int {
	for i := len(ec.startTagBuffer) - 1; i >= 0; i-- {
		if string(ec.startTagBuffer[i].name) == string(et.name) {
			return i
		}
	}
	return -1
}

func (ec *elemCorrector) isSkipMatch(et *endTag) bool {
	if len(ec.startTagBuffer) < 2 {
		return false
	}
	if string(ec.startTagBuffer[len(ec.startTagBuffer)-2].name) == string(et.name) {
		return true
	}
	return false
}

func (ec *elemCorrector) calculateBalance(l []element, et *endTag) int {
	var balance int
	for _, elem := range l {
		if string(elem.getName()) == string(et.name) {
			if _, ok := elem.(*startTag); ok {
				balance += 1
			} else {
				balance -= 1
			}
		}
	}
	return balance
}
