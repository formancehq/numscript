package jsonrpc2

// This is mostly intented as a testing utility
type ChanObjStream struct {
	in  <-chan Message
	out chan<- Message
}

var _ MessageStream = (*ChanObjStream)(nil)

func NewChanObjStream(in <-chan Message, out chan<- Message) *ChanObjStream {
	return &ChanObjStream{
		in:  in,
		out: out,
	}
}

func (c *ChanObjStream) Close() error {
	close(c.out)
	return nil
}

func (c *ChanObjStream) ReadMessage() (Message, error) {
	msg := <-c.in
	return msg, nil
}

func (c *ChanObjStream) WriteMessage(obj Message) error {
	c.out <- obj
	return nil
}
