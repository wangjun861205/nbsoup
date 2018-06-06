package nbsoup

type bufElemChan struct {
	channel chan element
	buffer  []element
}

func newBufElemChan() *bufElemChan {
	return &bufElemChan{
		make(chan element),
		make([]element, 0, 8),
	}
}

func (c *bufElemChan) read() (element, bool) {
	var elem element
	if len(c.buffer) > 0 {
		elem, c.buffer = c.buffer[len(c.buffer)-1], c.buffer[:len(c.buffer)-1]
		return elem, true
	}
	var ok bool
	elem, ok = <-c.channel
	return elem, ok
}

func (c *bufElemChan) push(elem element) {
	c.buffer = append(c.buffer, elem)
}

func (c *bufElemChan) write(elem element) {
	c.channel <- elem
}

func (c *bufElemChan) close() {
	close(c.channel)
}
