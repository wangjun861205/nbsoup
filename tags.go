package nbsoup

type commentTag struct {
	content []byte
}

func (ct *commentTag) getName() []byte {
	return nil
}

func (ct *commentTag) getAttrList() [][]byte {
	return nil
}

type voidTag struct {
	name        []byte
	attrStrList [][]byte
}

func (ct *voidTag) getName() []byte {
	return ct.name
}

func (ct *voidTag) getAttrList() [][]byte {
	return ct.attrStrList
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

type endTag struct {
	name []byte
}

func (et *endTag) getName() []byte {
	return et.name
}

func (et *endTag) getAttrList() [][]byte {
	return nil
}
