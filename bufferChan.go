package nbsoup

type bufElemChan struct {
	channel chan element
	buffer  []element
	bufSize int
	i       int
}

func newBufElemChan(bufSize int) *bufElemChan {
	return &bufElemChan{
		make(chan element),
		make([]element, 0, bufSize),
		bufSize,
		-1,
	}
}

func (c *bufElemChan) read() (element, bool) {
	if c.i+1 == len(c.buffer) {
		elem, ok := <-c.channel
		if !ok {
			return elem, false
		}
		if len(c.buffer) < c.bufSize {
			c.buffer = append(c.buffer, elem)
			c.i += 1
		} else {
			c.buffer = append(c.buffer[1:], elem)
		}
		return elem, true
	} else {
		c.i += 1
		return c.buffer[c.i], true
	}
}

func (c *bufElemChan) unread() bool {
	if c.i == -1 {
		return false
	}
	c.i -= 1
	return true
}

func (c *bufElemChan) write(elem element) {
	c.channel <- elem
}

func (c *bufElemChan) close() {
	close(c.channel)
}
