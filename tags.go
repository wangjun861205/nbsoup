package nbsoup

import "fmt"

type commentTag struct {
	content []byte
}

func (ct *commentTag) getName() []byte {
	return nil
}

func (ct *commentTag) getAttrList() [][]byte {
	return nil
}

func (ct *commentTag) String() string {
	return "<!--comment-->"
}

type voidTag struct {
	name        []byte
	attrStrList [][]byte
}

func (vt *voidTag) getName() []byte {
	return vt.name
}

func (vt *voidTag) getAttrList() [][]byte {
	return vt.attrStrList
}

func (vt *voidTag) String() string {
	return fmt.Sprintf("<%s />", vt.name)
}

type startTag struct {
	name        []byte
	attrStrList [][]byte
}

func (st *startTag) getName() []byte {
	return st.name
}

func (st *startTag) getAttrList() [][]byte {
	return st.attrStrList
}

func (st *startTag) String() string {
	return fmt.Sprintf("<%s>", st.name)
}

type endTag struct {
	name []byte
}

func (et *endTag) getName() []byte {
	return et.name
}

func (et *endTag) getAttrList() [][]byte {
	return nil
}

func (et *endTag) String() string {
	return fmt.Sprintf("</%s>", et.name)
}
