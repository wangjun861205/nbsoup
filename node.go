package nbsoup

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
)

var ErrEndTagNotMatch = errors.New("end tag not match start tag")
var ErrEmptyNode = errors.New("empty node")

type Node struct {
	Name     string
	AttrMap  map[string]string
	Content  string
	Parent   *Node
	Children []*Node
	Next     *Node
	Previous *Node
}

func (n *Node) matchQ(q q) bool {
	if q.name == "@content" {
		switch q.operator {
		case equal:
			return n.Content == q.value
		case notEqual:
			return n.Content != q.value
		case include:
			return strings.Contains(n.Content, q.value)
		case except:
			return !strings.Contains(n.Content, q.value)
		case reg:
			re := regexp.MustCompile(q.value)
			return re.MatchString(n.Content)
		}
	}
	attr, ok := n.AttrMap[q.name]
	if !ok {
		return false
	}
	switch q.operator {
	case equal:
		if attr == q.value {
			return true
		}
		return false
	case notEqual:
		if attr != q.value {
			return true
		}
		return false
	case include:
		return strings.Contains(attr, q.value)
	case except:
		return !strings.Contains(attr, q.value)
	case reg:
		re := regexp.MustCompile(q.value)
		return re.MatchString(attr)
	default:
		return false
	}
}

func (n *Node) matchQuery(query *query) bool {
	if query.name != "" && n.Name != query.name {
		return false
	}
	if query.queryList == nil {
		return query.name == n.Name
	}
	var isMatch bool
	for _, qList := range query.queryList {
		subMatch := true
		for _, q := range qList {
			if !n.matchQ(q) {
				subMatch = false
			}
		}
		if subMatch {
			isMatch = true
			break
		}
	}
	return isMatch
}

func (n *Node) findNodes(query *query, parentMatched bool) []*Node {
	currentMatch := make([]*Node, 0, 32)
	for _, child := range n.Children {
		if child.matchQuery(query) {
			if query.next != nil {
				subMatchList := child.findNodes(query.next, true)
				if subMatchList != nil {
					currentMatch = append(currentMatch, subMatchList...)
				}
			} else {
				currentMatch = append(currentMatch, child)
				subMatchList := child.findNodes(query, false)
				if subMatchList != nil {
					currentMatch = append(currentMatch, subMatchList...)
				}
			}
		} else {
			if !parentMatched {
				subMatchList := child.findNodes(query, false)
				if subMatchList != nil {
					currentMatch = append(currentMatch, subMatchList...)
				}
			}
		}
	}
	if len(currentMatch) == 0 {
		return nil
	}
	return currentMatch
}

func FindAll(node *Node, queryStr string) ([]*Node, error) {
	query, err := parseQuery(queryStr)
	if err != nil {
		return nil, err
	}
	return node.findNodes(query, false), nil
}

// func process(ep *elemProcessor) (*Node, error) {
// 	var rootTag *startTag
// OUTER:
// 	for {
// 		select {
// 		case err := <-ep.errChan:
// 			return nil, err
// 		case rootElem, ok := <-ep.elemChan:
// 			if !ok {
// 				return nil, ErrEmptyNode
// 			}
// 			switch r := rootElem.(type) {
// 			case *startTag:
// 				rootTag = r
// 				break OUTER
// 			default:
// 				continue
// 			}
// 		}
// 	}
// 	return parseTag(rootTag, ep)
// }

func process(ec *elemCorrector) (*Node, error) {
	var rootTag *startTag
OUTER:
	for {
		select {
		case err := <-ec.errChan:
			return nil, err
		default:
			rootElem, ok := ec.elemChan.read()
			if !ok {
				return nil, ErrEmptyNode
			}
			switch r := rootElem.(type) {
			case *startTag:
				rootTag = r
				break OUTER
			default:
				continue
			}
		}
	}
	return parseTag(rootTag, ec)
}

// func parseTag(tag *startTag, ep *elemProcessor) (*Node, error) {
// 	children := make([]*Node, 0, 32)
// 	node := fromTag(element(tag))
// 	for {
// 		select {
// 		case err := <-ep.errChan:
// 			return nil, err
// 		case nextElem, ok := <-ep.elemChan:
// 			if !ok {
// 				switch len(children) {
// 				case 0:
// 					return node, nil
// 				case 1:
// 					children[0].Parent = node
// 					node.Children = children
// 					return node, nil
// 				default:
// 					for i := 0; i < len(children)-1; i++ {
// 						children[i].Next = children[i+1]
// 						children[i+1].Previous = children[i]
// 						children[i].Parent = node
// 					}
// 					children[len(children)-1].Parent = node
// 					node.Children = children
// 					return node, nil
// 				}
// 			}
// 			switch elem := nextElem.(type) {
// 			case content:
// 				node.Content += string(elem)
// 			case *voidTag:
// 				children = append(children, fromTag(elem))
// 			case *startTag:
// 				childNode, err := parseTag(elem, ep)
// 				if err != nil {
// 					return nil, err
// 				}
// 				children = append(children, childNode)
// 			case *endTag:
// 				if string(elem.name) == string(tag.name) {
// 					switch len(children) {
// 					case 0:
// 						return node, nil
// 					case 1:
// 						children[0].Parent = node
// 						node.Children = children
// 						return node, nil
// 					default:
// 						for i := 0; i < len(children)-1; i++ {
// 							children[i].Next = children[i+1]
// 							children[i+1].Previous = children[i]
// 							children[i].Parent = node
// 						}
// 						children[len(children)-1].Parent = node
// 						node.Children = children
// 						return node, nil
// 					}
// 				} else {
// 					fmt.Println(node.Name)
// 					continue
// 				}
// 			}
// 		}
// 	}
// }

