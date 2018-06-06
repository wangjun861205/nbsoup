package nbsoup

type content []byte

func (c content) getName() []byte {
	return nil
}

func (c content) getAttrList() [][]byte {
	return nil
}

func (c content) String() string {
	return "<text content>"
}