func parseTag(tag *startTag, ec *elemCorrector) (*Node, error) {
	children := make([]*Node, 0, 32)
	node := fromTag(element(tag))
	for {
		select {
		case err := <-ec.errChan:
			return nil, err
		default:
			nextElem, ok := ec.elemChan.read()
			if !ok {
				switch len(children) {
				case 0:
					return node, nil
				case 1:
					children[0].Parent = node
					node.Children = children
					return node, nil
				default:
					for i := 0; i < len(children)-1; i++ {
						children[i].Next = children[i+1]
						children[i+1].Previous = children[i]
						children[i].Parent = node
					}
					children[len(children)-1].Parent = node
					node.Children = children
					return node, nil
				}
			}
			switch elem := nextElem.(type) {
			case content:
				node.Content += string(elem)
			case *voidTag:
				children = append(children, fromTag(elem))
			case *startTag:
				childNode, err := parseTag(elem, ec)
				if err != nil {
					return nil, err
				}
				children = append(children, childNode)
			case *endTag:
				switch len(children) {
				case 0:
					return node, nil
				case 1:
					children[0].Parent = node
					node.Children = children
					return node, nil
				default:
					for i := 0; i < len(children)-1; i++ {
						children[i].Next = children[i+1]
						children[i+1].Previous = children[i]
						children[i].Parent = node
					}
					children[len(children)-1].Parent = node
					node.Children = children
					return node, nil
				}
			}
		}
	}

}

func fromTag(tag element) *Node {
	switch t := tag.(type) {
	case *startTag, *voidTag:
		return &Node{
			Name:    string(t.getName()),
			AttrMap: parseAttrs(t.getAttrList()),
		}
	default:
		return &Node{}
	}
}

func parseAttrs(attrList [][]byte) map[string]string {
	attrMap := make(map[string]string)
	for _, bAttr := range attrList {
		l := bytes.Split(bAttr, []byte("="))
		if len(l) == 1 {
			attrMap[string(l[0])] = "true"
			continue
		}
		attrMap[string(l[0])] = string(bytes.Trim(l[1], "\""))
	}
	return attrMap
}

func Parse(html []byte) (*Node, error) {
	hp := newHTMLProcessor()
	ep := newElemProcessor(hp)
	ec := newElemCorrector(ep)
	go hp.process(html)
	go ep.process()
	go ec.process()
	return process(ec)
}

func (n *Node) FindAll(queryStr string) ([]*Node, error) {
	query, err := parseQuery(queryStr)
	if err != nil {
		return nil, err
	}
	chanList := []chan *Node{n.allChildren()}
	for {
		chanList = append(chanList, filter(chanList[len(chanList)-1], query))
		if query.next == nil {
			break
		} else {
			chanList = append(chanList, bridgeFilter(chanList[len(chanList)-1]))
			query = query.next
		}
	}
	nodeList := make([]*Node, 0, 128)
	for node := range chanList[len(chanList)-1] {
		nodeList = append(nodeList, node)
	}
	return nodeList, nil
}

func (n *Node) allChildren() chan *Node {
	childChan := make(chan *Node)
	go func() {
		for _, child := range n.Children {
			childChan <- child
			subChan := child.allChildren()
			for subChild := range subChan {
				childChan <- subChild
			}
		}
		close(childChan)
	}()
	return childChan
}

func filter(input chan *Node, query *query) chan *Node {
	outChan := make(chan *Node)
	go func() {
		for node := range input {
			if node.matchQuery(query) {
				outChan <- node
			}
		}
		close(outChan)
	}()
	return outChan
}

func bridgeFilter(input chan *Node) chan *Node {
	outChan := make(chan *Node)
	go func() {
		for n := range input {
			for _, child := range n.Children {
				outChan <- child
			}
		}
		close(outChan)
	}()
	return outChan
}

// func (n *Node) rMatch(query *query) chan *Node {
// 	matchChan := make(chan *Node)
// 	var wg sync.WaitGroup
// 	for _, child := range n.Children {
// 		wg.Add(1)
// 		go func() {
// 			if child.matchQuery(query) && child.matchParent(query) {
// 				fmt.Println(child.Name)
// 				matchChan <- child
// 			}
// 			subChan := child.rMatch(query)
// 			for subNode := range subChan {
// 				matchChan <- subNode
// 			}
// 			wg.Done()
// 		}()
// 	}
// 	go func() {
// 		wg.Wait()
// 		close(matchChan)
// 	}()
// 	return matchChan
// }

func (n *Node) matchParent(query *query) bool {
	if query.prev == nil {
		return true
	}
	if n.Parent == nil {
		return false
	}
	if !n.Parent.matchQuery(query.prev) {
		return false
	} else {
		return n.Parent.matchParent(query.prev)
	}
}

func (n *Node) GetAllContent() string {
	c := n.Content
	for _, child := range n.Children {
		c += child.GetAllContent()
	}
	return c
}
